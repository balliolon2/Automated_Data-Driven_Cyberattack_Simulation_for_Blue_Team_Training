package dto

type CreateCommentRequest struct {
	Content         string  `json:"content" binding:"required,min=1,max=2000"`
	ParentCommentID *string `json:"parent_comment_id,omitempty"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

type CommentAuthorDTO struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

type CommentResponse struct {
	CommentID       string            `json:"comment_id"`
	ThreadID        string            `json:"thread_id"`
	Author          CommentAuthorDTO  `json:"author"`
	ParentCommentID *string           `json:"parent_comment_id,omitempty"`
	Content         string            `json:"content"`
	UpvoteCount     int               `json:"upvote_count"`
	IsPinned        bool              `json:"is_pinned"`
	IsHidden        bool              `json:"is_hidden"`
	HasUpvoted      bool              `json:"has_upvoted"`
	Replies         []CommentResponse `json:"replies"`
	CreatedAt       string            `json:"created_at"`
	UpdatedAt       string            `json:"updated_at"`
}
