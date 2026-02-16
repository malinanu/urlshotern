package handlers

import (
	"fmt"
	"log"
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
	adminService *services.AdminService
	smsService   *services.SMSService
	emailService *services.EmailService
	validator    *validator.Validate
}

// NewAuthHandlers creates new auth handlers
func NewAuthHandlers(authService *services.AuthService, adminService *services.AdminService, smsService *services.SMSService, emailService *services.EmailService) *AuthHandlers {
	return &AuthHandlers{
		authService:  authService,
		adminService: adminService,
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
	response, err := h.authService.Register(&req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		middleware.LogError(c, err, "User registration failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log new user registration (async)
	go func() {
		_ = h.adminService.LogActivity(
			&response.User.ID,
			response.User.Name,
			"user_registered",
			fmt.Sprintf("New user registered: %s", req.Email),
			c.ClientIP(),
			c.Request.UserAgent(),
		)
	}()

	middleware.LogInfo(c, "User registered successfully")
	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful",
		"user_id": response.User.ID,
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

	// Log successful login (async)
	go func() {
		_ = h.adminService.LogActivity(
			&response.User.ID,
			response.User.Name,
			"user_login",
			fmt.Sprintf("User logged in from %s", c.ClientIP()),
			c.ClientIP(),
			c.Request.UserAgent(),
		)
	}()

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
	userID, exists := middleware.GetUserID(c)
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
			"details": err.Error(),
		})
		return
	}

	// Additional phone validation if phone is provided
	if req.Phone != nil && *req.Phone != "" {
		// Phone validation is already handled by validator tag, but we can add extra checks here if needed
		middleware.LogInfo(c, "Profile update requested with phone number")
	}

	// Update profile using auth service
	updatedUser, err := h.authService.UpdateProfile(userID, &req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		log.Printf("[ERROR] Profile update failed for user %d: %v", userID, err)
		middleware.LogError(c, err, "Failed to update profile")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update profile",
			"message": err.Error(),
			"details": err.Error(),
		})
		return
	}

	log.Printf("[INFO] Profile updated successfully for user %d", userID)

	middleware.LogInfo(c, "Profile updated successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    updatedUser.ToPublic(),
	})
}

// GetPreferences returns user preferences
func (h *AuthHandlers) GetPreferences(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get preferences using auth service
	preferences, err := h.authService.GetUserPreferences(userID)
	if err != nil {
		middleware.LogError(c, err, "Failed to get user preferences")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get preferences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"preferences": preferences,
	})
}

// UpdatePreferences handles updating user preferences
func (h *AuthHandlers) UpdatePreferences(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UserPreferences
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Update preferences using auth service
	err := h.authService.UpdateUserPreferences(userID, &req)
	if err != nil {
		middleware.LogError(c, err, "Failed to update preferences")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preferences"})
		return
	}

	// Get updated preferences to return
	updatedPreferences, err := h.authService.GetUserPreferences(userID)
	if err != nil {
		middleware.LogError(c, err, "Failed to get updated preferences")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated preferences"})
		return
	}

	middleware.LogInfo(c, "User preferences updated successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Preferences updated successfully",
		"preferences": updatedPreferences,
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

	user, err := h.adminService.GetUserByIDWithStats(userID)
	if err != nil {
		middleware.LogError(c, err, "Failed to get user")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
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

// Admin Panel Handlers

// GetAdminStats returns system-wide statistics
func (h *AuthHandlers) GetAdminStats(c *gin.Context) {
	stats, err := h.adminService.GetSystemStats()
	if err != nil {
		middleware.LogError(c, err, "Failed to get admin stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetUserList returns paginated list of users
func (h *AuthHandlers) GetUserList(c *gin.Context) {
	// Parse query params (limit, offset)
	limit := 20 // default
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	users, err := h.adminService.GetUserList(limit, offset)
	if err != nil {
		middleware.LogError(c, err, "Failed to get user list")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Get total count for pagination
	totalCount, err := h.adminService.GetTotalUsersCount()
	if err != nil {
		middleware.LogError(c, err, "Failed to get total users count")
		totalCount = 0 // Fallback to 0 if count fails
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  users,
		"total":  totalCount,
		"limit":  limit,
		"offset": offset,
	})
}

// ToggleUserStatus activates or deactivates a user
func (h *AuthHandlers) ToggleUserStatus(c *gin.Context) {
	userIDParam := c.Param("id")
	userID, err := strconv.ParseInt(userIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	err = h.adminService.ToggleUserStatus(userID, req.IsActive)
	if err != nil {
		middleware.LogError(c, err, "Failed to toggle user status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}

	middleware.LogInfo(c, "User status updated successfully")
	c.JSON(http.StatusOK, gin.H{"message": "User status updated"})
}

// DeleteUserAdmin deletes a user (admin endpoint)
func (h *AuthHandlers) DeleteUserAdmin(c *gin.Context) {
	userIDParam := c.Param("id")
	userID, err := strconv.ParseInt(userIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.adminService.DeleteUser(userID)
	if err != nil {
		middleware.LogError(c, err, "Failed to delete user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	middleware.LogInfo(c, "User deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// GetRecentActivity returns recent system activity
func (h *AuthHandlers) GetRecentActivity(c *gin.Context) {
	limit := 50 // default

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	activities, err := h.adminService.GetRecentActivity(limit)
	if err != nil {
		middleware.LogError(c, err, "Failed to get recent activity")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"activities": activities})
}

// CMS Management Handlers

// GetCMSPages returns list of CMS pages
func (h *AuthHandlers) GetCMSPages(c *gin.Context) {
	limit := 20 // default
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	pages, err := h.adminService.GetCMSPages(limit, offset)
	if err != nil {
		middleware.LogError(c, err, "Failed to get CMS pages")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pages"})
		return
	}

	// Get total count for pagination
	totalCount, err := h.adminService.GetTotalCMSPagesCount()
	if err != nil {
		middleware.LogError(c, err, "Failed to get total pages count")
		totalCount = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"pages":  pages,
		"total":  totalCount,
		"limit":  limit,
		"offset": offset,
	})
}

// GetCMSBlocks returns list of content blocks
func (h *AuthHandlers) GetCMSBlocks(c *gin.Context) {
	limit := 20 // default
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	blocks, err := h.adminService.GetCMSBlocks(limit, offset)
	if err != nil {
		middleware.LogError(c, err, "Failed to get CMS blocks")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch blocks"})
		return
	}

	// Get total count for pagination
	totalCount, err := h.adminService.GetTotalCMSBlocksCount()
	if err != nil {
		middleware.LogError(c, err, "Failed to get total blocks count")
		totalCount = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"blocks": blocks,
		"total":  totalCount,
		"limit":  limit,
		"offset": offset,
	})
}

// GetCMSTemplates returns list of page templates
func (h *AuthHandlers) GetCMSTemplates(c *gin.Context) {
	limit := 20 // default
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	templates, err := h.adminService.GetCMSTemplates(limit, offset)
	if err != nil {
		middleware.LogError(c, err, "Failed to get CMS templates")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch templates"})
		return
	}

	// Get total count for pagination
	totalCount, err := h.adminService.GetTotalCMSTemplatesCount()
	if err != nil {
		middleware.LogError(c, err, "Failed to get total templates count")
		totalCount = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"total":      totalCount,
		"limit":      limit,
		"offset":     offset,
	})
}

// SEO Tools Handlers

// GetSEOSettings returns SEO-related settings
func (h *AuthHandlers) GetSEOSettings(c *gin.Context) {
	settings, err := h.adminService.GetSEOSettings()
	if err != nil {
		middleware.LogError(c, err, "Failed to get SEO settings")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch SEO settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// GetSEOMetrics returns SEO-related metrics
func (h *AuthHandlers) GetSEOMetrics(c *gin.Context) {
	metrics, err := h.adminService.GetSEOMetrics()
	if err != nil {
		middleware.LogError(c, err, "Failed to get SEO metrics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch SEO metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetSEORecommendations returns SEO improvement recommendations
func (h *AuthHandlers) GetSEORecommendations(c *gin.Context) {
	recommendations, err := h.adminService.GetSEORecommendations()
	if err != nil {
		middleware.LogError(c, err, "Failed to get SEO recommendations")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch SEO recommendations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"recommendations": recommendations})
}

// Content Block CRUD Handlers

// CreateCMSBlock creates a new content block
func (h *AuthHandlers) CreateCMSBlock(c *gin.Context) {
	// TODO: Implement content block creation
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Content block creation not yet implemented"})
}

// UpdateCMSBlock updates an existing content block
func (h *AuthHandlers) UpdateCMSBlock(c *gin.Context) {
	// TODO: Implement content block update
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Content block update not yet implemented"})
}

// DeleteCMSBlock deletes a content block
func (h *AuthHandlers) DeleteCMSBlock(c *gin.Context) {
	// TODO: Implement content block deletion
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Content block deletion not yet implemented"})
}

// RenderCMSBlock renders a content block preview
func (h *AuthHandlers) RenderCMSBlock(c *gin.Context) {
	// TODO: Implement content block rendering
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Content block rendering not yet implemented"})
}

// SEO Analysis Handlers

// GetSEOAnalyses returns list of SEO analysis results
func (h *AuthHandlers) GetSEOAnalyses(c *gin.Context) {
	limit := 20 // default
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	analyses, err := h.adminService.GetSEOAnalyses(limit, offset)
	if err != nil {
		middleware.LogError(c, err, "Failed to get SEO analyses")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "fetch_failed",
			Message: "Failed to fetch SEO analyses",
		})
		return
	}

	// Get total count for pagination
	totalCount, err := h.adminService.GetTotalSEOAnalysesCount()
	if err != nil {
		middleware.LogError(c, err, "Failed to get total SEO analyses count")
		totalCount = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"analyses": analyses,
		"total":    totalCount,
		"limit":    limit,
		"offset":   offset,
	})
}

// AnalyzeSEOURL runs SEO analysis on a URL
func (h *AuthHandlers) AnalyzeSEOURL(c *gin.Context) {
	var req models.AnalyzeSEORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format",
		})
		return
	}

	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	// Run SEO analysis
	analysis, err := h.adminService.AnalyzeURL(req.URL, userID)
	if err != nil {
		middleware.LogError(c, err, "Failed to analyze URL")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "analysis_failed",
			Message: "Failed to analyze URL",
		})
		return
	}

	middleware.LogInfo(c, "SEO analysis created successfully")
	c.JSON(http.StatusCreated, analysis)
}

// Sitemap Handlers

// GetSitemap returns sitemap entries
func (h *AuthHandlers) GetSitemap(c *gin.Context) {
	entries, err := h.adminService.GetSitemapEntries()
	if err != nil {
		middleware.LogError(c, err, "Failed to get sitemap entries")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "fetch_failed",
			Message: "Failed to fetch sitemap entries",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entries": entries,
		"total":   len(entries),
	})
}

// GenerateSitemap generates XML sitemap
func (h *AuthHandlers) GenerateSitemap(c *gin.Context) {
	xml, err := h.adminService.GenerateSitemap()
	if err != nil {
		middleware.LogError(c, err, "Failed to generate sitemap")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "generation_failed",
			Message: "Failed to generate sitemap",
		})
		return
	}

	middleware.LogInfo(c, "Sitemap generated successfully")
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(xml))
}

// Meta Tags Handlers

// GetMetaTags returns list of meta tags
func (h *AuthHandlers) GetMetaTags(c *gin.Context) {
	limit := 20 // default
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	metaTags, err := h.adminService.GetMetaTags(limit, offset)
	if err != nil {
		middleware.LogError(c, err, "Failed to get meta tags")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "fetch_failed",
			Message: "Failed to fetch meta tags",
		})
		return
	}

	// Get total count for pagination
	totalCount, err := h.adminService.GetTotalMetaTagsCount()
	if err != nil {
		middleware.LogError(c, err, "Failed to get total meta tags count")
		totalCount = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"meta_tags": metaTags,
		"total":     totalCount,
		"limit":     limit,
		"offset":    offset,
	})
}

// CreateMetaTag creates a new meta tag
func (h *AuthHandlers) CreateMetaTag(c *gin.Context) {
	var req models.CreateMetaTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format",
		})
		return
	}

	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	// Create meta tag
	metaTag, err := h.adminService.CreateMetaTag(&req, userID)
	if err != nil {
		middleware.LogError(c, err, "Failed to create meta tag")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "creation_failed",
			Message: "Failed to create meta tag",
		})
		return
	}

	middleware.LogInfo(c, "Meta tag created successfully")
	c.JSON(http.StatusCreated, metaTag)
}

// DeleteMetaTag deletes a meta tag
func (h *AuthHandlers) DeleteMetaTag(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid meta tag ID",
		})
		return
	}

	err = h.adminService.DeleteMetaTag(id)
	if err != nil {
		if err.Error() == "meta tag not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "Meta tag not found",
			})
			return
		}
		middleware.LogError(c, err, "Failed to delete meta tag")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "deletion_failed",
			Message: "Failed to delete meta tag",
		})
		return
	}

	middleware.LogInfo(c, "Meta tag deleted successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Meta tag deleted successfully",
	})
}

// Password Reset Handlers

// ForgotPassword handles forgot password requests
func (h *AuthHandlers) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Password reset requested for email")

	// Request password reset (always returns success for security)
	err := h.authService.RequestPasswordReset(req.Email, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		// Log error but don't expose it to user
		middleware.LogError(c, err, "Failed to process password reset request")
	}

	// Always return success to prevent email enumeration
	c.JSON(http.StatusOK, gin.H{
		"message": "If an account exists with that email, a password reset link has been sent.",
	})
}

// ResetPassword handles password reset with token
func (h *AuthHandlers) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Password reset attempt")

	// Reset password using token
	err := h.authService.ResetPassword(req.Token, req.Password, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		middleware.LogError(c, err, "Password reset failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Password reset successful")
	c.JSON(http.StatusOK, gin.H{
		"message": "Password has been reset successfully. You can now log in with your new password.",
	})
}

// EmailHealthCheck checks email service configuration and status
func (h *AuthHandlers) EmailHealthCheck(c *gin.Context) {
	// Get email service health info
	healthInfo := h.emailService.GetHealthInfo()

	// Return health information
	c.JSON(http.StatusOK, healthInfo)
}