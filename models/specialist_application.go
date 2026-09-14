package models

import (
	"cybersim/dto"
	"time"
)

const (
	ApplicationStatusPending  = "pending"
	ApplicationStatusApproved = "approved"
	ApplicationStatusRejected = "rejected"
)

type SpecialistApplication struct {
	ApplicationID   string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"application_id"`
	UserID          string     `gorm:"type:uuid;not null;index" json:"user_id"`
	User            *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Status          string     `gorm:"type:varchar;not null;default:'pending'" json:"status"` // pending, approved, rejected
	Bio             string     `gorm:"type:text;not null" json:"bio"`
	ResumePath      string     `gorm:"type:varchar;not null" json:"resume_path"`
	CertificatePath string     `gorm:"type:varchar" json:"certificate_path,omitempty"`
	LinkedInURL     string     `gorm:"type:varchar" json:"linkedin_url,omitempty"`
	PortfolioURL    string     `gorm:"type:varchar" json:"portfolio_url,omitempty"`
	RejectionReason string     `gorm:"type:text" json:"rejection_reason,omitempty"`
	ReviewedBy      *string    `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	Reviewer        *User      `gorm:"foreignKey:ReviewedBy" json:"reviewer,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt       time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"default:now()" json:"updated_at"`
}

// ToDTO converts a SpecialistApplication database model into a public DTO
func (app *SpecialistApplication) ToDTO(nickname, email string) dto.SpecialistApplicationDTO {
	var reviewedAtStr *string
	if app.ReviewedAt != nil {
		s := app.ReviewedAt.Format(time.RFC3339)
		reviewedAtStr = &s
	}
	return dto.SpecialistApplicationDTO{
		ApplicationID:   app.ApplicationID,
		UserID:          app.UserID,
		Nickname:        nickname,
		Email:           email,
		Status:          app.Status,
		Bio:             app.Bio,
		ResumePath:      app.ResumePath,
		CertificatePath: app.CertificatePath,
		LinkedInURL:     app.LinkedInURL,
		PortfolioURL:    app.PortfolioURL,
		RejectionReason: app.RejectionReason,
		ReviewedBy:      app.ReviewedBy,
		ReviewedAt:      reviewedAtStr,
		CreatedAt:       app.CreatedAt.Format(time.RFC3339),
	}
}
