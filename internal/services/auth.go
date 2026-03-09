package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/URLshorter/url-shortener/internal/utils"
	"github.com/google/uuid"
)

type AuthService struct {
	userService  *UserService
	jwtService   *JWTService
	smsService   *SMSService
	emailService *EmailService
	db           *storage.PostgresStorage
	redis        *storage.RedisStorage
	config       *configs.Config
	googleConfig *oauth2.Config
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userService *UserService,
	jwtService *JWTService,
	smsService *SMSService,
	emailService *EmailService,
	db *storage.PostgresStorage,
	redis *storage.RedisStorage,
	config *configs.Config,
) *AuthService {

	// Setup Google OAuth config
	googleConfig := &oauth2.Config{
		ClientID:     config.GoogleClientID,
		ClientSecret: config.GoogleClientSecret,
		RedirectURL:  config.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	return &AuthService{
		userService:  userService,
		jwtService:   jwtService,
		smsService:   smsService,
		emailService: emailService,
		db:           db,
		redis:        redis,
		config:       config,
		googleConfig: googleConfig,
	}
}

// Register creates a new user account
func (a *AuthService) Register(req *models.RegisterRequest, ipAddress, userAgent string) (*UserRegistrationData, error) {
	// Validate input
	if err := a.validateRegistrationRequest(req); err != nil {
		return nil, err
	}

	// Use UserService to create user
	registrationData, err := a.userService.CreateUser(req)
	if err != nil {
		return nil, err
	}

	// Send email verification
	if a.emailService != nil {
		err = a.emailService.SendEmailVerification(registrationData.User.Email, registrationData.EmailToken)
		if err != nil {
			fmt.Printf("Failed to send email verification: %v\n", err)
			// Don't fail registration if email fails
		}
	}

	// Log registration event
	a.logAuditEvent(&registrationData.User.ID, "user_registered", "user", fmt.Sprintf("%d", registrationData.User.ID),
		map[string]interface{}{
			"email":    registrationData.User.Email,
			"provider": "email",
		}, ipAddress, userAgent)

	return registrationData, nil
}

// Login authenticates a user
func (a *AuthService) Login(req *models.LoginRequest, ipAddress, userAgent string) (*models.AuthResponse, error) {
	// Use UserService to authenticate
	user, sessionID, err := a.userService.AuthenticateUser(req)
	if err != nil {
		// Log failed login attempt
		a.logAuditEvent(nil, "login_failed", "authentication", "",
			map[string]interface{}{
				"email":  req.Email,
				"reason": err.Error(),
			}, ipAddress, userAgent)
		return nil, err
	}

	// Generate JWT tokens
	tokenPair, err := a.jwtService.GenerateTokenPair(user, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Log successful login
	a.logAuditEvent(&user.ID, "login_success", "authentication", fmt.Sprintf("%d", user.ID),
		map[string]interface{}{
			"email":      user.Email,
			"session_id": sessionID,
		}, ipAddress, userAgent)

	return &models.AuthResponse{
		User:         user.ToPublic(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int64(tokenPair.ExpiresAt.Sub(time.Now()).Seconds()),
	}, nil
}

// RefreshTokens generates new tokens using refresh token
func (a *AuthService) RefreshTokens(req *models.RefreshTokenRequest, ipAddress, userAgent string) (*models.AuthResponse, error) {
	// Validate and extract claims from refresh token
	claims, err := a.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Verify this is a refresh token
	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type for refresh")
	}

	// Get user
	user, err := a.userService.GetUserByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, fmt.Errorf("account is not active")
	}

	// Use JWT service to refresh tokens with existing session ID
	tokenPair, err := a.jwtService.RefreshToken(req.RefreshToken, user, claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Log token refresh
	a.logAuditEvent(&user.ID, "token_refreshed", "authentication", fmt.Sprintf("%d", user.ID),
		map[string]interface{}{
			"email":      user.Email,
			"session_id": claims.SessionID,
		}, ipAddress, userAgent)

	return &models.AuthResponse{
		User:         user.ToPublic(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int64(tokenPair.ExpiresAt.Sub(time.Now()).Seconds()),
	}, nil
}

// VerifyEmail verifies user email with token
func (a *AuthService) VerifyEmail(token string, ipAddress, userAgent string) (*models.User, error) {
	user, err := a.userService.VerifyEmail(token)
	if err != nil {
		return nil, err
	}

	// Log email verification
	a.logAuditEvent(&user.ID, "email_verified", "user", fmt.Sprintf("%d", user.ID),
		map[string]interface{}{
			"email": user.Email,
		}, ipAddress, userAgent)

	return user, nil
}

// SendOTP generates and sends OTP to user
func (a *AuthService) SendOTP(req *models.ResendOTPRequest, ipAddress, userAgent string) error {
	err := a.userService.GenerateAndSendOTP(req.UserID)
	if err != nil {
		return err
	}

	// Log OTP sent
	a.logAuditEvent(&req.UserID, "otp_sent", "verification", fmt.Sprintf("%d", req.UserID),
		map[string]interface{}{
			"type": "phone",
		}, ipAddress, userAgent)

	return nil
}

// VerifyOTP verifies the OTP code
func (a *AuthService) VerifyOTP(req *models.OTPRequest, ipAddress, userAgent string) error {
	err := a.userService.VerifyOTP(req.UserID, req.OTP)
	if err != nil {
		// Log failed OTP verification
		a.logAuditEvent(&req.UserID, "otp_verification_failed", "verification", fmt.Sprintf("%d", req.UserID),
			map[string]interface{}{
				"reason": err.Error(),
			}, ipAddress, userAgent)
		return err
	}

	// Log successful OTP verification
	a.logAuditEvent(&req.UserID, "otp_verified", "verification", fmt.Sprintf("%d", req.UserID),
		map[string]interface{}{
			"type": "phone",
		}, ipAddress, userAgent)

	return nil
}

// RequestPasswordReset initiates password reset process
func (a *AuthService) RequestPasswordReset(email, ipAddress, userAgent string) error {
	token, err := a.createPasswordResetToken(email)
	if err != nil {
		// Don't reveal if email exists or not - log error but return nil
		log.Printf("ERROR [AUTH] Failed to create password reset token for email: %v", err)
		return nil
	}

	// Send password reset email
	if a.emailService != nil {
		err = a.emailService.SendPasswordReset(email, token)
		if err != nil {
			// Log the error with details for debugging
			log.Printf("ERROR [AUTH] Failed to send password reset email to %s: %v", email, err)
			// Don't reveal if email exists - return nil for security
			// But audit the failure for monitoring
			a.logAuditEvent(nil, "password_reset_email_failed", "authentication", "",
				map[string]interface{}{
					"email": email,
					"error": err.Error(),
				}, ipAddress, userAgent)
		} else {
			log.Printf("INFO [AUTH] Password reset email sent successfully to %s", email)
		}
	} else {
		log.Printf("WARN [AUTH] Email service not available, cannot send password reset email")
	}

	// Log password reset request (don't log user ID to prevent enumeration)
	a.logAuditEvent(nil, "password_reset_requested", "authentication", "",
		map[string]interface{}{
			"email": email,
		}, ipAddress, userAgent)

	return nil
}

// ResetPassword resets password using token
func (a *AuthService) ResetPassword(token, newPassword, ipAddress, userAgent string) error {
	err := a.resetPasswordWithToken(token, newPassword)
	if err != nil {
		return err
	}

	// Log password reset
	a.logAuditEvent(nil, "password_reset_completed", "authentication", "",
		map[string]interface{}{
			"token_used": true,
		}, ipAddress, userAgent)

	return nil
}

// UpdateProfile updates user profile information
func (a *AuthService) UpdateProfile(userID int64, req *models.UpdateProfileRequest, ipAddress, userAgent string) (*models.User, error) {
	// Verify user exists
	_, err := a.userService.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Build update fields
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}

	// Update user in database
	err = a.db.UpdateUserProfile(userID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Fetch updated user
	updatedUser, err := a.userService.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated user: %w", err)
	}

	// Log profile update
	a.logAuditEvent(&userID, "profile_updated", "user", fmt.Sprintf("%d", userID),
		map[string]interface{}{
			"fields_updated": updates,
		}, ipAddress, userAgent)

	return updatedUser, nil
}

// ChangePassword changes user password
func (a *AuthService) ChangePassword(userID int64, req *models.ChangePasswordRequest, ipAddress, userAgent string) error {
	err := a.userService.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return err
	}

	// Invalidate all user sessions (force re-login for security)
	err = a.userService.InvalidateUserSessions(userID)
	if err != nil {
		// Log error but don't fail the password change
		fmt.Printf("Warning: failed to invalidate user sessions after password change: %v\n", err)
	}

	// Log password change
	a.logAuditEvent(&userID, "password_changed", "user", fmt.Sprintf("%d", userID),
		map[string]interface{}{
			"method": "self_service",
		}, ipAddress, userAgent)

	return nil
}

// GetUserPreferences retrieves user preferences
func (a *AuthService) GetUserPreferences(userID int64) (*models.UserPreferences, error) {
	return a.userService.GetUserPreferences(userID)
}

// UpdateUserPreferences updates user preferences
func (a *AuthService) UpdateUserPreferences(userID int64, preferences *models.UserPreferences) error {
	return a.userService.UpdateUserPreferences(userID, preferences)
}

// googleUserInfo represents the response from Google's userinfo endpoint
type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// GenerateOAuthState generates a random state for CSRF protection and stores it in Redis
func (a *AuthService) GenerateOAuthState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	state := hex.EncodeToString(bytes)

	// Store state in Redis with 10 minute TTL
	key := fmt.Sprintf("oauth_state:%s", state)
	err := a.redis.Set(key, "1", 10*time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed to store state: %w", err)
	}

	return state, nil
}

// ValidateOAuthState checks if the state exists in Redis and deletes it
func (a *AuthService) ValidateOAuthState(state string) bool {
	key := fmt.Sprintf("oauth_state:%s", state)
	_, err := a.redis.Get(key)
	if err != nil {
		return false
	}
	// Delete state after validation (one-time use)
	_ = a.redis.Del(key)
	return true
}

// GetGoogleAuthURL returns Google OAuth authorization URL
func (a *AuthService) GetGoogleAuthURL(state string) string {
	return a.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// HandleGoogleCallback exchanges the authorization code for tokens, fetches user info,
// and finds or creates the user in the database
func (a *AuthService) HandleGoogleCallback(code string, ipAddress, userAgent string) (*models.AuthResponse, error) {
	// Exchange authorization code for token
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	token, err := a.googleConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Fetch user info from Google
	client := a.googleConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user info response: %w", err)
	}

	var gUser googleUserInfo
	if err := json.Unmarshal(body, &gUser); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	if gUser.Email == "" {
		return nil, fmt.Errorf("no email received from Google")
	}

	// Try to find existing user by email
	user, err := a.db.GetUserByEmail(gUser.Email)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to look up user: %w", err)
	}

	if user != nil {
		// Existing user: update provider info if they were a local user
		if user.Provider == "local" || user.Provider == "email" {
			user.Provider = "google"
			user.ProviderID = &gUser.ID
			if gUser.Picture != "" {
				user.AvatarURL = &gUser.Picture
			}
			user.EmailVerified = true
			if err := a.db.UpdateUser(user); err != nil {
				log.Printf("WARN [AUTH] Failed to update user provider info: %v", err)
			}
		}
	} else {
		// New user: create account
		userID, err := utils.GenerateSnowflakeID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate user ID: %w", err)
		}

		user = &models.User{
			ID:            userID,
			Name:          gUser.Name,
			Email:         gUser.Email,
			PasswordHash:  nil, // No password for OAuth users
			EmailVerified: true,
			Provider:      "google",
			ProviderID:    &gUser.ID,
			AvatarURL:     &gUser.Picture,
			AccountType:   "free",
			IsActive:      true,
			IsAdmin:       false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := a.db.CreateUser(user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		// Create default preferences
		preferences := &models.UserPreferences{
			UserID:             user.ID,
			AnalyticsPublic:    false,
			EmailNotifications: true,
			MarketingEmails:    false,
			Timezone:           "UTC",
			Theme:              "light",
		}
		if err := a.db.CreateUserPreferences(preferences); err != nil {
			log.Printf("WARN [AUTH] Failed to create user preferences: %v", err)
		}

		// Log registration
		a.logAuditEvent(&user.ID, "user_registered", "user", fmt.Sprintf("%d", user.ID),
			map[string]interface{}{
				"email":    user.Email,
				"provider": "google",
			}, ipAddress, userAgent)
	}

	// Create session
	sessionID := uuid.New()
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days for OAuth
	session := &models.UserSession{
		ID:        sessionID,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	if ipAddress != "" {
		session.IPAddress = &ipAddress
	}
	if userAgent != "" {
		session.UserAgent = &userAgent
	}

	if err := a.db.CreateSession(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login
	_ = a.db.UpdateUserLastLogin(user.ID, time.Now())

	// Generate JWT tokens
	tokenPair, err := a.jwtService.GenerateTokenPair(user, sessionID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Log successful login
	a.logAuditEvent(&user.ID, "login_success", "authentication", fmt.Sprintf("%d", user.ID),
		map[string]interface{}{
			"email":      user.Email,
			"provider":   "google",
			"session_id": sessionID.String(),
		}, ipAddress, userAgent)

	return &models.AuthResponse{
		User:         user.ToPublic(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int64(tokenPair.ExpiresAt.Sub(time.Now()).Seconds()),
	}, nil
}

// Logout invalidates user session
func (a *AuthService) Logout(userID int64, sessionID string, ipAddress, userAgent string) error {
	// Invalidate session in database
	// This would be implemented in the UserService or directly here

	// Log logout
	a.logAuditEvent(&userID, "logout", "authentication", fmt.Sprintf("%d", userID),
		map[string]interface{}{
			"session_id": sessionID,
		}, ipAddress, userAgent)

	return nil
}

// InvalidateToken blacklists a token by adding it to Redis with TTL matching token expiry
func (a *AuthService) InvalidateToken(token string) error {
	// Validate token to get expiry time
	claims, err := a.jwtService.ValidateToken(token)
	if err != nil {
		// Token is already invalid or expired, no need to blacklist
		if err == ErrExpiredToken {
			return nil
		}
		return fmt.Errorf("failed to validate token: %w", err)
	}

	// Blacklist token in Redis until its natural expiry
	if claims.ExpiresAt != nil {
		err = a.redis.BlacklistToken(token, claims.ExpiresAt.Time)
		if err != nil {
			return fmt.Errorf("failed to blacklist token: %w", err)
		}
	}

	return nil
}

// Helper methods

func (a *AuthService) validateRegistrationRequest(req *models.RegisterRequest) error {
	if len(req.Name) < 2 || len(req.Name) > 100 {
		return fmt.Errorf("name must be between 2 and 100 characters")
	}

	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	if !req.TermsAccepted {
		return fmt.Errorf("terms and conditions must be accepted")
	}

	// Additional validation can be added here
	return nil
}

func (a *AuthService) logAuditEvent(userID *int64, action, resourceType, resourceID string, details map[string]interface{}, ipAddress, userAgent string) {
	// This would log to the audit_logs table
	// Implementation would depend on your storage layer
	auditLog := &models.AuditLog{
		UserID:       userID,
		Action:       action,
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
		IPAddress:    &ipAddress,
		UserAgent:    &userAgent,
		CreatedAt:    time.Now(),
	}

	// Convert details to JSON string
	if details != nil {
		// Implementation would serialize details to JSON
		_ = auditLog // Use the auditLog
	}

	// Save to database (this would be implemented in the storage layer)
}

// GenerateSecureToken generates a cryptographically secure random token
func (a *AuthService) GenerateSecureToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// createPasswordResetToken creates a password reset token for a user
func (a *AuthService) createPasswordResetToken(email string) (string, error) {
	// Get user by email
	user, err := a.db.GetUserByEmail(email)
	if err != nil {
		// Return nil error to prevent email enumeration
		return "", fmt.Errorf("user not found")
	}

	// Generate secure token
	token := a.GenerateSecureToken()

	// Create password reset record
	resetRecord := &models.PasswordReset{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour expiry
		CreatedAt: time.Now(),
	}

	// Save to database
	err = a.db.CreatePasswordReset(resetRecord)
	if err != nil {
		return "", fmt.Errorf("failed to create password reset token: %w", err)
	}

	return token, nil
}

// resetPasswordWithToken resets a user's password using a reset token
func (a *AuthService) resetPasswordWithToken(token, newPassword string) error {
	// Get password reset by token
	reset, err := a.db.GetPasswordResetByToken(token)
	if err != nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	// Check if token is expired
	if time.Now().After(reset.ExpiresAt) {
		return fmt.Errorf("reset token has expired")
	}

	// Check if token has been used
	if reset.UsedAt != nil {
		return fmt.Errorf("reset token has already been used")
	}

	// Get user
	user, err := a.db.GetUserByID(reset.UserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Hash new password
	hashedPassword, err := a.userService.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	err = a.db.UpdateUserPassword(user.ID, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	now := time.Now()
	err = a.db.MarkPasswordResetAsUsed(reset.ID, now)
	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to mark password reset token as used: %v\n", err)
	}

	return nil
}

// DeleteAccount permanently deletes a user and all associated data.
func (a *AuthService) DeleteAccount(userID int64) error {
	tx, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete analytics rows that reference url_mappings for this user
	analyticsSubquery := `(SELECT id FROM url_mappings WHERE user_id = $1)`
	for _, table := range []string{
		"geographic_analytics",
		"device_analytics",
		"time_analytics",
		"referrer_analytics",
		"click_events",
	} {
		if _, err := tx.Exec(`DELETE FROM `+table+` WHERE link_id IN `+analyticsSubquery, userID); err != nil {
			return fmt.Errorf("failed to delete %s: %w", table, err)
		}
	}

	// Delete AI insights by short_code
	if _, err := tx.Exec(`DELETE FROM ai_insights WHERE short_code IN (SELECT short_code FROM url_mappings WHERE user_id = $1)`, userID); err != nil {
		return fmt.Errorf("failed to delete ai_insights: %w", err)
	}

	// Delete URL mappings
	if _, err := tx.Exec(`DELETE FROM url_mappings WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("failed to delete url_mappings: %w", err)
	}

	// Delete polar subscriptions
	if _, err := tx.Exec(`DELETE FROM polar_subscriptions WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("failed to delete polar_subscriptions: %w", err)
	}

	// Delete custom domains
	if _, err := tx.Exec(`DELETE FROM custom_domains WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("failed to delete custom_domains: %w", err)
	}

	// Delete brand settings
	if _, err := tx.Exec(`DELETE FROM brand_settings WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("failed to delete brand_settings: %w", err)
	}

	// Delete team memberships
	if _, err := tx.Exec(`DELETE FROM team_members WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("failed to delete team_members: %w", err)
	}

	// Delete teams owned by the user
	if _, err := tx.Exec(`DELETE FROM teams WHERE owner_id = $1`, userID); err != nil {
		return fmt.Errorf("failed to delete teams: %w", err)
	}

	// Delete the user row
	if _, err := tx.Exec(`DELETE FROM users WHERE id = $1`, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return tx.Commit()
}
