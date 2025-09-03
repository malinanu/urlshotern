package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/URLshorter/url-shortener/internal/services"
)

type BillingHandler struct {
	billingService *services.BillingService
}

func NewBillingHandler(db *gorm.DB) *BillingHandler {
	return &BillingHandler{
		billingService: services.NewBillingService(db),
	}
}

// GetSubscription returns the current user's subscription details
func (bh *BillingHandler) GetSubscription(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	subscription, err := bh.billingService.GetUserSubscription(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get subscription details",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscription": subscription,
	})
}

// GetPlans returns all available subscription plans
func (bh *BillingHandler) GetPlans(c *gin.Context) {
	plans := bh.billingService.GetAllPlans()
	
	c.JSON(http.StatusOK, gin.H{
		"plans": plans,
	})
}

// GetUsage returns current usage statistics and limits for the user
func (bh *BillingHandler) GetUsage(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	usage, err := bh.billingService.CheckUsageLimits(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get usage statistics",
			"details": err.Error(),
		})
		return
	}

	subscription, err := bh.billingService.GetUserSubscription(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get subscription",
			"details": err.Error(),
		})
		return
	}

	plan, err := bh.billingService.GetPlan(subscription.PlanType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get plan details",
			"details": err.Error(),
		})
		return
	}

	// Calculate usage percentages for UI
	urlUsagePercent := 0
	clickUsagePercent := 0
	
	if plan.Limits.URLs > 0 {
		urlUsagePercent = (usage.CurrentURLs * 100) / plan.Limits.URLs
		if urlUsagePercent > 100 {
			urlUsagePercent = 100
		}
	}
	
	if plan.Limits.Clicks > 0 {
		clickUsagePercent = (usage.MonthlyClicks * 100) / plan.Limits.Clicks
		if clickUsagePercent > 100 {
			clickUsagePercent = 100
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"usage": usage,
		"limits": plan.Limits,
		"subscription": subscription,
		"plan": plan,
		"usage_percentages": gin.H{
			"urls": urlUsagePercent,
			"clicks": clickUsagePercent,
		},
		"warnings": gin.H{
			"url_limit_approaching": plan.Limits.URLs > 0 && urlUsagePercent >= 80,
			"click_limit_approaching": plan.Limits.Clicks > 0 && clickUsagePercent >= 80,
			"url_limit_reached": plan.Limits.URLs > 0 && usage.CurrentURLs >= plan.Limits.URLs,
			"click_limit_reached": plan.Limits.Clicks > 0 && usage.MonthlyClicks >= plan.Limits.Clicks,
		},
	})
}

// CreateCheckoutSession creates a new Stripe checkout session for upgrading
func (bh *BillingHandler) CreateCheckoutSession(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		PlanID     string `json:"plan_id" binding:"required"`
		SuccessURL string `json:"success_url"`
		CancelURL  string `json:"cancel_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate plan exists
	_, err := bh.billingService.GetPlan(req.PlanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID",
			"details": err.Error(),
		})
		return
	}

	// Set default URLs if not provided
	if req.SuccessURL == "" {
		req.SuccessURL = "https://yourdomain.com/billing/success"
	}
	if req.CancelURL == "" {
		req.CancelURL = "https://yourdomain.com/billing/cancel"
	}

	checkoutURL, err := bh.billingService.CreateCheckoutSession(userID, req.PlanID, req.SuccessURL, req.CancelURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create checkout session",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"checkout_url": checkoutURL,
		"expires_at": "2024-01-01T00:00:00Z", // Mock expiration
	})
}

// CancelSubscription marks a subscription for cancellation at period end
func (bh *BillingHandler) CancelSubscription(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if user has an active paid subscription
	subscription, err := bh.billingService.GetUserSubscription(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get subscription",
			"details": err.Error(),
		})
		return
	}

	if subscription.PlanType == "free" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot cancel free subscription",
		})
		return
	}

	if subscription.Status == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Subscription is already cancelled",
		})
		return
	}

	err = bh.billingService.UpdateSubscription(userID, map[string]interface{}{
		"cancel_at_end": true,
		"status": "cancelled",
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to cancel subscription",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription has been cancelled and will not renew",
		"current_period_end": subscription.CurrentPeriodEnd,
	})
}

// ReactivateSubscription reactivates a cancelled subscription
func (bh *BillingHandler) ReactivateSubscription(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	subscription, err := bh.billingService.GetUserSubscription(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get subscription",
			"details": err.Error(),
		})
		return
	}

	if !subscription.CancelAtEnd {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Subscription is not set to cancel",
		})
		return
	}

	err = bh.billingService.UpdateSubscription(userID, map[string]interface{}{
		"cancel_at_end": false,
		"status": "active",
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to reactivate subscription",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription has been reactivated",
	})
}

// GetBillingHistory returns the user's billing history
func (bh *BillingHandler) GetBillingHistory(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Mock billing history - in real implementation, fetch from Stripe
	history := []gin.H{
		{
			"id": "inv_001",
			"amount": 999,
			"currency": "usd",
			"status": "paid",
			"created": "2024-01-01T00:00:00Z",
			"period_start": "2024-01-01T00:00:00Z",
			"period_end": "2024-02-01T00:00:00Z",
			"description": "Premium Plan - Monthly",
			"hosted_invoice_url": "https://invoice.stripe.com/mock",
			"invoice_pdf": "https://invoice.stripe.com/mock.pdf",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"invoices": history,
		"has_more": false,
	})
}

// UpdatePaymentMethod updates the user's payment method
func (bh *BillingHandler) UpdatePaymentMethod(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		PaymentMethodID string `json:"payment_method_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Mock update - in real implementation, update via Stripe
	c.JSON(http.StatusOK, gin.H{
		"message": "Payment method updated successfully",
		"payment_method": gin.H{
			"id": req.PaymentMethodID,
			"type": "card",
			"card": gin.H{
				"brand": "visa",
				"last4": "4242",
				"exp_month": 12,
				"exp_year": 2025,
			},
		},
	})
}

// StripeWebhook handles Stripe webhook events
func (bh *BillingHandler) StripeWebhook(c *gin.Context) {
	bh.billingService.WebhookHandler(c)
}

// CheckFeatureAccess middleware to check if user has access to a feature
func (bh *BillingHandler) CheckFeatureAccess(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		subscription, err := bh.billingService.GetUserSubscription(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check subscription"})
			c.Abort()
			return
		}

		plan, err := bh.billingService.GetPlan(subscription.PlanType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get plan details"})
			c.Abort()
			return
		}

		// Check specific features
		hasAccess := false
		switch feature {
		case "custom_domain":
			hasAccess = plan.Limits.CustomDomain
		case "api_access":
			hasAccess = plan.Limits.APIAccess
		case "bulk_import":
			hasAccess = plan.Limits.BulkImport
		case "advanced_features":
			hasAccess = plan.Limits.Advanced
		default:
			hasAccess = true // Allow by default for unknown features
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Feature not available in your current plan",
				"feature": feature,
				"current_plan": subscription.PlanType,
				"upgrade_required": true,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}