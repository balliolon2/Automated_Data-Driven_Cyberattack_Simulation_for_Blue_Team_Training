package controllers

import (
	"fmt"
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

// validateThreadContent validates and cleans title and content
func validateThreadContent(title, content string) (string, string, error) {
	cleanTitle := strings.TrimSpace(title)
	if len(cleanTitle) < 5 {
		return "", "", fmt.Errorf("Title must be at least 5 characters long")
	}
	cleanContent := strings.TrimSpace(content)
	if len(cleanContent) < 20 {
		return "", "", fmt.Errorf("Content must be at least 20 characters long")
	}
	return cleanTitle, cleanContent, nil
}

// buildThreadFilterScope constructs reusable query filters for listing threads
func buildThreadFilterScope(scenarioID, domainID, authorID, tag string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if scenarioID != "" {
			db = db.Where("analysis_threads.scenario_id = ?", scenarioID)
		}
		if domainID != "" {
			db = db.Joins("JOIN scenarios ON scenarios.scenario_id = analysis_threads.scenario_id").
				Where("scenarios.domain_id = ?", domainID)
		}
		if authorID != "" {
			db = db.Where("analysis_threads.author_id = ?", authorID)
		}
		if tag != "" {
			db = db.Where("analysis_threads.tags::text LIKE ?", "%\""+tag+"\"%")
		}
		return db
	}
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

	cleanTitle, cleanContent, err := validateThreadContent(req.Title, req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	domainID := strings.TrimSpace(c.Query("domain_id"))
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

	filterScope := buildThreadFilterScope(scenarioID, domainID, authorID, tag)

	var total int64
	tc.DB.Model(&models.AnalysisThread{}).Scopes(filterScope).Count(&total)

	query := tc.DB.Model(&models.AnalysisThread{}).
		Preload("Author").
		Preload("Scenario").
		Scopes(filterScope)

	if sort == "popular" || sort == "upvotes" {
		query = query.Order("analysis_threads.is_pinned desc, analysis_threads.upvote_count desc, analysis_threads.created_at desc")
	} else {
		// Default: latest/recent
		query = query.Order("analysis_threads.is_pinned desc, analysis_threads.created_at desc")
	}

	var threads []models.AnalysisThread
	if err := query.Offset(offset).Limit(pageSize).Find(&threads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch threads"})
		return
	}

	var reqUserID string
	if uid, exists := c.Get("user_id"); exists {
		if s, ok := uid.(string); ok {
			reqUserID = s
		}
	}

	userUpvotes := make(map[string]bool)
	if reqUserID != "" && len(threads) > 0 {
		var threadIDs []string
		for _, t := range threads {
			threadIDs = append(threadIDs, t.ThreadID)
		}
		var upvotedIDs []string
		tc.DB.Model(&models.ThreadUpvote{}).Where("thread_id IN ? AND user_id = ?", threadIDs, reqUserID).Pluck("thread_id", &upvotedIDs)
		for _, id := range upvotedIDs {
			userUpvotes[id] = true
		}
	}

	items := make([]dto.ThreadResponse, 0, len(threads))
	for _, t := range threads {
		resp := t.ToResponseDTO()
		if userUpvotes[t.ThreadID] {
			resp.UserHasUpvoted = true
		}
		items = append(items, resp)
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

	resp := thread.ToResponseDTO()
	if uid, exists := c.Get("user_id"); exists {
		if s, ok := uid.(string); ok && s != "" {
			var count int64
			tc.DB.Model(&models.ThreadUpvote{}).Where("thread_id = ? AND user_id = ?", thread.ThreadID, s).Count(&count)
			resp.UserHasUpvoted = count > 0
		}
	}

	c.JSON(http.StatusOK, resp)
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

	cleanTitle, cleanContent, err := validateThreadContent(req.Title, req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
