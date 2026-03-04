package services

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BillingService struct {
	db           *gorm.DB
	polarService *PolarService
	config       *configs.Config
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
	URLs                int  `json:"max_urls"`           // -1 for unlimited
	Clicks              int  `json:"max_clicks_month"`   // -1 for unlimited
	Analytics           int  `json:"analytics_days"`     // days of analytics retention
	CustomDomain        bool `json:"custom_domain"`
	APIAccess           bool `json:"api_access"`
	BulkImport          bool `json:"bulk_import"`
	QRCodes             bool `json:"qr_codes"`
	CustomShortCodes    bool `json:"custom_short_codes"`
	TeamMembers         int  `json:"team_members"`       // 0 = none, -1 = unlimited
	PasswordProtection  bool `json:"password_protection"`
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
			"10 short links",
			"Basic analytics",
			"Standard support",
			"Link expiration",
			"QR codes",
		},
		Limits: PlanLimits{
			URLs:               10,
			Clicks:             5000,
			Analytics:          30,
			CustomDomain:       false,
			APIAccess:          false,
			BulkImport:         false,
			QRCodes:            true,
			CustomShortCodes:   false,
			TeamMembers:        0,
			PasswordProtection: false,
			Advanced:           false,
		},
	},
	"premium": {
		ID:          "premium",
		Name:        "Pro",
		Description: "For professionals and small teams",
		Price:       9.99,
		Currency:    "usd",
		Interval:    "month",
		Features: []string{
			"1,000 short links",
			"Advanced analytics",
			"Priority support",
			"Link expiration",
			"QR codes",
			"Custom short codes",
			"Custom domains",
			"Team collaboration (5 users)",
			"API access",
			"Password protection",
		},
		Limits: PlanLimits{
			URLs:               1000,
			Clicks:             100000,
			Analytics:          365,
			CustomDomain:       true,
			APIAccess:          true,
			BulkImport:         true,
			QRCodes:            true,
			CustomShortCodes:   true,
			TeamMembers:        5,
			PasswordProtection: true,
			Advanced:           true,
		},
	},
	"enterprise": {
		ID:          "enterprise",
		Name:        "Ultra",
		Description: "For large organizations",
		Price:       49.99,
		Currency:    "usd",
		Interval:    "month",
		Features: []string{
			"Unlimited short links",
			"Advanced analytics + AI insights",
			"24/7 premium support",
			"Link expiration",
			"QR codes",
			"Custom short codes",
			"Custom domains",
			"Unlimited team members",
			"Unlimited API access",
			"Password protection",
			"White-label branding",
		},
		Limits: PlanLimits{
			URLs:               -1,
			Clicks:             -1,
			Analytics:          -1,
			CustomDomain:       true,
			APIAccess:          true,
			BulkImport:         true,
			QRCodes:            true,
			CustomShortCodes:   true,
			TeamMembers:        -1,
			PasswordProtection: true,
			Advanced:           true,
		},
	},
	"premium_yearly": {
		ID:          "premium_yearly",
		Name:        "Pro (Yearly)",
		Description: "For professionals and small teams (2 months free)",
		Price:       99.90,
		Currency:    "usd",
		Interval:    "year",
		Features: []string{
			"1,000 short links",
			"Advanced analytics",
			"Priority support",
			"Link expiration",
			"QR codes",
			"Custom short codes",
			"Custom domains",
			"Team collaboration (5 users)",
			"API access",
			"Password protection",
		},
		Limits: PlanLimits{
			URLs:               1000,
			Clicks:             100000,
			Analytics:          365,
			CustomDomain:       true,
			APIAccess:          true,
			BulkImport:         true,
			QRCodes:            true,
			CustomShortCodes:   true,
			TeamMembers:        5,
			PasswordProtection: true,
			Advanced:           true,
		},
	},
	"enterprise_yearly": {
		ID:          "enterprise_yearly",
		Name:        "Ultra (Yearly)",
		Description: "For large organizations (2 months free)",
		Price:       499.90,
		Currency:    "usd",
		Interval:    "year",
		Features: []string{
			"Unlimited short links",
			"Advanced analytics + AI insights",
			"24/7 premium support",
			"Link expiration",
			"QR codes",
			"Custom short codes",
			"Custom domains",
			"Unlimited team members",
			"Unlimited API access",
			"Password protection",
			"White-label branding",
		},
		Limits: PlanLimits{
			URLs:               -1,
			Clicks:             -1,
			Analytics:          -1,
			CustomDomain:       true,
			APIAccess:          true,
			BulkImport:         true,
			QRCodes:            true,
			CustomShortCodes:   true,
			TeamMembers:        -1,
			PasswordProtection: true,
			Advanced:           true,
		},
	},
}

func NewBillingService(db *gorm.DB, config *configs.Config) *BillingService {
	polarService := NewPolarService(db, config)
	return &BillingService{
		db:           db,
		polarService: polarService,
		config:       config,
	}
}

// GetUserSubscription retrieves the current subscription for a user (Polar-based)
func (bs *BillingService) GetUserSubscription(userID int64) (*models.PolarSubscription, error) {
	var subscription models.PolarSubscription
	err := bs.db.Where("user_id = ? AND (status = ? OR status = ?)", userID, "active", "trialing").
		Order("created_at DESC").
		First(&subscription).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return free plan subscription
			return bs.CreateFreePolarSubscription(userID)
		}
		return nil, err
	}
	return &subscription, nil
}

// CreateFreePolarSubscription creates a free subscription for new users
func (bs *BillingService) CreateFreePolarSubscription(userID int64) (*models.PolarSubscription, error) {
	subscription := &models.PolarSubscription{
		UserID:           userID,
		PolarCustomerID:  fmt.Sprintf("free_user_%d", userID), // Placeholder for free users
		PolarProductID:   "free_plan",
		Status:           "active",
		PlanType:         "free",
		Benefits:         []models.PolarBenefit{},
		Metadata:         models.JSONB{"plan": "free"},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
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

// CreateCheckoutSession creates a Polar checkout session
func (bs *BillingService) CreateCheckoutSession(userID int64, email string, productID string, priceID *string, discountID *string, successURL, cancelURL string) (*PolarCheckoutSession, error) {
	// Get or create Polar customer
	customerID, err := bs.polarService.GetOrCreatePolarCustomer(userID, email, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create customer: %w", err)
	}

	// Create metadata
	metadata := map[string]interface{}{
		"user_id":    userID,
		"product_id": productID,
	}

	// Create checkout session via Polar
	checkout, err := bs.polarService.CreateCheckoutSession(productID, priceID, discountID, email, successURL, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout: %w", err)
	}

	// Save checkout ID to database for tracking
	var subscription models.PolarSubscription
	checkoutID := checkout.ID
	err = bs.db.Model(&subscription).
		Where("user_id = ? AND polar_customer_id = ?", userID, customerID).
		Update("polar_checkout_id", checkoutID).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Warning: failed to save checkout ID: %v", err)
	}

	log.Printf("Polar checkout session created for user %d, product %s: %s", userID, productID, checkout.URL)

	return checkout, nil
}

// GetPolarService returns the underlying Polar service
func (bs *BillingService) GetPolarService() *PolarService {
	return bs.polarService
}

// GetDB returns the database connection
func (bs *BillingService) GetDB() *gorm.DB {
	return bs.db
}

// ValidateCoupon validates a coupon code and returns it if valid
func (bs *BillingService) ValidateCoupon(code string, planID string) (*models.Coupon, error) {
	var coupon models.Coupon
	if err := bs.db.Where("code = ? AND is_active = true", code).First(&coupon).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid or expired coupon code")
		}
		return nil, err
	}

	now := time.Now()
	if coupon.ValidFrom.After(now) {
		return nil, fmt.Errorf("coupon is not yet valid")
	}

	if coupon.ValidUntil != nil && coupon.ValidUntil.Before(now) {
		return nil, fmt.Errorf("coupon has expired")
	}

	if coupon.MaxRedemptions != nil && coupon.CurrentRedemptions >= *coupon.MaxRedemptions {
		return nil, fmt.Errorf("coupon redemption limit reached")
	}

	if len(coupon.ApplicablePlans) > 0 && planID != "" {
		planIsApplicable := false
		for _, p := range coupon.ApplicablePlans {
			if p == planID {
				planIsApplicable = true
				break
			}
		}
		if !planIsApplicable {
			return nil, fmt.Errorf("coupon is not applicable for this plan")
		}
	}

	return &coupon, nil
}

func (bs *BillingService) RecordCouponRedemption(tx *gorm.DB, couponID int64, userID int64, subscriptionID *int64) error {
	redemption := models.CouponRedemption{
		CouponID:       couponID,
		UserID:         userID,
		SubscriptionID: subscriptionID,
		RedeemedAt:     time.Now(),
	}

	if err := tx.Create(&redemption).Error; err != nil {
		return err
	}

	return tx.Model(&models.Coupon{}).Where("id = ?", couponID).UpdateColumn("current_redemptions", gorm.Expr("current_redemptions + ?", 1)).Error
}

// RedeemFreeMonths immediately activates a subscription without a checkout session
func (bs *BillingService) RedeemFreeMonths(userID int64, planID string, coupon *models.Coupon) error {
	plan, err := bs.GetPlan(planID)
	if err != nil {
		return err
	}

	monthsToAdd := coupon.DiscountValue
	if monthsToAdd <= 0 {
		return fmt.Errorf("invalid free months value")
	}

	return bs.db.Transaction(func(tx *gorm.DB) error {
		// Calculate the end period based on current sub if it exists and is active, else start from now
		var currentSub models.PolarSubscription
		err := tx.Where("user_id = ? AND status IN ?", userID, []string{"active", "trialing"}).Order("created_at DESC").First(&currentSub).Error
		
		now := time.Now()
		var startPeriod time.Time = now
		var endPeriod time.Time

		if err == nil {
			if currentSub.CurrentPeriodEnd != nil && currentSub.CurrentPeriodEnd.After(now) {
				startPeriod = *currentSub.CurrentPeriodEnd
			} else {
				startPeriod = now
			}
		}

		endPeriod = startPeriod.AddDate(0, monthsToAdd, 0)
		
		// Map planID to a generic polar plan ID for free override or something
		// Assuming we bypass Polar for this totally free period.
		
		fakePolarSubID := fmt.Sprintf("sub_free_pass_%d_%d", userID, time.Now().Unix())

		newSub := models.PolarSubscription{
			UserID:              userID,
			PolarCustomerID:     fmt.Sprintf("free_user_%d", userID),
			PolarSubscriptionID: &fakePolarSubID,
			PolarProductID:      planID,
			Status:              "active",
			PlanType:            plan.ID,
			CurrentPeriodStart:  &startPeriod,
			CurrentPeriodEnd:    &endPeriod,
			Metadata:            models.JSONB{"coupon_id": coupon.ID, "free_months": true},
			CreatedAt:           now,
			UpdatedAt:           now,
		}

		if err := tx.Create(&newSub).Error; err != nil {
			return err
		}

		return bs.RecordCouponRedemption(tx, coupon.ID, userID, &newSub.ID)
	})
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

	// Get user email from database
	var user models.User
	if err := bs.db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return
	}

	checkoutURL, err := bs.CreateCheckoutSession(userID, user.Email, req.PlanID, nil, nil, req.SuccessURL, req.CancelURL)
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