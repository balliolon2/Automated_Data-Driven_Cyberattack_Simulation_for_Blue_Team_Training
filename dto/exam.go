package dto

import "time"

// QuestionDTO represents a single question delivered to the learner
type QuestionDTO struct {
	QuestionID   string `json:"question_id" example:"d1-q01"`
	DomainID     string `json:"domain_id" example:"domain1"`
	Type         string `json:"type" example:"multiple_choice"`
	QuestionText string `json:"question_text" example:"Which security control provides non-repudiation?"`
	Options      any    `json:"options"`
	Translations any    `json:"translations,omitempty"`
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
	DomainName   string  `json:"domain_name" example:"General Security Concepts"`
	TotalCount   int64   `json:"total" example:"4"`
	CorrectCount int64   `json:"correct" example:"3"`
	Incorrect    int64   `json:"incorrect" example:"1"`
	Unanswered   int64   `json:"unanswered" example:"0"`
	Percentage   float64 `json:"percentage" example:"75.0"`
	Threshold    float64 `json:"threshold" example:"70.0"`
	Passed       bool    `json:"passed" example:"true"`
}

// DetailedExamResultDTO provides review data for completed exam questions
type DetailedExamResultDTO struct {
	QuestionID    string  `json:"question_id" example:"d1-q01"`
	DomainID      string  `json:"domain_id" example:"domain1"`
	DomainName    string  `json:"domain_name" example:"General Security Concepts"`
	SubTopicID    *string `json:"sub_topic_id,omitempty" example:"1.1"`
	SubTopicTitle string  `json:"sub_topic_title" example:"Compare and contrast various types of security controls"`
	Type          string  `json:"type" example:"multiple_choice"`
	QuestionText  string  `json:"question_text" example:"Which security control provides non-repudiation?"`
	Options       any     `json:"options"`
	CorrectAnswer string  `json:"correct_answer" example:"A"`
	Explanation   string  `json:"explanation" example:"Correct Answer: A. Digital signatures ensure non-repudiation."`
	Translations  any     `json:"translations,omitempty"`
	MitreRef      *string `json:"mitre_ref,omitempty" example:"T1078"`
	OrderIndex    int     `json:"order_index" example:"0"`
	UserAnswer    string  `json:"user_answer" example:"A"`
	IsCorrect     bool    `json:"is_correct" example:"true"`
	IsUnanswered  bool    `json:"is_unanswered" example:"false"`
}

// SubmitAnswerResponse represents the progress or final score outcome
type SubmitAnswerResponse struct {
	SessionID        string                  `json:"session_id,omitempty" example:"987e6543-e21b-12d3-a456-426614174000"`
	ExamType         string                  `json:"exam_type,omitempty" example:"pre"`
	Completed        bool                    `json:"completed" example:"true"`
	Message          string                  `json:"message,omitempty" example:"Exam completed successfully!"`
	Score            float64                 `json:"score,omitempty" example:"83.33"`
	CorrectCount     int64                   `json:"correct_count,omitempty" example:"25"`
	TotalQuestions   int                     `json:"total_questions,omitempty" example:"30"`
	TimeSpentSeconds *int                    `json:"time_spent_seconds,omitempty" example:"842"`
	Results          []DetailedExamResultDTO `json:"results,omitempty"`
	DomainScores     []DomainScoreDTO        `json:"domain_scores,omitempty"`
}

// FinalizeExamRequest defines optional payload when explicitly finalizing an exam session
type FinalizeExamRequest struct {
	TimeSpentSeconds int `json:"time_spent_seconds" example:"842"`
}

// FinalizeExamResponse represents the finalized outcome returned by POST /exams/sessions/:sessionId/finalize
type FinalizeExamResponse struct {
	SessionID        string                  `json:"session_id" example:"987e6543-e21b-12d3-a456-426614174000"`
	ExamType         string                  `json:"exam_type" example:"pre"`
	Message          string                  `json:"message" example:"Exam completed successfully!"`
	Completed        bool                    `json:"completed" example:"true"`
	Score            float64                 `json:"score" example:"73.33"`
	CorrectCount     int64                   `json:"correct_count" example:"22"`
	IncorrectCount   int64                   `json:"incorrect_count" example:"6"`
	UnansweredCount  int64                   `json:"unanswered_count" example:"2"`
	TotalQuestions   int                     `json:"total_questions" example:"30"`
	TimeSpentSeconds int                     `json:"time_spent_seconds" example:"842"`
	DomainScores     []DomainScoreDTO        `json:"domain_scores"`
	Results          []DetailedExamResultDTO `json:"results"`
}

// ExamHistoryItemDTO summarizes one completed Pre-Test or Post-Test exam attempt
type ExamHistoryItemDTO struct {
	SessionID        string     `json:"session_id" example:"987e6543-e21b-12d3-a456-426614174000"`
	ExamType         string     `json:"exam_type" example:"pre"`
	Status           string     `json:"status" example:"completed"`
	Score            float64    `json:"score" example:"73.33"`
	CorrectCount     int64      `json:"correct_count" example:"22"`
	IncorrectCount   int64      `json:"incorrect_count" example:"6"`
	UnansweredCount  int64      `json:"unanswered_count" example:"2"`
	TotalQuestions   int        `json:"total_questions" example:"30"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	TimeSpentSeconds *int       `json:"time_spent_seconds,omitempty" example:"842"`
}

// ExamHistoryResponse lists completed exam attempts and latest Pre/Post session shortcuts
type ExamHistoryResponse struct {
	LatestPreSessionID   *string              `json:"latest_pre_session_id" example:"987e6543-e21b-12d3-a456-426614174000"`
	LatestPostSessionID  *string              `json:"latest_post_session_id,omitempty"`
	LatestPreToPostDelta *float64             `json:"latest_pre_to_post_delta,omitempty" example:"26.67"`
	Sessions             []ExamHistoryItemDTO `json:"sessions"`
}

// ExamDomainDeltaDTO compares domain performance between Pre-Test and Post-Test
type ExamDomainDeltaDTO struct {
	DomainID       string  `json:"domain_id" example:"domain1"`
	DomainName     string  `json:"domain_name" example:"General Security Concepts"`
	PrePercentage  float64 `json:"pre_percentage" example:"50.0"`
	PostPercentage float64 `json:"post_percentage" example:"100.0"`
	Delta          float64 `json:"delta" example:"50.0"`
}

// ExamPreComparisonDTO summarizes overall and per-domain learning gains on a completed Post-Test
type ExamPreComparisonDTO struct {
	PreSessionID     string               `json:"pre_session_id" example:"987e6543-e21b-12d3-a456-426614174000"`
	PreScore         float64              `json:"pre_score" example:"60.0"`
	PostScore        float64              `json:"post_score" example:"86.67"`
	ScoreDelta       float64              `json:"score_delta" example:"26.67"`
	PreCorrectCount  int64                `json:"pre_correct_count" example:"18"`
	PostCorrectCount int64                `json:"post_correct_count" example:"26"`
	DomainDeltas     []ExamDomainDeltaDTO `json:"domain_deltas"`
}

// ExamReviewResponse provides the complete learning-aligned review payload for a completed exam session
type ExamReviewResponse struct {
	Session       ExamHistoryItemDTO      `json:"session"`
	DomainScores  []DomainScoreDTO        `json:"domain_scores"`
	PreComparison *ExamPreComparisonDTO   `json:"pre_comparison"`
	Results       []DetailedExamResultDTO `json:"results"`
}

// SampleQuestionDTO represents a single public sample exam question previewed on the landing workbench.
type SampleQuestionDTO struct {
	QuestionID    string  `json:"question_id" example:"d1-q01"`
	DomainID      string  `json:"domain_id" example:"domain1"`
	SubTopicID    *string `json:"sub_topic_id,omitempty" example:"1.1"`
	Type          string  `json:"type" example:"multiple_choice"`
	QuestionText  string  `json:"question_text" example:"Which security control provides non-repudiation?"`
	Options       any     `json:"options"`
	CorrectAnswer string  `json:"correct_answer" example:"B"`
	Explanation   string  `json:"explanation" example:"Correct Answer: B..."`
	Translations  any     `json:"translations,omitempty"`
	Difficulty    int     `json:"difficulty" example:"3"`
	MitreRef      *string `json:"mitre_ref,omitempty" example:"T1566"`
}
