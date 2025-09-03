package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
)

type CollaborationHandlers struct {
	collaborationService *services.CollaborationService
	rbacService         *services.RBACService
}

// NewCollaborationHandlers creates new collaboration handlers
func NewCollaborationHandlers(collaborationService *services.CollaborationService, rbacService *services.RBACService) *CollaborationHandlers {
	return &CollaborationHandlers{
		collaborationService: collaborationService,
		rbacService:         rbacService,
	}
}

// URL Sharing Endpoints

// ShareURL creates a new URL share
// POST /api/v1/urls/:short_code/share
func (h *CollaborationHandlers) ShareURL(c *gin.Context) {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	var req models.ShareURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Get URL ID from short code (simplified - in real implementation, query from database)
	urlID := int64(1) // This would be looked up from the shortener service

	shareInfo, err := h.collaborationService.ShareURL(urlID, &req, userID.(int64))
	if err != nil {
		if err.Error() == "you don't have permission to share this URL" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "share_denied",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "share_failed",
			Message: "Failed to share URL",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "URL shared successfully",
		"share":   shareInfo,
	})
}

// GetURLShares retrieves all shares for a URL
// GET /api/v1/urls/:short_code/shares
func (h *CollaborationHandlers) GetURLShares(c *gin.Context) {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Get URL ID from short code
	urlID := int64(1) // This would be looked up from the shortener service

	shares, err := h.collaborationService.GetURLShares(urlID, userID.(int64))
	if err != nil {
		if err.Error() == "you don't have permission to view this URL's shares" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "access_denied",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "get_shares_failed",
			Message: "Failed to retrieve shares",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"shares": shares,
		"count":  len(shares),
	})
}

// UpdateShare updates sharing permissions
// PUT /api/v1/shares/:id
func (h *CollaborationHandlers) UpdateShare(c *gin.Context) {
	shareIDStr := c.Param("id")
	shareID, err := strconv.ParseInt(shareIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_share_id",
			Message: "Invalid share ID",
		})
		return
	}

	var req models.UpdateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	shareInfo, err := h.collaborationService.UpdateShare(shareID, &req, userID.(int64))
	if err != nil {
		if err.Error() == "you can only update shares you created" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "update_denied",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_share_failed",
			Message: "Failed to update share",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Share updated successfully",
		"share":   shareInfo,
	})
}

// RevokeShare revokes a URL share
// DELETE /api/v1/shares/:id
func (h *CollaborationHandlers) RevokeShare(c *gin.Context) {
	shareIDStr := c.Param("id")
	shareID, err := strconv.ParseInt(shareIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_share_id",
			Message: "Invalid share ID",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	err = h.collaborationService.RevokeShare(shareID, userID.(int64))
	if err != nil {
		if err.Error() == "you can only revoke shares you created" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "revoke_denied",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "revoke_share_failed",
			Message: "Failed to revoke share",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Share revoked successfully",
	})
}

// Comment Endpoints

// CreateComment creates a new comment on a URL
// POST /api/v1/urls/:short_code/comments
func (h *CollaborationHandlers) CreateComment(c *gin.Context) {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Get URL ID from short code
	urlID := int64(1) // This would be looked up from the shortener service

	comment, err := h.collaborationService.CreateComment(urlID, &req, userID.(int64))
	if err != nil {
		if err.Error() == "you don't have permission to comment on this URL" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "comment_denied",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "comment_failed",
			Message: "Failed to create comment",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Comment created successfully",
		"comment": comment,
	})
}

// GetURLComments retrieves comments for a URL
// GET /api/v1/urls/:short_code/comments
func (h *CollaborationHandlers) GetURLComments(c *gin.Context) {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Get URL ID from short code
	urlID := int64(1) // This would be looked up from the shortener service

	comments, err := h.collaborationService.GetURLComments(urlID, userID.(int64))
	if err != nil {
		if err.Error() == "you don't have permission to view comments on this URL" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "access_denied",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "get_comments_failed",
			Message: "Failed to retrieve comments",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
		"count":    len(comments),
	})
}

// UpdateComment updates a comment
// PUT /api/v1/comments/:id
func (h *CollaborationHandlers) UpdateComment(c *gin.Context) {
	commentIDStr := c.Param("id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_comment_id",
			Message: "Invalid comment ID",
		})
		return
	}

	var req models.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	comment, err := h.collaborationService.UpdateComment(commentID, &req, userID.(int64))
	if err != nil {
		if err.Error() == "you can only edit your own comments" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "edit_denied",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_comment_failed",
			Message: "Failed to update comment",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment updated successfully",
		"comment": comment,
	})
}

// DeleteComment deletes a comment
// DELETE /api/v1/comments/:id
func (h *CollaborationHandlers) DeleteComment(c *gin.Context) {
	commentIDStr := c.Param("id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_comment_id",
			Message: "Invalid comment ID",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	err = h.collaborationService.DeleteComment(commentID, userID.(int64))
	if err != nil {
		if err.Error() == "you can only delete your own comments" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "delete_denied",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "delete_comment_failed",
			Message: "Failed to delete comment",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment deleted successfully",
	})
}

// Bookmark Endpoints

// CreateBookmark creates a bookmark for a URL
// POST /api/v1/urls/:short_code/bookmark
func (h *CollaborationHandlers) CreateBookmark(c *gin.Context) {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	var req models.CreateBookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Get URL ID from short code
	urlID := int64(1) // This would be looked up from the shortener service

	bookmark, err := h.collaborationService.CreateBookmark(urlID, &req, userID.(int64))
	if err != nil {
		if err.Error() == "URL is already bookmarked" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "already_bookmarked",
				Message: err.Error(),
			})
			return
		}

		if err.Error() == "you don't have permission to bookmark this URL" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "bookmark_denied",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "bookmark_failed",
			Message: "Failed to create bookmark",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Bookmark created successfully",
		"bookmark": bookmark,
	})
}

// GetUserBookmarks retrieves all bookmarks for the current user
// GET /api/v1/bookmarks
func (h *CollaborationHandlers) GetUserBookmarks(c *gin.Context) {
	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	bookmarks, err := h.collaborationService.GetUserBookmarks(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "get_bookmarks_failed",
			Message: "Failed to retrieve bookmarks",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bookmarks": bookmarks,
		"count":     len(bookmarks),
	})
}

// UpdateBookmark updates a bookmark
// PUT /api/v1/bookmarks/:id
func (h *CollaborationHandlers) UpdateBookmark(c *gin.Context) {
	bookmarkIDStr := c.Param("id")
	bookmarkID, err := strconv.ParseInt(bookmarkIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_bookmark_id",
			Message: "Invalid bookmark ID",
		})
		return
	}

	var req models.UpdateBookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	bookmark, err := h.collaborationService.UpdateBookmark(bookmarkID, &req, userID.(int64))
	if err != nil {
		if err.Error() == "you can only update your own bookmarks" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "update_denied",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_bookmark_failed",
			Message: "Failed to update bookmark",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Bookmark updated successfully",
		"bookmark": bookmark,
	})
}

// DeleteBookmark deletes a bookmark
// DELETE /api/v1/bookmarks/:id
func (h *CollaborationHandlers) DeleteBookmark(c *gin.Context) {
	bookmarkIDStr := c.Param("id")
	bookmarkID, err := strconv.ParseInt(bookmarkIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_bookmark_id",
			Message: "Invalid bookmark ID",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	err = h.collaborationService.DeleteBookmark(bookmarkID, userID.(int64))
	if err != nil {
		if err.Error() == "you can only delete your own bookmarks" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "delete_denied",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "delete_bookmark_failed",
			Message: "Failed to delete bookmark",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bookmark deleted successfully",
	})
}

// Collection Endpoints

// CreateCollection creates a new URL collection
// POST /api/v1/collections
func (h *CollaborationHandlers) CreateCollection(c *gin.Context) {
	var req models.CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	collection, err := h.collaborationService.CreateCollection(&req, userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "create_collection_failed",
			Message: "Failed to create collection",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Collection created successfully",
		"collection": collection,
	})
}

// GetUserCollections retrieves all collections for the current user
// GET /api/v1/collections
func (h *CollaborationHandlers) GetUserCollections(c *gin.Context) {
	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	collections, err := h.collaborationService.GetUserCollections(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "get_collections_failed",
			Message: "Failed to retrieve collections",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"collections": collections,
		"count":       len(collections),
	})
}

// GetCollection retrieves a specific collection
// GET /api/v1/collections/:id
func (h *CollaborationHandlers) GetCollection(c *gin.Context) {
	collectionIDStr := c.Param("id")
	collectionID, err := strconv.ParseInt(collectionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_collection_id",
			Message: "Invalid collection ID",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	collection, err := h.collaborationService.GetCollectionInfo(collectionID, userID.(int64))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "collection_not_found",
			Message: "Collection not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"collection": collection,
	})
}

// AddToCollection adds URLs to a collection
// POST /api/v1/collections/:id/urls
func (h *CollaborationHandlers) AddToCollection(c *gin.Context) {
	collectionIDStr := c.Param("id")
	collectionID, err := strconv.ParseInt(collectionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_collection_id",
			Message: "Invalid collection ID",
		})
		return
	}

	var req models.AddToCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	err = h.collaborationService.AddToCollection(collectionID, &req, userID.(int64))
	if err != nil {
		if err.Error() == "you don't have permission to edit this collection" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "edit_denied",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "add_to_collection_failed",
			Message: "Failed to add URLs to collection",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "URLs added to collection successfully",
	})
}

// Activity and Summary Endpoints

// GetURLActivity retrieves activity log for a URL
// GET /api/v1/urls/:short_code/activity
func (h *CollaborationHandlers) GetURLActivity(c *gin.Context) {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	// Parse limit parameter
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Get URL ID from short code
	urlID := int64(1) // This would be looked up from the shortener service

	activities, err := h.collaborationService.GetURLActivity(urlID, userID.(int64), limit)
	if err != nil {
		if err.Error() == "you don't have permission to view activity for this URL" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "access_denied",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "get_activity_failed",
			Message: "Failed to retrieve activity",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
		"count":      len(activities),
	})
}

// GetCollaborationSummary retrieves collaboration summary for a URL
// GET /api/v1/urls/:short_code/collaboration
func (h *CollaborationHandlers) GetCollaborationSummary(c *gin.Context) {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Get URL ID from short code
	urlID := int64(1) // This would be looked up from the shortener service

	summary, err := h.collaborationService.GetCollaborationSummary(urlID, userID.(int64))
	if err != nil {
		if err.Error() == "you don't have permission to view this URL" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "access_denied",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "get_summary_failed",
			Message: "Failed to retrieve collaboration summary",
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}