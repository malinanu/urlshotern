package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

type EmailService struct {
	db     *storage.PostgresStorage
	config *configs.Config
}

// NewEmailService creates a new email service
func NewEmailService(db *storage.PostgresStorage, config *configs.Config) *EmailService {
	return &EmailService{
		db:     db,
		config: config,
	}
}

// SendVerificationEmail creates and sends email verification
func (e *EmailService) SendVerificationEmail(userID int64, email string) (*models.EmailVerification, error) {
	// Generate verification token
	token, err := e.generateVerificationToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification token: %w", err)
	}

	// Create email verification record
	verification := &models.EmailVerification{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours expiry
		CreatedAt: time.Now(),
	}

	// Save to database
	err = e.CreateEmailVerification(verification)
	if err != nil {
		return nil, fmt.Errorf("failed to save email verification: %w", err)
	}

	// Send email
	err = e.sendEmail(email, token)
	if err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	return verification, nil
}

// VerifyEmail verifies the provided token
func (e *EmailService) VerifyEmail(token string) error {
	// Get verification by token
	verification, err := e.GetEmailVerificationByToken(token)
	if err != nil {
		return fmt.Errorf("invalid verification token")
	}

	// Check if already verified
	if verification.VerifiedAt != nil {
		return fmt.Errorf("email already verified")
	}

	// Check expiration
	if time.Now().After(verification.ExpiresAt) {
		return fmt.Errorf("verification token expired")
	}

	// Mark as verified
	now := time.Now()
	err = e.MarkEmailVerified(verification.ID, verification.UserID, now)
	if err != nil {
		return fmt.Errorf("failed to mark email as verified: %w", err)
	}

	return nil
}

// ResendVerificationEmail resends verification email
func (e *EmailService) ResendVerificationEmail(userID int64) (*models.EmailVerification, error) {
	// Get user's email
	user, err := e.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if already verified
	if user.EmailVerified {
		return nil, fmt.Errorf("email already verified")
	}

	// Check rate limiting (1 email per 5 minutes)
	lastVerification, err := e.GetLatestEmailVerification(userID)
	if err == nil && time.Since(lastVerification.CreatedAt) < 5*time.Minute {
		return nil, fmt.Errorf("please wait before requesting another verification email")
	}

	// Invalidate previous verifications
	err = e.InvalidateEmailVerifications(userID)
	if err != nil {
		fmt.Printf("Failed to invalidate previous email verifications: %v\n", err)
	}

	// Send new verification email
	return e.SendVerificationEmail(userID, user.Email)
}

// SendEmailVerification sends email verification with token (updated interface)
func (e *EmailService) SendEmailVerification(email, token string) error {
	// For development, just log the email instead of actually sending
	if e.config.Environment == "development" {
		verificationURL := fmt.Sprintf("http://localhost:3000/verify-email?token=%s", token)
		fmt.Printf("Email verification to %s: Click here to verify your email: %s\n", email, verificationURL)
		return nil
	}

	// TODO: Implement actual email sending (SMTP, SendGrid, etc.)
	fmt.Printf("Email sending not configured for production environment\n")
	return nil
}

// SendPasswordReset sends password reset email
func (e *EmailService) SendPasswordReset(email, token string) error {
	// For development, just log the email instead of actually sending
	if e.config.Environment == "development" {
		resetURL := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", token)
		fmt.Printf("Password reset for %s: Click here to reset your password: %s\n", email, resetURL)
		return nil
	}

	// TODO: Implement actual email sending (SMTP, SendGrid, etc.)
	fmt.Printf("Email sending not configured for production environment\n")
	return nil
}

// SendTeamInvitation sends team invitation email
func (e *EmailService) SendTeamInvitation(email, teamName, inviterName, token string) error {
	// For development, just log the email instead of actually sending
	if e.config.Environment == "development" {
		inviteURL := fmt.Sprintf("http://localhost:3000/team-invite?token=%s", token)
		fmt.Printf("Team invitation to %s: %s invited you to join team '%s'. Click here to accept: %s\n", 
			email, inviterName, teamName, inviteURL)
		return nil
	}

	// TODO: Implement actual email sending (SMTP, SendGrid, etc.)
	fmt.Printf("Team invitation email sending not configured for production environment\n")
	return nil
}

// sendEmail sends verification email (mock implementation for development)
func (e *EmailService) sendEmail(email, token string) error {
	// For development, just log the email instead of actually sending
	if e.config.Environment == "development" {
		verificationURL := fmt.Sprintf("http://localhost:3000/verify-email?token=%s", token)
		fmt.Printf("Email verification to %s: Click here to verify your email: %s\n", email, verificationURL)
		return nil
	}

	// TODO: Implement actual email sending (SMTP, SendGrid, etc.)
	fmt.Printf("Email sending not configured for production environment\n")
	return nil
}

// generateVerificationToken generates a random verification token
func (e *EmailService) generateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Database methods

// CreateEmailVerification creates an email verification record
func (e *EmailService) CreateEmailVerification(verification *models.EmailVerification) error {
	query := `
		INSERT INTO email_verifications (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	
	_, err := e.db.Exec(query, verification.ID, verification.UserID, verification.Token,
		verification.ExpiresAt, verification.CreatedAt)
	return err
}

// GetEmailVerificationByToken gets email verification by token
func (e *EmailService) GetEmailVerificationByToken(token string) (*models.EmailVerification, error) {
	verification := &models.EmailVerification{}
	query := `
		SELECT id, user_id, token, expires_at, verified_at, created_at
		FROM email_verifications
		WHERE token = $1`
	
	err := e.db.QueryRow(query, token).Scan(
		&verification.ID, &verification.UserID, &verification.Token,
		&verification.ExpiresAt, &verification.VerifiedAt, &verification.CreatedAt,
	)
	
	return verification, err
}

// GetLatestEmailVerification gets the latest email verification for a user
func (e *EmailService) GetLatestEmailVerification(userID int64) (*models.EmailVerification, error) {
	verification := &models.EmailVerification{}
	query := `
		SELECT id, user_id, token, expires_at, verified_at, created_at
		FROM email_verifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1`
	
	err := e.db.QueryRow(query, userID).Scan(
		&verification.ID, &verification.UserID, &verification.Token,
		&verification.ExpiresAt, &verification.VerifiedAt, &verification.CreatedAt,
	)
	
	return verification, err
}

// MarkEmailVerified marks the email as verified
func (e *EmailService) MarkEmailVerified(verificationID uuid.UUID, userID int64, verifiedAt time.Time) error {
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Mark verification as completed
	_, err = tx.Exec("UPDATE email_verifications SET verified_at = $1 WHERE id = $2", verifiedAt, verificationID)
	if err != nil {
		return err
	}

	// Update user email_verified status
	_, err = tx.Exec("UPDATE users SET email_verified = true WHERE id = $1", userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// InvalidateEmailVerifications invalidates all pending verifications for a user
func (e *EmailService) InvalidateEmailVerifications(userID int64) error {
	query := "UPDATE email_verifications SET verified_at = NOW() WHERE user_id = $1 AND verified_at IS NULL"
	_, err := e.db.Exec(query, userID)
	return err
}

// GetUserByID gets a user by ID (simplified version for email service)
func (e *EmailService) GetUserByID(userID int64) (*models.User, error) {
	user := &models.User{}
	query := "SELECT id, email, email_verified FROM users WHERE id = $1"
	
	err := e.db.QueryRow(query, userID).Scan(&user.ID, &user.Email, &user.EmailVerified)
	return user, err
}