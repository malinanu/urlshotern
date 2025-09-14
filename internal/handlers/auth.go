package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
)

type AuthHandlers struct {
	authService  *services.AuthService
	smsService   *services.SMSService
	emailService *services.EmailService
	validator    *validator.Validate
}

// NewAuthHandlers creates new auth handlers
func NewAuthHandlers(authService *services.AuthService, smsService *services.SMSService, emailService *services.EmailService) *AuthHandlers {
	return &AuthHandlers{
		authService:  authService,
		smsService:   smsService,
		emailService: emailService,
		validator:    validator.New(),
	}
}

// Register handles user registration
func (h *AuthHandlers) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Register user
	_, err := h.authService.Register(&req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		middleware.LogError(c, err, "User registration failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "User registered successfully")
	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful",
		"user_id": 1, // Temporary placeholder
	})
}

// Login handles user authentication
func (h *AuthHandlers) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Login attempt")
	c.JSON(http.StatusOK, gin.H{
		"access_token":  "dummy-access-token",
		"refresh_token": "dummy-refresh-token",
		"expires_in":    3600,
		"user": gin.H{
			"id":             1,
			"name":          "Test User",
			"email":         req.Email,
			"phone":         "",
			"phone_verified": false,
			"email_verified": true,
			"provider":      "local",
			"avatar_url":    "",
			"account_type":  "user",
			"is_active":     true,
			"created_at":    "2025-01-01T00:00:00Z",
		},
	})
}

// RefreshToken handles token refresh
func (h *AuthHandlers) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Token refresh attempt")
	c.JSON(http.StatusOK, gin.H{
		"message": "Token refresh successful",
		"token":   "new-dummy-token", // Temporary placeholder
	})
}

// SendOTP handles OTP sending
func (h *AuthHandlers) SendOTP(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	verification, err := h.smsService.ResendOTP(userID)
	if err != nil {
		middleware.LogError(c, err, "Failed to send OTP")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "OTP sent successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "OTP sent successfully",
		"expires_at": verification.ExpiresAt,
	})
}

// VerifyOTP handles OTP verification
func (h *AuthHandlers) VerifyOTP(c *gin.Context) {
	var req models.OTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify OTP
	err := h.smsService.VerifyOTP(req.UserID, req.OTP)
	if err != nil {
		middleware.LogError(c, err, "OTP verification failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Phone number verified successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Phone number verified successfully",
		"verified": true,
	})
}

// GetProfile returns user profile
func (h *AuthHandlers) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    userID,
			"name":  "Test User",
			"email": "test@example.com",
		},
	})
}

// UpdateProfile handles profile updates
func (h *AuthHandlers) UpdateProfile(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement profile update in auth service
	middleware.LogInfo(c, "Profile update requested")
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile update functionality coming soon",
	})
}

// ChangePassword handles password changes
func (h *AuthHandlers) ChangePassword(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement password change in auth service
	middleware.LogInfo(c, "Password change requested")
	c.JSON(http.StatusOK, gin.H{
		"message": "Password change functionality coming soon",
	})
}

// Logout handles user logout (invalidate session)
func (h *AuthHandlers) Logout(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Implement session invalidation
	middleware.LogInfo(c, "User logged out")
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// ResendOTP handles OTP resending with user ID from request
func (h *AuthHandlers) ResendOTP(c *gin.Context) {
	var req models.ResendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	verification, err := h.smsService.ResendOTP(req.UserID)
	if err != nil {
		middleware.LogError(c, err, "Failed to resend OTP")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "OTP resent successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "OTP sent successfully",
		"expires_at": verification.ExpiresAt,
	})
}

// GetUserByID handles admin user lookup
func (h *AuthHandlers) GetUserByID(c *gin.Context) {
	userIDParam := c.Param("id")
	userID, err := strconv.ParseInt(userIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    userID,
			"name":  "Test User",
			"email": "test@example.com",
		},
	})
}

// SendEmailVerification handles email verification sending
func (h *AuthHandlers) SendEmailVerification(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	verification, err := h.emailService.ResendVerificationEmail(userID)
	if err != nil {
		middleware.LogError(c, err, "Failed to send email verification")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Email verification sent successfully")
	c.JSON(http.StatusOK, gin.H{
		"message":    "Email verification sent successfully",
		"expires_at": verification.ExpiresAt,
	})
}

// VerifyEmail handles email verification
func (h *AuthHandlers) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification token is required"})
		return
	}

	err := h.emailService.VerifyEmail(token)
	if err != nil {
		middleware.LogError(c, err, "Email verification failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Email verified successfully")
	c.JSON(http.StatusOK, gin.H{
		"message":  "Email verified successfully",
		"verified": true,
	})
}