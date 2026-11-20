package controllers

import (
	"errors"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"cybersim/dto"
	"cybersim/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ExamDomainScore struct {
	DomainID     string  `json:"domain_id"`
	TotalCount   int64   `json:"total"`
	CorrectCount int64   `json:"correct"`
	Percentage   float64 `json:"percentage"`
}

type ExamDetailedResult struct {
	models.Question
	UserAnswer string `json:"user_answer"`
	IsCorrect  bool   `json:"is_correct"`
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
	FindQuestionByID(questionID string) (*models.Question, error)
	FindSessionQuestion(sessionID, questionID string) (*models.ExamSessionQuestion, error)
	SaveExamAnswer(session *models.ExamSession, esq *models.ExamSessionQuestion) (int64, error)
	FinalizeExamSession(session *models.ExamSession, userID string) (int64, []ExamDomainScore, []ExamDetailedResult, error)
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
			OrderIndex:   r.OrderIndex,
			UserAnswer:   r.UserAnswer,
		}
	}
	return resp, nil
}

func (s *gormExamStore) FindActiveSessionByIDAndUser(sessionID, userID string) (*models.ExamSession, error) {
	var session models.ExamSession
	if err := s.DB.Where("session_id = ? AND user_id = ? AND status = 'in_progress'", sessionID, userID).First(&session).Error; err != nil {
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

func (s *gormExamStore) FinalizeExamSession(session *models.ExamSession, userID string) (int64, []ExamDomainScore, []ExamDetailedResult, error) {
	var correctCount int64
	var domainScores []ExamDomainScore

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.ExamSessionQuestion{}).
			Where("session_id = ? AND is_correct = true", session.SessionID).
			Count(&correctCount).Error; err != nil {
			return err
		}

		scorePercent := math.Round((float64(correctCount)/float64(session.TotalQuestions))*100*100) / 100
		session.Score = &scorePercent
		session.Status = "completed"
		now := time.Now()
		session.CompletedAt = &now

		if err := tx.Save(session).Error; err != nil {
			return err
		}

		if session.ExamType == "pre" {
			var domainResults []struct {
				DomainID     string `gorm:"column:domain_id"`
				TotalCount   int64  `gorm:"column:total_count"`
				CorrectCount int64  `gorm:"column:correct_count"`
			}

			tx.Table("exam_session_questions").
				Select("questions.domain_id, COUNT(*) as total_count, SUM(CASE WHEN exam_session_questions.is_correct = true THEN 1 ELSE 0 END) as correct_count").
				Joins("JOIN questions ON questions.question_id = exam_session_questions.question_id").
				Where("exam_session_questions.session_id = ?", session.SessionID).
				Group("questions.domain_id").
				Scan(&domainResults)

			domainScores = make([]ExamDomainScore, 0, len(domainResults))
			for _, dr := range domainResults {
				pct := 0.0
				if dr.TotalCount > 0 {
					pct = math.Round((float64(dr.CorrectCount)/float64(dr.TotalCount))*100*100) / 100
				}

				domainScores = append(domainScores, ExamDomainScore{
					DomainID:     dr.DomainID,
					TotalCount:   dr.TotalCount,
					CorrectCount: dr.CorrectCount,
					Percentage:   pct,
				})

				var profile models.UserSkillProfile
				if err := tx.Where("user_id = ? AND domain_id = ?", userID, dr.DomainID).First(&profile).Error; err != nil {
					now := time.Now()
					profile = models.UserSkillProfile{
						UserID:           userID,
						DomainID:         dr.DomainID,
						ProficiencyScore: pct,
						LastPracticed:    &now,
					}
					tx.Create(&profile)
				} else {
					profile.ProficiencyScore = pct
					now := time.Now()
					profile.LastPracticed = &now
					tx.Save(&profile)
				}
			}
		}

		return nil
	})
	if err != nil {
		return 0, nil, nil, err
	}

	var detailedResults []ExamDetailedResult
	if err := s.DB.Table("exam_session_questions").
		Select("questions.*, exam_session_questions.user_answer, exam_session_questions.is_correct").
		Joins("join questions on questions.question_id = exam_session_questions.question_id").
		Where("exam_session_questions.session_id = ?", session.SessionID).
		Scan(&detailedResults).Error; err != nil {
		return 0, nil, nil, err
	}

	return correctCount, domainScores, detailedResults, nil
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
	threshold := 70.0
	if envThreshold := os.Getenv("PROFICIENCY_THRESHOLD"); envThreshold != "" {
		if val, err := strconv.ParseFloat(envThreshold, 64); err == nil {
			threshold = val
		}
	}
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

	isCorrect := input.UserAnswer == question.CorrectAnswer
	esq.UserAnswer = input.UserAnswer
	esq.IsCorrect = &isCorrect

	if input.TimeSpentSeconds > 0 {
		var currentSpent int
		if session.TimeSpentSeconds != nil {
			currentSpent = *session.TimeSpentSeconds
		}
		newSpent := currentSpent + input.TimeSpentSeconds
		session.TimeSpentSeconds = &newSpent
	}

	unansweredCount, err := ec.store.SaveExamAnswer(session, esq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save answer", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	var responseData gin.H

	if unansweredCount == 0 {
		correctCount, domainScores, detailedResults, err := ec.store.FinalizeExamSession(session, userIDStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize session", "code": "INTERNAL_SERVER_ERROR"})
			return
		}

		scorePercent := 0.0
		if session.Score != nil {
			scorePercent = *session.Score
		}

		if session.ExamType == "pre" {
			responseData = gin.H{
				"message":       "Exam completed successfully!",
				"completed":     true,
				"score":         scorePercent,
				"results":       detailedResults,
				"correct_count": correctCount,
				"domain_scores": domainScores,
			}
		} else {
			responseData = gin.H{
				"message":       "Exam completed successfully!",
				"completed":     true,
				"score":         scorePercent,
				"results":       detailedResults,
				"correct_count": correctCount,
			}
		}
	} else {
		responseData = gin.H{
			"completed": false,
		}
	}

	c.JSON(http.StatusOK, responseData)
}

// GetSampleQuestion godoc
// @Summary Get a sample exam question
// @Description Retrieves a random question from the database question pool for preview.
// @Tags Exams
// @Accept json
// @Produce json
// @Success 200 {object} dto.QuestionDTO
// @Failure 500 {object} dto.ErrorResponse
// @Router /exams/sample [get]
func (ec *ExamController) GetSampleQuestion(c *gin.Context) {
	question, err := ec.store.GetRandomSampleQuestion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sample question", "code": "INTERNAL_SERVER_ERROR"})
		return
	}
	c.JSON(http.StatusOK, question)
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

