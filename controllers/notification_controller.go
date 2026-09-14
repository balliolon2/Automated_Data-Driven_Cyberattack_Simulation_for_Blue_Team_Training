package controllers

import (
	"net/http"
	"strconv"
	"time"

	"cybersim/dto"
	"cybersim/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NotificationController struct {
	DB *gorm.DB
}

func NewNotificationController(db *gorm.DB) *NotificationController {
	return &NotificationController{DB: db}
}

// ListNotifications retrieves paginated in-app alerts and unread count for current user
func (nc *NotificationController) ListNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "15"))
	if pageSize < 1 || pageSize > 50 {
		pageSize = 15
	}
	offset := (page - 1) * pageSize

	unreadOnly := c.Query("unread_only") == "true"

	query := nc.DB.Model(&models.UserNotification{}).Where("user_id = ?", userID.(string))
	if unreadOnly {
		query = query.Where("is_read = false")
	}

	var total int64
	query.Count(&total)

	var unreadCount int64
	nc.DB.Model(&models.UserNotification{}).Where("user_id = ? AND is_read = false", userID.(string)).Count(&unreadCount)

	var notifs []models.UserNotification
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&notifs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	items := make([]dto.NotificationResponse, 0, len(notifs))
	for _, n := range notifs {
		items = append(items, dto.NotificationResponse{
			NotificationID: n.NotificationID,
			Title:          n.Title,
			Message:        n.Message,
			LinkURL:        n.LinkURL,
			IsRead:         n.IsRead,
			CreatedAt:      n.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, dto.NotificationsListResponse{
		Notifications: items,
		UnreadCount:   unreadCount,
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
	})
}

// MarkAsRead marks a single notification as read
func (nc *NotificationController) MarkAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	notifID := c.Param("id")
	if err := nc.DB.Model(&models.UserNotification{}).
		Where("notification_id = ? AND user_id = ?", notifID, userID.(string)).
		Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllAsRead marks all unread notifications for current user as read
func (nc *NotificationController) MarkAllAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := nc.DB.Model(&models.UserNotification{}).
		Where("user_id = ? AND is_read = false", userID.(string)).
		Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}
