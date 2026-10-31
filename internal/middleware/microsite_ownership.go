package middleware

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/URLshorter/url-shortener/internal/models"
)

// MicrositeOwnerLookup resolves a page ID to its owning user.
// *services.MicrositeService satisfies this via GetOwnerID.
type MicrositeOwnerLookup interface {
	GetOwnerID(id int64) (int64, error)
}

// RequireMicrositeOwnership ensures the authenticated caller owns the page named
// in the ":id" route param. It mirrors RequireURLOwnership: 401 without a user,
// admin bypass, 404 for missing pages and 403 for pages owned by someone else.
// A lookup error matching unavailable (e.g. migration not applied) returns 503.
func RequireMicrositeOwnership(lookup MicrositeOwnerLookup, unavailable error) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists || userID == 0 {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_id",
				Message: "Page ID is required",
			})
			c.Abort()
			return
		}

		ownerID, err := lookup.GetOwnerID(id)
		if err != nil {
			if unavailable != nil && errors.Is(err, unavailable) {
				c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
					Error:   "unavailable",
					Message: "Pages are temporarily unavailable",
				})
				c.Abort()
				return
			}
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "Page not found",
			})
			c.Abort()
			return
		}

		// Admin bypass — mirrors RequireURLOwnership.
		if IsAdmin(c) {
			c.Set("microsite_id", id)
			c.Next()
			return
		}

		if ownerID != userID {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "forbidden",
				Message: "You do not have access to this page",
			})
			c.Abort()
			return
		}

		c.Set("microsite_id", id)
		c.Next()
	}
}
