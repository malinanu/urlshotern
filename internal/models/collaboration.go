package models

import (
	"time"
)

// URLShare represents a URL sharing relationship
type URLShare struct {
	ID             int64      `json:"id" db:"id"`
	URLID          int64      `json:"url_id" db:"url_id"`
	SharerID       int64      `json:"sharer_id" db:"sharer_id"`
	SharedWithType string     `json:"shared_with_type" db:"shared_with_type"` // user, team, public
	SharedWithID   *int64     `json:"shared_with_id,omitempty" db:"shared_with_id"`
	ShareType      string     `json:"share_type" db:"share_type"`       // view, edit, admin
	ExpiresAt      *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// URLShareInfo represents detailed information about a URL share
type URLShareInfo struct {
	*URLShare
	URL         *URL        `json:"url"`
	Sharer      *PublicUser `json:"sharer"`
	SharedWith  interface{} `json:"shared_with"` // Can be PublicUser or Team
	Permissions []string    `json:"permissions"`
}

// URLComment represents a comment on a URL
type URLComment struct {
	ID         int64      `json:"id" db:"id"`
	URLID      int64      `json:"url_id" db:"url_id"`
	UserID     int64      `json:"user_id" db:"user_id"`
	Content    string     `json:"content" db:"content"`
	ParentID   *int64     `json:"parent_id,omitempty" db:"parent_id"` // For threaded comments
	IsEdited   bool       `json:"is_edited" db:"is_edited"`
	IsDeleted  bool       `json:"is_deleted" db:"is_deleted"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// URLCommentInfo represents detailed comment information
type URLCommentInfo struct {
	*URLComment
	User        *PublicUser        `json:"user"`
	Replies     []*URLCommentInfo  `json:"replies,omitempty"`
	ReplyCount  int                `json:"reply_count"`
	CanEdit     bool               `json:"can_edit"`
	CanDelete   bool               `json:"can_delete"`
}

// URLActivity represents activity/event logs for a URL
type URLActivity struct {
	ID           int64                  `json:"id" db:"id"`
	URLID        int64                  `json:"url_id" db:"url_id"`
	UserID       *int64                 `json:"user_id,omitempty" db:"user_id"`
	ActivityType string                 `json:"activity_type" db:"activity_type"`
	Description  string                 `json:"description" db:"description"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	IPAddress    *string                `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string                `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// URLActivityInfo represents detailed activity information
type URLActivityInfo struct {
	*URLActivity
	User *PublicUser `json:"user,omitempty"`
}

// URLBookmark represents a user's bookmark of a URL
type URLBookmark struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	URLID     int64     `json:"url_id" db:"url_id"`
	Tags      string    `json:"tags" db:"tags"` // Comma-separated tags
	Notes     *string   `json:"notes,omitempty" db:"notes"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// URLBookmarkInfo represents detailed bookmark information
type URLBookmarkInfo struct {
	*URLBookmark
	URL     *URL      `json:"url"`
	TagList []string  `json:"tag_list"`
}

// URLFavorite represents a user's favorite URLs
type URLFavorite struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	URLID     int64     `json:"url_id" db:"url_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// URLCollection represents a collection/folder of URLs
type URLCollection struct {
	ID          int64      `json:"id" db:"id"`
	UserID      int64      `json:"user_id" db:"user_id"`
	TeamID      *int64     `json:"team_id,omitempty" db:"team_id"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description,omitempty" db:"description"`
	Color       string     `json:"color" db:"color"`
	Icon        string     `json:"icon" db:"icon"`
	IsPublic    bool       `json:"is_public" db:"is_public"`
	SortOrder   int        `json:"sort_order" db:"sort_order"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// URLCollectionItem represents a URL within a collection
type URLCollectionItem struct {
	ID           int64     `json:"id" db:"id"`
	CollectionID int64     `json:"collection_id" db:"collection_id"`
	URLID        int64     `json:"url_id" db:"url_id"`
	SortOrder    int       `json:"sort_order" db:"sort_order"`
	AddedBy      int64     `json:"added_by" db:"added_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// URLCollectionInfo represents detailed collection information
type URLCollectionInfo struct {
	*URLCollection
	URLCount    int     `json:"url_count"`
	URLs        []*URL  `json:"urls,omitempty"`
	Owner       *PublicUser `json:"owner"`
	Team        *Team   `json:"team,omitempty"`
	CanEdit     bool    `json:"can_edit"`
	CanDelete   bool    `json:"can_delete"`
}

// URLNote represents a private note on a URL
type URLNote struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	URLID     int64     `json:"url_id" db:"url_id"`
	Title     *string   `json:"title,omitempty" db:"title"`
	Content   string    `json:"content" db:"content"`
	IsPrivate bool      `json:"is_private" db:"is_private"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Request/Response Models

// ShareURLRequest represents a request to share a URL
type ShareURLRequest struct {
	SharedWithType string     `json:"shared_with_type" validate:"required,oneof=user team public"`
	SharedWithID   *int64     `json:"shared_with_id,omitempty"`
	ShareType      string     `json:"share_type" validate:"required,oneof=view edit admin"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	Message        *string    `json:"message,omitempty"`
}

// UpdateShareRequest represents a request to update sharing permissions
type UpdateShareRequest struct {
	ShareType string     `json:"share_type,omitempty" validate:"omitempty,oneof=view edit admin"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  *bool      `json:"is_active,omitempty"`
}

// CreateCommentRequest represents a request to create a comment
type CreateCommentRequest struct {
	Content  string `json:"content" validate:"required,min=1,max=1000"`
	ParentID *int64 `json:"parent_id,omitempty"`
}

// UpdateCommentRequest represents a request to update a comment
type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

// CreateBookmarkRequest represents a request to bookmark a URL
type CreateBookmarkRequest struct {
	Tags  string  `json:"tags"`
	Notes *string `json:"notes,omitempty"`
}

// UpdateBookmarkRequest represents a request to update a bookmark
type UpdateBookmarkRequest struct {
	Tags  *string `json:"tags,omitempty"`
	Notes *string `json:"notes,omitempty"`
}

// CreateCollectionRequest represents a request to create a collection
type CreateCollectionRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Color       string  `json:"color" validate:"required,hexcolor"`
	Icon        string  `json:"icon" validate:"required,min=1,max=50"`
	IsPublic    bool    `json:"is_public"`
	TeamID      *int64  `json:"team_id,omitempty"`
}

// UpdateCollectionRequest represents a request to update a collection
type UpdateCollectionRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Color       *string `json:"color,omitempty" validate:"omitempty,hexcolor"`
	Icon        *string `json:"icon,omitempty" validate:"omitempty,min=1,max=50"`
	IsPublic    *bool   `json:"is_public,omitempty"`
}

// AddToCollectionRequest represents a request to add URLs to a collection
type AddToCollectionRequest struct {
	URLIDs []int64 `json:"url_ids" validate:"required,min=1,max=100"`
}

// CreateNoteRequest represents a request to create a note
type CreateNoteRequest struct {
	Title     *string `json:"title,omitempty" validate:"omitempty,max=200"`
	Content   string  `json:"content" validate:"required,min=1,max=5000"`
	IsPrivate bool    `json:"is_private"`
}

// UpdateNoteRequest represents a request to update a note
type UpdateNoteRequest struct {
	Title     *string `json:"title,omitempty" validate:"omitempty,max=200"`
	Content   *string `json:"content,omitempty" validate:"omitempty,min=1,max=5000"`
	IsPrivate *bool   `json:"is_private,omitempty"`
}

// Collaboration Summary Models

// URLCollaborationSummary represents collaboration summary for a URL
type URLCollaborationSummary struct {
	URLID          int64              `json:"url_id"`
	ShareCount     int                `json:"share_count"`
	CommentCount   int                `json:"comment_count"`
	BookmarkCount  int                `json:"bookmark_count"`
	FavoriteCount  int                `json:"favorite_count"`
	NoteCount      int                `json:"note_count"`
	RecentActivity []*URLActivityInfo `json:"recent_activity"`
	TopContributors []*PublicUser     `json:"top_contributors"`
	IsSharedWithMe bool              `json:"is_shared_with_me"`
	MyPermissions  []string          `json:"my_permissions"`
}

// UserCollaborationStats represents user collaboration statistics
type UserCollaborationStats struct {
	UserID             int64 `json:"user_id"`
	URLsShared         int   `json:"urls_shared"`
	URLsSharedWithMe   int   `json:"urls_shared_with_me"`
	CommentsCreated    int   `json:"comments_created"`
	BookmarksCreated   int   `json:"bookmarks_created"`
	CollectionsCreated int   `json:"collections_created"`
	NotesCreated       int   `json:"notes_created"`
	FavoritesCreated   int   `json:"favorites_created"`
}

// Constants for collaboration

const (
	// Share types
	ShareTypeView  = "view"
	ShareTypeEdit  = "edit"
	ShareTypeAdmin = "admin"

	// Shared with types
	SharedWithUser   = "user"
	SharedWithTeam   = "team"
	SharedWithPublic = "public"

	// Activity types
	ActivityURLCreated    = "url_created"
	ActivityURLUpdated    = "url_updated"
	ActivityURLDeleted    = "url_deleted"
	ActivityURLShared     = "url_shared"
	ActivityShareRevoked  = "share_revoked"
	ActivityCommentAdded  = "comment_added"
	ActivityCommentEdited = "comment_edited"
	ActivityCommentDeleted = "comment_deleted"
	ActivityBookmarked    = "bookmarked"
	ActivityUnbookmarked  = "unbookmarked"
	ActivityFavorited     = "favorited"
	ActivityUnfavorited   = "unfavorited"
	ActivityNoteAdded     = "note_added"
	ActivityNoteUpdated   = "note_updated"
	ActivityNoteDeleted   = "note_deleted"
	ActivityCollectionCreated = "collection_created"
	ActivityCollectionUpdated = "collection_updated"
	ActivityCollectionDeleted = "collection_deleted"
	ActivityAddedToCollection = "added_to_collection"
	ActivityRemovedFromCollection = "removed_from_collection"

	// Default colors for collections
	ColorRed     = "#ef4444"
	ColorOrange  = "#f97316"
	ColorYellow  = "#eab308"
	ColorGreen   = "#22c55e"
	ColorBlue    = "#3b82f6"
	ColorIndigo  = "#6366f1"
	ColorPurple  = "#a855f7"
	ColorPink    = "#ec4899"
	ColorGray    = "#6b7280"

	// Default icons for collections
	IconFolder     = "folder"
	IconStar       = "star"
	IconHeart      = "heart"
	IconBookmark   = "bookmark"
	IconTag        = "tag"
	IconDocument   = "document"
	IconLink       = "link"
	IconGlobe      = "globe"
	IconUsers      = "users"
	IconLightbulb  = "lightbulb"
)

// Permission helpers

// GetSharePermissions returns the list of permissions for a share type
func GetSharePermissions(shareType string) []string {
	switch shareType {
	case ShareTypeView:
		return []string{"view"}
	case ShareTypeEdit:
		return []string{"view", "edit", "comment"}
	case ShareTypeAdmin:
		return []string{"view", "edit", "comment", "share", "delete"}
	default:
		return []string{}
	}
}

// CanPerformAction checks if a share type allows a specific action
func (s *URLShare) CanPerformAction(action string) bool {
	permissions := GetSharePermissions(s.ShareType)
	for _, permission := range permissions {
		if permission == action {
			return true
		}
	}
	return false
}