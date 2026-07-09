package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/URLshorter/url-shortener/internal/models"
)

// URLOwnerLookup is the minimal capability the ownership middleware needs:
// resolve a short code to its URL mapping (which carries the owning user_id).
// *services.ShortenerService satisfies this via GetOriginalURL.
type URLOwnerLookup interface {
	GetOriginalURL(shortCode string) (*models.URLMapping, error)
}

// RequireURLOwnership ensures the authenticated caller owns the short code named
// in the ":shortCode" route param before the handler runs. This closes the IDOR
// on analytics/insights endpoints where any authenticated user could previously
// read another user's link data by guessing or enumerating short codes.
//
// Admins (legacy is_admin flag set on the user in context) bypass the check so
// existing admin tooling keeps working. API-key callers are matched by the
// user_id the key belongs to, which is the correct ownership boundary.
func RequireURLOwnership(lookup URLOwnerLookup) gin.HandlerFunc {
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

		// Admin bypass — mirrors RequireAdmin/RequirePermission legacy behavior.
		if IsAdmin(c) {
			c.Next()
			return
		}

		shortCode := c.Param("shortCode")
		if shortCode == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_short_code",
				Message: "Short code is required",
			})
			c.Abort()
			return
		}

		mapping, err := lookup.GetOriginalURL(shortCode)
		if err != nil {
			// Do not distinguish "not found" from "not yours" beyond 404 —
			// avoids leaking which short codes exist.
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "URL not found",
			})
			c.Abort()
			return
		}

		if mapping.UserID == nil || *mapping.UserID != userID {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "forbidden",
				Message: "You do not have access to this URL",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
