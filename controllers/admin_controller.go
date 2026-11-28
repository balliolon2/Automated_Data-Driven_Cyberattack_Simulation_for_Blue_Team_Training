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

type AdminStore interface {
	GetApplicationByID(appID string) (*models.SpecialistApplication, error)
	GetUserByID(userID string) (*models.User, error)
	ApproveApplication(app *models.SpecialistApplication, adminID string, now time.Time) error
}

type gormAdminStore struct {
	DB *gorm.DB
}

func (s *gormAdminStore) GetApplicationByID(appID string) (*models.SpecialistApplication, error) {
	var app models.SpecialistApplication
	if err := s.DB.Where("application_id = ?", appID).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (s *gormAdminStore) GetUserByID(userID string) (*models.User, error) {
	var user models.User
	if err := s.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *gormAdminStore) ApproveApplication(app *models.SpecialistApplication, adminID string, now time.Time) error {
	tx := s.DB.Begin()
	app.Status = "approved"
	app.ReviewedBy = &adminID
	app.ReviewedAt = &now
	app.UpdatedAt = now
	if err := tx.Save(app).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&models.User{}).Where("user_id = ?", app.UserID).Update("role", "specialist").Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

type AdminController struct {
	DB    *gorm.DB
	store AdminStore
}

func NewAdminController(db *gorm.DB) *AdminController {
	return &AdminController{
		DB:    db,
		store: &gormAdminStore{DB: db},
	}
}

func NewAdminControllerWithStore(db *gorm.DB, store AdminStore) *AdminController {
	return &AdminController{
		DB:    db,
		store: store,
	}
}

func (ac *AdminController) effectiveStore() AdminStore {
	if ac.store != nil {
		return ac.store
	}
	if ac.DB == nil {
		return nil
	}
	return &gormAdminStore{DB: ac.DB}
}

func buildVerifiedApplicationsQuery(db *gorm.DB, status string) *gorm.DB {
	query := db.Model(&models.SpecialistApplication{}).
		Joins("JOIN users ON users.user_id = specialist_applications.user_id").
		Where("users.email_verified = ?", true).
		Preload("User")
	if status != "" {
		query = query.Where("specialist_applications.status = ?", status)
	}
	return query.Order("specialist_applications.created_at desc")
}

// ListApplications retrieves verified-applicant specialist applications filtered by status
func (ac *AdminController) ListApplications(c *gin.Context) {
	if ac.DB == nil {
		respondAPIError(c, http.StatusInternalServerError, "ADMIN_APPLICATIONS_FETCH_FAILED", "Database is unavailable")
		return
	}
	status := strings.TrimSpace(c.Query("status"))

	var applications []models.SpecialistApplication
	if err := buildVerifiedApplicationsQuery(ac.DB, status).Find(&applications).Error; err != nil {
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
	store := ac.effectiveStore()
	if store == nil {
		respondAPIError(c, http.StatusInternalServerError, "ADMIN_APPROVE_FAILED", "Admin store unavailable")
		return
	}

	app, err := store.GetApplicationByID(appID)
	if err != nil {
		respondAPIError(c, http.StatusNotFound, "ADMIN_APPLICATION_NOT_FOUND", "Application not found")
		return
	}

	if app.Status == "approved" {
		respondAPIError(c, http.StatusBadRequest, "ADMIN_ALREADY_APPROVED", "Application is already approved")
		return
	}

	applicant, err := store.GetUserByID(app.UserID)
	if err != nil {
		respondAPIError(c, http.StatusInternalServerError, "ADMIN_APPROVE_FAILED", "Failed to verify applicant status")
		return
	}
	if !applicant.EmailVerified {
		respondAPIError(c, http.StatusBadRequest, "ADMIN_APPLICANT_EMAIL_UNVERIFIED", "Applicant must verify their email address before approval")
		return
	}

	adminIDStr, _ := adminID.(string)
	now := time.Now()
	if err := store.ApproveApplication(app, adminIDStr, now); err != nil {
		respondAPIError(c, http.StatusInternalServerError, "ADMIN_APPROVE_FAILED", "Failed to approve application")
		return
	}

	// 3. Dispatch in-app notification to applicant
	if ac.DB != nil {
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
	}

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
