package models

import "time"

// RagDocument maps to the rag_documents table in init.sql and stores
// domain theory curriculum documents alongside RAG reference content.
type RagDocument struct {
	DocID     string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"doc_id"`
	Title     string    `gorm:"type:varchar;not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	DomainID  string    `gorm:"type:varchar" json:"domain_id"`
	SourceURL string    `gorm:"type:varchar" json:"source_url"`
	MitreRef  *string   `gorm:"type:varchar" json:"mitre_ref"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`
}

func (RagDocument) TableName() string {
	return "rag_documents"
}

// UserTheoryProgress records each domain Theory Module a learner has completed.
type UserTheoryProgress struct {
	UserID      string    `gorm:"type:uuid;primaryKey" json:"user_id"`
	DomainID    string    `gorm:"type:varchar;primaryKey" json:"domain_id"`
	CompletedAt time.Time `gorm:"default:now()" json:"completed_at"`
}

func (UserTheoryProgress) TableName() string {
	return "user_theory_progress"
}

// TheorySubtopic stores one CompTIA Security+ (SY0-701) objective sub-topic
// seeded from theory/domain{1..5}/*.md into PostgreSQL.
type TheorySubtopic struct {
	SubTopicID string    `gorm:"type:varchar(16);primaryKey" json:"sub_topic_id"`
	DomainID   string    `gorm:"type:varchar;not null" json:"domain_id"`
	OrderIndex int       `gorm:"not null" json:"order_index"`
	Title      string    `gorm:"type:varchar(255);not null" json:"title"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	SourceURL  string    `gorm:"type:text" json:"source_url"`
	UpdatedAt  time.Time `gorm:"default:now()" json:"updated_at"`
}

func (TheorySubtopic) TableName() string {
	return "theory_subtopics"
}

// UserTheorySubtopicProgress records each SY0-701 sub-topic completed by a learner.
type UserTheorySubtopicProgress struct {
	UserID      string    `gorm:"type:uuid;primaryKey" json:"user_id"`
	SubTopicID  string    `gorm:"type:varchar(16);primaryKey" json:"sub_topic_id"`
	CompletedAt time.Time `gorm:"default:now()" json:"completed_at"`
}

func (UserTheorySubtopicProgress) TableName() string {
	return "user_theory_subtopic_progress"
}

// UserTheorySampleQuestion locks the 2 formative sample questions assigned to a
// learner per Sub-topic and stores their first submitted attempt.
type UserTheorySampleQuestion struct {
	UserID          string     `gorm:"type:uuid;primaryKey" json:"user_id"`
	SubTopicID      string     `gorm:"type:varchar(16);primaryKey" json:"sub_topic_id"`
	QuestionID      string     `gorm:"type:uuid;primaryKey" json:"question_id"`
	OrderIndex      int        `gorm:"not null" json:"order_index"`
	IsFallback      bool       `gorm:"not null;default:false" json:"is_fallback"`
	SubmittedAnswer *string    `gorm:"type:varchar(10)" json:"submitted_answer,omitempty"`
	IsCorrect       *bool      `json:"is_correct,omitempty"`
	AssignedAt      time.Time  `gorm:"default:now()" json:"assigned_at"`
	AnsweredAt      *time.Time `json:"answered_at,omitempty"`
}

func (UserTheorySampleQuestion) TableName() string {
	return "user_theory_sample_questions"
}

// UserTheorySampleQuestionDetail pairs an assigned sample question row with its question body.
type UserTheorySampleQuestionDetail struct {
	Assignment UserTheorySampleQuestion
	Question   Question
}
