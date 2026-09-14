package controllers

import (
	"net/http"
	"time"

	"cybersim/dto"
	"cybersim/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UpvoteController struct {
	DB *gorm.DB
}

func NewUpvoteController(db *gorm.DB) *UpvoteController {
	return &UpvoteController{DB: db}
}

// ToggleThreadUpvote toggles Helpful upvote on an analysis thread (1-vote-per-user)
func (uc *UpvoteController) ToggleThreadUpvote(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	threadID := c.Param("id")
	var thread models.AnalysisThread
	if err := uc.DB.Where("thread_id = ?", threadID).First(&thread).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Analysis thread not found"})
		return
	}

	var hasUpvoted bool
	var newCount int

	err := uc.DB.Transaction(func(tx *gorm.DB) error {
		var existing models.ThreadUpvote
		if err := tx.Where("thread_id = ? AND user_id = ?", threadID, userID.(string)).First(&existing).Error; err == nil {
			// Already upvoted -> remove upvote
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.AnalysisThread{}).Where("thread_id = ?", threadID).
				UpdateColumn("upvote_count", gorm.Expr("GREATEST(0, upvote_count - 1)")).Error; err != nil {
				return err
			}
			hasUpvoted = false
		} else {
			// Not upvoted yet -> add upvote
			upvote := models.ThreadUpvote{
				ThreadID:  threadID,
				UserID:    userID.(string),
				CreatedAt: time.Now(),
			}
			if err := tx.Create(&upvote).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.AnalysisThread{}).Where("thread_id = ?", threadID).
				UpdateColumn("upvote_count", gorm.Expr("upvote_count + 1")).Error; err != nil {
				return err
			}
			hasUpvoted = true
		}

		// Retrieve refreshed count
		var refreshed models.AnalysisThread
		if err := tx.Select("upvote_count").Where("thread_id = ?", threadID).First(&refreshed).Error; err != nil {
			return err
		}
		newCount = refreshed.UpvoteCount
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle upvote"})
		return
	}

	c.JSON(http.StatusOK, dto.UpvoteResponse{
		UpvoteCount: newCount,
		HasUpvoted:  hasUpvoted,
	})
}

// ToggleCommentUpvote toggles Helpful upvote on a comment (1-vote-per-user)
func (uc *UpvoteController) ToggleCommentUpvote(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	commentID := c.Param("id")
	var comment models.ThreadComment
	if err := uc.DB.Where("comment_id = ?", commentID).First(&comment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	var hasUpvoted bool
	var newCount int

	err := uc.DB.Transaction(func(tx *gorm.DB) error {
		var existing models.CommentUpvote
		if err := tx.Where("comment_id = ? AND user_id = ?", commentID, userID.(string)).First(&existing).Error; err == nil {
			// Remove upvote
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.ThreadComment{}).Where("comment_id = ?", commentID).
				UpdateColumn("upvote_count", gorm.Expr("GREATEST(0, upvote_count - 1)")).Error; err != nil {
				return err
			}
			hasUpvoted = false
		} else {
			// Add upvote
			upvote := models.CommentUpvote{
				CommentID: commentID,
				UserID:    userID.(string),
				CreatedAt: time.Now(),
			}
			if err := tx.Create(&upvote).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.ThreadComment{}).Where("comment_id = ?", commentID).
				UpdateColumn("upvote_count", gorm.Expr("upvote_count + 1")).Error; err != nil {
				return err
			}
			hasUpvoted = true
		}

		var refreshed models.ThreadComment
		if err := tx.Select("upvote_count").Where("comment_id = ?", commentID).First(&refreshed).Error; err != nil {
			return err
		}
		newCount = refreshed.UpvoteCount
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle comment upvote"})
		return
	}

	c.JSON(http.StatusOK, dto.UpvoteResponse{
		UpvoteCount: newCount,
		HasUpvoted:  hasUpvoted,
	})
}
