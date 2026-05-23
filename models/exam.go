package models

import (
	"time"
)

type Question struct {
	QuestionID    string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"question_id"`
	DomainID      string    `gorm:"type:varchar;not null" json:"domain_id"`
	Type          string    `gorm:"type:question_type;not null" json:"type"`
	QuestionText  string    `gorm:"type:text;not null" json:"question_text"`
	Options       any       `gorm:"type:jsonb;serializer:json" json:"options"`
	CorrectAnswer string    `gorm:"type:varchar;not null" json:"correct_answer"`
	Explanation   string    `gorm:"type:text" json:"explanation"`
	Difficulty    int       `gorm:"default:3" json:"difficulty"`
	MitreRef      *string   `gorm:"type:varchar" json:"mitre_ref"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `gorm:"default:now()" json:"created_at"`
}

type ExamSession struct {
	SessionID        string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"session_id"`
	UserID           string     `gorm:"type:uuid;not null" json:"user_id"`
	ExamType         string     `gorm:"type:test_type;not null" json:"exam_type"`
	Status           string     `gorm:"type:session_status;default:'in_progress'" json:"status"`
	Score            *float64   `json:"score"`
	TotalQuestions   int        `gorm:"default:100" json:"total_questions"`
	StartedAt        time.Time  `gorm:"default:now()" json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	TimeSpentSeconds *int       `json:"time_spent_seconds"`
}

type ExamSessionQuestion struct {
	SessionID   string `gorm:"type:uuid;primaryKey" json:"session_id"`
	QuestionID  string `gorm:"type:uuid;primaryKey" json:"question_id"`
	OrderIndex  int    `gorm:"not null" json:"order_index"`
	UserAnswer  string `gorm:"type:varchar" json:"user_answer"`
	IsCorrect   *bool  `json:"is_correct"`
}
