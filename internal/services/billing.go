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
func (bs *BillingService) CreateCheckoutSession(userID int64, email string, productID string, priceID *string, successURL, cancelURL string) (*PolarCheckoutSession, error) {
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
	checkout, err := bs.polarService.CreateCheckoutSession(productID, priceID, email, successURL, metadata)
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

	checkoutURL, err := bs.CreateCheckoutSession(userID, user.Email, req.PlanID, nil, req.SuccessURL, req.CancelURL)
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

// ============================================================================
// ADMIN OPERATIONS
// ============================================================================

// AdminListSubscriptions lists all subscriptions with filtering and pagination
func (bs *BillingService) AdminListSubscriptions(filter models.SubscriptionFilter, page, limit int) ([]*models.PolarSubscription, int64, error) {
	var subscriptions []*models.PolarSubscription
	var total int64

	// Build base query
	query := bs.db.Model(&models.PolarSubscription{})

	// Apply filters
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.PlanType != "" {
		query = query.Where("plan_type = ?", filter.PlanType)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		query = query.Where("polar_customer_id LIKE ? OR polar_subscription_id LIKE ?",
			"%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := filter.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&subscriptions).Error; err != nil {
		return nil, 0, err
	}

	return subscriptions, total, nil
}

// AdminGetSubscription retrieves a specific subscription by ID
func (bs *BillingService) AdminGetSubscription(subscriptionID int64) (*models.PolarSubscription, error) {
	var subscription models.PolarSubscription
	err := bs.db.Where("id = ?", subscriptionID).First(&subscription).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, err
	}
	return &subscription, nil
}

// AdminUpdateSubscription updates a subscription (admin-initiated)
func (bs *BillingService) AdminUpdateSubscription(subscriptionID int64, req *models.AdminUpdateSubscriptionRequest, updatedBy int64) error {
	updates := make(map[string]interface{})

	if req.PlanType != nil {
		updates["plan_type"] = *req.PlanType
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.TrialEnd != nil {
		updates["trial_end"] = *req.TrialEnd
	}
	if req.CancelAtEnd != nil {
		updates["cancel_at_period_end"] = *req.CancelAtEnd
	}

	updates["updated_at"] = time.Now()

	result := bs.db.Model(&models.PolarSubscription{}).
		Where("id = ?", subscriptionID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}

// AdminCancelSubscription cancels a subscription (admin-initiated)
func (bs *BillingService) AdminCancelSubscription(subscriptionID int64, reason string, immediate bool, cancelledBy int64) error {
	updates := make(map[string]interface{})

	if immediate {
		updates["status"] = "cancelled"
		updates["cancelled_at"] = time.Now()
		updates["ended_at"] = time.Now()
	} else {
		updates["cancel_at_period_end"] = true
	}

	// Store reason in metadata
	var subscription models.PolarSubscription
	if err := bs.db.Where("id = ?", subscriptionID).First(&subscription).Error; err != nil {
		return fmt.Errorf("subscription not found")
	}

	if subscription.Metadata == nil {
		subscription.Metadata = make(models.JSONB)
	}
	subscription.Metadata["cancellation_reason"] = reason
	subscription.Metadata["cancelled_by_admin"] = cancelledBy
	updates["metadata"] = subscription.Metadata
	updates["updated_at"] = time.Now()

	result := bs.db.Model(&models.PolarSubscription{}).
		Where("id = ?", subscriptionID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}

// AdminRefundSubscription marks a refund (does not call Polar API)
func (bs *BillingService) AdminRefundSubscription(subscriptionID int64, req *models.RefundSubscriptionRequest, refundedBy int64) error {
	// Get subscription
	var subscription models.PolarSubscription
	if err := bs.db.Where("id = ?", subscriptionID).First(&subscription).Error; err != nil {
		return fmt.Errorf("subscription not found")
	}

	// Store refund in metadata
	if subscription.Metadata == nil {
		subscription.Metadata = make(models.JSONB)
	}

	refundData := map[string]interface{}{
		"amount":      req.Amount,
		"reason":      req.Reason,
		"refunded_by": refundedBy,
		"refunded_at": time.Now(),
	}

	subscription.Metadata["refund"] = refundData

	// Update subscription
	result := bs.db.Model(&models.PolarSubscription{}).
		Where("id = ?", subscriptionID).
		Update("metadata", subscription.Metadata)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// AdminGetRevenue calculates revenue analytics
func (bs *BillingService) AdminGetRevenue(startDate, endDate time.Time) (map[string]interface{}, error) {
	// Count active subscriptions
	var activeCount int64
	bs.db.Model(&models.PolarSubscription{}).
		Where("status = ?", "active").
		Count(&activeCount)

	// Count trial subscriptions
	var trialCount int64
	bs.db.Model(&models.PolarSubscription{}).
		Where("status = ?", "trialing").
		Count(&trialCount)

	// Count subscriptions by plan type
	var planStats []struct {
		PlanType string
		Count    int64
	}
	bs.db.Model(&models.PolarSubscription{}).
		Select("plan_type, COUNT(*) as count").
		Where("status IN (?)", []string{"active", "trialing"}).
		Group("plan_type").
		Scan(&planStats)

	// Calculate MRR (simplified - would need product prices)
	planRevenue := map[string]float64{
		"free":       0,
		"premium":    9.99,
		"enterprise": 49.99,
	}

	var mrr float64
	for _, stat := range planStats {
		if price, ok := planRevenue[stat.PlanType]; ok {
			mrr += price * float64(stat.Count)
		}
	}

	return map[string]interface{}{
		"active_subscriptions": activeCount,
		"trial_subscriptions":  trialCount,
		"mrr":                  mrr,
		"arr":                  mrr * 12,
		"plan_breakdown":       planStats,
		"period_start":         startDate,
		"period_end":           endDate,
	}, nil
}

// AdminListPlans lists all billing plans (from polar_products)
func (bs *BillingService) AdminListPlans(includeArchived bool) ([]*models.PolarProduct, error) {
	var products []*models.PolarProduct

	query := bs.db.Model(&models.PolarProduct{})
	if !includeArchived {
		query = query.Where("is_archived = ?", false)
	}

	if err := query.Order("plan_type ASC, created_at DESC").Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

// AdminArchivePlan archives a billing plan
func (bs *BillingService) AdminArchivePlan(productID string, archivedBy int64) error {
	result := bs.db.Model(&models.PolarProduct{}).
		Where("id = ? OR polar_product_id = ?", productID, productID).
		Update("is_archived", true)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("plan not found")
	}

	return nil
}

// AdminListWebhookEvents lists webhook events with filtering
func (bs *BillingService) AdminListWebhookEvents(filter map[string]interface{}, page, limit int) ([]*models.PolarWebhookEvent, int64, error) {
	var events []*models.PolarWebhookEvent
	var total int64

	query := bs.db.Model(&models.PolarWebhookEvent{})

	// Apply filters
	if eventType, ok := filter["event_type"].(string); ok && eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if status, ok := filter["status"].(string); ok && status != "" {
		query = query.Where("processing_status = ?", status)
	}
	if startDate, ok := filter["start_date"].(time.Time); ok {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filter["end_date"].(time.Time); ok {
		query = query.Where("created_at <= ?", endDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// AdminRetryWebhook retries a failed webhook event
func (bs *BillingService) AdminRetryWebhook(eventID int64, retriedBy int64) error {
	var event models.PolarWebhookEvent
	if err := bs.db.Where("id = ?", eventID).First(&event).Error; err != nil {
		return fmt.Errorf("webhook event not found")
	}

	// Update retry information
	now := time.Now()
	updates := map[string]interface{}{
		"processing_status": "pending",
		"retry_count":       event.RetryCount + 1,
		"last_retry_at":     &now,
		"error_message":     nil,
	}

	result := bs.db.Model(&models.PolarWebhookEvent{}).
		Where("id = ?", eventID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	// Note: Actual webhook reprocessing would be triggered by webhook retry worker
	log.Printf("Admin %d triggered retry for webhook event %d (retry count: %d)", retriedBy, eventID, event.RetryCount+1)

	return nil
}

// AdminSyncPolar manually triggers sync with Polar API
func (bs *BillingService) AdminSyncPolar(triggeredBy int64) (map[string]interface{}, error) {
	log.Printf("Admin %d triggered Polar product sync", triggeredBy)

	// Call existing PolarService sync method
	err := bs.polarService.SyncProductsFromPolar()
	if err != nil {
		return nil, fmt.Errorf("failed to sync products: %w", err)
	}

	// Count synced products
	var productCount int64
	bs.db.Model(&models.PolarProduct{}).Count(&productCount)

	return map[string]interface{}{
		"synced_products": productCount,
		"synced_at":       time.Now(),
		"triggered_by":    triggeredBy,
	}, nil
}