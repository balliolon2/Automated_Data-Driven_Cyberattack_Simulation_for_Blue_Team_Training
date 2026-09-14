package models

import (
	"time"
)

type ThreadComment struct {
	CommentID       string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"comment_id"`
	ThreadID        string          `gorm:"type:uuid;not null;index" json:"thread_id"`
	Thread          *AnalysisThread `gorm:"foreignKey:ThreadID;references:ThreadID;constraint:-" json:"thread,omitempty"`
	UserID          string          `gorm:"type:uuid;not null;index" json:"user_id"`
	User            *User           `gorm:"foreignKey:UserID;references:UserID;constraint:-" json:"user,omitempty"`
	ParentCommentID *string         `gorm:"type:uuid;index" json:"parent_comment_id,omitempty"`
	ParentComment   *ThreadComment  `gorm:"foreignKey:ParentCommentID;references:CommentID;constraint:-" json:"parent_comment,omitempty"`
	Content         string          `gorm:"type:text;not null" json:"content"`
	UpvoteCount     int             `gorm:"default:0" json:"upvote_count"`
	IsPinned        bool            `gorm:"default:false" json:"is_pinned"`
	IsHidden        bool            `gorm:"default:false" json:"is_hidden"`
	CreatedAt       time.Time       `gorm:"default:now()" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"default:now()" json:"updated_at"`
}
