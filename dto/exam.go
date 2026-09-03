package dto

import "time"

// QuestionDTO represents a single question delivered to the learner
type QuestionDTO struct {
	QuestionID   string `json:"question_id" example:"d1-q01"`
	DomainID     string `json:"domain_id" example:"domain1"`
	Type         string `json:"type" example:"multiple_choice"`
	QuestionText string `json:"question_text" example:"Which security control provides non-repudiation?"`
	Options      any    `json:"options"`
	OrderIndex   int    `json:"order_index,omitempty" example:"1"`
	UserAnswer   string `json:"user_answer,omitempty" example:"A"`
}

// ExamSessionDTO represents the state and metadata of an exam session
type ExamSessionDTO struct {
	SessionID      string     `json:"session_id" example:"987e6543-e21b-12d3-a456-426614174000"`
	UserID         string     `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ExamType       string     `json:"exam_type" example:"pre"`
	Status         string     `json:"status" example:"in_progress"`
	Score          float64    `json:"score" example:"80.0"`
	TotalQuestions int        `json:"total_questions" example:"30"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

// ExamSessionResponse represents the response when starting or resuming an exam
type ExamSessionResponse struct {
	Session   any           `json:"session"`
	Questions []QuestionDTO `json:"questions"`
}

// SubmitAnswerRequest defines the payload for submitting an answer to an exam question
type SubmitAnswerRequest struct {
	SessionID        string `json:"session_id" binding:"required" example:"987e6543-e21b-12d3-a456-426614174000"`
	QuestionID       string `json:"question_id" binding:"required" example:"d1-q01"`
	UserAnswer       string `json:"user_answer" binding:"required" example:"A"`
	TimeSpentSeconds int    `json:"time_spent_seconds" example:"45"`
}

// DomainScoreDTO represents performance breakdown per CompTIA domain
type DomainScoreDTO struct {
	DomainID     string  `json:"domain_id" example:"domain1"`
	TotalCount   int     `json:"total_count" example:"4"`
	CorrectCount int     `json:"correct_count" example:"3"`
	Percentage   float64 `json:"percentage" example:"75.0"`
}

// DetailedExamResultDTO provides review data for completed exam questions
type DetailedExamResultDTO struct {
	QuestionID    string `json:"question_id" example:"d1-q01"`
	DomainID      string `json:"domain_id" example:"domain1"`
	Type          string `json:"type" example:"multiple_choice"`
	QuestionText  string `json:"question_text" example:"Which security control provides non-repudiation?"`
	Options       any    `json:"options"`
	CorrectAnswer string `json:"correct_answer" example:"A"`
	Explanation   string `json:"explanation" example:"Digital signatures ensure non-repudiation."`
	UserAnswer    string `json:"user_answer" example:"A"`
	IsCorrect     bool   `json:"is_correct" example:"true"`
}

// SubmitAnswerResponse represents the progress or final score outcome
type SubmitAnswerResponse struct {
	Completed    bool                    `json:"completed" example:"true"`
	Message      string                  `json:"message,omitempty" example:"Exam completed successfully!"`
	Score        float64                 `json:"score,omitempty" example:"83.33"`
	CorrectCount int                     `json:"correct_count,omitempty" example:"25"`
	Results      []DetailedExamResultDTO `json:"results,omitempty"`
	DomainScores []DomainScoreDTO        `json:"domain_scores,omitempty"`
}
