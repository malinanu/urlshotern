package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
)

type AuthHandlers struct {
	authService  *services.AuthService
	userService  *services.UserService
}

func NewAuthHandlers(authService *services.AuthService, userService *services.UserService) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
		userService: userService,
	}
}

// Register handles user registration
func (h *AuthHandlers) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Details: map[string]interface{}{
				"validation_error": err.Error(),
			},
		})
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Register user
	registrationData, err := h.authService.Register(&req, ipAddress, userAgent)
	if err != nil {
		switch err {
		case services.ErrEmailAlreadyExists:
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "email_already_exists",
				Message: "An account with this email already exists",
			})
		case services.ErrPhoneAlreadyExists:
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "phone_already_exists",
				Message: "An account with this phone number already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "registration_failed",
				Message: "Failed to create account",
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			})
		}
		return
	}

	// Return registration success with next steps
	c.JSON(http.StatusCreated, gin.H{
		"message": "Account created successfully",
		"user":    registrationData.User.ToPublic(),
		"next_steps": map[string]interface{}{
			"email_verification_required": !registrationData.User.EmailVerified,
			"phone_verification_required": !registrationData.User.PhoneVerified,
		},
	})
}

// Login handles user authentication
func (h *AuthHandlers) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Details: map[string]interface{}{
				"validation_error": err.Error(),
			},
		})
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Authenticate user
	authResponse, err := h.authService.Login(&req, ipAddress, userAgent)
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "invalid_credentials",
				Message: "Invalid email or password",
			})
		case services.ErrAccountNotActive:
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "account_suspended",
				Message: "Your account has been suspended",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "login_failed",
				Message: "Login failed",
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// RefreshToken handles token refresh
func (h *AuthHandlers) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Refresh tokens
	authResponse, err := h.authService.RefreshTokens(&req, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "invalid_refresh_token",
			Message: "Invalid or expired refresh token",
		})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// VerifyEmail handles email verification
func (h *AuthHandlers) VerifyEmail(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "missing_token",
			Message: "Verification token is required",
		})
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Verify email
	user, err := h.authService.VerifyEmail(token, ipAddress, userAgent)
	if err != nil {
		switch err {
		case services.ErrInvalidToken:
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_token",
				Message: "Invalid or expired verification token",
			})
		case services.ErrTokenUsed:
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "token_already_used",
				Message: "This verification token has already been used",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "verification_failed",
				Message: "Email verification failed",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
		"user":    user.ToPublic(),
	})
}

// ResendEmailVerification resends email verification
func (h *AuthHandlers) ResendEmailVerification(c *gin.Context) {
	userIDStr := c.Param("userID")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	// Get user
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "user_not_found",
			Message: "User not found",
		})
		return
	}

	if user.EmailVerified {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "email_already_verified",
			Message: "Email is already verified",
		})
		return
	}

	// Generate new verification token
	token, err := h.userService.CreateEmailVerificationToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "verification_failed",
			Message: "Failed to generate verification token",
		})
		return
	}

	// Send email (implement based on your email service)
	_ = token // TODO: Send email with token

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification email sent successfully",
	})
}

// SendOTP sends OTP to user's phone
func (h *AuthHandlers) SendOTP(c *gin.Context) {
	var req models.ResendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Send OTP
	err := h.authService.SendOTP(&req, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "otp_send_failed",
			Message: "Failed to send OTP",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP sent successfully",
	})
}

// VerifyOTP verifies OTP code
func (h *AuthHandlers) VerifyOTP(c *gin.Context) {
	var req models.OTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Verify OTP
	err := h.authService.VerifyOTP(&req, ipAddress, userAgent)
	if err != nil {
		switch err {
		case services.ErrInvalidOTP:
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_otp",
				Message: "Invalid OTP code",
			})
		case services.ErrOTPExpired:
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "otp_expired",
				Message: "OTP has expired",
			})
		case services.ErrTooManyOTPAttempts:
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
				Error:   "too_many_attempts",
				Message: "Too many OTP attempts. Please request a new OTP.",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "otp_verification_failed",
				Message: "OTP verification failed",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Phone number verified successfully",
	})
}

// ForgotPassword initiates password reset
func (h *AuthHandlers) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Initiate password reset
	err := h.authService.ForgotPassword(&req, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "password_reset_failed",
			Message: "Failed to initiate password reset",
		})
		return
	}

	// Always return success to prevent email enumeration
	c.JSON(http.StatusOK, gin.H{
		"message": "If an account with that email exists, a password reset link has been sent",
	})
}

// ResetPassword resets password using token
func (h *AuthHandlers) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Reset password
	err := h.authService.ResetPassword(&req, ipAddress, userAgent)
	if err != nil {
		switch err {
		case services.ErrInvalidToken:
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_token",
				Message: "Invalid or expired reset token",
			})
		case services.ErrTokenUsed:
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "token_already_used",
				Message: "This reset token has already been used",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "password_reset_failed",
				Message: "Password reset failed",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

// ChangePassword changes user password (requires authentication)
func (h *AuthHandlers) ChangePassword(c *gin.Context) {
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get user from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Change password
	err := h.authService.ChangePassword(userID.(int64), &req, ipAddress, userAgent)
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_current_password",
				Message: "Current password is incorrect",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "password_change_failed",
				Message: "Password change failed",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// GoogleAuth handles Google OAuth authentication
func (h *AuthHandlers) GoogleAuth(c *gin.Context) {
	var req models.GoogleAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Authenticate with Google
	authResponse, err := h.authService.GoogleAuth(&req, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "google_auth_failed",
			Message: "Google authentication failed",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// GetGoogleAuthURL returns Google OAuth URL
func (h *AuthHandlers) GetGoogleAuthURL(c *gin.Context) {
	state := c.Query("state")
	if state == "" {
		// Generate random state if not provided
		state = h.authService.GenerateSecureToken()[:32]
	}

	url := h.authService.GetGoogleAuthURL(state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": url,
		"state":    state,
	})
}

// Logout invalidates user session
func (h *AuthHandlers) Logout(c *gin.Context) {
	// Get user and session info from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	sessionID, exists := c.Get("session_id")
	if !exists {
		sessionID = ""
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Logout user
	err := h.authService.Logout(userID.(int64), sessionID.(string), ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "logout_failed",
			Message: "Logout failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// GetProfile returns current user profile
func (h *AuthHandlers) GetProfile(c *gin.Context) {
	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user.(*models.User).ToPublic(),
	})
}

// UpdateProfile updates user profile
func (h *AuthHandlers) UpdateProfile(c *gin.Context) {
	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Update profile
	user, err := h.userService.UpdateProfile(userID.(int64), &req)
	if err != nil {
		switch err {
		case services.ErrPhoneAlreadyExists:
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "phone_already_exists",
				Message: "This phone number is already associated with another account",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "profile_update_failed",
				Message: "Failed to update profile",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user.ToPublic(),
	})
}