package controllers

import (
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cybersim/dto"
	"cybersim/models"
	"cybersim/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ExamDomainScore = dto.DomainScoreDTO
type ExamDetailedResult = dto.DetailedExamResultDTO

var (
	examUUIDPattern         = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	errExamSessionAbandoned = errors.New("exam session is abandoned")
)

func isValidExamUUID(id string) bool {
	return examUUIDPattern.MatchString(strings.TrimSpace(id))
}

func examGateThreshold() float64 {
	threshold := ProficiencyThreshold
	if envThreshold := os.Getenv("PROFICIENCY_THRESHOLD"); envThreshold != "" {
		if val, err := strconv.ParseFloat(envThreshold, 64); err == nil {
			threshold = val
		}
	}
	return threshold
}

func examDisplayThreshold() float64 {
	threshold := examGateThreshold()
	if threshold <= 0 {
		return ProficiencyThreshold
	}
	return threshold
}

// ExamStore abstracts persistence for ExamController handlers so the HTTP
// layer can be tested cleanly without a live PostgreSQL instance.
type ExamStore interface {
	GetActiveSessionForUser(userID string) (*models.ExamSession, error)
	HasCompletedPreTest(userID string) (bool, error)
	GetUserSkillProfiles(userID string) ([]models.UserSkillProfile, error)
	SelectQuestionsByDomains(userID string, domainCounts map[string]int, excludePreAndLearn bool) ([]models.Question, error)
	CreateExamSession(session *models.ExamSession, questions []models.Question) error
	GetSessionQuestions(sessionID string) ([]QuestionResponse, error)
	FindActiveSessionByIDAndUser(sessionID, userID string) (*models.ExamSession, error)
	FindSessionByIDAndUser(sessionID, userID string) (*models.ExamSession, error)
	FindQuestionByID(questionID string) (*models.Question, error)
	FindSessionQuestion(sessionID, questionID string) (*models.ExamSessionQuestion, error)
	SaveExamAnswer(session *models.ExamSession, esq *models.ExamSessionQuestion) (int64, error)
	FinalizeExamSession(session *models.ExamSession, userID string) (*services.ExamReviewData, error)
	MarkUnansweredAndFinalize(session *models.ExamSession, userID string) (*services.ExamReviewData, error)
	GetExamSessionReviewData(session *models.ExamSession) (*services.ExamReviewData, error)
	GetCompletedExamHistory(userID string) ([]dto.ExamHistoryItemDTO, error)
	GetLatestCompletedPreTestSession(userID string) (*models.ExamSession, error)
	GetRandomSampleQuestion() (*models.Question, error)
}

type gormExamStore struct {
	DB *gorm.DB
}

func (s *gormExamStore) GetActiveSessionForUser(userID string) (*models.ExamSession, error) {
	var activeSession models.ExamSession
	if err := s.DB.Where("user_id = ? AND status = 'in_progress'", userID).First(&activeSession).Error; err != nil {
		return nil, err
	}
	return &activeSession, nil
}

func (s *gormExamStore) HasCompletedPreTest(userID string) (bool, error) {
	var preTestSession models.ExamSession
	if err := s.DB.Where("user_id = ? AND exam_type = 'pre' AND status = 'completed'", userID).First(&preTestSession).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *gormExamStore) GetUserSkillProfiles(userID string) ([]models.UserSkillProfile, error) {
	var profiles []models.UserSkillProfile
	if err := s.DB.Where("user_id = ?", userID).Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

func (s *gormExamStore) SelectQuestionsByDomains(userID string, domainCounts map[string]int, excludePreAndLearn bool) ([]models.Question, error) {
	var allSelectedQuestions []models.Question
	for domain, count := range domainCounts {
		var qList []models.Question
		query := s.DB.Where("domain_id = ? AND is_active = true", domain)
		if excludePreAndLearn {
			query = s.DB.Where("domain_id = ? AND is_active = true AND question_id NOT IN (?)", domain, BuildPostTestExclusionQuery(s.DB, userID))
		}
		if err := query.Order("RANDOM()").Limit(count).Find(&qList).Error; err != nil {
			return nil, err
		}
		allSelectedQuestions = append(allSelectedQuestions, qList...)
	}
	return allSelectedQuestions, nil
}

func (s *gormExamStore) CreateExamSession(session *models.ExamSession, questions []models.Question) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(session).Error; err != nil {
			return err
		}
		for orderIndex, q := range questions {
			esq := models.ExamSessionQuestion{
				SessionID:  session.SessionID,
				QuestionID: q.QuestionID,
				OrderIndex: orderIndex,
			}
			if err := tx.Create(&esq).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *gormExamStore) GetSessionQuestions(sessionID string) ([]QuestionResponse, error) {
	var results []struct {
		models.Question
		OrderIndex int    `gorm:"column:order_index"`
		UserAnswer string `gorm:"column:user_answer"`
	}

	err := s.DB.Table("exam_session_questions").
		Select("questions.*, exam_session_questions.order_index, exam_session_questions.user_answer").
		Joins("join questions on questions.question_id = exam_session_questions.question_id").
		Where("exam_session_questions.session_id = ?", sessionID).
		Order("exam_session_questions.order_index asc").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	resp := make([]QuestionResponse, len(results))
	for i, r := range results {
		resp[i] = QuestionResponse{
			QuestionID:   r.QuestionID,
			DomainID:     r.DomainID,
			Type:         r.Type,
			QuestionText: r.QuestionText,
			Options:      r.Options,
			Translations: services.SanitizeQuestionTranslations(r.Translations, false),
			OrderIndex:   r.OrderIndex,
			UserAnswer:   r.UserAnswer,
		}
	}
	return resp, nil
}

func (s *gormExamStore) FindActiveSessionByIDAndUser(sessionID, userID string) (*models.ExamSession, error) {
	if !isValidExamUUID(sessionID) {
		return nil, gorm.ErrRecordNotFound
	}
	var session models.ExamSession
	if err := s.DB.Where("session_id = ? AND user_id = ? AND status = 'in_progress'", sessionID, userID).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *gormExamStore) FindSessionByIDAndUser(sessionID, userID string) (*models.ExamSession, error) {
	if !isValidExamUUID(sessionID) {
		return nil, gorm.ErrRecordNotFound
	}
	var session models.ExamSession
	if err := s.DB.Where("session_id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *gormExamStore) FindQuestionByID(questionID string) (*models.Question, error) {
	var question models.Question
	if err := s.DB.Where("question_id = ?", questionID).First(&question).Error; err != nil {
		return nil, err
	}
	return &question, nil
}

func (s *gormExamStore) FindSessionQuestion(sessionID, questionID string) (*models.ExamSessionQuestion, error) {
	var esq models.ExamSessionQuestion
	if err := s.DB.Where("session_id = ? AND question_id = ?", sessionID, questionID).First(&esq).Error; err != nil {
		return nil, err
	}
	return &esq, nil
}

func (s *gormExamStore) SaveExamAnswer(session *models.ExamSession, esq *models.ExamSessionQuestion) (int64, error) {
	var unansweredCount int64
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(esq).Error; err != nil {
			return err
		}
		if session.TimeSpentSeconds != nil {
			if err := tx.Save(session).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&models.ExamSessionQuestion{}).
			Where("session_id = ? AND (user_answer IS NULL OR user_answer = '')", session.SessionID).
			Count(&unansweredCount).Error; err != nil {
			return err
		}
		return nil
	})
	return unansweredCount, err
}

func (s *gormExamStore) queryDetailedResults(db *gorm.DB, sessionID string) ([]dto.DetailedExamResultDTO, error) {
	var rows []struct {
		models.Question
		OrderIndex int    `gorm:"column:order_index"`
		UserAnswer string `gorm:"column:user_answer"`
		IsCorrect  bool   `gorm:"column:is_correct"`
	}
	if err := db.Table("exam_session_questions").
		Select("questions.*, exam_session_questions.order_index, COALESCE(exam_session_questions.user_answer, '') AS user_answer, COALESCE(exam_session_questions.is_correct, false) AS is_correct").
		Joins("JOIN questions ON questions.question_id = exam_session_questions.question_id").
		Where("exam_session_questions.session_id = ?", sessionID).
		Order("exam_session_questions.order_index ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	detailedResults := make([]dto.DetailedExamResultDTO, len(rows))
	for i, r := range rows {
		detailedResults[i] = dto.DetailedExamResultDTO{
			QuestionID:    r.QuestionID,
			DomainID:      r.DomainID,
			SubTopicID:    r.SubTopicID,
			Type:          r.Type,
			QuestionText:  r.QuestionText,
			Options:       r.Options,
			CorrectAnswer: r.CorrectAnswer,
			Explanation:   r.Explanation,
			Translations:  services.SanitizeQuestionTranslations(r.Translations, true),
			MitreRef:      r.MitreRef,
			OrderIndex:    r.OrderIndex,
			UserAnswer:    r.UserAnswer,
			IsCorrect:     r.IsCorrect,
		}
	}
	return detailedResults, nil
}

func (s *gormExamStore) finalizeSessionInternal(session *models.ExamSession, userID string, markBlankUnanswered bool) (*services.ExamReviewData, error) {
	var reviewData *services.ExamReviewData
	threshold := examDisplayThreshold()

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var locked models.ExamSession
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("session_id = ? AND user_id = ?", session.SessionID, userID).
			First(&locked).Error; err != nil {
			return err
		}
		if locked.Status == "abandoned" {
			return errExamSessionAbandoned
		}
		if locked.Status == "completed" {
			*session = locked
			results, qErr := s.queryDetailedResults(tx, locked.SessionID)
			if qErr != nil {
				return qErr
			}
			reviewData = services.SummarizeExamResults(results, threshold)
			return nil
		}

		if session.TimeSpentSeconds != nil {
			existing := 0
			if locked.TimeSpentSeconds != nil {
				existing = *locked.TimeSpentSeconds
			}
			clamped := services.ClampExamDuration(*session.TimeSpentSeconds, existing)
			locked.TimeSpentSeconds = &clamped
		}

		if markBlankUnanswered {
			if err := tx.Model(&models.ExamSessionQuestion{}).
				Where("session_id = ? AND (user_answer IS NULL OR TRIM(user_answer) = '')", locked.SessionID).
				Updates(map[string]any{
					"user_answer": "",
					"is_correct":  false,
				}).Error; err != nil {
				return err
			}
		}

		detailedResults, err := s.queryDetailedResults(tx, locked.SessionID)
		if err != nil {
			return err
		}

		reviewData = services.SummarizeExamResults(detailedResults, threshold)

		denom := locked.TotalQuestions
		if denom <= 0 {
			denom = len(detailedResults)
		}
		scorePercent := services.ComputeSessionScore(nil, reviewData.CorrectCount, denom)
		locked.Score = &scorePercent
		locked.Status = "completed"
		now := time.Now()
		locked.CompletedAt = &now

		if err := tx.Save(&locked).Error; err != nil {
			return err
		}
		*session = locked

		// Strictly preserve Proficiency Invariance: only Pre-Test updates user_skill_profiles.
		if locked.ExamType == "pre" {
			for _, ds := range reviewData.DomainScores {
				var profile models.UserSkillProfile
				if err := tx.Where("user_id = ? AND domain_id = ?", userID, ds.DomainID).First(&profile).Error; err != nil {
					profileNow := time.Now()
					profile = models.UserSkillProfile{
						UserID:           userID,
						DomainID:         ds.DomainID,
						ProficiencyScore: ds.Percentage,
						LastPracticed:    &profileNow,
					}
					if createErr := tx.Create(&profile).Error; createErr != nil {
						return createErr
					}
				} else {
					profile.ProficiencyScore = ds.Percentage
					profileNow := time.Now()
					profile.LastPracticed = &profileNow
					if saveErr := tx.Save(&profile).Error; saveErr != nil {
						return saveErr
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return reviewData, nil
}

func (s *gormExamStore) FinalizeExamSession(session *models.ExamSession, userID string) (*services.ExamReviewData, error) {
	return s.finalizeSessionInternal(session, userID, false)
}

func (s *gormExamStore) MarkUnansweredAndFinalize(session *models.ExamSession, userID string) (*services.ExamReviewData, error) {
	return s.finalizeSessionInternal(session, userID, true)
}

func (s *gormExamStore) GetExamSessionReviewData(session *models.ExamSession) (*services.ExamReviewData, error) {
	detailedResults, err := s.queryDetailedResults(s.DB, session.SessionID)
	if err != nil {
		return nil, err
	}
	return services.SummarizeExamResults(detailedResults, examDisplayThreshold()), nil
}

func (s *gormExamStore) GetCompletedExamHistory(userID string) ([]dto.ExamHistoryItemDTO, error) {
	var sessions []models.ExamSession
	if err := s.DB.Where("user_id = ? AND status = 'completed'", userID).
		Order("completed_at DESC, started_at DESC").
		Find(&sessions).Error; err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return []dto.ExamHistoryItemDTO{}, nil
	}

	sessionIDs := make([]string, len(sessions))
	for i, sess := range sessions {
		sessionIDs[i] = sess.SessionID
	}

	type aggRow struct {
		SessionID       string `gorm:"column:session_id"`
		TotalRows       int64  `gorm:"column:total_rows"`
		CorrectCount    int64  `gorm:"column:correct_count"`
		UnansweredCount int64  `gorm:"column:unanswered_count"`
	}
	var aggs []aggRow
	if err := s.DB.Table("exam_session_questions").
		Select("session_id, COUNT(*) AS total_rows, SUM(CASE WHEN is_correct = true THEN 1 ELSE 0 END) AS correct_count, SUM(CASE WHEN user_answer IS NULL OR TRIM(user_answer) = '' THEN 1 ELSE 0 END) AS unanswered_count").
		Where("session_id IN ?", sessionIDs).
		Group("session_id").
		Scan(&aggs).Error; err != nil {
		return nil, err
	}
	aggMap := make(map[string]aggRow, len(aggs))
	for _, row := range aggs {
		aggMap[row.SessionID] = row
	}

	items := make([]dto.ExamHistoryItemDTO, 0, len(sessions))
	for _, sess := range sessions {
		agg := aggMap[sess.SessionID]
		total := sess.TotalQuestions
		if agg.TotalRows > 0 {
			total = int(agg.TotalRows)
		}
		incorrect := int64(total) - agg.CorrectCount - agg.UnansweredCount
		if incorrect < 0 {
			incorrect = 0
		}
		items = append(items, dto.ExamHistoryItemDTO{
			SessionID:        sess.SessionID,
			ExamType:         sess.ExamType,
			Status:           sess.Status,
			Score:            services.ComputeSessionScore(sess.Score, agg.CorrectCount, total),
			CorrectCount:     agg.CorrectCount,
			IncorrectCount:   incorrect,
			UnansweredCount:  agg.UnansweredCount,
			TotalQuestions:   total,
			StartedAt:        sess.StartedAt,
			CompletedAt:      sess.CompletedAt,
			TimeSpentSeconds: sess.TimeSpentSeconds,
		})
	}
	return items, nil
}

func (s *gormExamStore) GetLatestCompletedPreTestSession(userID string) (*models.ExamSession, error) {
	var session models.ExamSession
	if err := s.DB.Where("user_id = ? AND exam_type = 'pre' AND status = 'completed'", userID).
		Order("completed_at DESC, started_at DESC").
		First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *gormExamStore) GetRandomSampleQuestion() (*models.Question, error) {
	var question models.Question
	if err := s.DB.Order("RANDOM()").First(&question).Error; err != nil {
		return nil, err
	}
	return &question, nil
}

type ExamController struct {
	DB    *gorm.DB
	store ExamStore
}

func NewExamController(db *gorm.DB) *ExamController {
	return &ExamController{
		DB:    db,
		store: &gormExamStore{DB: db},
	}
}

func NewExamControllerWithStore(store ExamStore) *ExamController {
	return &ExamController{store: store}
}

type QuestionResponse = dto.QuestionDTO

// Helper to load questions for a session
func (ec *ExamController) getSessionQuestions(sessionID string) ([]QuestionResponse, error) {
	return ec.store.GetSessionQuestions(sessionID)
}

// StartPreTest godoc
// @Summary Start or resume pre-test exam session
// @Description Initializes a 30-question Pre-Test with required domain distribution (4/7/5/8/6). If an in-progress pre-test already exists, returns the existing session.
// @Tags Exams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} dto.ExamSessionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /exams/pre-test [post]
func (ec *ExamController) StartPreTest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "code": "AUTH_UNAUTHORIZED"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "code": "AUTH_UNAUTHORIZED"})
		return
	}

	// Check if there is already an active session
	if activeSession, err := ec.store.GetActiveSessionForUser(userIDStr); err == nil && activeSession != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "You already have an active exam session",
			"session_id": activeSession.SessionID,
			"code":       "EXAM_SESSION_ACTIVE",
		})
		return
	}

	// Query randomized questions per domain requirements
	// Domain 1: 4, Domain 2: 7, Domain 3: 5, Domain 4: 8, Domain 5: 6
	domains := map[string]int{
		"domain1": 4,
		"domain2": 7,
		"domain3": 5,
		"domain4": 8,
		"domain5": 6,
	}

	allSelectedQuestions, err := ec.store.SelectQuestionsByDomains(userIDStr, domains, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve questions", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	if len(allSelectedQuestions) < 30 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Insufficient questions in database pool to construct 30 questions exam", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	session := models.ExamSession{
		UserID:         userIDStr,
		ExamType:       "pre",
		Status:         "in_progress",
		TotalQuestions: 30,
		StartedAt:      time.Now(),
	}

	if err := ec.store.CreateExamSession(&session, allSelectedQuestions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create exam session", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	questions, err := ec.getSessionQuestions(session.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load session questions", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session":   session,
		"questions": questions,
	})
}

// StartPostTest godoc
// @Summary Start post-test exam session
// @Description Initializes a 30-question Post-Test with zero overlap from the Pre-Test questions.
// @Tags Exams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} dto.ExamSessionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /exams/post-test [post]
func (ec *ExamController) StartPostTest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "code": "AUTH_UNAUTHORIZED"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "code": "AUTH_UNAUTHORIZED"})
		return
	}

	if activeSession, err := ec.store.GetActiveSessionForUser(userIDStr); err == nil && activeSession != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "You already have an active exam session",
			"session_id": activeSession.SessionID,
			"code":       "EXAM_SESSION_ACTIVE",
		})
		return
	}

	// 1. Verify Pre-test is completed
	completedPre, err := ec.store.HasCompletedPreTest(userIDStr)
	if err != nil || !completedPre {
		c.JSON(http.StatusForbidden, gin.H{
			"error":    "You must complete the Pre-Test before starting the Post-Test.",
			"eligible": false,
			"reason":   "pre_test_required",
			"code":     "EXAM_PRETEST_REQUIRED",
		})
		return
	}

	// 2. Verify all 5 domains have reached the proficiency threshold (Default 70%)
	threshold := examGateThreshold()
	bypassThreshold := os.Getenv("BYPASS_POST_TEST_THRESHOLD") == "true"

	if !bypassThreshold && threshold > 0 {
		profiles, err := ec.store.GetUserSkillProfiles(userIDStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify skill profiles", "code": "INTERNAL_SERVER_ERROR"})
			return
		}

		if len(profiles) < 5 {
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "You are not eligible for the Post-Test. Skill profiles for all 5 security domains must be established.",
				"eligible": false,
				"reason":   "threshold_not_reached",
				"code":     "EXAM_POSTTEST_LOCKED",
				"params":   gin.H{"threshold": threshold},
			})
			return
		}

		for _, p := range profiles {
			if p.ProficiencyScore < threshold {
				c.JSON(http.StatusForbidden, gin.H{
					"error":             "You are not eligible for the Post-Test. All 5 CompTIA Security+ domains must reach the 70% proficiency threshold in simulation training.",
					"eligible":          false,
					"reason":            "threshold_not_reached",
					"domain_id":         p.DomainID,
					"proficiency_score": p.ProficiencyScore,
					"threshold":         threshold,
					"code":              "EXAM_POSTTEST_LOCKED",
					"params": gin.H{
						"domain_id":         p.DomainID,
						"proficiency_score": p.ProficiencyScore,
						"threshold":         threshold,
					},
				})
				return
			}
		}
	}

	// Exclude all questions from user's PRE test and Learn sample questions (Zero-Overlap)
	domains := map[string]int{
		"domain1": 4,
		"domain2": 7,
		"domain3": 5,
		"domain4": 8,
		"domain5": 6,
	}

	allSelectedQuestions, err := ec.store.SelectQuestionsByDomains(userIDStr, domains, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve questions", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	if len(allSelectedQuestions) < 30 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Insufficient unused questions in database pool to construct 30 questions post-test", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	session := models.ExamSession{
		UserID:         userIDStr,
		ExamType:       "post",
		Status:         "in_progress",
		TotalQuestions: 30,
		StartedAt:      time.Now(),
	}

	if err := ec.store.CreateExamSession(&session, allSelectedQuestions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create exam session", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	questions, err := ec.getSessionQuestions(session.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load session questions", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session":   session,
		"questions": questions,
	})
}

// GetActiveSession godoc
// @Summary Get active exam session
// @Description Recovers the active exam session and questions for the authenticated learner.
// @Tags Exams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ExamSessionResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /exams/session [get]
func (ec *ExamController) GetActiveSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "code": "AUTH_UNAUTHORIZED"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "code": "AUTH_UNAUTHORIZED"})
		return
	}

	session, err := ec.store.GetActiveSessionForUser(userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "No active session found", "code": "EXAM_SESSION_NOT_FOUND"})
		return
	}

	questions, err := ec.getSessionQuestions(session.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load session questions", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session":   session,
		"questions": questions,
	})
}

type SubmitAnswerInput = dto.SubmitAnswerRequest

// SubmitAnswer godoc
// @Summary Submit an exam answer
// @Description Commits an answer for an exam question and returns progression or final score results.
// @Tags Exams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.SubmitAnswerRequest true "Exam Answer Submission"
// @Success 200 {object} dto.SubmitAnswerResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /exams/submit-answer [post]
func (ec *ExamController) SubmitAnswer(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "code": "AUTH_UNAUTHORIZED"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "code": "AUTH_UNAUTHORIZED"})
		return
	}

	var input SubmitAnswerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		code := "AUTH_INVALID_PAYLOAD"
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			code = "EXAM_INVALID_ANSWER"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": code})
		return
	}
	if strings.TrimSpace(input.UserAnswer) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Answer cannot be empty", "code": "EXAM_INVALID_ANSWER"})
		return
	}

	// Verify session ownership and status
	session, err := ec.store.FindActiveSessionByIDAndUser(input.SessionID, userIDStr)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid or inactive session", "code": "EXAM_SESSION_NOT_FOUND"})
		return
	}

	// Get the question to check correctness
	question, err := ec.store.FindQuestionByID(input.QuestionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found", "code": "EXAM_QUESTION_NOT_FOUND"})
		return
	}

	esq, err := ec.store.FindSessionQuestion(input.SessionID, input.QuestionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Question not part of this exam session", "code": "EXAM_QUESTION_NOT_FOUND"})
		return
	}

	trimmedAns := strings.TrimSpace(input.UserAnswer)
	isCorrect := strings.EqualFold(trimmedAns, strings.TrimSpace(question.CorrectAnswer))
	esq.UserAnswer = trimmedAns
	esq.IsCorrect = &isCorrect

	if input.TimeSpentSeconds > 0 {
		var currentSpent int
		if session.TimeSpentSeconds != nil {
			currentSpent = *session.TimeSpentSeconds
		}
		newSpent := services.ClampExamDuration(currentSpent+input.TimeSpentSeconds, currentSpent)
		session.TimeSpentSeconds = &newSpent
	}

	unansweredCount, err := ec.store.SaveExamAnswer(session, esq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to save answer", Code: "INTERNAL_SERVER_ERROR"})
		return
	}

	if unansweredCount == 0 {
		reviewData, err := ec.store.FinalizeExamSession(session, userIDStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to finalize session", Code: "INTERNAL_SERVER_ERROR"})
			return
		}

		totalQ := effectiveExamTotal(session, reviewData)
		c.JSON(http.StatusOK, dto.SubmitAnswerResponse{
			SessionID:        session.SessionID,
			ExamType:         session.ExamType,
			Completed:        true,
			Message:          "Exam completed successfully!",
			Score:            services.ComputeSessionScore(session.Score, reviewData.CorrectCount, totalQ),
			CorrectCount:     reviewData.CorrectCount,
			TotalQuestions:   totalQ,
			TimeSpentSeconds: session.TimeSpentSeconds,
			Results:          reviewData.Results,
			DomainScores:     reviewData.DomainScores,
		})
		return
	}

	c.JSON(http.StatusOK, dto.SubmitAnswerResponse{Completed: false})
}

func effectiveExamTotal(session *models.ExamSession, reviewData *services.ExamReviewData) int {
	if session.TotalQuestions > 0 {
		return session.TotalQuestions
	}
	if reviewData != nil {
		return len(reviewData.Results)
	}
	return 0
}

func buildExamHistoryItemDTO(session *models.ExamSession, reviewData *services.ExamReviewData) dto.ExamHistoryItemDTO {
	totalQ := effectiveExamTotal(session, reviewData)
	return dto.ExamHistoryItemDTO{
		SessionID:        session.SessionID,
		ExamType:         session.ExamType,
		Status:           session.Status,
		Score:            services.ComputeSessionScore(session.Score, reviewData.CorrectCount, totalQ),
		CorrectCount:     reviewData.CorrectCount,
		IncorrectCount:   reviewData.IncorrectCount,
		UnansweredCount:  reviewData.UnansweredCount,
		TotalQuestions:   totalQ,
		StartedAt:        session.StartedAt,
		CompletedAt:      session.CompletedAt,
		TimeSpentSeconds: session.TimeSpentSeconds,
	}
}

func extractAuthUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized", Code: "AUTH_UNAUTHORIZED"})
		return "", false
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized", Code: "AUTH_UNAUTHORIZED"})
		return "", false
	}
	return userIDStr, true
}

// FinalizeSession godoc
// @Summary Finalize an active exam session
// @Description Marks unanswered questions as incorrect, records elapsed time, transitions the session to completed, and returns full review results. Idempotent when already completed.
// @Tags Exams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sessionId path string true "Exam Session ID"
// @Param request body dto.FinalizeExamRequest false "Optional elapsed exam time"
// @Success 200 {object} dto.FinalizeExamResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /exams/sessions/{sessionId}/finalize [post]
func (ec *ExamController) FinalizeSession(c *gin.Context) {
	userIDStr, ok := extractAuthUserID(c)
	if !ok {
		return
	}

	sessionID := c.Param("sessionId")
	var input dto.FinalizeExamRequest
	if c.Request != nil && c.Request.Body != nil {
		if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error(), Code: "AUTH_INVALID_PAYLOAD"})
			return
		}
	}

	session, err := ec.store.FindSessionByIDAndUser(sessionID, userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Exam session not found", Code: "EXAM_SESSION_NOT_FOUND"})
		return
	}

	if session.Status == "abandoned" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Exam session is abandoned", Code: "EXAM_SESSION_NOT_FOUND"})
		return
	}

	var reviewData *services.ExamReviewData
	if session.Status == "completed" {
		reviewData, err = ec.store.GetExamSessionReviewData(session)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to load completed session", Code: "INTERNAL_SERVER_ERROR"})
			return
		}
	} else {
		existingSpent := 0
		if session.TimeSpentSeconds != nil {
			existingSpent = *session.TimeSpentSeconds
		}
		finalSpent := services.ClampExamDuration(input.TimeSpentSeconds, existingSpent)
		session.TimeSpentSeconds = &finalSpent

		reviewData, err = ec.store.MarkUnansweredAndFinalize(session, userIDStr)
		if err != nil {
			if errors.Is(err, errExamSessionAbandoned) {
				c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Exam session is abandoned", Code: "EXAM_SESSION_NOT_FOUND"})
				return
			}
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to finalize session", Code: "INTERNAL_SERVER_ERROR"})
			return
		}
	}

	totalQ := effectiveExamTotal(session, reviewData)
	spentOut := 0
	if session.TimeSpentSeconds != nil {
		spentOut = *session.TimeSpentSeconds
	}

	c.JSON(http.StatusOK, dto.FinalizeExamResponse{
		SessionID:        session.SessionID,
		ExamType:         session.ExamType,
		Message:          "Exam completed successfully!",
		Completed:        true,
		Score:            services.ComputeSessionScore(session.Score, reviewData.CorrectCount, totalQ),
		CorrectCount:     reviewData.CorrectCount,
		IncorrectCount:   reviewData.IncorrectCount,
		UnansweredCount:  reviewData.UnansweredCount,
		TotalQuestions:   totalQ,
		TimeSpentSeconds: spentOut,
		DomainScores:     reviewData.DomainScores,
		Results:          reviewData.Results,
	})
}

// GetExamHistory godoc
// @Summary List completed exam sessions for the authenticated learner
// @Description Returns all completed Pre-Test and Post-Test sessions ordered most-recent-first along with latest session shortcuts.
// @Tags Exams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ExamHistoryResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /exams/history [get]
func (ec *ExamController) GetExamHistory(c *gin.Context) {
	userIDStr, ok := extractAuthUserID(c)
	if !ok {
		return
	}

	items, err := ec.store.GetCompletedExamHistory(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to load exam history", Code: "INTERNAL_SERVER_ERROR"})
		return
	}
	if items == nil {
		items = []dto.ExamHistoryItemDTO{}
	}

	var latestPre *string
	var latestPost *string
	var latestPreScore, latestPostScore *float64
	for i := range items {
		if items[i].ExamType == "pre" && latestPre == nil {
			id := items[i].SessionID
			s := items[i].Score
			latestPre = &id
			latestPreScore = &s
		}
		if items[i].ExamType == "post" && latestPost == nil {
			id := items[i].SessionID
			s := items[i].Score
			latestPost = &id
			latestPostScore = &s
		}
		if latestPre != nil && latestPost != nil {
			break
		}
	}

	var latestDelta *float64
	if latestPreScore != nil && latestPostScore != nil {
		d := services.RoundExamScore(*latestPostScore - *latestPreScore)
		latestDelta = &d
	}

	c.JSON(http.StatusOK, dto.ExamHistoryResponse{
		LatestPreSessionID:   latestPre,
		LatestPostSessionID:  latestPost,
		LatestPreToPostDelta: latestDelta,
		Sessions:             items,
	})
}

// GetExamReview godoc
// @Summary Get complete Q&A review for a completed exam session
// @Description Returns score metrics, domain breakdown, ordered 30-question Q&A details with explanations and sub-topic metadata, and Pre-to-Post learning gains on Post-Tests. Strictly rejects in-progress sessions with 403 EXAM_NOT_COMPLETED.
// @Tags Exams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sessionId path string true "Completed Exam Session ID"
// @Success 200 {object} dto.ExamReviewResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /exams/sessions/{sessionId}/review [get]
func (ec *ExamController) GetExamReview(c *gin.Context) {
	userIDStr, ok := extractAuthUserID(c)
	if !ok {
		return
	}

	sessionID := c.Param("sessionId")
	session, err := ec.store.FindSessionByIDAndUser(sessionID, userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Exam session not found", Code: "EXAM_SESSION_NOT_FOUND"})
		return
	}

	if session.Status != "completed" {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:  "Exam session is not completed yet. Complete the exam before reviewing answers.",
			Code:   "EXAM_NOT_COMPLETED",
			Params: map[string]any{"exam_type": session.ExamType},
		})
		return
	}

	reviewData, err := ec.store.GetExamSessionReviewData(session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to load exam review", Code: "INTERNAL_SERVER_ERROR"})
		return
	}

	totalQ := effectiveExamTotal(session, reviewData)
	var preComparison *dto.ExamPreComparisonDTO
	if session.ExamType == "post" {
		if preSession, preErr := ec.store.GetLatestCompletedPreTestSession(userIDStr); preErr == nil && preSession != nil {
			if preReview, revErr := ec.store.GetExamSessionReviewData(preSession); revErr == nil {
				preComparison = services.BuildPreComparison(
					preSession.SessionID,
					preSession.Score,
					effectiveExamTotal(preSession, preReview),
					preReview,
					session.Score,
					totalQ,
					reviewData,
				)
			}
		}
	}

	c.JSON(http.StatusOK, dto.ExamReviewResponse{
		Session:       buildExamHistoryItemDTO(session, reviewData),
		DomainScores:  reviewData.DomainScores,
		PreComparison: preComparison,
		Results:       reviewData.Results,
	})
}

// GetSampleQuestion godoc
// @Summary Get a sample exam question
// @Description Retrieves a random question from the database question pool for preview.
// @Tags Exams
// @Accept json
// @Produce json
// @Success 200 {object} dto.SampleQuestionDTO
// @Failure 500 {object} dto.ErrorResponse
// @Router /exams/sample [get]
func (ec *ExamController) GetSampleQuestion(c *gin.Context) {
	question, err := ec.store.GetRandomSampleQuestion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sample question", "code": "INTERNAL_SERVER_ERROR"})
		return
	}
	c.JSON(http.StatusOK, dto.SampleQuestionDTO{
		QuestionID:    question.QuestionID,
		DomainID:      question.DomainID,
		SubTopicID:    question.SubTopicID,
		Type:          question.Type,
		QuestionText:  question.QuestionText,
		Options:       question.Options,
		CorrectAnswer: question.CorrectAnswer,
		Explanation:   question.Explanation,
		Translations:  services.SanitizeQuestionTranslations(question.Translations, true),
		Difficulty:    question.Difficulty,
		MitreRef:      question.MitreRef,
	})
}

// BuildPostTestExclusionQuery returns a subquery of all question_ids already seen
// by userID in Pre-Test exam sessions OR assigned as Learn sample questions,
// enforcing Pre-Test ∪ Learn ∪ Post-Test = ∅.
func BuildPostTestExclusionQuery(db *gorm.DB, userID any) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(
		`SELECT esq.question_id
		 FROM exam_session_questions esq
		 JOIN exam_sessions es ON es.session_id = esq.session_id
		 WHERE es.user_id = ? AND es.exam_type = 'pre'
		 UNION
		 SELECT utsq.question_id
		 FROM user_theory_sample_questions utsq
		 WHERE utsq.user_id = ?`,
		userID, userID,
	)
}

