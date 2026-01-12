package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
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
	a.LogAuditEvent(&registrationData.User.ID, "user_registered", "user", fmt.Sprintf("%d", registrationData.User.ID), 
		map[string]interface{}{
			"email": registrationData.User.Email,
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
		a.LogAuditEvent(nil, "login_failed", "authentication", "", 
			map[string]interface{}{
				"email": req.Email,
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
	a.LogAuditEvent(&user.ID, "login_success", "authentication", fmt.Sprintf("%d", user.ID), 
		map[string]interface{}{
			"email": user.Email,
			"session_id": sessionID,
		}, ipAddress, userAgent)

	return &models.AuthResponse{
		User:         user.ToPublic(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int64(tokenPair.ExpiresAt.Sub(time.Now()).Seconds()),
	}, nil
}

// AdminLogin validates admin credentials and returns tokens
func (a *AuthService) AdminLogin(req *models.LoginRequest, ipAddress, userAgent string) (*models.AuthResponse, error) {
	// Get user first to check admin status BEFORE authenticating
	user, err := a.userService.GetUserByEmail(req.Email)
	if err != nil {
		// Don't reveal whether user exists - return generic error
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check admin status BEFORE authentication (fail fast for non-admin users)
	if !user.IsAdmin && !isAdminAccountType(user.AccountType) {
		a.LogAuditEvent(nil, "admin_login_rejected", "authentication", "",
			map[string]interface{}{
				"email":  req.Email,
				"reason": "not_admin",
			}, ipAddress, userAgent)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Now authenticate (will fetch user again internally, but ensures fresh data and proper password check)
	response, err := a.Login(req, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}

	// Log successful admin login
	a.LogAuditEvent(&user.ID, "admin_login_success", "authentication", fmt.Sprintf("%d", user.ID),
		map[string]interface{}{
			"email":      user.Email,
			"admin_type": user.AccountType,
		}, ipAddress, userAgent)

	return response, nil
}

// isAdminAccountType checks if the account type is an admin type
func isAdminAccountType(accountType string) bool {
	adminTypes := []string{
		"super_admin",
		"system_admin",
		"support_admin",
		"billing_admin",
		"content_admin",
	}
	for _, t := range adminTypes {
		if accountType == t {
			return true
		}
	}
	return false
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
	a.LogAuditEvent(&user.ID, "token_refreshed", "authentication", fmt.Sprintf("%d", user.ID),
		map[string]interface{}{
			"email": user.Email,
			"session_id": claims.SessionID,
		}, ipAddress, userAgent)

	return &models.AuthResponse{
		User:         user.ToPublic(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int64(tokenPair.ExpiresAt.Sub(time.Now()).Seconds()),
	}, nil
}

// RefreshToken is a convenience method that takes just the refresh token string
func (a *AuthService) RefreshToken(refreshToken string) (*models.AuthResponse, error) {
	req := &models.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}
	return a.RefreshTokens(req, "", "")
}

// VerifyEmail verifies user email with token
func (a *AuthService) VerifyEmail(token string, ipAddress, userAgent string) (*models.User, error) {
	user, err := a.userService.VerifyEmail(token)
	if err != nil {
		return nil, err
	}

	// Log email verification
	a.LogAuditEvent(&user.ID, "email_verified", "user", fmt.Sprintf("%d", user.ID),
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
	a.LogAuditEvent(&req.UserID, "otp_sent", "verification", fmt.Sprintf("%d", req.UserID),
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
		a.LogAuditEvent(&req.UserID, "otp_verification_failed", "verification", fmt.Sprintf("%d", req.UserID),
			map[string]interface{}{
				"reason": err.Error(),
			}, ipAddress, userAgent)
		return err
	}

	// Log successful OTP verification
	a.LogAuditEvent(&req.UserID, "otp_verified", "verification", fmt.Sprintf("%d", req.UserID),
		map[string]interface{}{
			"type": "phone",
		}, ipAddress, userAgent)

	return nil
}

// ForgotPassword initiates password reset process
func (a *AuthService) ForgotPassword(req *models.ForgotPasswordRequest, ipAddress, userAgent string) error {
	token, err := a.userService.CreatePasswordResetToken(req.Email)
	if err != nil {
		// Don't reveal if email exists or not
		return nil
	}

	// Send password reset email
	if a.emailService != nil {
		err = a.emailService.SendPasswordReset(req.Email, token)
		if err != nil {
			fmt.Printf("Failed to send password reset email: %v\n", err)
		}
	}

	// Log password reset request (don't log user ID to prevent enumeration)
	a.LogAuditEvent(nil, "password_reset_requested", "authentication", "",
		map[string]interface{}{
			"email": req.Email,
		}, ipAddress, userAgent)

	return nil
}

// ResetPassword resets password using token
func (a *AuthService) ResetPassword(req *models.ResetPasswordRequest, ipAddress, userAgent string) error {
	err := a.userService.ResetPassword(req.Token, req.Password)
	if err != nil {
		return err
	}

	// Log password reset
	a.LogAuditEvent(nil, "password_reset_completed", "authentication", "",
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
	a.LogAuditEvent(&userID, "profile_updated", "user", fmt.Sprintf("%d", userID),
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
	a.LogAuditEvent(&userID, "password_changed", "user", fmt.Sprintf("%d", userID),
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

// GoogleAuth handles Google OAuth authentication
func (a *AuthService) GoogleAuth(req *models.GoogleAuthRequest, ipAddress, userAgent string) (*models.AuthResponse, error) {
	// Implement Google OAuth verification
	// This would verify the ID token with Google and extract user information
	// For now, return an error indicating it's not implemented
	return nil, fmt.Errorf("Google OAuth not yet implemented")
}

// GetGoogleAuthURL returns Google OAuth authorization URL
func (a *AuthService) GetGoogleAuthURL(state string) string {
	return a.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// Logout invalidates user session
func (a *AuthService) Logout(userID int64, sessionID string, ipAddress, userAgent string) error {
	// Invalidate session in database
	// This would be implemented in the UserService or directly here

	// Log logout
	a.LogAuditEvent(&userID, "logout", "authentication", fmt.Sprintf("%d", userID),
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

// LogAuditEvent logs an audit event (exported for use in handlers)
func (a *AuthService) LogAuditEvent(userID *int64, action, resourceType, resourceID string, details map[string]interface{}, ipAddress, userAgent string) {
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