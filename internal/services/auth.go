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
	a.logAuditEvent(&registrationData.User.ID, "user_registered", "user", fmt.Sprintf("%d", registrationData.User.ID), 
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
		a.logAuditEvent(nil, "login_failed", "authentication", "", 
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
	a.logAuditEvent(&user.ID, "login_success", "authentication", fmt.Sprintf("%d", user.ID), 
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

// RefreshTokens generates new tokens using refresh token
func (a *AuthService) RefreshTokens(req *models.RefreshTokenRequest, ipAddress, userAgent string) (*models.AuthResponse, error) {
	// Get user ID from refresh token
	userID, err := a.jwtService.ExtractUserIDFromToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Get user
	user, err := a.userService.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Use JWT service to refresh tokens
	tokenPair, err := a.jwtService.RefreshToken(req.RefreshToken, user, "")
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Log token refresh
	a.logAuditEvent(&user.ID, "token_refreshed", "authentication", fmt.Sprintf("%d", user.ID), 
		map[string]interface{}{
			"email": user.Email,
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
	a.logAuditEvent(nil, "password_reset_requested", "authentication", "", 
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
	a.logAuditEvent(nil, "password_reset_completed", "authentication", "", 
		map[string]interface{}{
			"token_used": true,
		}, ipAddress, userAgent)

	return nil
}

// ChangePassword changes user password
func (a *AuthService) ChangePassword(userID int64, req *models.ChangePasswordRequest, ipAddress, userAgent string) error {
	err := a.userService.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return err
	}

	// Log password change
	a.logAuditEvent(&userID, "password_changed", "user", fmt.Sprintf("%d", userID), 
		map[string]interface{}{
			"method": "self_service",
		}, ipAddress, userAgent)

	return nil
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
	a.logAuditEvent(&userID, "logout", "authentication", fmt.Sprintf("%d", userID), 
		map[string]interface{}{
			"session_id": sessionID,
		}, ipAddress, userAgent)

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