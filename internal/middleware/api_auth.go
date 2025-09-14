package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
)

// APIKeyAuthMiddleware validates API key authentication
func APIKeyAuthMiddleware(apiKeyService *services.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for API key in Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			c.Abort()
			return
		}

		// Extract API key from "Bearer <key>" format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format. Use 'Bearer <api_key>'"})
			c.Abort()
			return
		}

		apiKey := parts[1]

		// Validate API key
		keyData, err := apiKeyService.ValidateAPIKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired API key"})
			c.Abort()
			return
		}

		// Update API key usage
		ipAddress := c.ClientIP()
		go func() {
			_ = apiKeyService.UpdateAPIKeyUsage(keyData.ID, ipAddress)
		}()

		// Set user and API key info in context
		c.Set("user_id", keyData.UserID)
		c.Set("api_key_id", keyData.ID)
		c.Set("api_key_permissions", keyData.Permissions)
		c.Set("auth_type", "api_key")

		c.Next()

		// Record API key usage after request
		go func() {
			statusCode := c.Writer.Status()
			endpoint := c.Request.URL.Path
			method := c.Request.Method
			userAgent := c.GetHeader("User-Agent")
			
			// Calculate response time (simplified - you might want to use a more accurate method)
			responseTime := int(time.Since(time.Now()).Milliseconds())
			
			_ = apiKeyService.RecordAPIKeyUsage(
				keyData.ID,
				endpoint,
				method,
				statusCode,
				responseTime,
				int(c.Request.ContentLength),
				c.Writer.Size(),
				ipAddress,
				userAgent,
			)
		}()
	}
}

// APIKeyPermissionMiddleware checks if the API key has the required permission
func APIKeyPermissionMiddleware(apiKeyService *services.APIKeyService, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if this is API key authentication
		authType, exists := c.Get("auth_type")
		if !exists || authType != "api_key" {
			c.Next()
			return
		}

		// Get API key permissions
		permissions, exists := c.Get("api_key_permissions")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "No permissions found"})
			c.Abort()
			return
		}

		apiKeyPerms, ok := permissions.(models.APIKeyPermissions)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid permissions format"})
			c.Abort()
			return
		}

		// Check permission
		hasPermission := false
		for _, perm := range apiKeyPerms {
			if perm == "*" || perm == permission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions for this action"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// FlexibleAuthMiddleware allows both JWT and API key authentication
func FlexibleAuthMiddleware(authMiddleware *AuthMiddleware, apiKeyService *services.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try API key authentication first
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer us_") {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 {
				apiKey := parts[1]
				keyData, err := apiKeyService.ValidateAPIKey(apiKey)
				if err == nil {
					// Valid API key
					ipAddress := c.ClientIP()
					go func() {
						_ = apiKeyService.UpdateAPIKeyUsage(keyData.ID, ipAddress)
					}()

					c.Set("user_id", keyData.UserID)
					c.Set("api_key_id", keyData.ID)
					c.Set("api_key_permissions", keyData.Permissions)
					c.Set("auth_type", "api_key")
					c.Next()
					return
				}
			}
		}

		// Fall back to JWT authentication
		authMiddleware.RequireAuth()(c)
	}
}