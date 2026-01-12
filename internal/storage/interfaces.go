package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
)

// PostgresStorageInterface defines the interface for PostgreSQL storage operations
type PostgresStorageInterface interface {
	Close() error
	SaveURLMapping(mapping *models.URLMapping) error
	GetURLMappingByShortCode(shortCode string) (*models.URLMapping, error)
	ShortCodeExists(shortCode string) (bool, error)
	IncrementClickCount(shortCode string) error
	SaveClickEvent(event *models.ClickEvent) error
	GetAnalytics(shortCode string, days int) (*models.AnalyticsResponse, error)

	// Database access methods
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Begin() (*sql.Tx, error)

	// User management methods
	CreateUser(user *models.User) error
	GetUserByID(id int64) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByPhone(phone string) (*models.User, error)
	UpdateUserLastLogin(userID int64, loginTime time.Time) error
	UpdateUserPassword(userID int64, passwordHash string) error
	UpdateUser(user *models.User) error
	UpdateUserProfile(userID int64, updates map[string]interface{}) error
	UpdateUserStatus(userID int64, isActive bool) error

	// Session management
	CreateSession(session *models.UserSession) error
	GetSessionByID(id string) (*models.UserSession, error)
	DeleteUserSessions(userID int64) error

	// Email verification
	CreateEmailVerification(verification *models.EmailVerification) error
	GetEmailVerificationByToken(token string) (*models.EmailVerification, error)
	UpdateEmailVerification(id string, verifiedAt time.Time) error
	UpdateUserEmailVerified(userID int64, verified bool) error

	// Phone verification
	CreatePhoneVerification(verification *models.PhoneVerification) error
	DeletePhoneVerificationByUserID(userID int64) error
	GetPhoneVerificationByUserID(userID int64) (*models.PhoneVerification, error)
	UpdatePhoneVerificationAttempts(userID int64, attempts int) error
	UpdatePhoneVerification(id string, verifiedAt time.Time) error
	UpdateUserPhoneVerified(userID int64, verified bool) error

	// Password reset
	CreatePasswordReset(reset *models.PasswordReset) error
	GetPasswordResetByToken(token string) (*models.PasswordReset, error)
	UpdatePasswordReset(id string, usedAt time.Time) error

	// User preferences
	CreateUserPreferences(prefs *models.UserPreferences) error
	GetUserPreferences(userID int64) (*models.UserPreferences, error)
	UpdateUserPreferences(prefs *models.UserPreferences) error

	// Pool statistics
	GetPoolStats() map[string]interface{}
}

// RedisStorageInterface defines the interface for Redis storage operations
type RedisStorageInterface interface {
	Close() error
	SetURLMapping(shortCode string, mapping *models.URLMapping, ttl time.Duration) error
	GetURLMapping(shortCode string) (*models.URLMapping, error)
	DeleteURLMapping(shortCode string) error
	IncrementClickCount(shortCode string) (int64, error)
}
