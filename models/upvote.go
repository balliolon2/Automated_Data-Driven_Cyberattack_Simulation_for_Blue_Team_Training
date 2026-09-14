package models

import "time"

type ThreadUpvote struct {
	ThreadID  string          `gorm:"type:uuid;primaryKey" json:"thread_id"`
	Thread    *AnalysisThread `gorm:"foreignKey:ThreadID" json:"thread,omitempty"`
	UserID    string          `gorm:"type:uuid;primaryKey" json:"user_id"`
	User      *User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt time.Time       `gorm:"default:now()" json:"created_at"`
}

type CommentUpvote struct {
	CommentID string         `gorm:"type:uuid;primaryKey" json:"comment_id"`
	Comment   *ThreadComment `gorm:"foreignKey:CommentID" json:"comment,omitempty"`
	UserID    string         `gorm:"type:uuid;primaryKey" json:"user_id"`
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt time.Time      `gorm:"default:now()" json:"created_at"`
}
