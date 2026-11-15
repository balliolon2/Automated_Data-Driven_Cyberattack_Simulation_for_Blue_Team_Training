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
