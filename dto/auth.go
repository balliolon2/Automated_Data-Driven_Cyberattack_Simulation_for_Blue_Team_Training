package dto

// RegisterRequest defines the payload for learner registration
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"analyst@soc.local"`
	Password string `json:"password" binding:"required,min=6" example:"Secret123!"`
}

// LoginRequest defines the payload for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"analyst@soc.local"`
	Password string `json:"password" binding:"required" example:"Secret123!"`
}

// AuthResponse defines the successful authentication response containing JWT
type AuthResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	Email string `json:"email" example:"analyst@soc.local"`
}
