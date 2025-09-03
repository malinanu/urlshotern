package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/URLshorter/url-shortener/internal/utils"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrPhoneAlreadyExists = errors.New("phone already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountNotActive   = errors.New("account not active")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrPhoneNotVerified   = errors.New("phone not verified")
	ErrInvalidOTP         = errors.New("invalid or expired OTP")
	ErrOTPExpired         = errors.New("OTP has expired")
	ErrTooManyOTPAttempts = errors.New("too many OTP attempts")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTokenUsed          = errors.New("token already used")
)

type UserService struct {
	db          storage.PostgresStorageInterface
	redis       storage.RedisStorageInterface
	jwtService  *JWTService
	smsService  *SMSService
	emailService *EmailService
	config      *Config
}

type Config struct {
	BCryptCost           int
	OTPExpiryMinutes     int
	MaxOTPAttempts       int
	EmailTokenExpiryHours int
	PasswordResetExpiryHours int
}

type UserRegistrationData struct {
	User            *models.User
	EmailToken      string
	RequiresOTP     bool
	TempPassword    string // For OAuth users
}

func NewUserService(db storage.PostgresStorageInterface, redis storage.RedisStorageInterface, 
	jwtService *JWTService, smsService *SMSService, emailService *EmailService, config *Config) *UserService {
	
	if config == nil {
		config = &Config{
			BCryptCost:               12,
			OTPExpiryMinutes:         5,
			MaxOTPAttempts:          3,
			EmailTokenExpiryHours:    24,
			PasswordResetExpiryHours: 2,
		}
	}

	return &UserService{
		db:           db,
		redis:        redis,
		jwtService:   jwtService,
		smsService:   smsService,
		emailService: emailService,
		config:       config,
	}
}

// CreateUser creates a new user account
func (u *UserService) CreateUser(req *models.RegisterRequest) (*UserRegistrationData, error) {
	// Check if email already exists
	existingUser, err := u.GetUserByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Check if phone already exists
	if req.Phone != "" {
		existingPhone, err := u.GetUserByPhone(req.Phone)
		if err == nil && existingPhone != nil {
			return nil, ErrPhoneAlreadyExists
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), u.config.BCryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate Snowflake ID
	userID := utils.GenerateSnowflakeID()

	// Create user
	user := &models.User{
		ID:              userID,
		Name:            req.Name,
		Email:           req.Email,
		PasswordHash:    stringPtr(string(hashedPassword)),
		Phone:           &req.Phone,
		PhoneVerified:   false,
		EmailVerified:   false,
		Provider:        "email",
		AccountType:     "free",
		IsActive:        true,
		IsAdmin:         false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save user to database
	err = u.db.CreateUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default preferences
	preferences := &models.UserPreferences{
		UserID:             user.ID,
		DefaultExpiration:  nil, // No expiration by default
		AnalyticsPublic:    false,
		EmailNotifications: true,
		MarketingEmails:    req.MarketingConsent,
		Timezone:          "UTC",
		Theme:             "light",
	}
	
	err = u.db.CreateUserPreferences(preferences)
	if err != nil {
		// Log error but don't fail user creation
		fmt.Printf("Failed to create user preferences: %v\n", err)
	}

	// Generate email verification token
	emailToken, err := u.CreateEmailVerificationToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create email verification token: %w", err)
	}

	return &UserRegistrationData{
		User:         user,
		EmailToken:   emailToken,
		RequiresOTP:  true,
		TempPassword: "",
	}, nil
}

// AuthenticateUser authenticates user with email and password
func (u *UserService) AuthenticateUser(req *models.LoginRequest) (*models.User, string, error) {
	// Get user by email
	user, err := u.GetUserByEmail(req.Email)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// Check if account is active
	if !user.IsActive {
		return nil, "", ErrAccountNotActive
	}

	// Check if user has password (for OAuth-only users)
	if user.PasswordHash == nil {
		return nil, "", errors.New("please use social login for this account")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// Create session
	sessionID, err := u.CreateSession(user, req.RememberMe, "", "")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login time
	user.LastLoginAt = timePtr(time.Now())
	err = u.db.UpdateUserLastLogin(user.ID, time.Now())
	if err != nil {
		// Log but don't fail authentication
		fmt.Printf("Failed to update last login: %v\n", err)
	}

	return user, sessionID, nil
}

// GetUserByID retrieves user by ID
func (u *UserService) GetUserByID(userID int64) (*models.User, error) {
	return u.db.GetUserByID(userID)
}

// GetUserByEmail retrieves user by email
func (u *UserService) GetUserByEmail(email string) (*models.User, error) {
	return u.db.GetUserByEmail(email)
}

// GetUserByPhone retrieves user by phone
func (u *UserService) GetUserByPhone(phone string) (*models.User, error) {
	return u.db.GetUserByPhone(phone)
}

// CreateSession creates a new user session
func (u *UserService) CreateSession(user *models.User, rememberMe bool, ipAddress, userAgent string) (string, error) {
	sessionID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour) // Default 24 hours
	
	if rememberMe {
		expiresAt = time.Now().Add(7 * 24 * time.Hour) // 7 days for remember me
	}

	session := &models.UserSession{
		ID:         sessionID,
		UserID:     user.ID,
		IPAddress:  &ipAddress,
		UserAgent:  &userAgent,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
	}

	err := u.db.CreateSession(session)
	if err != nil {
		return "", err
	}

	return sessionID.String(), nil
}

// IsSessionValid checks if session is valid
func (u *UserService) IsSessionValid(userID int64, sessionID string) (bool, error) {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return false, err
	}

	session, err := u.db.GetSessionByID(sessionUUID)
	if err != nil {
		return false, err
	}

	if session.UserID != userID {
		return false, nil
	}

	if session.ExpiresAt.Before(time.Now()) {
		return false, nil
	}

	return true, nil
}

// CreateEmailVerificationToken creates an email verification token
func (u *UserService) CreateEmailVerificationToken(userID int64) (string, error) {
	token := u.generateSecureToken()
	expiresAt := time.Now().Add(time.Duration(u.config.EmailTokenExpiryHours) * time.Hour)

	verification := &models.EmailVerification{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	err := u.db.CreateEmailVerification(verification)
	if err != nil {
		return "", err
	}

	return token, nil
}

// VerifyEmail verifies user email with token
func (u *UserService) VerifyEmail(token string) (*models.User, error) {
	verification, err := u.db.GetEmailVerificationByToken(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if verification.ExpiresAt.Before(time.Now()) {
		return nil, ErrInvalidToken
	}

	if verification.VerifiedAt != nil {
		return nil, ErrTokenUsed
	}

	// Mark email as verified
	now := time.Now()
	verification.VerifiedAt = &now
	err = u.db.UpdateEmailVerification(verification)
	if err != nil {
		return nil, err
	}

	// Update user email_verified status
	err = u.db.UpdateUserEmailVerified(verification.UserID, true)
	if err != nil {
		return nil, err
	}

	return u.GetUserByID(verification.UserID)
}

// GenerateAndSendOTP generates and sends OTP to user's phone
func (u *UserService) GenerateAndSendOTP(userID int64) error {
	user, err := u.GetUserByID(userID)
	if err != nil {
		return err
	}

	if user.Phone == nil || *user.Phone == "" {
		return errors.New("user has no phone number")
	}

	// Generate 6-digit OTP
	otp := u.generateOTP()
	expiresAt := time.Now().Add(time.Duration(u.config.OTPExpiryMinutes) * time.Minute)

	phoneVerification := &models.PhoneVerification{
		ID:        uuid.New(),
		UserID:    userID,
		Phone:     *user.Phone,
		OTPCode:   otp,
		Attempts:  0,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	// Delete any existing OTP for this user
	u.db.DeletePhoneVerificationByUserID(userID)

	// Create new OTP
	err = u.db.CreatePhoneVerification(phoneVerification)
	if err != nil {
		return err
	}

	// Send SMS
	if u.smsService != nil {
		err = u.smsService.SendOTP(*user.Phone, otp)
		if err != nil {
			fmt.Printf("Failed to send SMS: %v\n", err)
			// Don't fail the entire process if SMS fails
		}
	}

	return nil
}

// VerifyOTP verifies the OTP code
func (u *UserService) VerifyOTP(userID int64, otpCode string) error {
	verification, err := u.db.GetPhoneVerificationByUserID(userID)
	if err != nil {
		return ErrInvalidOTP
	}

	// Check if expired
	if verification.ExpiresAt.Before(time.Now()) {
		return ErrOTPExpired
	}

	// Check if already verified
	if verification.VerifiedAt != nil {
		return ErrTokenUsed
	}

	// Check attempts
	if verification.Attempts >= u.config.MaxOTPAttempts {
		return ErrTooManyOTPAttempts
	}

	// Verify OTP
	if verification.OTPCode != otpCode {
		// Increment attempts
		verification.Attempts++
		u.db.UpdatePhoneVerificationAttempts(verification.ID, verification.Attempts)
		return ErrInvalidOTP
	}

	// Mark as verified
	now := time.Now()
	verification.VerifiedAt = &now
	err = u.db.UpdatePhoneVerification(verification)
	if err != nil {
		return err
	}

	// Update user phone_verified status
	err = u.db.UpdateUserPhoneVerified(userID, true)
	if err != nil {
		return err
	}

	return nil
}

// CreatePasswordResetToken creates a password reset token
func (u *UserService) CreatePasswordResetToken(email string) (string, error) {
	user, err := u.GetUserByEmail(email)
	if err != nil {
		return "", ErrUserNotFound
	}

	token := u.generateSecureToken()
	expiresAt := time.Now().Add(time.Duration(u.config.PasswordResetExpiryHours) * time.Hour)

	passwordReset := &models.PasswordReset{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	err = u.db.CreatePasswordReset(passwordReset)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ResetPassword resets password using token
func (u *UserService) ResetPassword(token, newPassword string) error {
	passwordReset, err := u.db.GetPasswordResetByToken(token)
	if err != nil {
		return ErrInvalidToken
	}

	if passwordReset.ExpiresAt.Before(time.Now()) {
		return ErrInvalidToken
	}

	if passwordReset.UsedAt != nil {
		return ErrTokenUsed
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), u.config.BCryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	err = u.db.UpdateUserPassword(passwordReset.UserID, string(hashedPassword))
	if err != nil {
		return err
	}

	// Mark token as used
	now := time.Now()
	passwordReset.UsedAt = &now
	err = u.db.UpdatePasswordReset(passwordReset)
	if err != nil {
		return err
	}

	return nil
}

// ChangePassword changes user password (requires current password)
func (u *UserService) ChangePassword(userID int64, currentPassword, newPassword string) error {
	user, err := u.GetUserByID(userID)
	if err != nil {
		return err
	}

	if user.PasswordHash == nil {
		return errors.New("user has no password set")
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(currentPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), u.config.BCryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	return u.db.UpdateUserPassword(userID, string(hashedPassword))
}

// UpdateProfile updates user profile information
func (u *UserService) UpdateProfile(userID int64, req *models.UpdateProfileRequest) (*models.User, error) {
	user, err := u.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	updateData := make(map[string]interface{})
	
	if req.Name != nil && *req.Name != user.Name {
		updateData["name"] = *req.Name
	}

	if req.Phone != nil && (user.Phone == nil || *req.Phone != *user.Phone) {
		// Check if phone is already taken
		existingPhone, err := u.GetUserByPhone(*req.Phone)
		if err == nil && existingPhone != nil && existingPhone.ID != userID {
			return nil, ErrPhoneAlreadyExists
		}
		updateData["phone"] = *req.Phone
		updateData["phone_verified"] = false // Need to re-verify new phone
	}

	if len(updateData) > 0 {
		updateData["updated_at"] = time.Now()
		err = u.db.UpdateUser(userID, updateData)
		if err != nil {
			return nil, err
		}
	}

	return u.GetUserByID(userID)
}

// DeactivateUser deactivates a user account
func (u *UserService) DeactivateUser(userID int64) error {
	return u.db.UpdateUserStatus(userID, false)
}

// ActivateUser activates a user account
func (u *UserService) ActivateUser(userID int64) error {
	return u.db.UpdateUserStatus(userID, true)
}

// DeleteUser deletes a user account (soft delete by deactivating)
func (u *UserService) DeleteUser(userID int64) error {
	// For now, we'll just deactivate the account
	// In a real implementation, you might want to anonymize data or hard delete after a grace period
	return u.DeactivateUser(userID)
}

// GetUserPreferences retrieves user preferences
func (u *UserService) GetUserPreferences(userID int64) (*models.UserPreferences, error) {
	return u.db.GetUserPreferences(userID)
}

// UpdateUserPreferences updates user preferences
func (u *UserService) UpdateUserPreferences(userID int64, preferences *models.UserPreferences) error {
	preferences.UserID = userID
	return u.db.UpdateUserPreferences(preferences)
}

// Helper functions

func (u *UserService) generateSecureToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (u *UserService) generateOTP() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	otp := int(bytes[0])<<16 | int(bytes[1])<<8 | int(bytes[2])
	return fmt.Sprintf("%06d", otp%1000000)
}

func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}