package middleware

import (
	"net/http"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// Analytics feature constants — used in plan gating checks.
const (
	FeatureCityGeo     = "city_geo"
	FeatureTimeSeries  = "time_series"
	FeatureHeatmap     = "heatmap"
	FeatureUTM         = "utm"
	FeatureRealtime    = "realtime"
	FeatureLiveMap     = "live_map"
	FeatureBotFilter   = "bot_filter"
	FeatureCSVExport   = "csv_export"
	FeatureAISummary   = "ai_summary"
	FeatureAIAnomaly   = "ai_anomaly"
	FeatureAIRecommend = "ai_recommend"
	FeatureForecast    = "ai_forecast"
	FeatureAudience    = "ai_audience"
	FeatureBranding    = "branding" // Enterprise white-label branding

	FeatureRemoveBranding = "remove_branding" // Hide the "Made with Trunc" badge on microsites
	FeatureMicrositeForms = "microsite_forms" // Newsletter / contact forms on microsites
)

// planFeatures maps DB plan names to allowed features.
// DB plan names: free | premium | enterprise
// Display names (frontend only): Free | Pro | Ultra
var planFeatures = map[string][]string{
	"free": {},
	"premium": {
		FeatureCityGeo, FeatureTimeSeries, FeatureHeatmap, FeatureUTM,
		FeatureRealtime, FeatureLiveMap, FeatureBotFilter, FeatureCSVExport,
		FeatureMicrositeForms,
	},
	"growth": {
		FeatureCityGeo, FeatureTimeSeries, FeatureHeatmap, FeatureUTM,
		FeatureRealtime, FeatureLiveMap, FeatureBotFilter, FeatureCSVExport,
		FeatureMicrositeForms, FeatureRemoveBranding,
	},
	"enterprise": {
		FeatureCityGeo, FeatureTimeSeries, FeatureHeatmap, FeatureUTM,
		FeatureRealtime, FeatureLiveMap, FeatureBotFilter, FeatureCSVExport,
		FeatureAISummary, FeatureAIAnomaly, FeatureAIRecommend, FeatureForecast, FeatureAudience,
		FeatureBranding, FeatureRemoveBranding, FeatureMicrositeForms,
	},
}

func hasFeature(plan, feature string) bool {
	features, ok := planFeatures[plan]
	if !ok {
		return false
	}
	for _, f := range features {
		if f == feature {
			return true
		}
	}
	return false
}

// PlanHasFeature reports whether a DB plan name (free | premium | enterprise)
// includes the given feature. Used where a feature is enforced on a field
// rather than on a whole route (e.g. microsite badge removal).
func PlanHasFeature(plan, feature string) bool {
	return hasFeature(plan, feature)
}

// CheckPlanFeature returns a gin middleware that enforces plan-level access to
// a specific analytics feature. Unauthenticated requests are rejected with 401.
// Insufficient plan returns 403 with upgrade information.
func CheckPlanFeature(billingService *services.BillingService, feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "authentication_required",
				"message": "Authentication required to access analytics",
			})
			return
		}

		userID, ok := userIDVal.(int64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "authentication_required",
				"message": "Invalid user session",
			})
			return
		}

		subscription, err := billingService.GetUserSubscription(userID)
		if err != nil {
			// Default to free plan on error rather than hard-failing
			subscription = &models.PolarSubscription{PlanType: "free"}
		}

		if !hasFeature(subscription.PlanType, feature) {
			upgradeMsg := "Upgrade to Pro to access this feature"
			if subscription.PlanType == "premium" {
				upgradeMsg = "Upgrade to Ultra to access this feature"
			}
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":        "plan_upgrade_required",
				"message":      upgradeMsg,
				"current_plan": subscription.PlanType,
				"feature":      feature,
			})
			return
		}

		c.Next()
	}
}
