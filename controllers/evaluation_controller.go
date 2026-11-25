package controllers

import (
	"errors"
	"math"
	"net/http"
	"strings"
	"time"

	"cybersim/dto"
	"cybersim/models"
	"cybersim/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ErrQuestionnaireAlreadySubmitted signals a duplicate submission attempt for a learner.
var ErrQuestionnaireAlreadySubmitted = errors.New("questionnaire already submitted")

// EvaluationStore abstracts persistence for EvaluationController handlers so
// HTTP tests can exercise every rule without a live PostgreSQL database.
type EvaluationStore interface {
	FindLatestCompletedExam(userID, examType string) (*models.ExamSession, error)
	FindEvaluationMetricByUser(userID string) (*models.EvaluationMetric, error)
	GetLearnerTrainingSnapshot(userID string) (int, []models.UserSkillProfile, error)
	CreateEvaluationMetricAtomic(metric *models.EvaluationMetric) error
	GetAdminLearnerEvaluations() ([]services.AdminLearnerEvaluationRecord, error)
	GetLearnerCohortFunnel() (dto.EvaluationCohortFunnelDTO, error)
}

type gormEvaluationStore struct {
	DB *gorm.DB
}

func (s *gormEvaluationStore) FindLatestCompletedExam(userID, examType string) (*models.ExamSession, error) {
	var session models.ExamSession
	if err := s.DB.Where("user_id = ? AND exam_type = ? AND status = 'completed'", userID, examType).
		Order("completed_at DESC, started_at DESC").
		First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *gormEvaluationStore) FindEvaluationMetricByUser(userID string) (*models.EvaluationMetric, error) {
	var metric models.EvaluationMetric
	if err := s.DB.Where("user_id = ?", userID).Order("eval_date DESC").First(&metric).Error; err != nil {
		return nil, err
	}
	return &metric, nil
}

func (s *gormEvaluationStore) GetLearnerTrainingSnapshot(userID string) (int, []models.UserSkillProfile, error) {
	var completedCount int64
	if err := s.DB.Model(&models.SimulationSession{}).
		Where("user_id = ? AND status = 'completed'", userID).
		Count(&completedCount).Error; err != nil {
		return 0, nil, err
	}
	var profiles []models.UserSkillProfile
	if err := s.DB.Where("user_id = ?", userID).Order("domain_id ASC").Find(&profiles).Error; err != nil {
		return 0, nil, err
	}
	return int(completedCount), profiles, nil
}

func (s *gormEvaluationStore) CreateEvaluationMetricAtomic(metric *models.EvaluationMetric) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		// Transaction-scoped advisory lock enforces 1 submission per learner without DDL changes.
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", "eval_metric:"+metric.UserID).Error; err != nil {
			return err
		}
		var existingCount int64
		if err := tx.Model(&models.EvaluationMetric{}).Where("user_id = ?", metric.UserID).Count(&existingCount).Error; err != nil {
			return err
		}
		if existingCount > 0 {
			return ErrQuestionnaireAlreadySubmitted
		}
		return tx.Create(metric).Error
	})
}

func (s *gormEvaluationStore) GetAdminLearnerEvaluations() ([]services.AdminLearnerEvaluationRecord, error) {
	var metrics []models.EvaluationMetric
	if err := s.DB.Model(&models.EvaluationMetric{}).
		Joins("JOIN users ON users.user_id = evaluation_metrics.user_id").
		Where("users.role = ?", "learner").
		Order("evaluation_metrics.eval_date DESC").
		Find(&metrics).Error; err != nil {
		return nil, err
	}
	if len(metrics) == 0 {
		return []services.AdminLearnerEvaluationRecord{}, nil
	}

	userIDs := make([]string, 0, len(metrics))
	seen := make(map[string]bool, len(metrics))
	for _, m := range metrics {
		if !seen[m.UserID] {
			seen[m.UserID] = true
			userIDs = append(userIDs, m.UserID)
		}
	}

	var users []struct {
		UserID   string `gorm:"column:user_id"`
		Nickname string `gorm:"column:nickname"`
	}
	if err := s.DB.Model(&models.User{}).
		Select("user_id, nickname").
		Where("user_id IN ?", userIDs).
		Scan(&users).Error; err != nil {
		return nil, err
	}

	nickByUser := make(map[string]string, len(users))
	for _, u := range users {
		nickByUser[u.UserID] = u.Nickname
	}

	records := make([]services.AdminLearnerEvaluationRecord, 0, len(metrics))
	for _, m := range metrics {
		records = append(records, services.AdminLearnerEvaluationRecord{
			UserID:   m.UserID,
			Nickname: nickByUser[m.UserID],
			Metric:   m,
		})
	}
	return records, nil
}

func (s *gormEvaluationStore) GetLearnerCohortFunnel() (dto.EvaluationCohortFunnelDTO, error) {
	var totalLearners int64
	if err := s.DB.Model(&models.User{}).Where("role = ?", "learner").Count(&totalLearners).Error; err != nil {
		return dto.EvaluationCohortFunnelDTO{}, err
	}

	var preCompleted int64
	if err := s.DB.Table("exam_sessions").
		Joins("JOIN users ON users.user_id = exam_sessions.user_id").
		Where("users.role = ? AND exam_sessions.exam_type = ? AND exam_sessions.status = ?", "learner", "pre", "completed").
		Distinct("exam_sessions.user_id").
		Count(&preCompleted).Error; err != nil {
		return dto.EvaluationCohortFunnelDTO{}, err
	}

	var postCompleted int64
	if err := s.DB.Table("exam_sessions").
		Joins("JOIN users ON users.user_id = exam_sessions.user_id").
		Where("users.role = ? AND exam_sessions.exam_type = ? AND exam_sessions.status = ?", "learner", "post", "completed").
		Distinct("exam_sessions.user_id").
		Count(&postCompleted).Error; err != nil {
		return dto.EvaluationCohortFunnelDTO{}, err
	}

	var respondents int64
	if err := s.DB.Table("evaluation_metrics").
		Joins("JOIN users ON users.user_id = evaluation_metrics.user_id").
		Where("users.role = ?", "learner").
		Distinct("evaluation_metrics.user_id").
		Count(&respondents).Error; err != nil {
		return dto.EvaluationCohortFunnelDTO{}, err
	}

	responseRate := 0.0
	if postCompleted > 0 {
		responseRate = roundMetric2((float64(respondents) / float64(postCompleted)) * 100.0)
	}

	return dto.EvaluationCohortFunnelDTO{
		TotalLearners:             int(totalLearners),
		CompletedPreTestLearners:  int(preCompleted),
		CompletedPostTestLearners: int(postCompleted),
		QuestionnaireRespondents:  int(respondents),
		ResponseRatePct:           responseRate,
	}, nil
}

// EvaluationController manages learner post-test evaluation questionnaires and admin reporting.
type EvaluationController struct {
	DB    *gorm.DB
	store EvaluationStore
}

// NewEvaluationController constructs an EvaluationController backed by GORM/PostgreSQL.
func NewEvaluationController(db *gorm.DB) *EvaluationController {
	return &EvaluationController{
		DB:    db,
		store: &gormEvaluationStore{DB: db},
	}
}

// NewEvaluationControllerWithStore constructs an EvaluationController with an injectable store for testing.
func NewEvaluationControllerWithStore(store EvaluationStore) *EvaluationController {
	return &EvaluationController{store: store}
}

func roundMetric2(val float64) float64 {
	return math.Round(val*100) / 100
}

func extractCallerIdentity(c *gin.Context) (string, string, bool) {
	rawID, exists := c.Get("user_id")
	if !exists {
		return "", "", false
	}
	userID, ok := rawID.(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return "", "", false
	}
	rawRole, _ := c.Get("role")
	role, _ := rawRole.(string)
	if role == "" {
		role = "learner"
	}
	return userID, role, true
}

// GetQuestionnaireStatus godoc
// @Summary Get post-test evaluation questionnaire eligibility and submission state
// @Description Returns whether the authenticated user can submit the post-test evaluation questionnaire, preview-only status for non-learners, and any submitted evaluation summary.
// @Tags Evaluation
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.QuestionnaireStatusResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /evaluation/questionnaire/status [get]
const (
	ineligiblePreviewOnlyNonLearner = "preview_only_non_learner"
	ineligibleAlreadySubmitted      = "already_submitted"
	ineligiblePostTestRequired      = "post_test_required"
)

func (ec *EvaluationController) fetchAdminCohortData(c *gin.Context) ([]services.AdminLearnerEvaluationRecord, dto.EvaluationCohortFunnelDTO, bool) {
	records, err := ec.store.GetAdminLearnerEvaluations()
	if err != nil {
		respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to load evaluation metrics")
		return nil, dto.EvaluationCohortFunnelDTO{}, false
	}
	cohort, err := ec.store.GetLearnerCohortFunnel()
	if err != nil {
		respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to load learner cohort metrics")
		return nil, dto.EvaluationCohortFunnelDTO{}, false
	}
	return records, cohort, true
}

func (ec *EvaluationController) GetQuestionnaireStatus(c *gin.Context) {
	userID, role, ok := extractCallerIdentity(c)
	if !ok {
		respondAPIError(c, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Unauthorized")
		return
	}

	if role != "learner" {
		c.JSON(http.StatusOK, dto.QuestionnaireStatusResponse{
			Role:             role,
			IsPreviewOnly:    true,
			Eligible:         false,
			IneligibleReason: ineligiblePreviewOnlyNonLearner,
		})
		return
	}

	postExam, err := ec.store.FindLatestCompletedExam(userID, "post")
	hasCompletedPost := err == nil && postExam != nil
	var latestPostID *string
	if hasCompletedPost {
		idCopy := postExam.SessionID
		latestPostID = &idCopy
	}

	existing, err := ec.store.FindEvaluationMetricByUser(userID)
	if err == nil && existing != nil {
		summary := services.EvaluationMetricToSummaryDTO(*existing)
		c.JSON(http.StatusOK, dto.QuestionnaireStatusResponse{
			Role:                 role,
			IsPreviewOnly:        false,
			HasCompletedPostTest: hasCompletedPost,
			LatestPostSessionID:  latestPostID,
			Eligible:             false,
			IneligibleReason:     ineligibleAlreadySubmitted,
			Submitted:            true,
			Submission:           &summary,
		})
		return
	}

	if !hasCompletedPost {
		c.JSON(http.StatusOK, dto.QuestionnaireStatusResponse{
			Role:                 role,
			IsPreviewOnly:        false,
			HasCompletedPostTest: false,
			Eligible:             false,
			IneligibleReason:     ineligiblePostTestRequired,
			Submitted:            false,
		})
		return
	}

	c.JSON(http.StatusOK, dto.QuestionnaireStatusResponse{
		Role:                 role,
		IsPreviewOnly:        false,
		HasCompletedPostTest: true,
		LatestPostSessionID:  latestPostID,
		Eligible:             true,
		Submitted:            false,
	})
}

// SubmitQuestionnaire godoc
// @Summary Submit post-test evaluation questionnaire
// @Description Validates and records one CSUQ v3 + Raw NASA-TLX + qualitative evaluation per learner after completing the Post-Test, snapshotting Pre/Post and Skill Profile data server-side.
// @Tags Evaluation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body dto.SubmitQuestionnaireRequest true "Questionnaire ratings and qualitative impressions"
// @Success 201 {object} dto.SubmitQuestionnaireResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /evaluation/questionnaire [post]
func (ec *EvaluationController) SubmitQuestionnaire(c *gin.Context) {
	userID, role, ok := extractCallerIdentity(c)
	if !ok {
		respondAPIError(c, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Unauthorized")
		return
	}

	if role != "learner" {
		respondAPIError(c, http.StatusForbidden, "QUESTIONNAIRE_LEARNER_ONLY", "Only learner accounts can submit research evaluation questionnaires")
		return
	}

	postExam, err := ec.store.FindLatestCompletedExam(userID, "post")
	if err != nil || postExam == nil {
		respondAPIError(c, http.StatusForbidden, "QUESTIONNAIRE_POST_TEST_REQUIRED", "Complete the Post-Test before submitting the evaluation questionnaire")
		return
	}

	var input dto.SubmitQuestionnaireRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondAPIError(c, http.StatusBadRequest, "QUESTIONNAIRE_INVALID_CSUQ", "Invalid questionnaire submission payload")
		return
	}

	csuqPayload, err := services.ValidateAndScoreCSUQ(input.CSUQ, input.Locale)
	if err != nil {
		if errors.Is(err, services.ErrQuestionnaireAllNA) {
			respondAPIError(c, http.StatusBadRequest, "QUESTIONNAIRE_ALL_NA", "Please rate at least one CSUQ item with a numeric score (1–7)")
			return
		}
		respondAPIError(c, http.StatusBadRequest, "QUESTIONNAIRE_INVALID_CSUQ", "All 16 CSUQ items must be answered with 1–7 or N/A")
		return
	}

	tlxPayload, err := services.ValidateAndScoreNASATLX(input.NASATLX)
	if err != nil {
		respondAPIError(c, http.StatusBadRequest, "QUESTIONNAIRE_INVALID_TLX", "All 6 NASA-TLX dimensions must be multiples of 5 between 0 and 100")
		return
	}

	cleanedFeedback, formattedNotes, err := services.SanitizeQualitativeFeedback(input.QualitativeFeedback)
	if err != nil {
		respondAPIError(c, http.StatusBadRequest, "QUESTIONNAIRE_FEEDBACK_TOO_LONG", "Each qualitative feedback field must be 2,000 characters or fewer")
		return
	}

	// Authoritative server-side snapshot of Pre-Test, Post-Test, and Skill Profiles.
	preScore := 0.0
	var preSessionID *string
	if preExam, preErr := ec.store.FindLatestCompletedExam(userID, "pre"); preErr == nil && preExam != nil {
		idCopy := preExam.SessionID
		preSessionID = &idCopy
		if preExam.Score != nil {
			preScore = roundMetric2(*preExam.Score)
		}
	}

	postScore := 0.0
	if postExam.Score != nil {
		postScore = roundMetric2(*postExam.Score)
	}
	improvementPct := roundMetric2(postScore - preScore)

	completedScenarios, profiles, err := ec.store.GetLearnerTrainingSnapshot(userID)
	if err != nil {
		respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to snapshot learner training context")
		return
	}

	domainStatuses := make([]dto.DomainStatusDTO, 0, len(profiles))
	sumProf := 0.0
	for _, p := range profiles {
		prof := roundMetric2(p.ProficiencyScore)
		sumProf += prof
		domainStatuses = append(domainStatuses, dto.DomainStatusDTO{
			DomainID:         p.DomainID,
			DomainName:       domainNames[p.DomainID],
			ProficiencyScore: prof,
			Threshold:        ProficiencyThreshold,
			Passed:           p.ProficiencyScore >= ProficiencyThreshold,
		})
	}
	avgProf := 0.0
	if len(profiles) > 0 {
		avgProf = roundMetric2(sumProf / float64(len(profiles)))
	}

	locale := services.NormalizeQuestionnaireLocale(input.Locale)
	snapshot := dto.EvaluationSkillGapSnapshotDTO{
		SnapshotVersion:     services.EvalSnapshotVersion,
		SubmittedLocale:     locale,
		PreSessionID:        preSessionID,
		PostSessionID:       postExam.SessionID,
		CompletedScenarios:  completedScenarios,
		AverageProficiency:  avgProf,
		Domains:             domainStatuses,
		QualitativeFeedback: cleanedFeedback,
	}

	metric := models.EvaluationMetric{
		UserID:          userID,
		PreScore:        preScore,
		PostScore:       postScore,
		ImprovementPct:  improvementPct,
		CSUQScores:      csuqPayload,
		NASATLXScores:   tlxPayload,
		SkillGapSummary: snapshot,
		EvalDate:        time.Now().UTC(),
		Notes:           formattedNotes,
	}

	if err := ec.store.CreateEvaluationMetricAtomic(&metric); err != nil {
		if errors.Is(err, ErrQuestionnaireAlreadySubmitted) {
			respondAPIError(c, http.StatusConflict, "QUESTIONNAIRE_ALREADY_SUBMITTED", "You have already submitted your post-test evaluation questionnaire")
			return
		}
		respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to save evaluation questionnaire")
		return
	}

	c.JSON(http.StatusCreated, dto.SubmitQuestionnaireResponse{
		Message:    "Post-test evaluation questionnaire submitted successfully.",
		Submission: services.EvaluationMetricToSummaryDTO(metric),
	})
}

// GetAdminEvaluations godoc
// @Summary List pseudonymized learner post-test evaluations and summary KPIs (Admin only)
// @Description Returns learner cohort funnel counts, aggregate CSUQ/NASA-TLX summary stats, and pseudonymized individual evaluation rows without email PII.
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.AdminEvaluationsResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/evaluations [get]
func (ec *EvaluationController) GetAdminEvaluations(c *gin.Context) {
	records, cohort, ok := ec.fetchAdminCohortData(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, services.BuildAdminEvaluationsResponse(records, cohort))
}

// GetAdminStatistics godoc
// @Summary Get paper-ready statistical analysis report (Admin only)
// @Description Computes descriptive statistics (Student-t 95% CI, R-7 IQR), Cronbach's alpha, Pre-to-Post Paired t-test (Cohen's d_z) and Wilcoxon signed-rank (rank-biserial r_rb), and Spearman rank correlations over learner evaluations.
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.AdminPaperStatisticsResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/evaluations/statistics [get]
func (ec *EvaluationController) GetAdminStatistics(c *gin.Context) {
	records, cohort, ok := ec.fetchAdminCohortData(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, services.BuildPaperStatisticsReport(records, cohort))
}

// ExportAdminEvaluations godoc
// @Summary Export pseudonymized learner evaluations as UTF-8 BOM CSV or JSON (Admin only)
// @Description Streams RFC 4180 CSV bytes (default format=csv, with UTF-8 BOM and formula-injection neutralization) or returns AdminEvaluationsExportJSONResponse (format=json) for Excel, SPSS, or Python analysis.
// @Tags Admin
// @Produce text/csv
// @Produce application/json
// @Security BearerAuth
// @Param format query string false "Export format (csv or json)" default(csv)
// @Success 200 {object} dto.AdminEvaluationsExportJSONResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/evaluations/export [get]
func (ec *EvaluationController) ExportAdminEvaluations(c *gin.Context) {
	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "csv")))
	if format != "csv" && format != "json" {
		respondAPIError(c, http.StatusBadRequest, "EVALUATION_INVALID_EXPORT_FORMAT", "Unsupported export format. Expected 'csv' or 'json'.")
		return
	}

	records, cohort, ok := ec.fetchAdminCohortData(c)
	if !ok {
		return
	}

	dashboard := services.BuildAdminEvaluationsResponse(records, cohort)
	dateTag := time.Now().UTC().Format("20060102")

	if format == "json" {
		c.Header("Content-Disposition", "attachment; filename=\"soc_trainer_evaluations_"+dateTag+".json\"")
		c.JSON(http.StatusOK, dto.AdminEvaluationsExportJSONResponse{
			ExportedAt: time.Now().UTC(),
			Cohort:     dashboard.Cohort,
			Summary:    dashboard.Summary,
			Responses:  dashboard.Responses,
		})
		return
	}

	csvBytes, err := services.BuildEvaluationsCSV(dashboard.Responses)
	if err != nil {
		respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to generate CSV export")
		return
	}
	c.Header("Content-Disposition", "attachment; filename=\"soc_trainer_evaluations_"+dateTag+".csv\"")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvBytes)
}

