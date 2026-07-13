package controllers

import (
	"math"
	"net/http"
	"time"

	"cybersim/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ExamController struct {
	DB *gorm.DB
}

func NewExamController(db *gorm.DB) *ExamController {
	return &ExamController{DB: db}
}

type QuestionResponse struct {
	QuestionID   string `json:"question_id"`
	DomainID     string `json:"domain_id"`
	Type         string `json:"type"`
	QuestionText string `json:"question_text"`
	Options      any    `json:"options"`
	OrderIndex   int    `json:"order_index"`
	UserAnswer   string `json:"user_answer,omitempty"`
}

// Helper to load questions for a session
func (ec *ExamController) getSessionQuestions(sessionID string) ([]QuestionResponse, error) {
	var results []struct {
		models.Question
		OrderIndex int    `gorm:"column:order_index"`
		UserAnswer string `gorm:"column:user_answer"`
	}

	err := ec.DB.Table("exam_session_questions").
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

// POST /api/exams/pre-test
func (ec *ExamController) StartPreTest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Check if there is already an active session
	var activeSession models.ExamSession
	err := ec.DB.Where("user_id = ? AND status = 'in_progress'", userID).First(&activeSession).Error
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You already have an active exam session", "session_id": activeSession.SessionID})
		return
	}

	tx := ec.DB.Begin()

	session := models.ExamSession{
		UserID:         userID.(string),
		ExamType:       "pre",
		Status:         "in_progress",
		TotalQuestions: 30,
		StartedAt:      time.Now(),
	}

	if err := tx.Create(&session).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create exam session"})
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

	var allSelectedQuestions []models.Question
	for domain, count := range domains {
		var qList []models.Question
		if err := tx.Where("domain_id = ? AND is_active = true", domain).
			Order("RANDOM()").
			Limit(count).
			Find(&qList).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve questions"})
			return
		}
		allSelectedQuestions = append(allSelectedQuestions, qList...)
	}

	if len(allSelectedQuestions) < 30 {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Insufficient questions in database pool to construct 30 questions exam"})
		return
	}

	// Bulk insert junction records
	orderIndex := 0
	for _, q := range allSelectedQuestions {
		esq := models.ExamSessionQuestion{
			SessionID:  session.SessionID,
			QuestionID: q.QuestionID,
			OrderIndex: orderIndex,
		}
		if err := tx.Create(&esq).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to map exam questions"})
			return
		}
		orderIndex++
	}

	tx.Commit()

	questions, err := ec.getSessionQuestions(session.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load session questions"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session":   session,
		"questions": questions,
	})
}

// POST /api/exams/post-test
func (ec *ExamController) StartPostTest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var activeSession models.ExamSession
	err := ec.DB.Where("user_id = ? AND status = 'in_progress'", userID).First(&activeSession).Error
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You already have an active exam session", "session_id": activeSession.SessionID})
		return
	}

	tx := ec.DB.Begin()

	session := models.ExamSession{
		UserID:         userID.(string),
		ExamType:       "post",
		Status:         "in_progress",
		TotalQuestions: 30,
		StartedAt:      time.Now(),
	}

	if err := tx.Create(&session).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create exam session"})
		return
	}

	// Exclude all questions from user's PRE test
	excludeSubquery := tx.Table("exam_session_questions").
		Select("exam_session_questions.question_id").
		Joins("join exam_sessions on exam_sessions.session_id = exam_session_questions.session_id").
		Where("exam_sessions.user_id = ? AND exam_sessions.exam_type = 'pre'", userID)

	domains := map[string]int{
		"domain1": 4,
		"domain2": 7,
		"domain3": 5,
		"domain4": 8,
		"domain5": 6,
	}

	var allSelectedQuestions []models.Question
	for domain, count := range domains {
		var qList []models.Question
		if err := tx.Where("domain_id = ? AND is_active = true AND question_id NOT IN (?)", domain, excludeSubquery).
			Order("RANDOM()").
			Limit(count).
			Find(&qList).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve questions"})
			return
		}
		allSelectedQuestions = append(allSelectedQuestions, qList...)
	}

	if len(allSelectedQuestions) < 30 {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Insufficient unused questions in database pool to construct 30 questions post-test"})
		return
	}

	orderIndex := 0
	for _, q := range allSelectedQuestions {
		esq := models.ExamSessionQuestion{
			SessionID:  session.SessionID,
			QuestionID: q.QuestionID,
			OrderIndex: orderIndex,
		}
		if err := tx.Create(&esq).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to map exam questions"})
			return
		}
		orderIndex++
	}

	tx.Commit()

	questions, err := ec.getSessionQuestions(session.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load session questions"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session":   session,
		"questions": questions,
	})
}

// GET /api/exams/session
func (ec *ExamController) GetActiveSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var session models.ExamSession
	err := ec.DB.Where("user_id = ? AND status = 'in_progress'", userID).First(&session).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "No active session found"})
		return
	}

	questions, err := ec.getSessionQuestions(session.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load session questions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session":   session,
		"questions": questions,
	})
}

type SubmitAnswerInput struct {
	SessionID        string `json:"session_id" binding:"required"`
	QuestionID       string `json:"question_id" binding:"required"`
	UserAnswer       string `json:"user_answer" binding:"required"`
	TimeSpentSeconds int    `json:"time_spent_seconds"`
}

// POST /api/exams/submit-answer
func (ec *ExamController) SubmitAnswer(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input SubmitAnswerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify session ownership and status
	var session models.ExamSession
	if err := ec.DB.Where("session_id = ? AND user_id = ? AND status = 'in_progress'", input.SessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid or inactive session"})
		return
	}

	// Get the question to check correctness
	var question models.Question
	if err := ec.DB.Where("question_id = ?", input.QuestionID).First(&question).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	isCorrect := input.UserAnswer == question.CorrectAnswer

	tx := ec.DB.Begin()

	// Update the junction table
	var esq models.ExamSessionQuestion
	if err := tx.Where("session_id = ? AND question_id = ?", input.SessionID, input.QuestionID).First(&esq).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Question not part of this exam session"})
		return
	}

	esq.UserAnswer = input.UserAnswer
	esq.IsCorrect = &isCorrect

	if err := tx.Save(&esq).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save answer"})
		return
	}

	// Update time spent
	if input.TimeSpentSeconds > 0 {
		var currentSpent int
		if session.TimeSpentSeconds != nil {
			currentSpent = *session.TimeSpentSeconds
		}
		newSpent := currentSpent + input.TimeSpentSeconds
		session.TimeSpentSeconds = &newSpent
		if err := tx.Save(&session).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update time spent"})
			return
		}
	}

	// Check if all questions are answered
	var unansweredCount int64
	if err := tx.Model(&models.ExamSessionQuestion{}).
		Where("session_id = ? AND (user_answer IS NULL OR user_answer = '')", input.SessionID).
		Count(&unansweredCount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check progress"})
		return
	}

	var responseData gin.H

	if unansweredCount == 0 {
		// Calculate final score
		var correctCount int64
		if err := tx.Model(&models.ExamSessionQuestion{}).
			Where("session_id = ? AND is_correct = true", input.SessionID).
			Count(&correctCount).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate score"})
			return
		}

		scorePercent := math.Round((float64(correctCount)/float64(session.TotalQuestions))*100*100) / 100
		session.Score = &scorePercent
		session.Status = "completed"
		now := time.Now()
		session.CompletedAt = &now

		if err := tx.Save(&session).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize session"})
			return
		}

		// Also populate evaluation_metrics if needed, but we keep it simple here

		// Calculate per-domain scores and seed user_skill_profiles (for pre-test)
		if session.ExamType == "pre" {
			type DomainScore struct {
				DomainID     string `json:"domain_id"`
				TotalCount   int64  `json:"total"`
				CorrectCount int64  `json:"correct"`
				Percentage   float64 `json:"percentage"`
			}

			var domainResults []struct {
				DomainID     string
				TotalCount   int64
				CorrectCount int64
			}

			tx.Table("exam_session_questions").
				Select("questions.domain_id, COUNT(*) as total_count, SUM(CASE WHEN exam_session_questions.is_correct = true THEN 1 ELSE 0 END) as correct_count").
				Joins("JOIN questions ON questions.question_id = exam_session_questions.question_id").
				Where("exam_session_questions.session_id = ?", session.SessionID).
				Group("questions.domain_id").
				Scan(&domainResults)

			domainScores := make([]DomainScore, 0, len(domainResults))
			for _, dr := range domainResults {
				pct := 0.0
				if dr.TotalCount > 0 {
					pct = math.Round((float64(dr.CorrectCount)/float64(dr.TotalCount))*100*100) / 100
				}

				domainScores = append(domainScores, DomainScore{
					DomainID:     dr.DomainID,
					TotalCount:   dr.TotalCount,
					CorrectCount: dr.CorrectCount,
					Percentage:   pct,
				})

				// Upsert user_skill_profiles
				var profile models.UserSkillProfile
				err := tx.Where("user_id = ? AND domain_id = ?", userID, dr.DomainID).First(&profile).Error
				if err != nil {
					// Create new profile
					now := time.Now()
					profile = models.UserSkillProfile{
						UserID:           userID.(string),
						DomainID:         dr.DomainID,
						ProficiencyScore: pct,
						LastPracticed:    &now,
					}
					tx.Create(&profile)
				} else {
					// Update existing profile
					profile.ProficiencyScore = pct
					now := time.Now()
					profile.LastPracticed = &now
					tx.Save(&profile)
				}
			}

			tx.Commit()

			// For completed exams, we can return the correct answer and explanations
			var detailedResults []struct {
				models.Question
				UserAnswer string `json:"user_answer"`
				IsCorrect  bool   `json:"is_correct"`
			}
			ec.DB.Table("exam_session_questions").
				Select("questions.*, exam_session_questions.user_answer, exam_session_questions.is_correct").
				Joins("join questions on questions.question_id = exam_session_questions.question_id").
				Where("exam_session_questions.session_id = ?", session.SessionID).
				Scan(&detailedResults)

			responseData = gin.H{
				"message":        "Exam completed successfully!",
				"completed":      true,
				"score":          scorePercent,
				"results":        detailedResults,
				"correct_count":  correctCount,
				"domain_scores":  domainScores,
			}
		} else {
			tx.Commit()

			// For completed exams, we can return the correct answer and explanations
			var detailedResults []struct {
				models.Question
				UserAnswer string `json:"user_answer"`
				IsCorrect  bool   `json:"is_correct"`
			}
			ec.DB.Table("exam_session_questions").
				Select("questions.*, exam_session_questions.user_answer, exam_session_questions.is_correct").
				Joins("join questions on questions.question_id = exam_session_questions.question_id").
				Where("exam_session_questions.session_id = ?", session.SessionID).
				Scan(&detailedResults)

			responseData = gin.H{
				"message":       "Exam completed successfully!",
				"completed":     true,
				"score":         scorePercent,
				"results":       detailedResults,
				"correct_count": correctCount,
			}
		}
	} else {
		tx.Commit()
		responseData = gin.H{
			"completed": false,
		}
	}

	c.JSON(http.StatusOK, responseData)
}

// GET /api/exams/sample
func (ec *ExamController) GetSampleQuestion(c *gin.Context) {
	var question models.Question
	err := ec.DB.Order("RANDOM()").First(&question).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sample question"})
		return
	}
	c.JSON(http.StatusOK, question)
}

