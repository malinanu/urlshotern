package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/URLshorter/url-shortener/internal/models"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidClaims = errors.New("invalid token claims")
)

type JWTService struct {
	secretKey       string
	issuer          string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type Claims struct {
	UserID      int64  `json:"user_id"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	AccountType string `json:"account_type"`
	SessionID   string `json:"session_id"`
	TokenType   string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

func NewJWTService(secretKey, issuer string, accessTTL, refreshTTL time.Duration) *JWTService {
	return &JWTService{
		secretKey:       secretKey,
		issuer:          issuer,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

// GenerateTokenPair creates both access and refresh tokens for a user
func (j *JWTService) GenerateTokenPair(user *models.User, sessionID string) (*TokenPair, error) {
	// Generate access token
	accessToken, accessExpiresAt, err := j.generateToken(user, sessionID, "access", j.accessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, _, err := j.generateToken(user, sessionID, "refresh", j.refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiresAt,
		TokenType:    "Bearer",
	}, nil
}

// generateToken creates a JWT token with the specified type and TTL
func (j *JWTService) generateToken(user *models.User, sessionID, tokenType string, ttl time.Duration) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)

	claims := &Claims{
		UserID:      user.ID,
		Email:       user.Email,
		Role:        j.getUserRole(user),
		AccountType: user.AccountType,
		SessionID:   sessionID,
		TokenType:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("%d", user.ID),
			Audience:  []string{"url-shortener"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        j.generateTokenID(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates and parses a JWT token
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// RefreshToken creates a new access token using a valid refresh token
func (j *JWTService) RefreshToken(refreshToken string, user *models.User, sessionID string) (*TokenPair, error) {
	claims, err := j.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Verify this is a refresh token
	if claims.TokenType != "refresh" {
		return nil, errors.New("invalid token type for refresh")
	}

	// Verify user and session match
	if claims.UserID != user.ID || claims.SessionID != sessionID {
		return nil, errors.New("token validation failed")
	}

	// Generate new token pair
	return j.GenerateTokenPair(user, sessionID)
}

// GetTokenClaims extracts claims from a token without full validation (for expired tokens)
func (j *JWTService) GetTokenClaims(tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// InvalidateToken adds token to blacklist (would need Redis implementation)
func (j *JWTService) InvalidateToken(tokenID string) error {
	// This would typically store the token ID in Redis with expiration
	// For now, we'll implement basic invalidation
	// In production, you'd want to store invalidated token IDs in Redis
	return nil
}

// generateTokenID creates a unique ID for the token
func (j *JWTService) generateTokenID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getUserRole determines the user's role for JWT claims
func (j *JWTService) getUserRole(user *models.User) string {
	if user.IsAdmin {
		return "admin"
	}
	
	switch user.AccountType {
	case "premium":
		return "premium_user"
	case "enterprise":
		return "enterprise_user"
	default:
		return "user"
	}
}

// IsTokenExpired checks if a token is expired without full validation
func (j *JWTService) IsTokenExpired(tokenString string) bool {
	claims, err := j.GetTokenClaims(tokenString)
	if err != nil {
		return true
	}

	return claims.ExpiresAt.Before(time.Now())
}

// GetTokenRemainingTime returns the remaining time until token expires
func (j *JWTService) GetTokenRemainingTime(tokenString string) (time.Duration, error) {
	claims, err := j.GetTokenClaims(tokenString)
	if err != nil {
		return 0, err
	}

	remaining := claims.ExpiresAt.Sub(time.Now())
	if remaining < 0 {
		return 0, ErrExpiredToken
	}

	return remaining, nil
}

// ValidateTokenType ensures the token is of the expected type
func (j *JWTService) ValidateTokenType(tokenString, expectedType string) error {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return err
	}

	if claims.TokenType != expectedType {
		return fmt.Errorf("expected token type %s, got %s", expectedType, claims.TokenType)
	}

	return nil
}

// ExtractUserIDFromToken extracts user ID from token without full validation
func (j *JWTService) ExtractUserIDFromToken(tokenString string) (int64, error) {
	claims, err := j.GetTokenClaims(tokenString)
	if err != nil {
		return 0, err
	}

	return claims.UserID, nil
}

// CreatePasswordResetToken creates a special token for password reset
func (j *JWTService) CreatePasswordResetToken(user *models.User, duration time.Duration) (string, error) {
	now := time.Now()
	expiresAt := now.Add(duration)

	claims := &Claims{
		UserID:    user.ID,
		Email:     user.Email,
		TokenType: "password_reset",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("%d", user.ID),
			Audience:  []string{"password-reset"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        j.generateTokenID(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// ValidatePasswordResetToken validates a password reset token
func (j *JWTService) ValidatePasswordResetToken(tokenString string) (*Claims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "password_reset" {
		return nil, errors.New("invalid token type for password reset")
	}

	return claims, nil
}

// CreateEmailVerificationToken creates a token for email verification
func (j *JWTService) CreateEmailVerificationToken(user *models.User, duration time.Duration) (string, error) {
	now := time.Now()
	expiresAt := now.Add(duration)

	claims := &Claims{
		UserID:    user.ID,
		Email:     user.Email,
		TokenType: "email_verification",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("%d", user.ID),
			Audience:  []string{"email-verification"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        j.generateTokenID(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// ValidateEmailVerificationToken validates an email verification token
func (j *JWTService) ValidateEmailVerificationToken(tokenString string) (*Claims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "email_verification" {
		return nil, errors.New("invalid token type for email verification")
	}

	return claims, nil
}