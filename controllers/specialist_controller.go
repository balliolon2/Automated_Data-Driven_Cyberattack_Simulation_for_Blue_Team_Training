package controllers

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cybersim/dto"
	"cybersim/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

type SpecialistController struct {
	DB *gorm.DB
}

func NewSpecialistController(db *gorm.DB) *SpecialistController {
	return &SpecialistController{DB: db}
}

const (
	MaxFileSize     = 10 * 1024 * 1024 // 10MB
	UploadBaseDir   = "./uploads/applications"
)

var allowedExtensions = map[string]bool{
	".pdf":  true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
}

// Apply handles specialist application submission with multipart form and documents
func (sc *SpecialistController) Apply(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(MaxFileSize * 2); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload too large or invalid multipart form"})
		return
	}

	bio := strings.TrimSpace(c.PostForm("bio"))
	if bio == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bio is required"})
		return
	}

	linkedinURL := strings.TrimSpace(c.PostForm("linkedin_url"))
	portfolioURL := strings.TrimSpace(c.PostForm("portfolio_url"))

	// Validate resume file
	resumeHeader, err := c.FormFile("resume")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Resume file is required"})
		return
	}

	if resumeHeader.Size > MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Resume file exceeds 10MB limit"})
		return
	}

	resumeExt := strings.ToLower(filepath.Ext(resumeHeader.Filename))
	if !allowedExtensions[resumeExt] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resume file type. Allowed: PDF, PNG, JPG, JPEG"})
		return
	}

	// Prepare application directory
	appID := generateUUID()
	appDir := filepath.Join(UploadBaseDir, appID)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload storage directory"})
		return
	}

	// Save resume
	resumeFilename := fmt.Sprintf("resume_%s%s", generateUUID(), resumeExt)
	resumeSavePath := filepath.Join(appDir, resumeFilename)
	if err := c.SaveUploadedFile(resumeHeader, resumeSavePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save resume file"})
		return
	}

	// Optional certificate file
	var certSavePath string
	certHeader, err := c.FormFile("certificate")
	if err == nil && certHeader != nil {
		if certHeader.Size > MaxFileSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Certificate file exceeds 10MB limit"})
			return
		}
		certExt := strings.ToLower(filepath.Ext(certHeader.Filename))
		if !allowedExtensions[certExt] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid certificate file type. Allowed: PDF, PNG, JPG, JPEG"})
			return
		}
		certFilename := fmt.Sprintf("certificate_%s%s", generateUUID(), certExt)
		certSavePath = filepath.Join(appDir, certFilename)
		if err := c.SaveUploadedFile(certHeader, certSavePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save certificate file"})
			return
		}
	}

	// Check if applicant already has a previous application
	var existing models.SpecialistApplication
	now := time.Now()
	if err := sc.DB.Where("user_id = ?", userID).Order("created_at desc").First(&existing).Error; err == nil {
		if existing.Status == models.ApplicationStatusPending {
			c.JSON(http.StatusBadRequest, gin.H{"error": "An application is already pending administrator review"})
			return
		}
		if existing.Status == models.ApplicationStatusApproved {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You are already an approved Specialist"})
			return
		}
		if existing.Status == models.ApplicationStatusRejected {
			// Update rejected application to pending (re-application)
			existing.Status = models.ApplicationStatusPending
			existing.Bio = bio
			existing.ResumePath = resumeSavePath
			if certSavePath != "" {
				existing.CertificatePath = certSavePath
			}
			existing.LinkedInURL = linkedinURL
			existing.PortfolioURL = portfolioURL
			existing.RejectionReason = ""
			existing.ReviewedBy = nil
			existing.ReviewedAt = nil
			existing.UpdatedAt = now
			if err := sc.DB.Save(&existing).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update application"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message":        "Application re-submitted successfully",
				"application_id": existing.ApplicationID,
				"status":         existing.Status,
			})
			return
		}
	}

	application := models.SpecialistApplication{
		ApplicationID:   appID,
		UserID:          userID.(string),
		Status:          models.ApplicationStatusPending,
		Bio:             bio,
		ResumePath:      resumeSavePath,
		CertificatePath: certSavePath,
		LinkedInURL:     linkedinURL,
		PortfolioURL:    portfolioURL,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := sc.DB.Create(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create application"})
		return
	}

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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var app models.SpecialistApplication
	if err := sc.DB.Where("user_id = ?", userID).Order("created_at desc").First(&app).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "No specialist application found"})
		return
	}

	var user models.User
	_ = sc.DB.Where("user_id = ?", userID).First(&user)

	c.JSON(http.StatusOK, app.ToDTO(user.Nickname, user.Email))
}


// GetSubmissions retrieves completed simulation sessions with anonymized learner nicknames
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
		Select("simulation_sessions.session_id, simulation_sessions.scenario_id, simulation_sessions.final_score, simulation_sessions.completed_at, simulation_sessions.total_actions, scenarios.title as scenario_title, scenarios.domain_id, users.nickname as learner_nickname, users.current_tier as learner_tier").
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
		SessionID       string     `json:"session_id"`
		ScenarioID      string     `json:"scenario_id"`
		ScenarioTitle   string     `json:"scenario_title"`
		DomainID        string     `json:"domain_id"`
		LearnerNickname string     `json:"learner_nickname"`
		LearnerTier     int        `json:"learner_tier"`
		FinalScore      *float64   `json:"final_score"`
		TotalActions    int        `json:"total_actions"`
		CompletedAt     *time.Time `json:"completed_at"`
	}

	var rows []submissionRow
	if err := query.Order("simulation_sessions.completed_at desc").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions"})
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

		// Count TP/FP correctness and findings for this session
		var tpfpCorrect bool
		var findingsFoundCount int
		var totalFindingsCount int

		var triageAction models.SessionAction
		if err := sc.DB.Where("session_id = ? AND action_type = 'triage_alert'", r.SessionID).First(&triageAction).Error; err == nil && triageAction.IsCorrect != nil {
			tpfpCorrect = *triageAction.IsCorrect
		}

		var findingsAction models.SessionAction
		if err := sc.DB.Where("session_id = ? AND action_type = 'submit_decision'", r.SessionID).First(&findingsAction).Error; err == nil && findingsAction.Payload != nil {
			if payloadMap, ok := findingsAction.Payload.(map[string]interface{}); ok {
				if fList, ok := payloadMap["discovered_findings"].([]interface{}); ok {
					findingsFoundCount = len(fList)
				}
			}
		}

		// Calculate total findings from scenario
		var scn models.Scenario
		if err := sc.DB.Select("expected_outcomes").Where("scenario_id = ?", r.ScenarioID).First(&scn).Error; err == nil && scn.ExpectedOutcomes != nil {
			if outList, ok := scn.ExpectedOutcomes.([]interface{}); ok {
				totalFindingsCount = len(outList)
			}
		}

		domainName := domainNames[r.DomainID]
		if domainName == "" {
			domainName = r.DomainID
		}

		result = append(result, dto.SpecialistSubmissionSummary{
			SessionID:          r.SessionID,
			LearnerNickname:    r.LearnerNickname,
			LearnerTier:        r.LearnerTier,
			ScenarioID:         r.ScenarioID,
			ScenarioTitle:      r.ScenarioTitle,
			DomainID:           r.DomainID,
			DomainName:         domainName,
			FinalScore:         score,
			TPFPCorrect:        tpfpCorrect,
			FindingsFoundCount: findingsFoundCount,
			TotalFindingsCount: totalFindingsCount,
			ActionsCount:       r.TotalActions,
			CompletedAt:        completedStr,
		})
	}

	c.JSON(http.StatusOK, result)
}

// GetSubmissionDetail provides deep-dive investigation trace for an attempt
func (sc *SpecialistController) GetSubmissionDetail(c *gin.Context) {
	sessionID := c.Param("session_id")

	var session models.SimulationSession
	if err := sc.DB.Where("session_id = ?", sessionID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	var scenario models.Scenario
	if err := sc.DB.Where("scenario_id = ?", session.ScenarioID).First(&scenario).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scenario not found"})
		return
	}

	var user models.User
	if err := sc.DB.Where("user_id = ?", session.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
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

	score := 0.0
	if session.FinalScore != nil {
		score = *session.FinalScore
	}
	completedStr := ""
	if session.CompletedAt != nil {
		completedStr = session.CompletedAt.Format(time.RFC3339)
	}

	detail := dto.SpecialistSubmissionDetail{
		SessionID:           session.SessionID,
		LearnerNickname:     user.Nickname,
		LearnerTier:         user.CurrentTier,
		ScenarioID:          scenario.ScenarioID,
		ScenarioTitle:       scenario.Title,
		ScenarioDescription: scenario.Description,
		DomainID:            scenario.DomainID,
		IsTruePositive:      scenario.IsTruePositive,
		TpFpExplanation:     scenario.TpFpExplanation,
		LearnerChoice:       learnerChoice,
		TPFPCorrect:         tpfpCorrect,
		TPFPPoints:          tpfpPoints,
		DiscoveredFindings:  discoveredFindings,
		ExpectedFindings:    scenario.ExpectedOutcomes,
		SelectedActions:     selectedActions,
		ExpectedActions:     scenario.PlaybookSteps,
		FinalScore:          score,
		Timeline:            timeline,
		CompletedAt:         completedStr,
	}

	c.JSON(http.StatusOK, detail)
}
