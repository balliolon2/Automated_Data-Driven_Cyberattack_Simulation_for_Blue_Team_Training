package models

import "time"

type UserNotification struct {
	NotificationID string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"notification_id"`
	UserID         string    `gorm:"type:uuid;not null;index" json:"user_id"`
	User           *User     `gorm:"foreignKey:UserID;references:UserID;constraint:-" json:"user,omitempty"`
	Title          string    `gorm:"type:varchar(120);not null" json:"title"`
	Message        string    `gorm:"type:text;not null" json:"message"`
	LinkURL        string    `gorm:"type:varchar(255)" json:"link_url"`
	IsRead         bool      `gorm:"default:false;index" json:"is_read"`
	CreatedAt      time.Time `gorm:"default:now()" json:"created_at"`
}
