package dto

type UpvoteResponse struct {
	UpvoteCount int  `json:"upvote_count"`
	HasUpvoted  bool `json:"has_upvoted"`
}

type NotificationResponse struct {
	NotificationID string `json:"notification_id"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	LinkURL        string `json:"link_url"`
	IsRead         bool   `json:"is_read"`
	CreatedAt      string `json:"created_at"`
}

type NotificationsListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	UnreadCount   int64                  `json:"unread_count"`
	Total         int64                  `json:"total"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"page_size"`
}
