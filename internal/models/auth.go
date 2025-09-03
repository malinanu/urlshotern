package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID              int64      `json:"id" db:"id"`
	Name            string     `json:"name" db:"name"`
	Email           string     `json:"email" db:"email"`
	PasswordHash    *string    `json:"-" db:"password_hash"` // Hidden from JSON
	Phone           *string    `json:"phone" db:"phone"`
	PhoneVerified   bool       `json:"phone_verified" db:"phone_verified"`
	EmailVerified   bool       `json:"email_verified" db:"email_verified"`
	Provider        string     `json:"provider" db:"provider"`
	ProviderID      *string    `json:"provider_id" db:"provider_id"`
	AvatarURL       *string    `json:"avatar_url" db:"avatar_url"`
	AccountType     string     `json:"account_type" db:"account_type"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	IsAdmin         bool       `json:"is_admin" db:"is_admin"`
	LastLoginAt     *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// PublicUser represents user data safe for public consumption
type PublicUser struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	Phone         *string    `json:"phone"`
	PhoneVerified bool       `json:"phone_verified"`
	EmailVerified bool       `json:"email_verified"`
	Provider      string     `json:"provider"`
	AvatarURL     *string    `json:"avatar_url"`
	AccountType   string     `json:"account_type"`
	IsActive      bool       `json:"is_active"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ToPublic converts User to PublicUser
func (u *User) ToPublic() *PublicUser {
	return &PublicUser{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		Phone:         u.Phone,
		PhoneVerified: u.PhoneVerified,
		EmailVerified: u.EmailVerified,
		Provider:      u.Provider,
		AvatarURL:     u.AvatarURL,
		AccountType:   u.AccountType,
		IsActive:      u.IsActive,
		LastLoginAt:   u.LastLoginAt,
		CreatedAt:     u.CreatedAt,
	}
}

// UserSession represents a user session
type UserSession struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	UserID           int64      `json:"user_id" db:"user_id"`
	DeviceInfo       *string    `json:"device_info" db:"device_info"` // JSON string
	IPAddress        *string    `json:"ip_address" db:"ip_address"`
	UserAgent        *string    `json:"user_agent" db:"user_agent"`
	RefreshTokenHash *string    `json:"-" db:"refresh_token_hash"`
	ExpiresAt        time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// EmailVerification represents email verification tokens
type EmailVerification struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	UserID     int64      `json:"user_id" db:"user_id"`
	Token      string     `json:"token" db:"token"`
	ExpiresAt  time.Time  `json:"expires_at" db:"expires_at"`
	VerifiedAt *time.Time `json:"verified_at" db:"verified_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// PhoneVerification represents phone verification (OTP)
type PhoneVerification struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	UserID     int64      `json:"user_id" db:"user_id"`
	Phone      string     `json:"phone" db:"phone"`
	OTPCode    string     `json:"-" db:"otp_code"` // Hidden from JSON
	Attempts   int        `json:"attempts" db:"attempts"`
	ExpiresAt  time.Time  `json:"expires_at" db:"expires_at"`
	VerifiedAt *time.Time `json:"verified_at" db:"verified_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// PasswordReset represents password reset tokens
type PasswordReset struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    int64      `json:"user_id" db:"user_id"`
	Token     string     `json:"token" db:"token"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	UsedAt    *time.Time `json:"used_at" db:"used_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// UserPreferences represents user preferences
type UserPreferences struct {
	UserID             int64  `json:"user_id" db:"user_id"`
	DefaultExpiration  *int   `json:"default_expiration" db:"default_expiration"` // days
	AnalyticsPublic    bool   `json:"analytics_public" db:"analytics_public"`
	EmailNotifications bool   `json:"email_notifications" db:"email_notifications"`
	MarketingEmails    bool   `json:"marketing_emails" db:"marketing_emails"`
	Timezone           string `json:"timezone" db:"timezone"`
	Theme              string `json:"theme" db:"theme"`
}

// APIKey represents API keys for programmatic access
type APIKey struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       int64      `json:"user_id" db:"user_id"`
	Name         string     `json:"name" db:"name"`
	KeyHash      string     `json:"-" db:"key_hash"` // Hidden from JSON
	LastUsedAt   *time.Time `json:"last_used_at" db:"last_used_at"`
	ExpiresAt    *time.Time `json:"expires_at" db:"expires_at"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	Permissions  string     `json:"permissions" db:"permissions"` // JSON string
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// AuditLog represents audit log entries
type AuditLog struct {
	ID           int64      `json:"id" db:"id"`
	UserID       *int64     `json:"user_id" db:"user_id"`
	Action       string     `json:"action" db:"action"`
	ResourceType *string    `json:"resource_type" db:"resource_type"`
	ResourceID   *string    `json:"resource_id" db:"resource_id"`
	Details      *string    `json:"details" db:"details"` // JSON string
	IPAddress    *string    `json:"ip_address" db:"ip_address"`
	UserAgent    *string    `json:"user_agent" db:"user_agent"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// Auth Request/Response Models

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Name            string `json:"name" validate:"required,min=2,max=100"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	Phone           string `json:"phone" validate:"required,min=10,max=20"`
	TermsAccepted   bool   `json:"terms_accepted" validate:"required,eq=true"`
	MarketingConsent bool  `json:"marketing_consent"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	RememberMe bool `json:"remember_me"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User         *PublicUser `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"` // seconds
}

// OTPRequest represents OTP verification request
type OTPRequest struct {
	UserID int64  `json:"user_id" validate:"required"`
	OTP    string `json:"otp" validate:"required,len=6"`
}

// ResendOTPRequest represents resend OTP request
type ResendOTPRequest struct {
	UserID int64 `json:"user_id" validate:"required"`
}

// ForgotPasswordRequest represents forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents reset password request
type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

// ChangePasswordRequest represents change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Phone *string `json:"phone,omitempty" validate:"omitempty,min=10,max=20"`
}

// GoogleAuthRequest represents Google OAuth request
type GoogleAuthRequest struct {
	IDToken string `json:"id_token" validate:"required"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}