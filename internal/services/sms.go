package services

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

type SMSService struct {
	db     *storage.PostgresStorage
	config *configs.Config
	client *http.Client
}

// TextLK API structures
type TextLKRequest struct {
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
	SenderID  string `json:"sender_id"`
	Type      string `json:"type"`
}

type TextLKResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// NewSMSService creates a new SMS service
func NewSMSService(db *storage.PostgresStorage, config *configs.Config) *SMSService {
	return &SMSService{
		db:     db,
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendOTP generates and sends an OTP to the user's phone
func (s *SMSService) SendOTP(userID int64, phone string) (*models.PhoneVerification, error) {
	// Generate 6-digit OTP
	otp, err := s.generateOTP()
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Create phone verification record
	verification := &models.PhoneVerification{
		UserID:    userID,
		Phone:     phone,
		OTPCode:   otp,
		Attempts:  0,
		ExpiresAt: time.Now().Add(5 * time.Minute), // 5 minutes expiry
		CreatedAt: time.Now(),
	}

	// Save to database
	err = s.CreatePhoneVerification(verification)
	if err != nil {
		return nil, fmt.Errorf("failed to save phone verification: %w", err)
	}

	// Send SMS
	message := fmt.Sprintf("Your URLShorter verification code is: %s. Valid for 5 minutes. Do not share this code.", otp)
	err = s.sendSMS(phone, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send SMS: %w", err)
	}

	// Don't return the actual OTP code
	verification.OTPCode = ""
	return verification, nil
}

// VerifyOTP verifies the provided OTP
func (s *SMSService) VerifyOTP(userID int64, otp string) error {
	// Get the latest verification for this user
	verification, err := s.GetLatestPhoneVerification(userID)
	if err != nil {
		return fmt.Errorf("no pending verification found")
	}

	// Check if already verified
	if verification.VerifiedAt != nil {
		return fmt.Errorf("phone number already verified")
	}

	// Check expiration
	if time.Now().After(verification.ExpiresAt) {
		return fmt.Errorf("OTP expired")
	}

	// Check attempts
	if verification.Attempts >= 3 {
		return fmt.Errorf("too many failed attempts")
	}

	// Verify OTP
	if verification.OTPCode != otp {
		// Increment attempts
		err = s.IncrementVerificationAttempts(verification.ID)
		if err != nil {
			fmt.Printf("Failed to increment verification attempts: %v\n", err)
		}
		return fmt.Errorf("invalid OTP")
	}

	// Mark as verified
	now := time.Now()
	err = s.MarkPhoneVerified(verification.ID, userID, now)
	if err != nil {
		return fmt.Errorf("failed to mark phone as verified: %w", err)
	}

	return nil
}

// ResendOTP resends OTP to the user's phone
func (s *SMSService) ResendOTP(userID int64) (*models.PhoneVerification, error) {
	// Get user's phone number
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.Phone == nil {
		return nil, fmt.Errorf("no phone number associated with account")
	}

	// Check rate limiting (1 SMS per minute)
	lastVerification, err := s.GetLatestPhoneVerification(userID)
	if err == nil && time.Since(lastVerification.CreatedAt) < time.Minute {
		return nil, fmt.Errorf("please wait before requesting another OTP")
	}

	// Invalidate previous verifications
	err = s.InvalidatePhoneVerifications(userID)
	if err != nil {
		fmt.Printf("Failed to invalidate previous verifications: %v\n", err)
	}

	// Send new OTP
	return s.SendOTP(userID, *user.Phone)
}

// sendSMS sends SMS using Text.lk API
func (s *SMSService) sendSMS(phone, message string) error {
	// For development, just log the SMS instead of actually sending
	if s.config.Environment == "development" {
		fmt.Printf("SMS to %s: %s\n", phone, message)
		return nil
	}

	// Text.lk API integration
	apiKey := s.getTextLKAPIKey()
	if apiKey == "" {
		return fmt.Errorf("SMS API key not configured")
	}

	request := TextLKRequest{
		Recipient: phone,
		Message:   message,
		SenderID:  "URLShorter", // This should be registered with Text.lk
		Type:      "plain",
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal SMS request: %w", err)
	}

	// Text.lk API endpoint
	apiURL := "https://textit.business/api/v1/send"

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send SMS request: %w", err)
	}
	defer resp.Body.Close()

	var smsResp TextLKResponse
	if err := json.NewDecoder(resp.Body).Decode(&smsResp); err != nil {
		return fmt.Errorf("failed to decode SMS response: %w", err)
	}

	if resp.StatusCode != 200 || smsResp.Status != "success" {
		return fmt.Errorf("SMS sending failed: %s", smsResp.Message)
	}

	return nil
}

// generateOTP generates a 6-digit OTP
func (s *SMSService) generateOTP() (string, error) {
	// Generate a random number between 100000 and 999999
	min := big.NewInt(100000)
	max := big.NewInt(999999)
	
	n, err := rand.Int(rand.Reader, new(big.Int).Sub(max, min))
	if err != nil {
		return "", err
	}
	
	otp := new(big.Int).Add(n, min)
	return fmt.Sprintf("%06d", otp.Int64()), nil
}

// getTextLKAPIKey gets the Text.lk API key from config
func (s *SMSService) getTextLKAPIKey() string {
	// In production, this should come from environment variables
	// For now, return empty string to use development mode
	return ""
}

// Database methods

// CreatePhoneVerification creates a phone verification record
func (s *SMSService) CreatePhoneVerification(verification *models.PhoneVerification) error {
	query := `
		INSERT INTO phone_verifications (user_id, phone, otp_code, attempts, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	
	return s.db.QueryRow(query, verification.UserID, verification.Phone, verification.OTPCode,
		verification.Attempts, verification.ExpiresAt, verification.CreatedAt).Scan(&verification.ID)
}

// GetLatestPhoneVerification gets the latest phone verification for a user
func (s *SMSService) GetLatestPhoneVerification(userID int64) (*models.PhoneVerification, error) {
	verification := &models.PhoneVerification{}
	query := `
		SELECT id, user_id, phone, otp_code, attempts, expires_at, verified_at, created_at
		FROM phone_verifications
		WHERE user_id = $1 AND verified_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1`
	
	err := s.db.QueryRow(query, userID).Scan(
		&verification.ID, &verification.UserID, &verification.Phone, &verification.OTPCode,
		&verification.Attempts, &verification.ExpiresAt, &verification.VerifiedAt,
		&verification.CreatedAt,
	)
	
	return verification, err
}

// IncrementVerificationAttempts increments the attempt count
func (s *SMSService) IncrementVerificationAttempts(verificationID interface{}) error {
	query := "UPDATE phone_verifications SET attempts = attempts + 1 WHERE id = $1"
	_, err := s.db.Exec(query, verificationID)
	return err
}

// MarkPhoneVerified marks the phone as verified
func (s *SMSService) MarkPhoneVerified(verificationID interface{}, userID int64, verifiedAt time.Time) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Mark verification as completed
	_, err = tx.Exec("UPDATE phone_verifications SET verified_at = $1 WHERE id = $2", verifiedAt, verificationID)
	if err != nil {
		return err
	}

	// Update user phone_verified status
	_, err = tx.Exec("UPDATE users SET phone_verified = true WHERE id = $1", userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// InvalidatePhoneVerifications invalidates all pending verifications for a user
func (s *SMSService) InvalidatePhoneVerifications(userID int64) error {
	query := "UPDATE phone_verifications SET verified_at = NOW() WHERE user_id = $1 AND verified_at IS NULL"
	_, err := s.db.Exec(query, userID)
	return err
}

// GetUserByID gets a user by ID (simplified version for SMS service)
func (s *SMSService) GetUserByID(userID int64) (*models.User, error) {
	user := &models.User{}
	query := "SELECT id, email, phone, phone_verified FROM users WHERE id = $1"
	
	err := s.db.QueryRow(query, userID).Scan(&user.ID, &user.Email, &user.Phone, &user.PhoneVerified)
	return user, err
}