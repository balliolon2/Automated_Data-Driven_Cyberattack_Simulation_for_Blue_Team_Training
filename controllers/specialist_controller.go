package controllers

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cybersim/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

type SpecialistController struct {
	DB *gorm.DB
}

func NewSpecialistController(db *gorm.DB) *SpecialistController {
	return &SpecialistController{DB: db}
}

const (
	MaxFileSize     = 10 * 1024 * 1024 // 10MB
	UploadBaseDir   = "./uploads/applications"
)

var allowedExtensions = map[string]bool{
	".pdf":  true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
}

// Apply handles specialist application submission with multipart form and documents
func (sc *SpecialistController) Apply(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(MaxFileSize * 2); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload too large or invalid multipart form"})
		return
	}

	bio := strings.TrimSpace(c.PostForm("bio"))
	if bio == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bio is required"})
		return
	}

	linkedinURL := strings.TrimSpace(c.PostForm("linkedin_url"))
	portfolioURL := strings.TrimSpace(c.PostForm("portfolio_url"))

	// Validate resume file
	resumeHeader, err := c.FormFile("resume")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Resume file is required"})
		return
	}

	if resumeHeader.Size > MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Resume file exceeds 10MB limit"})
		return
	}

	resumeExt := strings.ToLower(filepath.Ext(resumeHeader.Filename))
	if !allowedExtensions[resumeExt] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resume file type. Allowed: PDF, PNG, JPG, JPEG"})
		return
	}

	// Prepare application directory
	appID := generateUUID()
	appDir := filepath.Join(UploadBaseDir, appID)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload storage directory"})
		return
	}

	// Save resume
	resumeFilename := fmt.Sprintf("resume_%s%s", generateUUID(), resumeExt)
	resumeSavePath := filepath.Join(appDir, resumeFilename)
	if err := c.SaveUploadedFile(resumeHeader, resumeSavePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save resume file"})
		return
	}

	// Optional certificate file
	var certSavePath string
	certHeader, err := c.FormFile("certificate")
	if err == nil && certHeader != nil {
		if certHeader.Size > MaxFileSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Certificate file exceeds 10MB limit"})
			return
		}
		certExt := strings.ToLower(filepath.Ext(certHeader.Filename))
		if !allowedExtensions[certExt] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid certificate file type. Allowed: PDF, PNG, JPG, JPEG"})
			return
		}
		certFilename := fmt.Sprintf("certificate_%s%s", generateUUID(), certExt)
		certSavePath = filepath.Join(appDir, certFilename)
		if err := c.SaveUploadedFile(certHeader, certSavePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save certificate file"})
			return
		}
	}

	// Check if applicant already has a previous application
	var existing models.SpecialistApplication
	now := time.Now()
	if err := sc.DB.Where("user_id = ?", userID).Order("created_at desc").First(&existing).Error; err == nil {
		if existing.Status == models.ApplicationStatusPending {
			c.JSON(http.StatusBadRequest, gin.H{"error": "An application is already pending administrator review"})
			return
		}
		if existing.Status == models.ApplicationStatusApproved {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You are already an approved Specialist"})
			return
		}
		if existing.Status == models.ApplicationStatusRejected {
			// Update rejected application to pending (re-application)
			existing.Status = models.ApplicationStatusPending
			existing.Bio = bio
			existing.ResumePath = resumeSavePath
			if certSavePath != "" {
				existing.CertificatePath = certSavePath
			}
			existing.LinkedInURL = linkedinURL
			existing.PortfolioURL = portfolioURL
			existing.RejectionReason = ""
			existing.ReviewedBy = nil
			existing.ReviewedAt = nil
			existing.UpdatedAt = now
			if err := sc.DB.Save(&existing).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update application"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message":        "Application re-submitted successfully",
				"application_id": existing.ApplicationID,
				"status":         existing.Status,
			})
			return
		}
	}

	application := models.SpecialistApplication{
		ApplicationID:   appID,
		UserID:          userID.(string),
		Status:          models.ApplicationStatusPending,
		Bio:             bio,
		ResumePath:      resumeSavePath,
		CertificatePath: certSavePath,
		LinkedInURL:     linkedinURL,
		PortfolioURL:    portfolioURL,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := sc.DB.Create(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create application"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Specialist application submitted successfully",
		"application_id": application.ApplicationID,
		"status":         application.Status,
	})
}

// GetApplicationStatus returns the status of current user's specialist application
func (sc *SpecialistController) GetApplicationStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var app models.SpecialistApplication
	if err := sc.DB.Where("user_id = ?", userID).Order("created_at desc").First(&app).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "No specialist application found"})
		return
	}

	var user models.User
	_ = sc.DB.Where("user_id = ?", userID).First(&user)

	c.JSON(http.StatusOK, app.ToDTO(user.Nickname, user.Email))
}
