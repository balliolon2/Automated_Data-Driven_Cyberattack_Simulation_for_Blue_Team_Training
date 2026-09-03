package dto

// ErrorResponse represents a standardized error response body
type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid request payload"`
	Message string `json:"message,omitempty" example:"Detailed error description"`
	Code    string `json:"code,omitempty" example:"BAD_REQUEST"`
}

// MessageResponse represents a generic success or informational response
type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}
