package dto

// RegisterRequest defines the payload for learner registration
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"analyst@soc.local"`
	Password string `json:"password" binding:"required,min=6" example:"Secret123!"`
	Nickname string `json:"nickname" example:"cyber_defender"`
}

// LoginRequest defines the payload for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"analyst@soc.local"`
	Password string `json:"password" binding:"required" example:"Secret123!"`
}

// AuthResponse defines the successful authentication response containing JWT
type AuthResponse struct {
	Token    string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	Email    string `json:"email" example:"analyst@soc.local"`
	Nickname string `json:"nickname" example:"cyber_defender"`
	Role     string `json:"role" example:"learner"`
	UserID   string `json:"user_id" example:"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"`
}

// UserProfileResponse defines user profile details and application status
type UserProfileResponse struct {
	UserID            string                    `json:"user_id"`
	Email             string                    `json:"email"`
	Nickname          string                    `json:"nickname"`
	Role              string                    `json:"role"`
	CurrentTier       int                       `json:"current_tier"`
	LatestApplication *SpecialistApplicationDTO `json:"latest_application,omitempty"`
}

// SpecialistApplicationDTO represents the applicant review record
type SpecialistApplicationDTO struct {
	ApplicationID   string  `json:"application_id"`
	UserID          string  `json:"user_id"`
	Nickname        string  `json:"nickname,omitempty"`
	Email           string  `json:"email,omitempty"`
	Status          string  `json:"status"`
	Bio             string  `json:"bio"`
	ResumePath      string  `json:"resume_path"`
	CertificatePath string  `json:"certificate_path,omitempty"`
	LinkedInURL     string  `json:"linkedin_url,omitempty"`
	PortfolioURL    string  `json:"portfolio_url,omitempty"`
	RejectionReason string  `json:"rejection_reason,omitempty"`
	ReviewedBy      *string `json:"reviewed_by,omitempty"`
	ReviewedAt      *string `json:"reviewed_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
}

// UpdateNicknameRequest defines the payload for changing user nickname
type UpdateNicknameRequest struct {
	Nickname string `json:"nickname" binding:"required,min=2,max=30" example:"soc_rookie"`
}

// AdminRejectRequest defines the payload for rejecting a specialist application
type AdminRejectRequest struct {
	Reason string `json:"reason" binding:"required,min=3" example:"Requires at least 2 years SOC experience or relevant Security+ cert"`
}
