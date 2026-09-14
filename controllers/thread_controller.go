package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"cybersim/dto"
	"cybersim/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ThreadController struct {
	DB *gorm.DB
}

func NewThreadController(db *gorm.DB) *ThreadController {
	return &ThreadController{DB: db}
}

// CreateThread publishes a new analysis thread (Specialist or Admin)
func (tc *ThreadController) CreateThread(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.CreateThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cleanTitle := strings.TrimSpace(req.Title)
	if len(cleanTitle) < 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title must be at least 5 characters long"})
		return
	}
	cleanContent := strings.TrimSpace(req.Content)
	if len(cleanContent) < 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content must be at least 20 characters long"})
		return
	}

	if req.ScenarioID != nil && *req.ScenarioID != "" {
		var scenario models.Scenario
		if err := tc.DB.Where("scenario_id = ?", *req.ScenarioID).First(&scenario).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Referenced scenario does not exist"})
			return
		}
	} else {
		req.ScenarioID = nil
	}

	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	thread := models.AnalysisThread{
		AuthorID:    userID.(string),
		ScenarioID:  req.ScenarioID,
		Title:       cleanTitle,
		Content:     cleanContent,
		Tags:        tags,
		UpvoteCount: 0,
		ViewCount:   0,
		IsPinned:    false,
		IsLocked:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := tc.DB.Create(&thread).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create thread"})
		return
	}

	// Preload Author and Scenario
	tc.DB.Preload("Author").Preload("Scenario").First(&thread, "thread_id = ?", thread.ThreadID)

	c.JSON(http.StatusCreated, thread.ToResponseDTO())
}

// ListThreads returns a paginated list of analysis threads with filtering and sorting
func (tc *ThreadController) ListThreads(c *gin.Context) {
	scenarioID := strings.TrimSpace(c.Query("scenario_id"))
	tag := strings.TrimSpace(c.Query("tag"))
	authorID := strings.TrimSpace(c.Query("author_id"))
	sort := strings.ToLower(strings.TrimSpace(c.Query("sort")))

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := tc.DB.Model(&models.AnalysisThread{}).Preload("Author").Preload("Scenario")

	if scenarioID != "" {
		query = query.Where("scenario_id = ?", scenarioID)
	}
	if authorID != "" {
		query = query.Where("author_id = ?", authorID)
	}
	if tag != "" {
		query = query.Where("tags::text LIKE ?", "%\""+tag+"\"%")
	}

	var total int64
	tc.DB.Model(&models.AnalysisThread{}).Scopes(func(d *gorm.DB) *gorm.DB {
		if scenarioID != "" {
			d = d.Where("scenario_id = ?", scenarioID)
		}
		if authorID != "" {
			d = d.Where("author_id = ?", authorID)
		}
		if tag != "" {
			d = d.Where("tags::text LIKE ?", "%\""+tag+"\"%")
		}
		return d
	}).Count(&total)

	if sort == "popular" {
		query = query.Order("is_pinned desc, upvote_count desc, created_at desc")
	} else {
		query = query.Order("is_pinned desc, created_at desc")
	}

	var threads []models.AnalysisThread
	if err := query.Offset(offset).Limit(pageSize).Find(&threads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch threads"})
		return
	}

	items := make([]dto.ThreadResponse, 0, len(threads))
	for _, t := range threads {
		items = append(items, t.ToResponseDTO())
	}

	c.JSON(http.StatusOK, gin.H{
		"threads":   items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetThread returns thread detail and increments view count
func (tc *ThreadController) GetThread(c *gin.Context) {
	threadID := c.Param("id")

	var thread models.AnalysisThread
	if err := tc.DB.Preload("Author").Preload("Scenario").Where("thread_id = ?", threadID).First(&thread).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	tc.DB.Model(&models.AnalysisThread{}).Where("thread_id = ?", threadID).UpdateColumn("view_count", gorm.Expr("view_count + 1"))
	thread.ViewCount++

	c.JSON(http.StatusOK, thread.ToResponseDTO())
}

// UpdateThread updates title, content, or tags (author or admin only)
func (tc *ThreadController) UpdateThread(c *gin.Context) {
	userID, exists := c.Get("user_id")
	userRole, _ := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	threadID := c.Param("id")
	var thread models.AnalysisThread
	if err := tc.DB.Where("thread_id = ?", threadID).First(&thread).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	if thread.AuthorID != userID.(string) && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to edit this thread"})
		return
	}

	var req dto.UpdateThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cleanTitle := strings.TrimSpace(req.Title)
	if len(cleanTitle) < 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title must be at least 5 characters long"})
		return
	}
	cleanContent := strings.TrimSpace(req.Content)
	if len(cleanContent) < 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content must be at least 20 characters long"})
		return
	}

	thread.Title = cleanTitle
	thread.Content = cleanContent
	if req.Tags != nil {
		thread.Tags = req.Tags
	}
	thread.UpdatedAt = time.Now()

	if err := tc.DB.Save(&thread).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update thread"})
		return
	}

	tc.DB.Preload("Author").Preload("Scenario").First(&thread, "thread_id = ?", thread.ThreadID)
	c.JSON(http.StatusOK, thread.ToResponseDTO())
}

// DeleteThread removes a thread (author or admin only)
func (tc *ThreadController) DeleteThread(c *gin.Context) {
	userID, exists := c.Get("user_id")
	userRole, _ := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	threadID := c.Param("id")
	var thread models.AnalysisThread
	if err := tc.DB.Where("thread_id = ?", threadID).First(&thread).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	if thread.AuthorID != userID.(string) && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this thread"})
		return
	}

	if err := tc.DB.Delete(&thread).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete thread"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Thread deleted successfully"})
}
