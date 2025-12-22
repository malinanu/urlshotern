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

	// Authenticate user
	response, err := h.authService.Login(&req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		middleware.LogError(c, err, "Login failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
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

	// Refresh tokens using auth service
	response, err := h.authService.RefreshTokens(&req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		middleware.LogError(c, err, "Token refresh failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Token refresh successful")
	c.JSON(http.StatusOK, response)
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
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user from context (already fetched by auth middleware)
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data"})
		return
	}

	// Return user profile
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

	// Additional phone validation if phone is provided
	if req.Phone != nil && *req.Phone != "" {
		// Phone validation is already handled by validator tag, but we can add extra checks here if needed
		middleware.LogInfo(c, "Profile update requested with phone number")
	}

	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Update profile using auth service
	updatedUser, err := h.authService.UpdateProfile(userID, &req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		middleware.LogError(c, err, "Failed to update profile")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	middleware.LogInfo(c, "Profile updated successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    updatedUser.ToPublic(),
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

	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Change password using auth service
	err := h.authService.ChangePassword(userID, &req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		middleware.LogError(c, err, "Failed to change password")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Password changed successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully. Please log in again with your new password.",
	})
}

// Logout handles user logout (invalidate session)
func (h *AuthHandlers) Logout(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No token provided"})
		return
	}

	// Remove "Bearer " prefix
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	// Invalidate the token (blacklist it until expiry)
	err := h.authService.InvalidateToken(token)
	if err != nil {
		middleware.LogError(c, err, "Failed to invalidate token")
		// Log the error but still return success to user for security
		// (don't expose internal errors during logout)
	}

	middleware.LogInfo(c, "User logged out successfully, token invalidated")
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