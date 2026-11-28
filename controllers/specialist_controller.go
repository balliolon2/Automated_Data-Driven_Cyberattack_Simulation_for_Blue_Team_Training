package controllers

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cybersim/dto"
	"cybersim/models"
	"cybersim/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func decodeAnyJSONMap(raw any) any {
	switch v := raw.(type) {
	case nil:
		return nil
	case map[string]any:
		return v
	case []byte:
		if len(v) == 0 {
			return nil
		}
		var out map[string]any
		if err := json.Unmarshal(v, &out); err == nil {
			return out
		}
		return nil
	case json.RawMessage:
		if len(v) == 0 {
			return nil
		}
		var out map[string]any
		if err := json.Unmarshal(v, &out); err == nil {
			return out
		}
		return nil
	case string:
		if strings.TrimSpace(v) == "" {
			return nil
		}
		var out map[string]any
		if err := json.Unmarshal([]byte(v), &out); err == nil {
			return out
		}
		return nil
	default:
		return v
	}
}

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// SpecialistRegStore persists atomic new-user + specialist-application registrations cleanly.
type SpecialistRegStore interface {
	FindUserByEmail(email string) (*models.User, error)
	FindUserByNickname(nickname string) (*models.User, error)
	CreateUserAndApplication(user *models.User, app *models.SpecialistApplication) error
}

// gormSpecialistRegStore embeds gormAuthStore to reuse FindUserByEmail/FindUserByNickname
// without duplicating queries, while adding the atomic user + application transaction.
type gormSpecialistRegStore struct {
	*gormAuthStore
}

func (s *gormSpecialistRegStore) CreateUserAndApplication(user *models.User, app *models.SpecialistApplication) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		app.UserID = user.UserID
		if err := tx.Create(app).Error; err != nil {
			return err
		}
		return nil
	})
}

type SpecialistController struct {
	DB         *gorm.DB
	regStore   SpecialistRegStore
	mailSender services.MailSender
	uploadDir  string
}

func NewSpecialistController(db *gorm.DB) *SpecialistController {
	return &SpecialistController{
		DB:         db,
		regStore:   &gormSpecialistRegStore{gormAuthStore: &gormAuthStore{DB: db}},
		mailSender: services.NewSMTPMailSenderFromEnv(),
		uploadDir:  UploadBaseDir,
	}
}

func NewSpecialistControllerWithDeps(
	db *gorm.DB,
	regStore SpecialistRegStore,
	mailSender services.MailSender,
	uploadDir string,
) *SpecialistController {
	if strings.TrimSpace(uploadDir) == "" {
		uploadDir = UploadBaseDir
	}
	return &SpecialistController{
		DB:         db,
		regStore:   regStore,
		mailSender: mailSender,
		uploadDir:  uploadDir,
	}
}

func (sc *SpecialistController) effectiveRegStore() SpecialistRegStore {
	if sc.regStore != nil {
		return sc.regStore
	}
	if sc.DB != nil {
		return &gormSpecialistRegStore{gormAuthStore: &gormAuthStore{DB: sc.DB}}
	}
	return nil
}

func (sc *SpecialistController) effectiveUploadDir() string {
	if strings.TrimSpace(sc.uploadDir) != "" {
		return sc.uploadDir
	}
	return UploadBaseDir
}

const (
	MaxFileSize   = 10 * 1024 * 1024 // 10MB
	UploadBaseDir = "./uploads/applications"
)

var allowedExtensions = map[string]bool{
	".pdf":  true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
}

type validatedAppFiles struct {
	bio             string
	linkedinURL     string
	portfolioURL    string
	resumeHeader    *multipart.FileHeader
	resumeExt       string
	certHeader      *multipart.FileHeader
	certExt         string
}

func validateApplicationFormFiles(c *gin.Context) (validatedAppFiles, bool) {
	bio := strings.TrimSpace(c.PostForm("bio"))
	if bio == "" {
		respondAPIError(c, http.StatusBadRequest, "SPECIALIST_BIO_REQUIRED", "Bio is required")
		return validatedAppFiles{}, false
	}

	resumeHeader, err := c.FormFile("resume")
	if err != nil {
		respondAPIError(c, http.StatusBadRequest, "SPECIALIST_RESUME_REQUIRED", "Resume file is required")
		return validatedAppFiles{}, false
	}
	if resumeHeader.Size > MaxFileSize {
		respondAPIError(c, http.StatusBadRequest, "SPECIALIST_FILE_TOO_LARGE", "Resume file exceeds 10MB limit")
		return validatedAppFiles{}, false
	}
	resumeExt := strings.ToLower(filepath.Ext(resumeHeader.Filename))
	if !allowedExtensions[resumeExt] {
		respondAPIError(c, http.StatusBadRequest, "SPECIALIST_INVALID_FILE_TYPE", "Invalid resume file type. Allowed: PDF, PNG, JPG, JPEG")
		return validatedAppFiles{}, false
	}

	var certHeader *multipart.FileHeader
	var certExt string
	if ch, chErr := c.FormFile("certificate"); chErr == nil && ch != nil {
		if ch.Size > MaxFileSize {
			respondAPIError(c, http.StatusBadRequest, "SPECIALIST_FILE_TOO_LARGE", "Certificate file exceeds 10MB limit")
			return validatedAppFiles{}, false
		}
		certExt = strings.ToLower(filepath.Ext(ch.Filename))
		if !allowedExtensions[certExt] {
			respondAPIError(c, http.StatusBadRequest, "SPECIALIST_INVALID_FILE_TYPE", "Invalid certificate file type. Allowed: PDF, PNG, JPG, JPEG")
			return validatedAppFiles{}, false
		}
		certHeader = ch
	}

	return validatedAppFiles{
		bio:          bio,
		linkedinURL:  strings.TrimSpace(c.PostForm("linkedin_url")),
		portfolioURL: strings.TrimSpace(c.PostForm("portfolio_url")),
		resumeHeader: resumeHeader,
		resumeExt:    resumeExt,
		certHeader:   certHeader,
		certExt:      certExt,
	}, true
}

func saveApplicationFormFiles(c *gin.Context, baseDir, appID string, files validatedAppFiles) (appDir, resumePath, certPath string, ok bool) {
	appDir = filepath.Join(baseDir, appID)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		respondAPIError(c, http.StatusInternalServerError, "SPECIALIST_APPLICATION_FAILED", "Failed to create upload storage directory")
		return "", "", "", false
	}

	resumePath = filepath.Join(appDir, fmt.Sprintf("resume_%s%s", generateUUID(), files.resumeExt))
	if err := c.SaveUploadedFile(files.resumeHeader, resumePath); err != nil {
		_ = os.RemoveAll(appDir)
		respondAPIError(c, http.StatusInternalServerError, "SPECIALIST_APPLICATION_FAILED", "Failed to save resume file")
		return "", "", "", false
	}

	if files.certHeader != nil {
		certPath = filepath.Join(appDir, fmt.Sprintf("certificate_%s%s", generateUUID(), files.certExt))
		if err := c.SaveUploadedFile(files.certHeader, certPath); err != nil {
			_ = os.RemoveAll(appDir)
			respondAPIError(c, http.StatusInternalServerError, "SPECIALIST_APPLICATION_FAILED", "Failed to save certificate file")
			return "", "", "", false
		}
	}

	return appDir, resumePath, certPath, true
}

// RegisterSpecialist godoc
// @Summary Atomically register a new user and submit specialist application documents
// @Description Creates a learner account (unverified when REQUIRE_EMAIL_VERIFICATION=true) and stores specialist application credentials atomically, cleaning up files on failure.
// @Tags Auth
// @Accept multipart/form-data
// @Produce json
// @Param email formData string true "Account Email"
// @Param password formData string true "Account Password"
// @Param nickname formData string false "Display Nickname"
// @Param preferred_language formData string false "Preferred UI Language (en or th)"
// @Param bio formData string true "Specialist Biography"
// @Param linkedin_url formData string false "LinkedIn Profile URL"
// @Param portfolio_url formData string false "Portfolio URL"
// @Param resume formData file true "Resume/CV PDF"
// @Param certificate formData file false "Optional Certificate Document"
// @Success 201 {object} dto.RegisterSpecialistResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /register-specialist [post]
func (sc *SpecialistController) RegisterSpecialist(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(MaxFileSize * 2); err != nil {
		respondAPIError(c, http.StatusBadRequest, "SPECIALIST_INVALID_PAYLOAD", "Payload too large or invalid multipart form")
		return
	}

	files, ok := validateApplicationFormFiles(c)
	if !ok {
		return
	}

	regStore := sc.effectiveRegStore()
	if regStore == nil {
		respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Registration store is unavailable")
		return
	}

	prep, err := validateAndPrepareNewUser(
		regStore,
		c.PostForm("email"),
		c.PostForm("password"),
		c.PostForm("nickname"),
		c.PostForm("preferred_language"),
	)
	if err != nil {
		respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
		return
	}
	if prep.ValidationErrCode != "" {
		respondAPIError(c, prep.ValidationStatus, prep.ValidationErrCode, prep.ValidationErrMsg)
		return
	}

	appID := generateUUID()
	appDir, resumeSavePath, certSavePath, savedOK := saveApplicationFormFiles(c, sc.effectiveUploadDir(), appID, files)
	if !savedOK {
		return
	}

	cleanupOnError := true
	defer func() {
		if cleanupOnError && appDir != "" {
			_ = os.RemoveAll(appDir)
		}
	}()

	now := time.Now().UTC()
	application := models.SpecialistApplication{
		ApplicationID:   appID,
		Status:          models.ApplicationStatusPending,
		Bio:             files.bio,
		ResumePath:      resumeSavePath,
		CertificatePath: certSavePath,
		LinkedInURL:     files.linkedinURL,
		PortfolioURL:    files.portfolioURL,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := regStore.CreateUserAndApplication(&prep.User, &application); err != nil {
		log.Printf("Atomic specialist registration failed: %v", err)
		respondAPIError(c, http.StatusInternalServerError, "SPECIALIST_APPLICATION_FAILED", "Failed to create specialist registration")
		return
	}

	cleanupOnError = false

	if prep.RequireVerify && prep.RawToken != "" {
		services.DispatchVerificationEmailAsync(services.EffectiveMailSender(sc.mailSender), prep.User.Email, prep.User.Nickname, prep.RawToken)
	}

	msg := "Account created and specialist application submitted."
	if prep.RequireVerify {
		msg = "Account and specialist application created. Please verify your email before signing in."
	}

	c.JSON(http.StatusCreated, dto.RegisterSpecialistResponse{
		Message:              msg,
		Nickname:             prep.User.Nickname,
		Email:                prep.User.Email,
		PreferredLanguage:    prep.NormalizedLang,
		RequiresVerification: prep.RequireVerify,
		ApplicationID:        application.ApplicationID,
		Status:               application.Status,
	})
}

// Apply handles specialist application submission with multipart form and documents
func (sc *SpecialistController) Apply(c *gin.Context) {
	userID, exists := c.Get("user_id")
	userIDStr, ok := userID.(string)
	if !exists || !ok || strings.TrimSpace(userIDStr) == "" {
		respondAPIError(c, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Unauthorized")
		return
	}

	if sc.DB == nil {
		respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Database connection is unavailable")
		return
	}

	if err := c.Request.ParseMultipartForm(MaxFileSize * 2); err != nil {
		respondAPIError(c, http.StatusBadRequest, "SPECIALIST_INVALID_PAYLOAD", "Payload too large or invalid multipart form")
		return
	}

	files, ok := validateApplicationFormFiles(c)
	if !ok {
		return
	}

	// Reject already-pending or already-approved applicants before writing any files to disk.
	var existing models.SpecialistApplication
	existingErr := sc.DB.Where("user_id = ?", userIDStr).Order("created_at desc").First(&existing).Error
	if existingErr != nil && !errors.Is(existingErr, gorm.ErrRecordNotFound) {
		respondAPIError(c, http.StatusInternalServerError, "SPECIALIST_APPLICATION_FAILED", "Failed to check existing application")
		return
	}
	hasExisting := existingErr == nil
	if hasExisting {
		if existing.Status == models.ApplicationStatusPending {
			respondAPIError(c, http.StatusBadRequest, "SPECIALIST_ALREADY_PENDING", "An application is already pending administrator review")
			return
		}
		if existing.Status == models.ApplicationStatusApproved {
			respondAPIError(c, http.StatusBadRequest, "SPECIALIST_ALREADY_APPROVED", "You are already an approved Specialist")
			return
		}
	}

	appID := generateUUID()
	appDir, resumeSavePath, certSavePath, savedOK := saveApplicationFormFiles(c, sc.effectiveUploadDir(), appID, files)
	if !savedOK {
		return
	}
	cleanupOnError := true
	defer func() {
		if cleanupOnError && appDir != "" {
			_ = os.RemoveAll(appDir)
		}
	}()

	now := time.Now()
	if hasExisting && existing.Status == models.ApplicationStatusRejected {
		oldResumeDir := ""
		if strings.TrimSpace(existing.ResumePath) != "" {
			oldResumeDir = filepath.Dir(existing.ResumePath)
		}
		// Update rejected application to pending (re-application)
		existing.Status = models.ApplicationStatusPending
		existing.Bio = files.bio
		existing.ResumePath = resumeSavePath
		if certSavePath != "" {
			existing.CertificatePath = certSavePath
		}
		existing.LinkedInURL = files.linkedinURL
		existing.PortfolioURL = files.portfolioURL
		existing.RejectionReason = ""
		existing.ReviewedBy = nil
		existing.ReviewedAt = nil
		existing.UpdatedAt = now
		if err := sc.DB.Save(&existing).Error; err != nil {
			respondAPIError(c, http.StatusInternalServerError, "SPECIALIST_APPLICATION_FAILED", "Failed to update application")
			return
		}
		cleanupOnError = false
		cleanBase := filepath.Clean(sc.effectiveUploadDir())
		if oldResumeDir != "" {
			cleanOld := filepath.Clean(oldResumeDir)
			if cleanOld != filepath.Clean(appDir) && strings.HasPrefix(cleanOld, cleanBase+string(os.PathSeparator)) {
				_ = os.RemoveAll(cleanOld)
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"message":        "Application re-submitted successfully",
			"application_id": existing.ApplicationID,
			"status":         existing.Status,
		})
		return
	}

	application := models.SpecialistApplication{
		ApplicationID:   appID,
		UserID:          userIDStr,
		Status:          models.ApplicationStatusPending,
		Bio:             files.bio,
		ResumePath:      resumeSavePath,
		CertificatePath: certSavePath,
		LinkedInURL:     files.linkedinURL,
		PortfolioURL:    files.portfolioURL,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := sc.DB.Create(&application).Error; err != nil {
		log.Printf("Failed to create specialist application: %v", err)
		respondAPIError(c, http.StatusInternalServerError, "SPECIALIST_APPLICATION_FAILED", "Failed to create application")
		return
	}

	cleanupOnError = false
	c.JSON(http.StatusCreated, gin.H{
		"message":        "Specialist application submitted successfully",
		"application_id": application.ApplicationID,
		"status":         application.Status,
	})
}

// GetApplicationStatus returns the status of current user's specialist application
func (sc *SpecialistController) GetApplicationStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondAPIError(c, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Unauthorized")
		return
	}

	var app models.SpecialistApplication
	if err := sc.DB.Where("user_id = ?", userID).Order("created_at desc").First(&app).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "No specialist application found",
			"code":    "SPECIALIST_APPLICATION_NOT_FOUND",
		})
		return
	}

	var user models.User
	_ = sc.DB.Where("user_id = ?", userID).First(&user)

	c.JSON(http.StatusOK, app.ToDTO(user.Nickname, user.Email))
}

// GetSubmissions godoc
// @Summary List anonymized learner submissions
// @Description Returns completed simulation sessions with evidence review counters for specialist triage.
// @Tags Specialist
// @Produce json
// @Security BearerAuth
// @Param scenario_id query string false "Filter by scenario ID"
// @Param domain_id query string false "Filter by security domain ID"
// @Param max_score query number false "Only include submissions at or below this final score"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size (1-100)" default(20)
// @Success 200 {array} dto.SpecialistSubmissionSummary
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /specialist/submissions [get]
func (sc *SpecialistController) GetSubmissions(c *gin.Context) {
	scenarioID := strings.TrimSpace(c.Query("scenario_id"))
	domainID := strings.TrimSpace(c.Query("domain_id"))
	maxScoreStr := strings.TrimSpace(c.Query("max_score"))

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := sc.DB.Table("simulation_sessions").
		Select("simulation_sessions.session_id, simulation_sessions.scenario_id, simulation_sessions.final_score, simulation_sessions.completed_at, simulation_sessions.total_actions, simulation_sessions.rubric_version, simulation_sessions.skill_gap, scenarios.title as scenario_title, scenarios.translations as scenario_translations, scenarios.domain_id, users.nickname as learner_nickname, users.current_tier as learner_tier").
		Joins("JOIN scenarios ON scenarios.scenario_id = simulation_sessions.scenario_id").
		Joins("JOIN users ON users.user_id = simulation_sessions.user_id").
		Where("simulation_sessions.status = 'completed'")

	if scenarioID != "" {
		query = query.Where("simulation_sessions.scenario_id = ?", scenarioID)
	}
	if domainID != "" {
		query = query.Where("scenarios.domain_id = ?", domainID)
	}
	if maxScoreStr != "" {
		var maxScore float64
		if _, err := fmt.Sscanf(maxScoreStr, "%f", &maxScore); err == nil {
			query = query.Where("simulation_sessions.final_score <= ?", maxScore)
		}
	}

	type submissionRow struct {
		SessionID            string     `json:"session_id"`
		ScenarioID           string     `json:"scenario_id"`
		ScenarioTitle        string     `json:"scenario_title"`
		ScenarioTranslations any        `json:"scenario_translations"`
		DomainID             string     `json:"domain_id"`
		LearnerNickname      string     `json:"learner_nickname"`
		LearnerTier          int        `json:"learner_tier"`
		FinalScore           *float64   `json:"final_score"`
		TotalActions         int        `json:"total_actions"`
		RubricVersion        string     `json:"rubric_version"`
		SkillGap             any        `json:"skill_gap"`
		CompletedAt          *time.Time `json:"completed_at"`
	}

	var rows []submissionRow
	if err := query.Order("simulation_sessions.completed_at desc").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		respondAPIError(c, http.StatusInternalServerError, "SPECIALIST_SUBMISSIONS_FETCH_FAILED", "Failed to fetch submissions")
		return
	}

	result := make([]dto.SpecialistSubmissionSummary, 0, len(rows))
	for _, r := range rows {
		score := 0.0
		if r.FinalScore != nil {
			score = *r.FinalScore
		}
		completedStr := ""
		if r.CompletedAt != nil {
			completedStr = r.CompletedAt.Format(time.RFC3339)
		}

		// Count TP/FP correctness and rehearse the evidence selection for this
		// session. The immutable ScenarioSnapshot is authoritative; when it is
		// absent, the recorded SkillGap counters are used. Only legacy sessions
		// that have neither fall back to the live scenarios row.
		var tpfpCorrect bool

		var triageAction models.SessionAction
		if err := sc.DB.Where("session_id = ? AND action_type = 'triage_alert'", r.SessionID).First(&triageAction).Error; err == nil && triageAction.IsCorrect != nil {
			tpfpCorrect = *triageAction.IsCorrect
		}

		var actions []models.SessionAction
		if err := sc.DB.Where("session_id = ?", r.SessionID).Order("step_order asc").Find(&actions).Error; err != nil {
			log.Printf("Failed to load actions for session %s: %v", r.SessionID, err)
		}

		session := models.SimulationSession{
			SessionID:     r.SessionID,
			ScenarioID:    r.ScenarioID,
			RubricVersion: r.RubricVersion,
			SkillGap:      r.SkillGap,
		}

		var review services.EvidenceReview
		scenario, hasSnapshot, scenarioErr := sessionScenarioFromDB(sc.DB, session)
		switch {
		case scenarioErr == nil && hasSnapshot:
			review = services.BuildEvidenceReviewForSession(scenario, session, actions)
		case scenarioErr == nil:
			if recorded, ok := services.ScenarioResultFromRecorded(r.SkillGap); ok {
				review = services.EvidenceReviewFromResult(recorded)
			} else {
				review = services.BuildEvidenceReviewForSession(scenario, session, actions)
			}
		default:
			if recorded, ok := services.ScenarioResultFromRecorded(r.SkillGap); ok {
				review = services.EvidenceReviewFromResult(recorded)
			} else {
				log.Printf("Failed to load scenario for session %s: %v", r.SessionID, scenarioErr)
			}
		}

		scenarioTranslations := decodeAnyJSONMap(r.ScenarioTranslations)
		if scenarioErr == nil && scenario.Translations != nil {
			if decoded := decodeAnyJSONMap(scenario.Translations); decoded != nil {
				scenarioTranslations = decoded
			}
		}

		domainName := domainNames[r.DomainID]
		if domainName == "" {
			domainName = r.DomainID
		}

		result = append(result, dto.SpecialistSubmissionSummary{
			SessionID:             r.SessionID,
			LearnerNickname:       r.LearnerNickname,
			LearnerTier:           r.LearnerTier,
			ScenarioID:            r.ScenarioID,
			ScenarioTitle:         r.ScenarioTitle,
			ScenarioTranslations:  scenarioTranslations,
			DomainID:              r.DomainID,
			DomainName:            domainName,
			FinalScore:            score,
			TPFPCorrect:           tpfpCorrect,
			FindingsFoundCount:    review.FindingsFoundCount,
			TotalFindingsCount:    review.TotalFindingsCount,
			EvidenceSelectedCount: review.EvidenceSelectedCount,
			NoiseSelectedCount:    review.NoiseSelectedCount,
			EvidencePenalty:       review.EvidencePenalty,
			ActionsCount:          r.TotalActions,
			CompletedAt:           completedStr,
		})
	}

	c.JSON(http.StatusOK, result)
}

// GetSubmissionDetail godoc
// @Summary Get specialist submission detail
// @Description Returns the full investigation timeline plus the evidence review for one completed simulation session.
// @Tags Specialist
// @Produce json
// @Security BearerAuth
// @Param session_id path string true "Simulation Session UUID"
// @Success 200 {object} dto.SpecialistSubmissionDetail
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /specialist/submissions/{session_id} [get]
func (sc *SpecialistController) GetSubmissionDetail(c *gin.Context) {
	sessionID := c.Param("session_id")

	var session models.SimulationSession
	if err := sc.DB.Where("session_id = ?", sessionID).First(&session).Error; err != nil {
		respondAPIError(c, http.StatusNotFound, "SPECIALIST_SESSION_NOT_FOUND", "Session not found")
		return
	}

	// The session snapshot is the immutable evidence index space the learner
	// was scored against. Legacy sessions without a snapshot fall back to the
	// live scenario row inside sessionScenarioFromDB.
	scenario, _, err := sessionScenarioFromDB(sc.DB, session)
	if err != nil {
		respondAPIError(c, http.StatusNotFound, "SPECIALIST_SCENARIO_NOT_FOUND", "Scenario not found")
		return
	}

	var user models.User
	if err := sc.DB.Where("user_id = ?", session.UserID).First(&user).Error; err != nil {
		respondAPIError(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}

	var actions []models.SessionAction
	sc.DB.Where("session_id = ?", sessionID).Order("step_order asc").Find(&actions)

	timeline := make([]dto.SubmissionTimelineAction, 0, len(actions))
	var learnerChoice *bool
	var tpfpCorrect bool
	var tpfpPoints float64
	var selectedActions []string
	var discoveredFindings []string

	for _, a := range actions {
		timeline = append(timeline, dto.SubmissionTimelineAction{
			ActionID:   a.ActionID,
			StepOrder:  a.StepOrder,
			ActionType: string(a.ActionType),
			Payload:    a.Payload,
			IsCorrect:  a.IsCorrect,
			Points:     a.Points,
			Timestamp:  a.Timestamp,
		})

		if a.ActionType == "triage_alert" && a.Payload != nil {
			if m, ok := a.Payload.(map[string]interface{}); ok {
				if choice, ok := m["user_choice"].(bool); ok {
					learnerChoice = &choice
				}
			}
			if a.IsCorrect != nil {
				tpfpCorrect = *a.IsCorrect
			}
			tpfpPoints = float64(a.Points)
		} else if a.ActionType == "respond" && a.Payload != nil {
			if m, ok := a.Payload.(map[string]interface{}); ok {
				if acts, ok := m["selected_actions"].([]interface{}); ok {
					for _, act := range acts {
						if s, ok := act.(string); ok {
							selectedActions = append(selectedActions, s)
						}
					}
				}
			}
		} else if a.ActionType == "submit_decision" && a.Payload != nil {
			if m, ok := a.Payload.(map[string]interface{}); ok {
				if finds, ok := m["discovered_findings"].([]interface{}); ok {
					for _, f := range finds {
						if s, ok := f.(string); ok {
							discoveredFindings = append(discoveredFindings, s)
						}
					}
				}
			}
		}
	}

	review := services.BuildEvidenceReviewForSession(scenario, session, actions)

	score := 0.0
	if session.FinalScore != nil {
		score = *session.FinalScore
	}
	completedStr := ""
	if session.CompletedAt != nil {
		completedStr = session.CompletedAt.Format(time.RFC3339)
	}

	detail := dto.SpecialistSubmissionDetail{
		SessionID:               session.SessionID,
		LearnerNickname:         user.Nickname,
		LearnerTier:             user.CurrentTier,
		ScenarioID:              scenario.ScenarioID,
		ScenarioTitle:           scenario.Title,
		ScenarioDescription:     scenario.Description,
		ScenarioTranslations:    decodeAnyJSONMap(scenario.Translations),
		DomainID:                scenario.DomainID,
		IsTruePositive:          scenario.IsTruePositive,
		TpFpExplanation:         scenario.TpFpExplanation,
		LearnerChoice:           learnerChoice,
		TPFPCorrect:             tpfpCorrect,
		TPFPPoints:              tpfpPoints,
		DiscoveredFindings:      discoveredFindings,
		ExpectedFindings:        scenario.ExpectedOutcomes,
		SelectedActions:         selectedActions,
		ExpectedActions:         scenario.PlaybookSteps,
		FinalScore:              score,
		SelectedLogIndices:      review.SelectedLogIndices,
		EvidenceFindings:        evidenceFindingDTOs(review.EvidenceFindings),
		NoiseSelections:         evidenceLogDTOs(review.NoiseSelections),
		GrossFindingsPoints:     review.GrossFindingsPoints,
		FindingsPoints:          review.FindingsPoints,
		EvidencePenalty:         review.EvidencePenalty,
		EvidenceReviewAvailable: !review.Legacy,
		Timeline:                timeline,
		CompletedAt:             completedStr,
	}

	c.JSON(http.StatusOK, detail)
}

func evidenceLogDTOs(snippets []services.EvidenceLogSnippet) []dto.EvidenceLogDTO {
	if snippets == nil {
		return nil
	}
	logs := make([]dto.EvidenceLogDTO, 0, len(snippets))
	for _, snippet := range snippets {
		logs = append(logs, dto.EvidenceLogDTO{
			Index:     snippet.Index,
			Timestamp: snippet.Timestamp,
			Source:    snippet.Source,
			Severity:  snippet.Severity,
			Message:   snippet.Message,
		})
	}
	return logs
}

func evidenceFindingDTOs(findings []services.EvidenceFindingReview) []dto.EvidenceFindingDTO {
	if findings == nil {
		return nil
	}
	items := make([]dto.EvidenceFindingDTO, 0, len(findings))
	for _, finding := range findings {
		items = append(items, dto.EvidenceFindingDTO{
			ID:                 finding.ID,
			Description:        finding.Description,
			DomainID:           finding.DomainID,
			Found:              finding.Found,
			Points:             finding.Points,
			Earned:             finding.Earned,
			EvidenceLogIndices: finding.EvidenceLogIndices,
			MatchedLogIndices:  finding.MatchedLogIndices,
			EvidenceLogs:       evidenceLogDTOs(finding.EvidenceLogs),
			Explanation:        finding.Explanation,
		})
	}
	return items
}
