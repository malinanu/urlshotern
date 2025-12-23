package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BillingHandler struct {
	billingService *services.BillingService
	config         *configs.Config
}

func NewBillingHandler(db *gorm.DB, config *configs.Config) *BillingHandler {
	return &BillingHandler{
		billingService: services.NewBillingService(db, config),
		config:         config,
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

// CreateCheckoutSession creates a new Polar checkout session for upgrading
func (bh *BillingHandler) CreateCheckoutSession(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user from context (set by auth middleware)
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
		return
	}

	// DEBUG: Read and log the request body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	log.Printf("[DEBUG] Checkout request body: %s", string(bodyBytes))

	// Restore the request body for binding
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req struct {
		ProductID  string  `json:"product_id" binding:"required"`
		PriceID    *string `json:"price_id,omitempty"`
		SuccessURL string  `json:"success_url"`
		CancelURL  string  `json:"cancel_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ERROR] JSON binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	log.Printf("[DEBUG] Parsed checkout request - ProductID: %s, SuccessURL: %s, CancelURL: %s", req.ProductID, req.SuccessURL, req.CancelURL)

	// Set default URLs if not provided
	if req.SuccessURL == "" {
		req.SuccessURL = bh.config.BaseURL + "/billing/success"
	}
	if req.CancelURL == "" {
		req.CancelURL = bh.config.BaseURL + "/billing/cancelled"
	}

	checkout, err := bh.billingService.CreateCheckoutSession(
		userID,
		user.Email,
		req.ProductID,
		req.PriceID,
		req.SuccessURL,
		req.CancelURL,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create checkout session",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"checkout_url":    checkout.URL,
		"checkout_id":     checkout.ID,
		"client_secret":   checkout.ClientSecret,
		"expires_at":      checkout.ExpiresAt,
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

	if !subscription.CancelAtPeriodEnd {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Subscription is not set to cancel",
		})
		return
	}

	err = bh.billingService.UpdateSubscription(userID, map[string]interface{}{
		"cancel_at_period_end": false,
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

// GetBillingHistory returns the user's billing history from Polar
func (bh *BillingHandler) GetBillingHistory(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user's subscription to find customer ID
	subscription, err := bh.billingService.GetUserSubscription(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get subscription",
			"details": err.Error(),
		})
		return
	}

	// If user doesn't have a Polar customer ID, return empty history
	if subscription.PolarCustomerID == "" {
		c.JSON(http.StatusOK, gin.H{
			"invoices": []gin.H{},
			"has_more": false,
		})
		return
	}

	// Fetch orders from Polar API
	polarService := bh.billingService.GetPolarService()
	orders, err := polarService.ListOrders(subscription.PolarCustomerID, 20)
	if err != nil {
		log.Printf("Failed to fetch orders from Polar: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch billing history",
			"details": err.Error(),
		})
		return
	}

	// Transform Polar orders into invoice format
	history := make([]gin.H, 0, len(orders))
	for _, order := range orders {
		// Get product details for description
		product, err := polarService.GetProduct(order.ProductID)
		description := "Subscription"
		if err == nil {
			description = product.Name
		}

		// Build invoice URL (Polar may provide this in the future)
		invoiceURL := fmt.Sprintf("https://polar.sh/orders/%s", order.ID)

		invoice := gin.H{
			"id":          order.ID,
			"amount":      order.Amount,
			"currency":    order.Currency,
			"status":      "paid",
			"created":     order.CreatedAt.Format(time.RFC3339),
			"description": description,
			"product_id":  order.ProductID,
			"hosted_invoice_url": invoiceURL,
		}

		if order.SubscriptionID != nil {
			invoice["subscription_id"] = *order.SubscriptionID
		}

		history = append(history, invoice)
	}

	c.JSON(http.StatusOK, gin.H{
		"invoices": history,
		"has_more": len(orders) >= 20,
	})
}

// UpdatePaymentMethod redirects user to Polar customer portal for payment method management
// Polar manages payment methods through their customer portal rather than direct API
func (bh *BillingHandler) UpdatePaymentMethod(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user's subscription to find customer ID
	subscription, err := bh.billingService.GetUserSubscription(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get subscription",
			"details": err.Error(),
		})
		return
	}

	// Payment methods are managed through Polar customer portal
	portalURL := "https://polar.sh/customer-portal"
	if subscription.PolarCustomerID != "" {
		portalURL = fmt.Sprintf("https://polar.sh/customer-portal?customer_id=%s", subscription.PolarCustomerID)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment methods are managed through Polar customer portal",
		"portal_url": portalURL,
		"redirect_required": true,
	})
}

// GetCustomerPortalURL returns the Polar customer portal URL
func (bh *BillingHandler) GetCustomerPortalURL(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	subscription, err := bh.billingService.GetUserSubscription(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get subscription",
			"details": err.Error(),
		})
		return
	}

	// Polar customer portal URL
	portalURL := "https://polar.sh/customer-portal"
	if subscription.PolarCustomerID != "" {
		portalURL = "https://polar.sh/customer-portal?customer_id=" + subscription.PolarCustomerID
	}

	c.JSON(http.StatusOK, gin.H{
		"portal_url": portalURL,
	})
}

// GetProducts returns all available Polar products/plans
func (bh *BillingHandler) GetProducts(c *gin.Context) {
	polarService := bh.billingService.GetPolarService()

	// Check if we need to sync (database is empty or force sync requested)
	var productCount int64
	bh.billingService.GetDB().
		Table("polar_products").
		Where("is_archived = ?", false).
		Count(&productCount)

	shouldSync := productCount == 0 || c.Query("sync") == "true"

	// Try to sync products from Polar if needed
	if shouldSync {
		log.Printf("[Billing] Syncing products from Polar (count=%d, force=%v)", productCount, c.Query("sync") == "true")
		err := polarService.SyncProductsFromPolar()
		if err != nil {
			log.Printf("[Billing] Failed to sync products from Polar: %v", err)
			// If database is empty and sync failed, this is a critical error
			if productCount == 0 {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error":   "Failed to fetch products from payment gateway",
					"details": "Unable to sync products and no cached data available. Please try again later.",
				})
				return
			}
			// Otherwise, log warning and continue with cached data
			log.Printf("[Billing] Using cached product data due to sync failure")
		} else {
			log.Printf("[Billing] Successfully synced products from Polar")
		}
	}

	// Get products from database
	var products []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		PlanType    string `json:"plan_type"`
		IsRecurring bool   `json:"is_recurring"`
		Prices      interface{} `json:"prices"`
		Benefits    interface{} `json:"benefits"`
	}

	err := bh.billingService.GetDB().
		Table("polar_products").
		Where("is_archived = ?", false).
		Scan(&products).Error

	if err != nil {
		log.Printf("[Billing] Failed to retrieve products from database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get products",
			"details": err.Error(),
		})
		return
	}

	log.Printf("[Billing] Returning %d products", len(products))
	c.JSON(http.StatusOK, gin.H{
		"products": products,
	})
}

// SyncProducts manually triggers a product sync from Polar (admin/debugging endpoint)
func (bh *BillingHandler) SyncProducts(c *gin.Context) {
	polarService := bh.billingService.GetPolarService()

	log.Printf("[Billing] Manual product sync triggered")

	err := polarService.SyncProductsFromPolar()
	if err != nil {
		log.Printf("[Billing] Manual sync failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to sync products",
			"details": err.Error(),
		})
		return
	}

	// Count synced products
	var productCount int64
	bh.billingService.GetDB().
		Table("polar_products").
		Where("is_archived = ?", false).
		Count(&productCount)

	log.Printf("[Billing] Manual sync successful: %d products", productCount)

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Products synced successfully",
		"product_count": productCount,
	})
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