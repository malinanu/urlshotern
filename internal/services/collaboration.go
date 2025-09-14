package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/URLshorter/url-shortener/internal/utils"
)

type CollaborationService struct {
	db           *storage.PostgresStorage
	redis        *storage.RedisStorage
	userService  *UserService
	teamService  *TeamService
	emailService *EmailService
}

// NewCollaborationService creates a new collaboration service
func NewCollaborationService(
	db *storage.PostgresStorage,
	redis *storage.RedisStorage,
	userService *UserService,
	teamService *TeamService,
	emailService *EmailService,
) *CollaborationService {
	return &CollaborationService{
		db:           db,
		redis:        redis,
		userService:  userService,
		teamService:  teamService,
		emailService: emailService,
	}
}

// URL Sharing

// ShareURL creates a new URL share
func (c *CollaborationService) ShareURL(urlID int64, req *models.ShareURLRequest, sharerID int64) (*models.URLShareInfo, error) {
	// Validate share request
	if err := c.validateShareRequest(req); err != nil {
		return nil, err
	}

	// Check if URL exists and if user has permission to share
	canShare, err := c.canUserShareURL(urlID, sharerID)
	if err != nil {
		return nil, err
	}
	if !canShare {
		return nil, fmt.Errorf("you don't have permission to share this URL")
	}

	// Create share record
	share := &models.URLShare{
		ID:             time.Now().Unix(), // Temporary ID generation
		URLID:          urlID,
		SharerID:       sharerID,
		SharedWithType: req.SharedWithType,
		SharedWithID:   req.SharedWithID,
		ShareType:      req.ShareType,
		ExpiresAt:      req.ExpiresAt,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Save to database
	// Implementation would save share to database

	// Log activity
	c.logActivity(urlID, &sharerID, models.ActivityURLShared, "URL shared", map[string]interface{}{
		"share_type":       req.ShareType,
		"shared_with_type": req.SharedWithType,
		"shared_with_id":   req.SharedWithID,
	})

	// Send notification if sharing with user
	if req.SharedWithType == models.SharedWithUser && req.SharedWithID != nil {
		go c.sendShareNotification(share, req.Message)
	}

	return c.GetShareInfo(share.ID)
}

// GetURLShares retrieves all shares for a URL
func (c *CollaborationService) GetURLShares(urlID int64, userID int64) ([]*models.URLShareInfo, error) {
	// Check if user has access to view shares
	canView, err := c.canUserViewURL(urlID, userID)
	if err != nil {
		return nil, err
	}
	if !canView {
		return nil, fmt.Errorf("you don't have permission to view this URL's shares")
	}

	// Implementation would query database
	// For now, return simulated shares
	shares := []*models.URLShareInfo{
		{
			URLShare: &models.URLShare{
				ID:             1,
				URLID:          urlID,
				SharerID:       userID,
				SharedWithType: models.SharedWithUser,
				SharedWithID:   utils.Int64Ptr(2),
				ShareType:      models.ShareTypeEdit,
				IsActive:       true,
				CreatedAt:      time.Now().Add(-2 * time.Hour),
				UpdatedAt:      time.Now().Add(-2 * time.Hour),
			},
			Permissions: models.GetSharePermissions(models.ShareTypeEdit),
		},
	}

	return shares, nil
}

// GetShareInfo retrieves detailed information about a share
func (c *CollaborationService) GetShareInfo(shareID int64) (*models.URLShareInfo, error) {
	// Implementation would query database
	// For now, return simulated share info
	share := &models.URLShareInfo{
		URLShare: &models.URLShare{
			ID:             shareID,
			URLID:          1,
			SharerID:       1,
			SharedWithType: models.SharedWithUser,
			SharedWithID:   utils.Int64Ptr(2),
			ShareType:      models.ShareTypeEdit,
			IsActive:       true,
			CreatedAt:      time.Now().Add(-2 * time.Hour),
			UpdatedAt:      time.Now().Add(-2 * time.Hour),
		},
		Permissions: models.GetSharePermissions(models.ShareTypeEdit),
	}

	return share, nil
}

// UpdateShare updates sharing permissions
func (c *CollaborationService) UpdateShare(shareID int64, req *models.UpdateShareRequest, userID int64) (*models.URLShareInfo, error) {
	// Get existing share
	share, err := c.GetShareInfo(shareID)
	if err != nil {
		return nil, err
	}

	// Check if user can update this share
	if share.SharerID != userID {
		return nil, fmt.Errorf("you can only update shares you created")
	}

	// Update fields
	if req.ShareType != "" {
		share.ShareType = req.ShareType
		share.Permissions = models.GetSharePermissions(req.ShareType)
	}
	if req.ExpiresAt != nil {
		share.ExpiresAt = req.ExpiresAt
	}
	if req.IsActive != nil {
		share.IsActive = *req.IsActive
	}

	share.UpdatedAt = time.Now()

	// Save to database
	// Implementation would update share in database

	// Log activity
	action := models.ActivityURLShared
	if req.IsActive != nil && !*req.IsActive {
		action = models.ActivityShareRevoked
	}
	
	c.logActivity(share.URLID, &userID, action, "Share permissions updated", map[string]interface{}{
		"share_id":   shareID,
		"share_type": share.ShareType,
		"is_active":  share.IsActive,
	})

	return share, nil
}

// RevokeShare revokes a URL share
func (c *CollaborationService) RevokeShare(shareID int64, userID int64) error {
	// Get share
	share, err := c.GetShareInfo(shareID)
	if err != nil {
		return err
	}

	// Check if user can revoke this share
	if share.SharerID != userID {
		return fmt.Errorf("you can only revoke shares you created")
	}

	// Deactivate share
	share.IsActive = false
	share.UpdatedAt = time.Now()

	// Save to database
	// Implementation would update share in database

	// Log activity
	c.logActivity(share.URLID, &userID, models.ActivityShareRevoked, "Share revoked", map[string]interface{}{
		"share_id": shareID,
	})

	return nil
}

// Comments

// CreateComment creates a new comment on a URL
func (c *CollaborationService) CreateComment(urlID int64, req *models.CreateCommentRequest, userID int64) (*models.URLCommentInfo, error) {
	// Check if user can comment on this URL
	canComment, err := c.canUserCommentOnURL(urlID, userID)
	if err != nil {
		return nil, err
	}
	if !canComment {
		return nil, fmt.Errorf("you don't have permission to comment on this URL")
	}

	// Create comment
	comment := &models.URLComment{
		ID:        time.Now().Unix(), // Temporary ID generation
		URLID:     urlID,
		UserID:    userID,
		Content:   req.Content,
		ParentID:  req.ParentID,
		IsEdited:  false,
		IsDeleted: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	// Implementation would save comment to database

	// Log activity
	c.logActivity(urlID, &userID, models.ActivityCommentAdded, "Comment added", map[string]interface{}{
		"comment_id": comment.ID,
		"parent_id":  comment.ParentID,
	})

	return c.GetCommentInfo(comment.ID, userID)
}

// GetURLComments retrieves comments for a URL
func (c *CollaborationService) GetURLComments(urlID int64, userID int64) ([]*models.URLCommentInfo, error) {
	// Check if user can view comments
	canView, err := c.canUserViewURL(urlID, userID)
	if err != nil {
		return nil, err
	}
	if !canView {
		return nil, fmt.Errorf("you don't have permission to view comments on this URL")
	}

	// Implementation would query database for comments
	// For now, return simulated comments
	user, _ := c.userService.GetUserByID(userID)
	comments := []*models.URLCommentInfo{
		{
			URLComment: &models.URLComment{
				ID:        1,
				URLID:     urlID,
				UserID:    userID,
				Content:   "This is a great URL!",
				IsEdited:  false,
				IsDeleted: false,
				CreatedAt: time.Now().Add(-2 * time.Hour),
				UpdatedAt: time.Now().Add(-2 * time.Hour),
			},
			User:       user.ToPublic(),
			Replies:    []*models.URLCommentInfo{},
			ReplyCount: 0,
			CanEdit:    true,
			CanDelete:  true,
		},
	}

	return comments, nil
}

// GetCommentInfo retrieves detailed information about a comment
func (c *CollaborationService) GetCommentInfo(commentID int64, requestorID int64) (*models.URLCommentInfo, error) {
	// Implementation would query database
	// For now, return simulated comment
	user, _ := c.userService.GetUserByID(requestorID)
	
	return &models.URLCommentInfo{
		URLComment: &models.URLComment{
			ID:        commentID,
			URLID:     1,
			UserID:    requestorID,
			Content:   "This is a comment",
			IsEdited:  false,
			IsDeleted: false,
			CreatedAt: time.Now().Add(-1 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
		User:       user.ToPublic(),
		Replies:    []*models.URLCommentInfo{},
		ReplyCount: 0,
		CanEdit:    true,
		CanDelete:  true,
	}, nil
}

// UpdateComment updates a comment
func (c *CollaborationService) UpdateComment(commentID int64, req *models.UpdateCommentRequest, userID int64) (*models.URLCommentInfo, error) {
	comment, err := c.GetCommentInfo(commentID, userID)
	if err != nil {
		return nil, err
	}

	// Check if user can edit this comment
	if comment.UserID != userID {
		return nil, fmt.Errorf("you can only edit your own comments")
	}

	// Update comment
	comment.Content = req.Content
	comment.IsEdited = true
	comment.UpdatedAt = time.Now()

	// Save to database
	// Implementation would update comment in database

	// Log activity
	c.logActivity(comment.URLID, &userID, models.ActivityCommentEdited, "Comment edited", map[string]interface{}{
		"comment_id": commentID,
	})

	return comment, nil
}

// DeleteComment deletes a comment
func (c *CollaborationService) DeleteComment(commentID int64, userID int64) error {
	comment, err := c.GetCommentInfo(commentID, userID)
	if err != nil {
		return err
	}

	// Check if user can delete this comment
	if comment.UserID != userID {
		return fmt.Errorf("you can only delete your own comments")
	}

	// Soft delete comment
	comment.IsDeleted = true
	now := time.Now()
	comment.DeletedAt = &now
	comment.UpdatedAt = now

	// Save to database
	// Implementation would update comment in database

	// Log activity
	c.logActivity(comment.URLID, &userID, models.ActivityCommentDeleted, "Comment deleted", map[string]interface{}{
		"comment_id": commentID,
	})

	return nil
}

// Bookmarks

// CreateBookmark creates a bookmark for a URL
func (c *CollaborationService) CreateBookmark(urlID int64, req *models.CreateBookmarkRequest, userID int64) (*models.URLBookmarkInfo, error) {
	// Check if user can access the URL
	canView, err := c.canUserViewURL(urlID, userID)
	if err != nil {
		return nil, err
	}
	if !canView {
		return nil, fmt.Errorf("you don't have permission to bookmark this URL")
	}

	// Check if bookmark already exists
	existing, err := c.getUserBookmark(urlID, userID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("URL is already bookmarked")
	}

	// Create bookmark
	bookmark := &models.URLBookmark{
		ID:        time.Now().Unix(), // Temporary ID generation
		UserID:    userID,
		URLID:     urlID,
		Tags:      req.Tags,
		Notes:     req.Notes,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	// Implementation would save bookmark to database

	// Log activity
	c.logActivity(urlID, &userID, models.ActivityBookmarked, "URL bookmarked", map[string]interface{}{
		"bookmark_id": bookmark.ID,
		"tags":        req.Tags,
	})

	return c.GetBookmarkInfo(bookmark.ID)
}

// GetUserBookmarks retrieves all bookmarks for a user
func (c *CollaborationService) GetUserBookmarks(userID int64) ([]*models.URLBookmarkInfo, error) {
	// Implementation would query database
	// For now, return simulated bookmarks
	bookmarks := []*models.URLBookmarkInfo{}

	return bookmarks, nil
}

// GetBookmarkInfo retrieves detailed bookmark information
func (c *CollaborationService) GetBookmarkInfo(bookmarkID int64) (*models.URLBookmarkInfo, error) {
	// Implementation would query database
	// For now, return simulated bookmark
	return &models.URLBookmarkInfo{
		URLBookmark: &models.URLBookmark{
			ID:        bookmarkID,
			UserID:    1,
			URLID:     1,
			Tags:      "important,work",
			CreatedAt: time.Now().Add(-1 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
		TagList: []string{"important", "work"},
	}, nil
}

// UpdateBookmark updates a bookmark
func (c *CollaborationService) UpdateBookmark(bookmarkID int64, req *models.UpdateBookmarkRequest, userID int64) (*models.URLBookmarkInfo, error) {
	bookmark, err := c.GetBookmarkInfo(bookmarkID)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if bookmark.UserID != userID {
		return nil, fmt.Errorf("you can only update your own bookmarks")
	}

	// Update fields
	if req.Tags != nil {
		bookmark.Tags = *req.Tags
		bookmark.TagList = strings.Split(*req.Tags, ",")
	}
	if req.Notes != nil {
		bookmark.Notes = req.Notes
	}

	bookmark.UpdatedAt = time.Now()

	// Save to database
	// Implementation would update bookmark in database

	return bookmark, nil
}

// DeleteBookmark deletes a bookmark
func (c *CollaborationService) DeleteBookmark(bookmarkID int64, userID int64) error {
	bookmark, err := c.GetBookmarkInfo(bookmarkID)
	if err != nil {
		return err
	}

	// Check ownership
	if bookmark.UserID != userID {
		return fmt.Errorf("you can only delete your own bookmarks")
	}

	// Delete from database
	// Implementation would delete bookmark from database

	// Log activity
	c.logActivity(bookmark.URLID, &userID, models.ActivityUnbookmarked, "Bookmark removed", map[string]interface{}{
		"bookmark_id": bookmarkID,
	})

	return nil
}

// Collections

// CreateCollection creates a new URL collection
func (c *CollaborationService) CreateCollection(req *models.CreateCollectionRequest, userID int64) (*models.URLCollectionInfo, error) {
	// Create collection
	collection := &models.URLCollection{
		ID:          time.Now().Unix(), // Temporary ID generation
		UserID:      userID,
		TeamID:      req.TeamID,
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		IsPublic:    req.IsPublic,
		SortOrder:   0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to database
	// Implementation would save collection to database

	// Log activity (use a sample URL ID for logging)
	c.logActivity(0, &userID, models.ActivityCollectionCreated, "Collection created", map[string]interface{}{
		"collection_id":   collection.ID,
		"collection_name": collection.Name,
	})

	return c.GetCollectionInfo(collection.ID, userID)
}

// GetUserCollections retrieves all collections for a user
func (c *CollaborationService) GetUserCollections(userID int64) ([]*models.URLCollectionInfo, error) {
	// Implementation would query database
	// For now, return simulated collections
	user, _ := c.userService.GetUserByID(userID)
	
	collections := []*models.URLCollectionInfo{
		{
			URLCollection: &models.URLCollection{
				ID:          1,
				UserID:      userID,
				Name:        "Work Links",
				Description: utils.StringPtr("Important work-related URLs"),
				Color:       models.ColorBlue,
				Icon:        models.IconFolder,
				IsPublic:    false,
				CreatedAt:   time.Now().Add(-7 * 24 * time.Hour),
				UpdatedAt:   time.Now().Add(-1 * time.Hour),
			},
			URLCount:  5,
			Owner:     user.ToPublic(),
			CanEdit:   true,
			CanDelete: true,
		},
	}

	return collections, nil
}

// GetCollectionInfo retrieves detailed collection information
func (c *CollaborationService) GetCollectionInfo(collectionID int64, requestorID int64) (*models.URLCollectionInfo, error) {
	// Implementation would query database
	// For now, return simulated collection
	user, _ := c.userService.GetUserByID(requestorID)
	
	return &models.URLCollectionInfo{
		URLCollection: &models.URLCollection{
			ID:          collectionID,
			UserID:      requestorID,
			Name:        "My Collection",
			Description: utils.StringPtr("A sample collection"),
			Color:       models.ColorBlue,
			Icon:        models.IconFolder,
			IsPublic:    false,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		},
		URLCount:  0,
		URLs:      []*models.URL{},
		Owner:     user.ToPublic(),
		CanEdit:   true,
		CanDelete: true,
	}, nil
}

// AddToCollection adds URLs to a collection
func (c *CollaborationService) AddToCollection(collectionID int64, req *models.AddToCollectionRequest, userID int64) error {
	// Get collection and verify permissions
	collection, err := c.GetCollectionInfo(collectionID, userID)
	if err != nil {
		return err
	}

	if !collection.CanEdit {
		return fmt.Errorf("you don't have permission to edit this collection")
	}

	// Add URLs to collection
	for _, urlID := range req.URLIDs {
		// Check if URL is already in collection
		exists, _ := c.isURLInCollection(collectionID, urlID)
		if exists {
			continue // Skip if already in collection
		}

		// Create collection item
		item := &models.URLCollectionItem{
			ID:           time.Now().Unix() + urlID, // Avoid ID collision
			CollectionID: collectionID,
			URLID:        urlID,
			SortOrder:    0,
			AddedBy:      userID,
			CreatedAt:    time.Now(),
		}

		// Save to database
		// Implementation would save collection item to database
		_ = item

		// Log activity
		c.logActivity(urlID, &userID, models.ActivityAddedToCollection, "URL added to collection", map[string]interface{}{
			"collection_id":   collectionID,
			"collection_name": collection.Name,
		})
	}

	return nil
}

// RemoveFromCollection removes URLs from a collection
func (c *CollaborationService) RemoveFromCollection(collectionID int64, urlID int64, userID int64) error {
	// Get collection and verify permissions
	collection, err := c.GetCollectionInfo(collectionID, userID)
	if err != nil {
		return err
	}

	if !collection.CanEdit {
		return fmt.Errorf("you don't have permission to edit this collection")
	}

	// Remove URL from collection
	// Implementation would delete collection item from database

	// Log activity
	c.logActivity(urlID, &userID, models.ActivityRemovedFromCollection, "URL removed from collection", map[string]interface{}{
		"collection_id":   collectionID,
		"collection_name": collection.Name,
	})

	return nil
}

// Activity and Statistics

// GetURLActivity retrieves activity log for a URL
func (c *CollaborationService) GetURLActivity(urlID int64, userID int64, limit int) ([]*models.URLActivityInfo, error) {
	// Check if user can view URL activity
	canView, err := c.canUserViewURL(urlID, userID)
	if err != nil {
		return nil, err
	}
	if !canView {
		return nil, fmt.Errorf("you don't have permission to view activity for this URL")
	}

	// Implementation would query database
	// For now, return simulated activity
	user, _ := c.userService.GetUserByID(userID)
	
	activities := []*models.URLActivityInfo{
		{
			URLActivity: &models.URLActivity{
				ID:           1,
				URLID:        urlID,
				UserID:       &userID,
				ActivityType: models.ActivityURLCreated,
				Description:  "URL created",
				CreatedAt:    time.Now().Add(-2 * time.Hour),
			},
			User: user.ToPublic(),
		},
	}

	return activities, nil
}

// GetCollaborationSummary retrieves collaboration summary for a URL
func (c *CollaborationService) GetCollaborationSummary(urlID int64, userID int64) (*models.URLCollaborationSummary, error) {
	// Check if user can view URL
	canView, err := c.canUserViewURL(urlID, userID)
	if err != nil {
		return nil, err
	}
	if !canView {
		return nil, fmt.Errorf("you don't have permission to view this URL")
	}

	// Get recent activity
	activities, _ := c.GetURLActivity(urlID, userID, 5)

	// Check user's permissions
	permissions := []string{"view"}
	isSharedWithMe := false
	
	// Get user's share if exists
	share, _ := c.getUserShareForURL(urlID, userID)
	if share != nil {
		permissions = models.GetSharePermissions(share.ShareType)
		isSharedWithMe = true
	}

	summary := &models.URLCollaborationSummary{
		URLID:          urlID,
		ShareCount:     3,
		CommentCount:   5,
		BookmarkCount:  12,
		FavoriteCount:  8,
		NoteCount:      2,
		RecentActivity: activities,
		IsSharedWithMe: isSharedWithMe,
		MyPermissions:  permissions,
	}

	return summary, nil
}

// Helper Methods

func (c *CollaborationService) validateShareRequest(req *models.ShareURLRequest) error {
	if req.SharedWithType == models.SharedWithUser && req.SharedWithID == nil {
		return fmt.Errorf("shared_with_id is required when sharing with a user")
	}
	if req.SharedWithType == models.SharedWithTeam && req.SharedWithID == nil {
		return fmt.Errorf("shared_with_id is required when sharing with a team")
	}
	return nil
}

func (c *CollaborationService) canUserShareURL(urlID int64, userID int64) (bool, error) {
	// Implementation would check if user owns URL or has share permissions
	// For now, assume user can share if they can view
	return c.canUserViewURL(urlID, userID)
}

func (c *CollaborationService) canUserViewURL(urlID int64, userID int64) (bool, error) {
	// Implementation would check URL ownership and sharing
	// For now, assume user can always view (in real implementation, check ownership and shares)
	return true, nil
}

func (c *CollaborationService) canUserCommentOnURL(urlID int64, userID int64) (bool, error) {
	// Implementation would check if user has comment permissions
	// For now, assume user can comment if they can view
	return c.canUserViewURL(urlID, userID)
}

func (c *CollaborationService) getUserShareForURL(urlID int64, userID int64) (*models.URLShare, error) {
	// Implementation would query database for user's share
	// For now, return nil (no share)
	return nil, fmt.Errorf("no share found")
}

func (c *CollaborationService) getUserBookmark(urlID int64, userID int64) (*models.URLBookmark, error) {
	// Implementation would query database for existing bookmark
	// For now, return nil (no existing bookmark)
	return nil, fmt.Errorf("no bookmark found")
}

func (c *CollaborationService) isURLInCollection(collectionID int64, urlID int64) (bool, error) {
	// Implementation would check if URL is already in collection
	// For now, assume URL is not in collection
	return false, nil
}

func (c *CollaborationService) logActivity(urlID int64, userID *int64, activityType, description string, metadata map[string]interface{}) {
	// Implementation would save activity to database
	activity := &models.URLActivity{
		ID:           time.Now().Unix(),
		URLID:        urlID,
		UserID:       userID,
		ActivityType: activityType,
		Description:  description,
		Metadata:     metadata,
		CreatedAt:    time.Now(),
	}

	// Save to database
	_ = activity
}

func (c *CollaborationService) sendShareNotification(share *models.URLShare, message *string) {
	if c.emailService == nil || share.SharedWithID == nil {
		return
	}

	// Get shared user
	sharedUser, err := c.userService.GetUserByID(*share.SharedWithID)
	if err != nil {
		return
	}

	// Get sharer
	sharer, err := c.userService.GetUserByID(share.SharerID)
	if err != nil {
		return
	}

	// Send notification email
	// Implementation would send actual email
	fmt.Printf("Share notification: %s shared a URL with %s\n", sharer.Name, sharedUser.Name)
}

// Helper functions

