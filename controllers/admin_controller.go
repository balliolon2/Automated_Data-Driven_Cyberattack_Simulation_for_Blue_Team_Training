package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cybersim/dto"
	"cybersim/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminController struct {
	DB *gorm.DB
}

func NewAdminController(db *gorm.DB) *AdminController {
	return &AdminController{DB: db}
}

// ListApplications retrieves specialist applications filtered by status
func (ac *AdminController) ListApplications(c *gin.Context) {
	status := strings.TrimSpace(c.Query("status"))

	query := ac.DB.Model(&models.SpecialistApplication{}).Preload("User")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var applications []models.SpecialistApplication
	if err := query.Order("created_at desc").Find(&applications).Error; err != nil {
		respondAPIError(c, http.StatusInternalServerError, "ADMIN_APPLICATIONS_FETCH_FAILED", "Failed to query applications")
		return
	}

	result := make([]dto.SpecialistApplicationDTO, 0, len(applications))
	for _, app := range applications {
		nickname := ""
		email := ""
		if app.User != nil {
			nickname = app.User.Nickname
			email = app.User.Email
		}
		result = append(result, app.ToDTO(nickname, email))
	}

	c.JSON(http.StatusOK, result)
}

// StreamDocument allows admins or the applicant to download/preview uploaded resume/certificate
func (ac *AdminController) StreamDocument(c *gin.Context) {
	appID := c.Param("id")
	fileType := c.Param("file_type") // "resume" or "certificate"

	callerID, _ := c.Get("user_id")
	callerRole, _ := c.Get("role")

	var app models.SpecialistApplication
	if err := ac.DB.Where("application_id = ?", appID).First(&app).Error; err != nil {
		respondAPIError(c, http.StatusNotFound, "ADMIN_APPLICATION_NOT_FOUND", "Application not found")
		return
	}

	// Security check: must be admin or applicant owner
	if callerRole != "admin" && callerID != app.UserID {
		respondAPIError(c, http.StatusForbidden, "AUTH_FORBIDDEN", "Access denied")
		return
	}

	var targetPath string
	if fileType == "resume" {
		targetPath = app.ResumePath
	} else if fileType == "certificate" {
		targetPath = app.CertificatePath
	} else {
		respondAPIError(c, http.StatusBadRequest, "ADMIN_INVALID_FILE_TYPE", "Invalid file type. Must be 'resume' or 'certificate'")
		return
	}

	if targetPath == "" {
		respondAPIError(c, http.StatusNotFound, "ADMIN_DOCUMENT_NOT_FOUND", "No document associated with this request")
		return
	}

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		respondAPIError(c, http.StatusNotFound, "ADMIN_DOCUMENT_MISSING", "Document file not found on server")
		return
	}

	filename := filepath.Base(targetPath)
	c.FileAttachment(targetPath, filename)
}

// Approve transitions application to approved and elevates applicant to specialist
func (ac *AdminController) Approve(c *gin.Context) {
	appID := c.Param("id")
	adminID, _ := c.Get("user_id")

	var app models.SpecialistApplication
	if err := ac.DB.Where("application_id = ?", appID).First(&app).Error; err != nil {
		respondAPIError(c, http.StatusNotFound, "ADMIN_APPLICATION_NOT_FOUND", "Application not found")
		return
	}

	if app.Status == "approved" {
		respondAPIError(c, http.StatusBadRequest, "ADMIN_ALREADY_APPROVED", "Application is already approved")
		return
	}

	adminIDStr := adminID.(string)
	now := time.Now()

	tx := ac.DB.Begin()

	// 1. Update application status
	app.Status = "approved"
	app.ReviewedBy = &adminIDStr
	app.ReviewedAt = &now
	app.UpdatedAt = now
	if err := tx.Save(&app).Error; err != nil {
		tx.Rollback()
		respondAPIError(c, http.StatusInternalServerError, "ADMIN_APPROVE_FAILED", "Failed to update application status")
		return
	}

	// 2. Elevate user role to specialist
	if err := tx.Model(&models.User{}).Where("user_id = ?", app.UserID).Update("role", "specialist").Error; err != nil {
		tx.Rollback()
		respondAPIError(c, http.StatusInternalServerError, "ADMIN_APPROVE_FAILED", "Failed to update user role")
		return
	}

	tx.Commit()

	// 3. Dispatch in-app notification to applicant
	go func(targetUserID string) {
		notif := newTemplatedNotification(
			targetUserID,
			"Specialist Application Approved",
			"Congratulations! Your application has been approved. You now have full specialist permissions.",
			"specialist_approved",
			nil,
			"/specialist/reviews",
		)
		ac.DB.Create(&notif)
	}(app.UserID)

	c.JSON(http.StatusOK, gin.H{
		"message":        "Application approved and user elevated to Specialist successfully",
		"application_id": app.ApplicationID,
		"user_id":        app.UserID,
		"new_role":       "specialist",
	})
}

// Reject transitions application to rejected with a mandatory reason note
func (ac *AdminController) Reject(c *gin.Context) {
	appID := c.Param("id")
	adminID, _ := c.Get("user_id")

	var input dto.AdminRejectRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondAPIError(c, http.StatusBadRequest, "ADMIN_REJECTION_REASON_REQUIRED", "A rejection reason is required (minimum 3 characters)")
		return
	}

	var app models.SpecialistApplication
	if err := ac.DB.Where("application_id = ?", appID).First(&app).Error; err != nil {
		respondAPIError(c, http.StatusNotFound, "ADMIN_APPLICATION_NOT_FOUND", "Application not found")
		return
	}

	adminIDStr := adminID.(string)
	now := time.Now()

	app.Status = "rejected"
	app.RejectionReason = input.Reason
	app.ReviewedBy = &adminIDStr
	app.ReviewedAt = &now
	app.UpdatedAt = now

	if err := ac.DB.Save(&app).Error; err != nil {
		respondAPIError(c, http.StatusInternalServerError, "ADMIN_REJECT_FAILED", "Failed to update application status")
		return
	}

	// Dispatch in-app notification to applicant
	go func(targetUserID, reason string) {
		notif := newTemplatedNotification(
			targetUserID,
			"Specialist Application Update",
			fmt.Sprintf("Your specialist application was reviewed and declined. Reason: %s", reason),
			"specialist_rejected",
			map[string]any{"reason": reason},
			"/profile",
		)
		ac.DB.Create(&notif)
	}(app.UserID, input.Reason)

	c.JSON(http.StatusOK, gin.H{
		"message":          "Application rejected",
		"application_id":   app.ApplicationID,
		"rejection_reason": input.Reason,
	})
}
