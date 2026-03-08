package services

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/google/uuid"
)

type SMSService struct {
	db     *storage.PostgresStorage
	config *configs.Config
	client *http.Client
}

func NewSMSService(db *storage.PostgresStorage, config *configs.Config) *SMSService {
	return &SMSService{
		db:     db,
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// SendOTP generates and sends an OTP to the user's phone.
func (s *SMSService) SendOTP(userID int64, phone string) (*models.PhoneVerification, error) {
	otp, err := s.generateOTP()
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %w", err)
	}

	verification := &models.PhoneVerification{
		ID:        uuid.New(),
		UserID:    userID,
		Phone:     phone,
		OTPCode:   otp,
		Attempts:  0,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}

	if err = s.CreatePhoneVerification(verification); err != nil {
		return nil, fmt.Errorf("failed to save phone verification: %w", err)
	}

	message := fmt.Sprintf("Your trunc.site verification code is: %s. Valid for 5 minutes. Do not share this code.", otp)
	if err = s.sendSMS(phone, message); err != nil {
		return nil, fmt.Errorf("failed to send SMS: %w", err)
	}

	verification.OTPCode = "" // never expose the OTP in the API response
	return verification, nil
}

// VerifyOTP validates the 6-digit code submitted by the user.
func (s *SMSService) VerifyOTP(userID int64, otp string) error {
	verification, err := s.GetLatestPhoneVerification(userID)
	if err != nil {
		return fmt.Errorf("no pending verification found")
	}
	if verification.VerifiedAt != nil {
		return fmt.Errorf("phone number already verified")
	}
	if time.Now().After(verification.ExpiresAt) {
		return fmt.Errorf("OTP expired")
	}
	if verification.Attempts >= 3 {
		return fmt.Errorf("too many failed attempts")
	}
	if verification.OTPCode != otp {
		_ = s.IncrementVerificationAttempts(verification.ID)
		return fmt.Errorf("invalid OTP")
	}
	if err = s.MarkPhoneVerified(verification.ID, userID, time.Now()); err != nil {
		return fmt.Errorf("failed to mark phone as verified: %w", err)
	}
	return nil
}

// ResendOTP invalidates any pending code and sends a fresh one.
func (s *SMSService) ResendOTP(userID int64) (*models.PhoneVerification, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	if user.Phone == nil {
		return nil, fmt.Errorf("no phone number associated with account")
	}

	// Rate-limit: one OTP per minute
	last, err := s.GetLatestPhoneVerification(userID)
	if err == nil && time.Since(last.CreatedAt) < time.Minute {
		return nil, fmt.Errorf("please wait before requesting another OTP")
	}

	_ = s.InvalidatePhoneVerifications(userID)
	return s.SendOTP(userID, *user.Phone)
}

// sendSMS routes to text.lk for Sri Lankan numbers, Twilio for everything else.
func (s *SMSService) sendSMS(phone, message string) error {
	if s.config.Environment == "development" {
		fmt.Printf("[SMS DEV] To %s: %s\n", phone, message)
		return nil
	}
	if isSriLankanNumber(phone) {
		return s.sendViaTextLK(phone, message)
	}
	return s.sendViaTwilio(phone, message)
}

// isSriLankanNumber detects +94 / 94 / 07x formats.
func isSriLankanNumber(phone string) bool {
	n := strings.ReplaceAll(phone, " ", "")
	return strings.HasPrefix(n, "+94") ||
		strings.HasPrefix(n, "94") ||
		strings.HasPrefix(n, "07")
}

// normalizeToE164SriLanka converts +94xxxxxxxxx or 07xxxxxxxx → 94xxxxxxxxx
func normalizeToE164SriLanka(phone string) string {
	n := strings.ReplaceAll(phone, " ", "")
	switch {
	case strings.HasPrefix(n, "+94"):
		return strings.TrimPrefix(n, "+")
	case strings.HasPrefix(n, "07"):
		return "94" + n[1:]
	default:
		return n
	}
}

// ── text.lk ──────────────────────────────────────────────────────────────────

type textLKRequest struct {
	APIToken  string `json:"api_token"`
	Recipient string `json:"recipient"`
	SenderID  string `json:"sender_id"`
	Type      string `json:"type"`
	Message   string `json:"message"`
}

type textLKResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func (s *SMSService) sendViaTextLK(phone, message string) error {
	if s.config.TextLKAPIKey == "" {
		return fmt.Errorf("TEXTLK_API_KEY not configured")
	}

	body, _ := json.Marshal(textLKRequest{
		APIToken:  s.config.TextLKAPIKey,
		Recipient: normalizeToE164SriLanka(phone),
		SenderID:  s.config.TextLKSenderID,
		Type:      "plain",
		Message:   message,
	})

	req, err := http.NewRequest("POST", "https://app.text.lk/api/http/sms/send", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to build text.lk request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("text.lk request failed: %w", err)
	}
	defer resp.Body.Close()

	var result textLKResponse
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode text.lk response: %w", err)
	}
	if result.Status != "success" {
		return fmt.Errorf("text.lk error: %s", result.Message)
	}
	return nil
}

// ── Twilio ────────────────────────────────────────────────────────────────────

type twilioResponse struct {
	SID          string `json:"sid"`
	Status       string `json:"status"`
	ErrorCode    *int   `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

func (s *SMSService) sendViaTwilio(phone, message string) error {
	if s.config.TwilioAccountSID == "" || s.config.TwilioAuthToken == "" {
		return fmt.Errorf("Twilio credentials not configured")
	}
	if s.config.TwilioFromNumber == "" {
		return fmt.Errorf("TWILIO_FROM_NUMBER not configured")
	}

	endpoint := fmt.Sprintf(
		"https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json",
		s.config.TwilioAccountSID,
	)

	form := url.Values{}
	form.Set("To", phone)
	form.Set("From", s.config.TwilioFromNumber)
	form.Set("Body", message)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to build Twilio request: %w", err)
	}
	req.SetBasicAuth(s.config.TwilioAccountSID, s.config.TwilioAuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("Twilio request failed: %w", err)
	}
	defer resp.Body.Close()

	var result twilioResponse
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode Twilio response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("Twilio error: %s", result.ErrorMessage)
	}
	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (s *SMSService) generateOTP() (string, error) {
	min := big.NewInt(100000)
	max := big.NewInt(999999)
	n, err := rand.Int(rand.Reader, new(big.Int).Sub(max, min))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", new(big.Int).Add(n, min).Int64()), nil
}

// ── DB methods ────────────────────────────────────────────────────────────────

func (s *SMSService) CreatePhoneVerification(verification *models.PhoneVerification) error {
	query := `INSERT INTO phone_verifications (id, user_id, phone, otp_code, attempts, expires_at, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := s.db.Exec(query, verification.ID, verification.UserID, verification.Phone, verification.OTPCode,
		verification.Attempts, verification.ExpiresAt, verification.CreatedAt)
	return err
}

func (s *SMSService) GetLatestPhoneVerification(userID int64) (*models.PhoneVerification, error) {
	v := &models.PhoneVerification{}
	query := `SELECT id, user_id, phone, otp_code, attempts, expires_at, verified_at, created_at
	          FROM phone_verifications
	          WHERE user_id = $1 AND verified_at IS NULL
	          ORDER BY created_at DESC LIMIT 1`
	err := s.db.QueryRow(query, userID).Scan(
		&v.ID, &v.UserID, &v.Phone, &v.OTPCode,
		&v.Attempts, &v.ExpiresAt, &v.VerifiedAt, &v.CreatedAt,
	)
	return v, err
}

func (s *SMSService) IncrementVerificationAttempts(verificationID interface{}) error {
	_, err := s.db.Exec("UPDATE phone_verifications SET attempts = attempts + 1 WHERE id = $1", verificationID)
	return err
}

func (s *SMSService) MarkPhoneVerified(verificationID interface{}, userID int64, verifiedAt time.Time) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec("UPDATE phone_verifications SET verified_at = $1 WHERE id = $2", verifiedAt, verificationID); err != nil {
		return err
	}
	if _, err = tx.Exec("UPDATE users SET phone_verified = true WHERE id = $1", userID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SMSService) InvalidatePhoneVerifications(userID int64) error {
	_, err := s.db.Exec("UPDATE phone_verifications SET verified_at = NOW() WHERE user_id = $1 AND verified_at IS NULL", userID)
	return err
}

func (s *SMSService) GetUserByID(userID int64) (*models.User, error) {
	user := &models.User{}
	err := s.db.QueryRow("SELECT id, email, phone, phone_verified FROM users WHERE id = $1", userID).
		Scan(&user.ID, &user.Email, &user.Phone, &user.PhoneVerified)
	return user, err
}
