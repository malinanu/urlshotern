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
	user, err := h.authService.Register(&req)
	if err != nil {
		middleware.LogError(c, err, "User registration failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send OTP for phone verification
	verification, err := h.smsService.SendOTP(user.ID, req.Phone)
	if err != nil {
		middleware.LogError(c, err, "Failed to send OTP")
		// Don't fail registration, user can resend OTP later
	}

	// Send email verification
	emailVerification, err := h.emailService.SendVerificationEmail(user.ID, req.Email)
	if err != nil {
		middleware.LogError(c, err, "Failed to send email verification")
		// Don't fail registration, user can resend email later
	}

	middleware.LogInfo(c, "User registered successfully")
	c.JSON(http.StatusCreated, gin.H{
		"message":      "Registration successful. Please verify your phone number and email.",
		"user":         user.ToPublic(),
		"otp_sent":     verification != nil,
		"email_sent":   emailVerification != nil,
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

	// Authenticate user
	authResponse, err := h.authService.Login(&req)
	if err != nil {
		middleware.LogError(c, err, "Login failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "User logged in successfully")
	c.JSON(http.StatusOK, authResponse)
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

	// Refresh tokens
	authResponse, err := h.authService.RefreshTokens(req.RefreshToken)
	if err != nil {
		middleware.LogError(c, err, "Token refresh failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Tokens refreshed successfully")
	c.JSON(http.StatusOK, authResponse)
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

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		middleware.LogError(c, err, "Failed to get user profile")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user.ToPublic(),
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

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		middleware.LogError(c, err, "Failed to get user")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user.ToPublic(),
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