package models

import (
	"time"
)

type User struct {
	UserID       string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"user_id"`
	Email        string     `gorm:"type:varchar;unique;not null" json:"email"`
	PasswordHash string     `gorm:"type:varchar;not null" json:"-"`
	Role         string     `gorm:"type:role;default:'learner'" json:"role"`
	CurrentTier  int        `gorm:"default:1" json:"current_tier"`
	CreatedAt    time.Time  `gorm:"default:now()" json:"created_at"`
	LastLogin    *time.Time `json:"last_login"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
}
