package services

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"
	"github.com/gin-gonic/gin"
)

type BillingService struct {
	db *gorm.DB
}

type Subscription struct {
	ID                int64                  `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID           int64                  `json:"user_id" gorm:"not null;index"`
	PlanType         string                 `json:"plan_type" gorm:"not null"` // free, premium, enterprise
	Status           string                 `json:"status" gorm:"not null"`    // active, cancelled, expired, trialing
	StripeCustomerID *string                `json:"stripe_customer_id,omitempty"`
	StripeSubID      *string                `json:"stripe_subscription_id,omitempty"`
	CurrentPeriodEnd *time.Time             `json:"current_period_end,omitempty"`
	CancelAtEnd      bool                   `json:"cancel_at_period_end" gorm:"default:false"`
	TrialEnd         *time.Time             `json:"trial_end,omitempty"`
	BillingCycle     string                 `json:"billing_cycle" gorm:"default:'monthly'"` // monthly, yearly
	CreatedAt        time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" gorm:"serializer:json"`
}

type Plan struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	Interval    string  `json:"interval"` // month, year
	Features    []string `json:"features"`
	Limits      PlanLimits `json:"limits"`
}

type PlanLimits struct {
	URLs         int `json:"max_urls"`          // -1 for unlimited
	Clicks       int `json:"max_clicks_month"`  // -1 for unlimited
	Analytics    int `json:"analytics_days"`    // days of analytics retention
	CustomDomain bool `json:"custom_domain"`
	APIAccess    bool `json:"api_access"`
	BulkImport   bool `json:"bulk_import"`
	Advanced     bool `json:"advanced_features"`
}

type BillingUsage struct {
	UserID       int64 `json:"user_id"`
	CurrentURLs  int   `json:"current_urls"`
	MonthlyClicks int   `json:"monthly_clicks"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
}

var PredefinedPlans = map[string]Plan{
	"free": {
		ID:          "free",
		Name:        "Free",
		Description: "Perfect for personal use",
		Price:       0,
		Currency:    "usd",
		Interval:    "month",
		Features: []string{
			"Up to 100 URLs",
			"5,000 clicks/month",
			"Basic analytics",
			"30 days data retention",
		},
		Limits: PlanLimits{
			URLs:         100,
			Clicks:       5000,
			Analytics:    30,
			CustomDomain: false,
			APIAccess:    false,
			BulkImport:   false,
			Advanced:     false,
		},
	},
	"premium": {
		ID:          "premium",
		Name:        "Premium",
		Description: "For professionals and small teams",
		Price:       9.99,
		Currency:    "usd",
		Interval:    "month",
		Features: []string{
			"Unlimited URLs",
			"100,000 clicks/month",
			"Advanced analytics",
			"1 year data retention",
			"Custom domains",
			"API access",
			"Bulk import",
		},
		Limits: PlanLimits{
			URLs:         -1, // unlimited
			Clicks:       100000,
			Analytics:    365,
			CustomDomain: true,
			APIAccess:    true,
			BulkImport:   true,
			Advanced:     true,
		},
	},
	"enterprise": {
		ID:          "enterprise",
		Name:        "Enterprise",
		Description: "For large organizations",
		Price:       49.99,
		Currency:    "usd",
		Interval:    "month",
		Features: []string{
			"Unlimited everything",
			"Advanced integrations",
			"Priority support",
			"Custom branding",
			"SSO integration",
			"Dedicated account manager",
		},
		Limits: PlanLimits{
			URLs:         -1,
			Clicks:       -1,
			Analytics:    -1, // unlimited
			CustomDomain: true,
			APIAccess:    true,
			BulkImport:   true,
			Advanced:     true,
		},
	},
}

func NewBillingService(db *gorm.DB) *BillingService {
	return &BillingService{db: db}
}

// GetUserSubscription retrieves the current subscription for a user
func (bs *BillingService) GetUserSubscription(userID int64) (*Subscription, error) {
	var subscription Subscription
	err := bs.db.Where("user_id = ?", userID).Order("created_at DESC").First(&subscription).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default free subscription
			return bs.CreateFreeSubscription(userID)
		}
		return nil, err
	}
	return &subscription, nil
}

// CreateFreeSubscription creates a free subscription for new users
func (bs *BillingService) CreateFreeSubscription(userID int64) (*Subscription, error) {
	subscription := &Subscription{
		UserID:       userID,
		PlanType:     "free",
		Status:       "active",
		BillingCycle: "monthly",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	err := bs.db.Create(subscription).Error
	if err != nil {
		return nil, err
	}
	
	return subscription, nil
}

// GetPlan returns plan details by plan ID
func (bs *BillingService) GetPlan(planID string) (*Plan, error) {
	plan, exists := PredefinedPlans[planID]
	if !exists {
		return nil, fmt.Errorf("plan not found: %s", planID)
	}
	return &plan, nil
}

// GetAllPlans returns all available plans
func (bs *BillingService) GetAllPlans() []Plan {
	plans := make([]Plan, 0, len(PredefinedPlans))
	for _, plan := range PredefinedPlans {
		plans = append(plans, plan)
	}
	return plans
}

// CheckUsageLimits checks if user has exceeded their plan limits
func (bs *BillingService) CheckUsageLimits(userID int64) (*BillingUsage, error) {
	subscription, err := bs.GetUserSubscription(userID)
	if err != nil {
		return nil, err
	}
	
	_, err = bs.GetPlan(subscription.PlanType)
	if err != nil {
		return nil, err
	}
	
	// Get current month start and end
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	
	// Count current URLs
	var urlCount int64
	bs.db.Table("url_mappings").Where("user_id = ? AND is_active = true", userID).Count(&urlCount)
	
	// Count monthly clicks
	var clickCount int64
	bs.db.Table("click_events").
		Joins("JOIN url_mappings ON url_mappings.short_code = click_events.short_code").
		Where("url_mappings.user_id = ? AND click_events.created_at >= ? AND click_events.created_at <= ?", userID, monthStart, monthEnd).
		Count(&clickCount)
	
	usage := &BillingUsage{
		UserID:       userID,
		CurrentURLs:  int(urlCount),
		MonthlyClicks: int(clickCount),
		PeriodStart:  monthStart,
		PeriodEnd:    monthEnd,
	}
	
	return usage, nil
}

// CanCreateURL checks if user can create a new URL within their plan limits
func (bs *BillingService) CanCreateURL(userID int64) (bool, string, error) {
	usage, err := bs.CheckUsageLimits(userID)
	if err != nil {
		return false, "Error checking usage limits", err
	}
	
	subscription, err := bs.GetUserSubscription(userID)
	if err != nil {
		return false, "Error getting subscription", err
	}
	
	plan, err := bs.GetPlan(subscription.PlanType)
	if err != nil {
		return false, "Error getting plan details", err
	}
	
	// Check URL limit
	if plan.Limits.URLs != -1 && usage.CurrentURLs >= plan.Limits.URLs {
		return false, fmt.Sprintf("URL limit reached (%d/%d). Upgrade to create more URLs.", usage.CurrentURLs, plan.Limits.URLs), nil
	}
	
	return true, "", nil
}

// CanProcessClick checks if user can process more clicks within their plan limits
func (bs *BillingService) CanProcessClick(userID int64) (bool, string, error) {
	usage, err := bs.CheckUsageLimits(userID)
	if err != nil {
		return false, "Error checking usage limits", err
	}
	
	subscription, err := bs.GetUserSubscription(userID)
	if err != nil {
		return false, "Error getting subscription", err
	}
	
	plan, err := bs.GetPlan(subscription.PlanType)
	if err != nil {
		return false, "Error getting plan details", err
	}
	
	// Check click limit
	if plan.Limits.Clicks != -1 && usage.MonthlyClicks >= plan.Limits.Clicks {
		return false, fmt.Sprintf("Monthly click limit reached (%d/%d). Upgrade for more capacity.", usage.MonthlyClicks, plan.Limits.Clicks), nil
	}
	
	return true, "", nil
}

// UpdateSubscription updates a user's subscription
func (bs *BillingService) UpdateSubscription(userID int64, updates map[string]interface{}) error {
	return bs.db.Model(&Subscription{}).Where("user_id = ?", userID).Updates(updates).Error
}

// CreateCheckoutSession creates a Stripe checkout session (mock implementation)
func (bs *BillingService) CreateCheckoutSession(userID int64, planID string, successURL, cancelURL string) (string, error) {
	_, err := bs.GetPlan(planID)
	if err != nil {
		return "", err
	}
	
	// In a real implementation, this would integrate with Stripe
	// For now, return a mock checkout URL
	checkoutURL := fmt.Sprintf("https://checkout.stripe.com/mock/%s/%d", planID, userID)
	
	log.Printf("Mock checkout session created for user %d, plan %s: %s", userID, planID, checkoutURL)
	
	return checkoutURL, nil
}

// WebhookHandler handles Stripe webhooks (mock implementation)
func (bs *BillingService) WebhookHandler(c *gin.Context) {
	// In a real implementation, this would verify the webhook signature
	// and process Stripe events
	
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}
	
	eventType, ok := payload["type"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing event type"})
		return
	}
	
	log.Printf("Received webhook event: %s", eventType)
	
	switch eventType {
	case "checkout.session.completed":
		// Handle successful subscription creation
		bs.handleCheckoutCompleted(payload)
	case "invoice.payment_succeeded":
		// Handle successful payment
		bs.handlePaymentSucceeded(payload)
	case "customer.subscription.updated":
		// Handle subscription updates
		bs.handleSubscriptionUpdated(payload)
	case "customer.subscription.deleted":
		// Handle subscription cancellation
		bs.handleSubscriptionDeleted(payload)
	}
	
	c.JSON(http.StatusOK, gin.H{"received": true})
}

func (bs *BillingService) handleCheckoutCompleted(payload map[string]interface{}) {
	// Mock implementation - in real app, extract customer and subscription info from Stripe
	log.Println("Processing checkout completion...")
}

func (bs *BillingService) handlePaymentSucceeded(payload map[string]interface{}) {
	log.Println("Processing successful payment...")
}

func (bs *BillingService) handleSubscriptionUpdated(payload map[string]interface{}) {
	log.Println("Processing subscription update...")
}

func (bs *BillingService) handleSubscriptionDeleted(payload map[string]interface{}) {
	log.Println("Processing subscription cancellation...")
}

// API Handlers

func (bs *BillingService) GetSubscriptionHandler(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	
	subscription, err := bs.GetUserSubscription(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription"})
		return
	}
	
	c.JSON(http.StatusOK, subscription)
}

func (bs *BillingService) GetPlansHandler(c *gin.Context) {
	plans := bs.GetAllPlans()
	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

func (bs *BillingService) GetUsageHandler(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	
	usage, err := bs.CheckUsageLimits(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get usage"})
		return
	}
	
	subscription, err := bs.GetUserSubscription(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription"})
		return
	}
	
	plan, err := bs.GetPlan(subscription.PlanType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get plan"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"usage": usage,
		"limits": plan.Limits,
		"subscription": subscription,
	})
}

func (bs *BillingService) CreateCheckoutHandler(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	checkoutURL, err := bs.CreateCheckoutSession(userID, req.PlanID, req.SuccessURL, req.CancelURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"checkout_url": checkoutURL})
}

func (bs *BillingService) CancelSubscriptionHandler(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	
	err := bs.UpdateSubscription(userID, map[string]interface{}{
		"cancel_at_end": true,
		"updated_at":   time.Now(),
	})
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel subscription"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Subscription will be cancelled at the end of the current period"})
}