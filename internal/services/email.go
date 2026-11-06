package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services/email"
	"github.com/URLshorter/url-shortener/internal/storage"
)

type EmailService struct {
	db       *storage.PostgresStorage
	config   *configs.Config
	provider email.Provider
}

// NewEmailService creates a new email service
func NewEmailService(db *storage.PostgresStorage, config *configs.Config) *EmailService {
	// Determine email provider based on environment
	providerType := getEmailProvider(config)

	// Log email service initialization
	log.Printf("INFO [EMAIL] Initializing email service (environment=%s, provider=%s)",
		config.Environment, providerType)

	// Create email provider based on environment
	emailConfig := email.EmailConfig{
		Provider:     providerType,
		SMTPHost:     config.SMTPHost,
		SMTPPort:     config.SMTPPort,
		SMTPUsername: config.SMTPUsername,
		SMTPPassword: config.SMTPPassword,
		FromEmail:    config.FromEmail,
		FromName:     config.FromName,
	}

	// Log SMTP configuration if using SMTP provider (mask password)
	if providerType == "smtp" {
		maskedPassword := maskPassword(config.SMTPPassword)
		log.Printf("INFO [EMAIL] SMTP configuration: host=%s, port=%d, username=%s, password=%s, from_email=%s, from_name=%s",
			config.SMTPHost, config.SMTPPort, config.SMTPUsername, maskedPassword,
			config.FromEmail, config.FromName)

		// Validate SMTP configuration
		validateSMTPConfig(config)
	}

	provider, err := email.NewProvider(emailConfig)
	if err != nil {
		log.Printf("ERROR [EMAIL] Failed to create %s email provider: %v - falling back to console provider",
			providerType, err)
		provider = email.NewConsoleProvider(emailConfig)
		providerType = "console" // Update provider type for logging
	}

	log.Printf("INFO [EMAIL] Email service initialized successfully using %s provider", providerType)

	return &EmailService{
		db:       db,
		config:   config,
		provider: provider,
	}
}

// getEmailProvider determines which email provider to use based on environment
func getEmailProvider(config *configs.Config) string {
	// Prioritize SMTP if explicitly configured
	if config.SMTPHost != "" {
		return "smtp"
	}

	// Fallback to environment-based defaults
	if config.Environment == "development" {
		return "console"
	}
	return "console"
}

// maskPassword masks the password for logging
func maskPassword(password string) string {
	if password == "" {
		return "<empty>"
	}
	if len(password) <= 4 {
		return "****"
	}
	return password[:2] + "****" + password[len(password)-2:]
}

// validateSMTPConfig validates SMTP configuration and logs warnings
func validateSMTPConfig(config *configs.Config) {
	warnings := []string{}

	// Check if required fields are set
	if config.SMTPHost == "" {
		warnings = append(warnings, "SMTP_HOST is empty")
	}
	if config.SMTPPort == 0 {
		warnings = append(warnings, "SMTP_PORT is 0 (should be 587, 465, or 25)")
	}
	if config.SMTPUsername == "" {
		warnings = append(warnings, "SMTP_USERNAME is empty")
	}
	if config.SMTPPassword == "" {
		warnings = append(warnings, "SMTP_PASSWORD is empty")
	}
	if config.FromEmail == "" {
		warnings = append(warnings, "FROM_EMAIL is empty")
	}

	// Gmail-specific validation
	if strings.Contains(strings.ToLower(config.SMTPHost), "gmail.com") {
		// Check if FROM_EMAIL matches SMTP_USERNAME
		if config.FromEmail != "" && config.SMTPUsername != "" &&
			config.FromEmail != config.SMTPUsername {
			log.Printf("WARN [EMAIL] Gmail SMTP detected: FROM_EMAIL (%s) differs from SMTP_USERNAME (%s)",
				config.FromEmail, config.SMTPUsername)
			log.Printf("WARN [EMAIL] Gmail requires FROM address to match authenticated account OR be configured as a 'Send mail as' alias in Gmail settings")
			log.Printf("WARN [EMAIL] If emails are not being sent, verify that '%s' is configured as a verified alias in Gmail", config.FromEmail)
		}
	}

	// Log all warnings
	if len(warnings) > 0 {
		for _, warning := range warnings {
			log.Printf("WARN [EMAIL] SMTP configuration issue: %s", warning)
		}
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
func (e *EmailService) SendEmailVerification(emailAddr, token string) error {
	// Build verification URL
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", e.config.BaseURL, token)

	// Generate email template
	template := email.VerificationEmailTemplate(verificationURL, emailAddr)

	// Send email using provider
	return e.deliver(emailAddr, template.Subject, template.HTMLBody, template.TextBody, false)
}

// SendPasswordReset sends password reset email
func (e *EmailService) SendPasswordReset(emailAddr, token string) error {
	// Build reset URL
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", e.config.BaseURL, token)

	// Generate email template
	template := email.PasswordResetEmailTemplate(resetURL, emailAddr)

	// Log email sending attempt
	log.Printf("INFO [EMAIL] Sending password reset email to %s", emailAddr)

	// Send email using provider
	err := e.deliver(emailAddr, template.Subject, template.HTMLBody, template.TextBody, false)
	if err != nil {
		log.Printf("ERROR [EMAIL] Failed to send password reset email to %s: %v", emailAddr, err)
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	log.Printf("INFO [EMAIL] Password reset email sent successfully to %s", emailAddr)
	return nil
}

// SendTeamInvitation sends team invitation email
func (e *EmailService) SendTeamInvitation(emailAddr, teamName, inviterName, token string) error {
	// Build invitation URL
	inviteURL := fmt.Sprintf("%s/team-invite?token=%s", e.config.BaseURL, token)

	log.Printf("INFO [EMAIL] Sending team invitation to %s for team %q — accept URL: %s", emailAddr, teamName, inviteURL)

	// Generate email template
	template := email.TeamInvitationEmailTemplate(inviteURL, teamName, inviterName, emailAddr)

	// Send email using provider
	err := e.deliver(emailAddr, template.Subject, template.HTMLBody, template.TextBody, false)
	if err != nil {
		log.Printf("ERROR [EMAIL] Failed to send team invitation email to %s: %v", emailAddr, err)
	}
	return err
}

// sendEmail sends verification email using the provider
func (e *EmailService) sendEmail(emailAddr, token string) error {
	// This is the same as SendEmailVerification - redirect to it
	return e.SendEmailVerification(emailAddr, token)
}

// SendEmail sends a generic email (for alerts, notifications, etc.)
func (e *EmailService) SendEmail(to, subject, htmlBody, textBody string) error {
	return e.deliver(to, subject, htmlBody, textBody, false)
}

// ErrEmailSuppressed is returned when a marketing email is addressed to someone who opted out.
var ErrEmailSuppressed = errors.New("recipient has unsubscribed from marketing emails")

// SendPromotionalEmail sends a marketing message. It refuses addresses that have
// unsubscribed and labels the footer as promotional (CAN-SPAM).
func (e *EmailService) SendPromotionalEmail(to, subject, htmlBody, textBody string) error {
	suppressed, err := e.IsEmailSuppressed(to)
	if err != nil {
		return fmt.Errorf("check suppression list: %w", err)
	}
	if suppressed {
		return ErrEmailSuppressed
	}
	return e.deliver(to, subject, htmlBody, textBody, true)
}

// SendDMCANotice forwards a takedown notice to the Designated Agent and sends the
// complaining party a receipt.
func (e *EmailService) SendDMCANotice(notice email.DMCANotice) error {
	tpl := email.DMCANoticeEmailTemplate(notice)
	if err := e.deliver(email.DMCAAgentEmail, tpl.Subject, tpl.HTMLBody, tpl.TextBody, false); err != nil {
		return err
	}
	receipt := email.DMCAReceiptEmailTemplate(notice.ReferenceID, notice.SubmittedAt)
	if err := e.deliver(notice.Email, receipt.Subject, receipt.HTMLBody, receipt.TextBody, false); err != nil {
		// The notice itself reached the agent; a failed receipt shouldn't fail the submission
		log.Printf("WARN [EMAIL] DMCA receipt for %s not sent: %v", notice.ReferenceID, err)
	}
	return nil
}

// deliver adds the CAN-SPAM footer (postal address, privacy policy and a
// per-recipient unsubscribe link) and hands the message to the provider.
func (e *EmailService) deliver(to, subject, htmlBody, textBody string, promotional bool) error {
	htmlBody, textBody = email.ApplyFooter(htmlBody, textBody, email.FooterOptions{
		SiteURL:        e.siteURL(),
		UnsubscribeURL: e.UnsubscribeURL(to),
		Promotional:    promotional,
	})
	return e.provider.SendEmail(to, subject, htmlBody, textBody)
}

func (e *EmailService) siteURL() string {
	if e.config != nil && e.config.BaseURL != "" {
		return strings.TrimRight(e.config.BaseURL, "/")
	}
	return email.DefaultSiteURL
}

// UnsubscribeURL returns the recipient's personal unsubscribe link.
func (e *EmailService) UnsubscribeURL(to string) string {
	return e.siteURL() + "/unsubscribe?token=" + url.QueryEscape(UnsubscribeToken(e.config.JWTSecret, to))
}

// IsEmailSuppressed reports whether marketing email must not be sent to addr: it is on
// the suppression list, or it belongs to a user who has marketing emails turned off.
func (e *EmailService) IsEmailSuppressed(addr string) (bool, error) {
	var suppressed bool
	err := e.db.QueryRow(`
		SELECT EXISTS (SELECT 1 FROM email_suppressions WHERE email = $1)
		    OR EXISTS (
		        SELECT 1 FROM users u
		        LEFT JOIN user_preferences p ON p.user_id = u.id
		        WHERE LOWER(u.email) = $1 AND COALESCE(p.marketing_emails, FALSE) = FALSE)`,
		normalizeEmail(addr)).Scan(&suppressed)
	return suppressed, err
}

// Unsubscribe applies an unsubscribe link: the address joins the suppression list and
// any matching user has marketing emails switched off. Account emails are unaffected.
func (e *EmailService) Unsubscribe(token string) (string, error) {
	addr, err := ParseUnsubscribeToken(e.config.JWTSecret, token)
	if err != nil {
		return "", err
	}
	if _, err := e.db.Exec(`
		INSERT INTO email_suppressions (email, reason) VALUES ($1, 'unsubscribe')
		ON CONFLICT (email) DO NOTHING`, addr); err != nil {
		return "", fmt.Errorf("record suppression: %w", err)
	}
	if _, err := e.db.Exec(`
		UPDATE user_preferences SET marketing_emails = FALSE
		WHERE user_id IN (SELECT id FROM users WHERE LOWER(email) = $1)`, addr); err != nil {
		return "", fmt.Errorf("update preferences: %w", err)
	}
	return addr, nil
}

// SendContactFormEmail sends contact form submission email
func (e *EmailService) SendContactFormEmail(name, emailAddr, subject, message string) error {
	// Generate email template for admin
	adminTemplate := email.ContactFormEmailTemplate(name, emailAddr, subject, message)

	// Send to contact email (admin)
	contactEmail := "contact@trunc.site"
	err := e.deliver(contactEmail, adminTemplate.Subject, adminTemplate.HTMLBody, adminTemplate.TextBody, false)
	if err != nil {
		return err
	}

	// Send a copy to the user
	userSubject := fmt.Sprintf("Copy of your message: %s", subject)
	userTemplate := email.ContactFormEmailTemplate(name, emailAddr, userSubject, message)
	return e.deliver(emailAddr, userTemplate.Subject, userTemplate.HTMLBody, userTemplate.TextBody, false)
}

// SendSupportInquiryEmail sends support inquiry email
func (e *EmailService) SendSupportInquiryEmail(name, emailAddr, subject, issue string) error {
	// Generate email template
	template := email.SupportInquiryEmailTemplate(name, emailAddr, subject, issue)

	// Send to support email
	supportEmail := "hello@trunc.site"
	return e.deliver(supportEmail, template.Subject, template.HTMLBody, template.TextBody, false)
}

// SendOTPEmail sends one-time password email
func (e *EmailService) SendOTPEmail(emailAddr, otp string, expiryMinutes int) error {
	// Generate email template
	template := email.OTPEmailTemplate(otp, emailAddr, expiryMinutes)

	// Send email using provider
	return e.deliver(emailAddr, template.Subject, template.HTMLBody, template.TextBody, false)
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

// GetHealthInfo returns email service health information
func (e *EmailService) GetHealthInfo() map[string]interface{} {
	// Determine provider type
	providerType := "unknown"
	switch e.provider.(type) {
	case *email.SMTPProvider:
		providerType = "smtp"
	case *email.ConsoleProvider:
		providerType = "console"
	}

	healthInfo := map[string]interface{}{
		"status":      "ok",
		"provider":    providerType,
		"environment": e.config.Environment,
	}

	// Add SMTP configuration if using SMTP provider
	if providerType == "smtp" {
		healthInfo["smtp_config"] = map[string]interface{}{
			"host":       e.config.SMTPHost,
			"port":       e.config.SMTPPort,
			"username":   e.config.SMTPUsername,
			"password":   maskPassword(e.config.SMTPPassword),
			"from_email": e.config.FromEmail,
			"from_name":  e.config.FromName,
		}

		// Add warnings if any
		warnings := []string{}

		// Check if required fields are set
		if e.config.SMTPHost == "" {
			warnings = append(warnings, "SMTP_HOST is empty")
		}
		if e.config.SMTPPort == 0 {
			warnings = append(warnings, "SMTP_PORT is 0")
		}
		if e.config.SMTPUsername == "" {
			warnings = append(warnings, "SMTP_USERNAME is empty")
		}
		if e.config.SMTPPassword == "" {
			warnings = append(warnings, "SMTP_PASSWORD is empty")
		}
		if e.config.FromEmail == "" {
			warnings = append(warnings, "FROM_EMAIL is empty")
		}

		// Gmail-specific warnings
		if strings.Contains(strings.ToLower(e.config.SMTPHost), "gmail.com") {
			if e.config.FromEmail != "" && e.config.SMTPUsername != "" &&
				e.config.FromEmail != e.config.SMTPUsername {
				warnings = append(warnings, fmt.Sprintf("Gmail SMTP: FROM_EMAIL (%s) differs from SMTP_USERNAME (%s). Ensure '%s' is configured as a verified 'Send mail as' alias in Gmail settings.",
					e.config.FromEmail, e.config.SMTPUsername, e.config.FromEmail))
			}
		}

		if len(warnings) > 0 {
			healthInfo["warnings"] = warnings
			healthInfo["status"] = "warning"
		}
	} else if providerType == "console" {
		healthInfo["warnings"] = []string{"Using console provider - emails will be logged to stdout instead of being sent"}
		healthInfo["status"] = "warning"
	}

	return healthInfo
}

// SendSpamAlert sends an email to the user when their URL is flagged as spam
func (e *EmailService) SendSpamAlert(toEmail, originalURL, shortCode string) error {
	subject := "Action Required: URL Disabled Due to Spam/Malware Detection"
	htmlBody := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2>Your URL has been disabled</h2>
			<p>Hello,</p>
			<p>Our security systems have flagged one of your shortened URLs as potentially malicious or spam.</p>
			<p><strong>Short Code:</strong> %s</p>
			<p><strong>Original URL:</strong> %s</p>
			<p>To protect our users, this short link has been disabled and will no longer redirect to the destination.</p>
			<p>If you believe this is an error, please contact support.</p>
			<br/>
			<p>Best regards,<br/>The URL Shortener Team</p>
		</div>
	`, shortCode, originalURL)

	textBody := fmt.Sprintf(
		"Hello,\n\nOur security systems have flagged one of your shortened URLs as potentially malicious or spam.\n\nShort Code: %s\nOriginal URL: %s\n\nTo protect our users, this short link has been disabled and will no longer redirect to the destination.\n\nBest regards,\nThe URL Shortener Team",
		shortCode, originalURL)

	return e.deliver(toEmail, subject, htmlBody, textBody, false)
}

// SendWelcomeEmail sends a welcome email to a newly registered user
func (e *EmailService) SendWelcomeEmail(name, emailAddr string) error {
	// Generate email template
	template := email.WelcomeEmailTemplate(name, emailAddr)

	// Send to user
	return e.deliver(emailAddr, template.Subject, template.HTMLBody, template.TextBody, false)
}
