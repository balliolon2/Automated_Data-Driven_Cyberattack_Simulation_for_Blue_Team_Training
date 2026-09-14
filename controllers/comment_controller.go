package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"cybersim/dto"
	"cybersim/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommentController struct {
	DB *gorm.DB
}

func NewCommentController(db *gorm.DB) *CommentController {
	return &CommentController{DB: db}
}

// CreateComment posts a top-level comment or nested reply with 2-level normalization
func (cc *CommentController) CreateComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userNickname, _ := c.Get("nickname")
	userRole, _ := c.Get("role")

	threadID := c.Param("id")
	var thread models.AnalysisThread
	if err := cc.DB.Where("thread_id = ?", threadID).First(&thread).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Analysis thread not found"})
		return
	}

	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cleanContent := strings.TrimSpace(req.Content)
	if len(cleanContent) < 1 || len(cleanContent) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comment must be between 1 and 2000 characters"})
		return
	}

	var parentID *string
	var parentComment models.ThreadComment
	var isReply bool

	if req.ParentCommentID != nil && *req.ParentCommentID != "" {
		if err := cc.DB.Where("comment_id = ? AND thread_id = ?", *req.ParentCommentID, threadID).First(&parentComment).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parent comment not found in this thread"})
			return
		}
		isReply = true

		// 2-Level Flattening: If parent is already a level-2 reply, normalize to its root ancestor
		if parentComment.ParentCommentID != nil && *parentComment.ParentCommentID != "" {
			parentID = parentComment.ParentCommentID
		} else {
			targetID := parentComment.CommentID
			parentID = &targetID
		}
	}

	comment := models.ThreadComment{
		ThreadID:        threadID,
		UserID:          userID.(string),
		ParentCommentID: parentID,
		Content:         cleanContent,
		UpvoteCount:     0,
		IsPinned:        false,
		IsHidden:        false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := cc.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	// Trigger in-app notification asynchronously
	go func() {
		nickStr := ""
		if userNickname != nil {
			nickStr = userNickname.(string)
		}
		if nickStr == "" {
			nickStr = "An analyst"
		}

		if isReply && parentComment.UserID != userID.(string) {
			// Notify parent comment author
			cc.DB.Create(&models.UserNotification{
				UserID:    parentComment.UserID,
				Title:     "New reply to your comment",
				Message:   fmt.Sprintf("%s replied to your comment on '%s'", nickStr, thread.Title),
				LinkURL:   fmt.Sprintf("/discussions/%s#comment-%s", thread.ThreadID, comment.CommentID),
				IsRead:    false,
				CreatedAt: time.Now(),
			})
		} else if !isReply && thread.AuthorID != userID.(string) {
			// Notify thread author
			cc.DB.Create(&models.UserNotification{
				UserID:    thread.AuthorID,
				Title:     "New comment on your analysis post",
				Message:   fmt.Sprintf("%s commented on '%s'", nickStr, thread.Title),
				LinkURL:   fmt.Sprintf("/discussions/%s#comment-%s", thread.ThreadID, comment.CommentID),
				IsRead:    false,
				CreatedAt: time.Now(),
			})
		}
	}()

	authorRole := "learner"
	if userRole != nil {
		authorRole = userRole.(string)
	}

	c.JSON(http.StatusCreated, dto.CommentResponse{
		CommentID:       comment.CommentID,
		ThreadID:        comment.ThreadID,
		ParentCommentID: comment.ParentCommentID,
		Author: dto.CommentAuthorDTO{
			UserID:   userID.(string),
			Nickname: userNickname.(string),
			Role:     authorRole,
		},
		Content:     comment.Content,
		UpvoteCount: 0,
		IsPinned:    false,
		IsHidden:    false,
		HasUpvoted:  false,
		Replies:     []dto.CommentResponse{},
		CreatedAt:   comment.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   comment.UpdatedAt.Format(time.RFC3339),
	})
}

// ListComments retrieves structured 2-level comments for an analysis thread
func (cc *CommentController) ListComments(c *gin.Context) {
	threadID := c.Param("id")
	userID, hasUser := c.Get("user_id")
	userRole, _ := c.Get("role")

	var thread models.AnalysisThread
	if err := cc.DB.Where("thread_id = ?", threadID).First(&thread).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Analysis thread not found"})
		return
	}

	// Fetch all Level 1 comments
	var rootComments []models.ThreadComment
	if err := cc.DB.Preload("User").
		Where("thread_id = ? AND parent_comment_id IS NULL", threadID).
		Order("is_pinned desc, upvote_count desc, created_at asc").
		Find(&rootComments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}

	// Map user upvotes if authenticated
	upvotedCommentIDs := make(map[string]bool)
	if hasUser {
		var upvotes []models.CommentUpvote
		cc.DB.Where("user_id = ?", userID.(string)).Find(&upvotes)
		for _, u := range upvotes {
			upvotedCommentIDs[u.CommentID] = true
		}
	}

	result := make([]dto.CommentResponse, 0, len(rootComments))
	for _, root := range rootComments {
		// Fetch Level 2 replies for this root comment
		var childComments []models.ThreadComment
		cc.DB.Preload("User").
			Where("parent_comment_id = ?", root.CommentID).
			Order("created_at asc").
			Find(&childComments)

		replies := make([]dto.CommentResponse, 0, len(childComments))
		for _, ch := range childComments {
			chContent := ch.Content
			if ch.IsHidden {
				canViewHidden := hasUser && (ch.UserID == userID.(string) || thread.AuthorID == userID.(string) || userRole == "admin")
				if !canViewHidden {
					chContent = "[This reply has been hidden by the moderator]"
				}
			}

			chNick := "Anonymous"
			chRole := "learner"
			if ch.User != nil {
				chNick = ch.User.Nickname
				chRole = ch.User.Role
			}

			replies = append(replies, dto.CommentResponse{
				CommentID:       ch.CommentID,
				ThreadID:        ch.ThreadID,
				ParentCommentID: ch.ParentCommentID,
				Author: dto.CommentAuthorDTO{
					UserID:   ch.UserID,
					Nickname: chNick,
					Role:     chRole,
				},
				Content:     chContent,
				UpvoteCount: ch.UpvoteCount,
				IsPinned:    ch.IsPinned,
				IsHidden:    ch.IsHidden,
				HasUpvoted:  upvotedCommentIDs[ch.CommentID],
				Replies:     []dto.CommentResponse{},
				CreatedAt:   ch.CreatedAt.Format(time.RFC3339),
				UpdatedAt:   ch.UpdatedAt.Format(time.RFC3339),
			})
		}

		rootContent := root.Content
		if root.IsHidden {
			canViewHidden := hasUser && (root.UserID == userID.(string) || thread.AuthorID == userID.(string) || userRole == "admin")
			if !canViewHidden {
				rootContent = "[This comment has been hidden by the moderator]"
			}
		}

		rNick := "Anonymous"
		rRole := "learner"
		if root.User != nil {
			rNick = root.User.Nickname
			rRole = root.User.Role
		}

		result = append(result, dto.CommentResponse{
			CommentID:       root.CommentID,
			ThreadID:        root.ThreadID,
			ParentCommentID: root.ParentCommentID,
			Author: dto.CommentAuthorDTO{
				UserID:   root.UserID,
				Nickname: rNick,
				Role:     rRole,
			},
			Content:     rootContent,
			UpvoteCount: root.UpvoteCount,
			IsPinned:    root.IsPinned,
			IsHidden:    root.IsHidden,
			HasUpvoted:  upvotedCommentIDs[root.CommentID],
			Replies:     replies,
			CreatedAt:   root.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   root.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, result)
}

// UpdateComment edits comment content (Author or Admin only)
func (cc *CommentController) UpdateComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	userRole, _ := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	commentID := c.Param("id")
	var comment models.ThreadComment
	if err := cc.DB.Where("comment_id = ?", commentID).First(&comment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	if comment.UserID != userID.(string) && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to edit this comment"})
		return
	}

	var req dto.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cleanContent := strings.TrimSpace(req.Content)
	if len(cleanContent) < 1 || len(cleanContent) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comment must be between 1 and 2000 characters"})
		return
	}

	comment.Content = cleanContent
	comment.UpdatedAt = time.Now()

	if err := cc.DB.Save(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment updated successfully", "content": comment.Content})
}

// DeleteComment removes a comment (Author, Thread Author Specialist, or Admin)
func (cc *CommentController) DeleteComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	userRole, _ := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	commentID := c.Param("id")
	var comment models.ThreadComment
	if err := cc.DB.Where("comment_id = ?", commentID).First(&comment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	var thread models.AnalysisThread
	cc.DB.Where("thread_id = ?", comment.ThreadID).First(&thread)

	isAuthor := comment.UserID == userID.(string)
	isThreadAuthor := thread.AuthorID == userID.(string)
	isAdmin := userRole == "admin"

	if !isAuthor && !isThreadAuthor && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this comment"})
		return
	}

	if err := cc.DB.Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}

// TogglePinComment toggles pin status of a comment (Thread Author or Admin only, max 3 pinned)
func (cc *CommentController) TogglePinComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	userRole, _ := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	commentID := c.Param("id")
	var comment models.ThreadComment
	if err := cc.DB.Where("comment_id = ?", commentID).First(&comment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	var thread models.AnalysisThread
	if err := cc.DB.Where("thread_id = ?", comment.ThreadID).First(&thread).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	if thread.AuthorID != userID.(string) && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the thread author or admin can pin comments"})
		return
	}

	if !comment.IsPinned {
		// Enforce max 3 pinned comments per thread
		var pinnedCount int64
		cc.DB.Model(&models.ThreadComment{}).Where("thread_id = ? AND is_pinned = true", comment.ThreadID).Count(&pinnedCount)
		if pinnedCount >= 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum of 3 comments can be pinned per thread"})
			return
		}
		comment.IsPinned = true
	} else {
		comment.IsPinned = false
	}

	if err := cc.DB.Save(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle pin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment pin status updated", "is_pinned": comment.IsPinned})
}

// ToggleHideComment toggles hidden state for spam moderation (Thread Author or Admin only)
func (cc *CommentController) ToggleHideComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	userRole, _ := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	commentID := c.Param("id")
	var comment models.ThreadComment
	if err := cc.DB.Where("comment_id = ?", commentID).First(&comment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	var thread models.AnalysisThread
	cc.DB.Where("thread_id = ?", comment.ThreadID).First(&thread)

	if thread.AuthorID != userID.(string) && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only thread author or admin can hide comments"})
		return
	}

	comment.IsHidden = !comment.IsHidden
	if err := cc.DB.Save(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment moderation state"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment visibility toggled", "is_hidden": comment.IsHidden})
}
