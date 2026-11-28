package models

import (
	"time"
)

type User struct {
	UserID            string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"user_id"`
	Email             string     `gorm:"type:varchar;unique;not null" json:"email"`
	Nickname          string     `gorm:"type:varchar;unique;not null" json:"nickname"`
	PasswordHash      string     `gorm:"type:varchar;not null" json:"-"`
	Role              string     `gorm:"type:role;default:'learner'" json:"role"`
	CurrentTier       int        `gorm:"default:1" json:"current_tier"`
	PreferredLanguage     string     `gorm:"type:varchar(10);not null;default:'en'" json:"preferred_language"`
	EmailVerified         bool       `gorm:"not null;default:false" json:"email_verified"`
	EmailVerifiedAt       *time.Time `json:"email_verified_at,omitempty"`
	VerificationTokenHash *string    `gorm:"type:varchar(64)" json:"-"`
	VerificationSentAt    *time.Time `json:"-"`
	VerificationExpiresAt *time.Time `json:"-"`
	CreatedAt             time.Time  `gorm:"default:now()" json:"created_at"`
	LastLogin             *time.Time `json:"last_login"`
	IsActive              bool       `gorm:"default:true" json:"is_active"`
}
