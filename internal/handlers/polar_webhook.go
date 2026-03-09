package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PolarWebhookHandler struct {
	db             *gorm.DB
	billingService *services.BillingService
	polarService   *services.PolarService
}

func NewPolarWebhookHandler(db *gorm.DB, billingService *services.BillingService) *PolarWebhookHandler {
	return &PolarWebhookHandler{
		db:             db,
		billingService: billingService,
		polarService:   billingService.GetPolarService(),
	}
}

// updateUserAccountType syncs users.account_type so JWT-based plan gating reflects the current plan.
func (pwh *PolarWebhookHandler) updateUserAccountType(userID int64, planType string) {
	if err := pwh.db.Exec("UPDATE users SET account_type = ?, updated_at = NOW() WHERE id = ?", planType, userID).Error; err != nil {
		log.Printf("[webhook] failed to update account_type for user %d: %v", userID, err)
	} else {
		log.Printf("[webhook] account_type updated to %q for user %d", planType, userID)
	}
}

// Polar webhook event structures

type PolarWebhookPayload struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	EventID   string          `json:"id"`
	CreatedAt time.Time       `json:"created_at"`
}

type CheckoutEventData struct {
	ID             string                 `json:"id"`
	Status         string                 `json:"status"`
	CustomerID     *string                `json:"customer_id,omitempty"`
	CustomerEmail  *string                `json:"customer_email,omitempty"`
	ProductID      string                 `json:"product_id"`
	ProductPriceID *string                `json:"product_price_id,omitempty"`
	SubscriptionID *string                `json:"subscription_id,omitempty"`
	Amount         *int                   `json:"amount,omitempty"`
	Currency       *string                `json:"currency,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

type SubscriptionEventData struct {
	ID                 string                 `json:"id"`
	Status             string                 `json:"status"`
	CurrentPeriodStart *time.Time             `json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time             `json:"current_period_end,omitempty"`
	CancelAtPeriodEnd  bool                   `json:"cancel_at_period_end"`
	StartedAt          *time.Time             `json:"started_at,omitempty"`
	EndedAt            *time.Time             `json:"ended_at,omitempty"`
	CustomerID         string                 `json:"customer_id"`
	ProductID          string                 `json:"product_id"`
	PriceID            *string                `json:"price_id,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
}

type OrderEventData struct {
	ID             string                 `json:"id"`
	Amount         int                    `json:"amount"`
	Currency       string                 `json:"currency"`
	CustomerID     string                 `json:"customer_id"`
	ProductID      string                 `json:"product_id"`
	ProductPriceID string                 `json:"product_price_id"`
	SubscriptionID *string                `json:"subscription_id,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// HandleWebhook processes incoming Polar webhooks
func (pwh *PolarWebhookHandler) HandleWebhook(c *gin.Context) {
	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Error reading webhook body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Verify webhook signature
	signature := c.GetHeader("Webhook-Signature")
	timestamp := c.GetHeader("Webhook-Timestamp")

	if !pwh.polarService.ValidateWebhookSignature(body, signature, timestamp) {
		log.Printf("Invalid webhook signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Parse webhook payload
	var payload PolarWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("Error parsing webhook payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	// Check for duplicate events (idempotency)
	var existingEvent models.PolarWebhookEvent
	err = pwh.db.Where("polar_event_id = ?", payload.EventID).First(&existingEvent).Error
	if err == nil {
		// Event already processed
		log.Printf("Duplicate webhook event %s, skipping", payload.EventID)
		c.JSON(http.StatusOK, gin.H{"received": true, "status": "duplicate"})
		return
	}

	// Store the webhook event
	webhookEvent := models.PolarWebhookEvent{
		PolarEventID:     payload.EventID,
		EventType:        payload.Type,
		Payload:          models.JSONB{"data": payload.Data, "type": payload.Type},
		ProcessingStatus: "pending",
		CreatedAt:        time.Now(),
	}

	if err := pwh.db.Create(&webhookEvent).Error; err != nil {
		log.Printf("Error storing webhook event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store event"})
		return
	}

	// Process the event
	var processingErr error
	switch payload.Type {
	// Checkout events
	case "checkout.created":
		processingErr = pwh.handleCheckoutCreated(payload.Data)
	case "checkout.updated":
		processingErr = pwh.handleCheckoutUpdated(payload.Data)
	case "checkout.expired":
		processingErr = pwh.handleCheckoutExpired(payload.Data)

	// Subscription events
	case "subscription.created":
		processingErr = pwh.handleSubscriptionCreated(payload.Data)
	case "subscription.updated":
		processingErr = pwh.handleSubscriptionUpdated(payload.Data)
	case "subscription.active":
		processingErr = pwh.handleSubscriptionActive(payload.Data)
	case "subscription.canceled": // FIXED: Polar uses American spelling
		processingErr = pwh.handleSubscriptionCanceled(payload.Data)
	case "subscription.uncanceled":
		processingErr = pwh.handleSubscriptionUncanceled(payload.Data)
	case "subscription.revoked":
		processingErr = pwh.handleSubscriptionRevoked(payload.Data)
	case "subscription.past_due":
		processingErr = pwh.handleSubscriptionPastDue(payload.Data)

	// Order events
	case "order.created":
		processingErr = pwh.handleOrderCreated(payload.Data)
	case "order.paid":
		processingErr = pwh.handleOrderPaid(payload.Data)
	case "order.refunded":
		processingErr = pwh.handleOrderRefunded(payload.Data)

	// Customer events
	case "customer.created":
		processingErr = pwh.handleCustomerCreated(payload.Data)
	case "customer.updated":
		processingErr = pwh.handleCustomerUpdated(payload.Data)

	// Benefit grant events
	case "benefit_grant.created":
		processingErr = pwh.handleBenefitGrantCreated(payload.Data)
	case "benefit_grant.revoked":
		processingErr = pwh.handleBenefitGrantRevoked(payload.Data)

	default:
		log.Printf("Unhandled webhook event type: %s", payload.Type)
		processingErr = nil // Not an error, just unhandled
	}

	// Update webhook event status
	now := time.Now()
	if processingErr != nil {
		errorMsg := processingErr.Error()
		webhookEvent.ProcessingStatus = "failed"
		webhookEvent.ErrorMessage = &errorMsg
		webhookEvent.LastRetryAt = &now // Initialize for retry worker
		log.Printf("⚠️  Error processing webhook %s: %v (queued for retry)", payload.Type, processingErr)
	} else {
		webhookEvent.ProcessingStatus = "processed"
		webhookEvent.ProcessedAt = &now
		log.Printf("✅ Successfully processed webhook %s", payload.Type)
	}

	pwh.db.Save(&webhookEvent)

	// IMPORTANT: Always return 200 to Polar to prevent their retry system
	// Our internal retry worker will handle retries with exponential backoff
	c.JSON(http.StatusOK, gin.H{
		"received": true,
		"status":   webhookEvent.ProcessingStatus,
	})
}

// Event Handlers

func (pwh *PolarWebhookHandler) handleCheckoutCreated(data json.RawMessage) error {
	var checkout CheckoutEventData
	if err := json.Unmarshal(data, &checkout); err != nil {
		return fmt.Errorf("failed to parse checkout data: %w", err)
	}

	log.Printf("Checkout created: %s, status: %s", checkout.ID, checkout.Status)
	return nil
}

func (pwh *PolarWebhookHandler) handleCheckoutUpdated(data json.RawMessage) error {
	var checkout CheckoutEventData
	if err := json.Unmarshal(data, &checkout); err != nil {
		return fmt.Errorf("failed to parse checkout data: %w", err)
	}

	log.Printf("Checkout updated: %s, status: %s", checkout.ID, checkout.Status)

	// If checkout is confirmed and we have a subscription ID, handle it
	if checkout.Status == "confirmed" && checkout.SubscriptionID != nil {
		return pwh.handleCheckoutConfirmed(&checkout)
	}

	return nil
}

func (pwh *PolarWebhookHandler) handleCheckoutConfirmed(checkout *CheckoutEventData) error {
	// Extract user ID from metadata
	userIDFloat, ok := checkout.Metadata["user_id"].(float64)
	if !ok {
		return fmt.Errorf("user_id not found in checkout metadata")
	}
	userID := int64(userIDFloat)

	// Fetch the subscription from Polar
	subscription, err := pwh.polarService.GetSubscription(*checkout.SubscriptionID)
	if err != nil {
		return fmt.Errorf("failed to fetch subscription: %w", err)
	}

	// Determine plan type from product metadata or default mapping
	planType := "premium" // fallback if product fetch fails
	productID := checkout.ProductID

	// Get product to extract benefits and determine plan type
	product, err := pwh.polarService.GetProduct(productID)
	if err != nil {
		log.Printf("Warning: failed to fetch product: %v", err)
	} else {
		planType = pwh.polarService.DeterminePlanType(*product)
	}

	// Convert benefits
	benefits := make([]models.PolarBenefit, 0)
	if product != nil {
		for _, b := range product.Benefits {
			benefitProps, _ := json.Marshal(b.Properties)
			var propsJSON models.JSONB
			json.Unmarshal(benefitProps, &propsJSON)

			benefits = append(benefits, models.PolarBenefit{
				ID:          b.ID,
				Type:        b.Type,
				Description: b.Description,
				Properties:  propsJSON,
			})
		}
	}

	// Create or update subscription in database
	polarSub := models.PolarSubscription{
		UserID:              userID,
		PolarCustomerID:     subscription.CustomerID,
		PolarSubscriptionID: &subscription.ID,
		PolarProductID:      subscription.ProductID,
		PolarPriceID:        subscription.PriceID,
		PolarCheckoutID:     &checkout.ID,
		Status:              subscription.Status,
		PlanType:            planType,
		CurrentPeriodStart:  subscription.CurrentPeriodStart,
		CurrentPeriodEnd:    subscription.CurrentPeriodEnd,
		CancelAtPeriodEnd:   subscription.CancelAtPeriodEnd,
		Benefits:            benefits,
		Metadata:            models.JSONB{"subscription_id": subscription.ID},
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Check if subscription already exists
	var existing models.PolarSubscription
	err = pwh.db.Where("polar_subscription_id = ?", subscription.ID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		// Create new
		err = pwh.db.Create(&polarSub).Error
	} else {
		// Update existing
		polarSub.ID = existing.ID
		err = pwh.db.Save(&polarSub).Error
	}

	if err != nil {
		return fmt.Errorf("failed to save subscription: %w", err)
	}

	pwh.updateUserAccountType(userID, planType)
	log.Printf("Subscription created/updated for user %d: %s", userID, subscription.ID)
	return nil
}

// resolveUserIDFromSubscription extracts userID via 3 fallbacks:
// 1. subscription.Metadata["user_id"] (usually empty — Polar doesn't propagate checkout metadata to subscriptions)
// 2. existing polar_subscriptions row for the same customer (populated by handleCheckoutConfirmed)
// 3. Polar customer API (GetOrCreatePolarCustomer always sets metadata.user_id on the customer object)
func (pwh *PolarWebhookHandler) resolveUserIDFromSubscription(sub *SubscriptionEventData) (int64, error) {
	// 1. Check subscription metadata
	if sub.Metadata != nil {
		if v, ok := sub.Metadata["user_id"].(float64); ok && v > 0 {
			return int64(v), nil
		}
	}

	// 2. Look up existing polar_subscriptions row by customer_id
	var existing models.PolarSubscription
	if err := pwh.db.Where("polar_customer_id = ?", sub.CustomerID).
		Order("created_at DESC").First(&existing).Error; err == nil && existing.UserID > 0 {
		return existing.UserID, nil
	}

	// 3. Fetch customer from Polar API and read metadata
	customer, err := pwh.polarService.GetCustomer(sub.CustomerID)
	if err != nil {
		return 0, fmt.Errorf("user_id not found in subscription metadata and customer lookup failed: %w", err)
	}
	if customer.Metadata != nil {
		if v, ok := customer.Metadata["user_id"].(float64); ok && v > 0 {
			return int64(v), nil
		}
	}
	return 0, fmt.Errorf("user_id not resolvable for customer %s", sub.CustomerID)
}

func (pwh *PolarWebhookHandler) handleSubscriptionCreated(data json.RawMessage) error {
	var subscription SubscriptionEventData
	if err := json.Unmarshal(data, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	log.Printf("Subscription created: %s, customer: %s", subscription.ID, subscription.CustomerID)

	// Resolve user ID via metadata → existing subscription row → Polar customer API
	userID, err := pwh.resolveUserIDFromSubscription(&subscription)
	if err != nil {
		return err
	}

	// Get product to determine plan type and benefits
	planType := "premium"
	benefits := make([]models.PolarBenefit, 0)

	product, err := pwh.polarService.GetProduct(subscription.ProductID)
	if err == nil {
		planType = pwh.polarService.DeterminePlanType(*product)

		// Extract benefits
		for _, b := range product.Benefits {
			benefitProps, _ := json.Marshal(b.Properties)
			var propsJSON models.JSONB
			json.Unmarshal(benefitProps, &propsJSON)

			benefits = append(benefits, models.PolarBenefit{
				ID:          b.ID,
				Type:        b.Type,
				Description: b.Description,
				Properties:  propsJSON,
			})
		}
	}

	// Determine billing cycle
	var billingCycle *string
	if product != nil && len(product.Prices) > 0 {
		for _, price := range product.Prices {
			if subscription.PriceID != nil && price.ID == *subscription.PriceID {
				billingCycle = price.RecurringInterval
				break
			}
		}
	}

	// Create subscription record
	polarSub := models.PolarSubscription{
		UserID:              userID,
		PolarCustomerID:     subscription.CustomerID,
		PolarSubscriptionID: &subscription.ID,
		PolarProductID:      subscription.ProductID,
		PolarPriceID:        subscription.PriceID,
		Status:              subscription.Status,
		PlanType:            planType,
		BillingCycle:        billingCycle,
		CurrentPeriodStart:  subscription.CurrentPeriodStart,
		CurrentPeriodEnd:    subscription.CurrentPeriodEnd,
		CancelAtPeriodEnd:   subscription.CancelAtPeriodEnd,
		Benefits:            benefits,
		Metadata:            models.JSONB{"subscription_id": subscription.ID},
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	err = pwh.db.Create(&polarSub).Error
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	pwh.updateUserAccountType(userID, planType)
	log.Printf("Subscription saved for user %d", userID)
	return nil
}

func (pwh *PolarWebhookHandler) handleSubscriptionUpdated(data json.RawMessage) error {
	var subscription SubscriptionEventData
	if err := json.Unmarshal(data, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	log.Printf("Subscription updated: %s, status: %s", subscription.ID, subscription.Status)

	// Find existing subscription
	var polarSub models.PolarSubscription
	err := pwh.db.Where("polar_subscription_id = ?", subscription.ID).First(&polarSub).Error
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Re-derive plan_type and billing_cycle if product/price changed
	newPlanType := polarSub.PlanType
	newBillingCycle := polarSub.BillingCycle

	if subscription.ProductID != "" && subscription.ProductID != polarSub.PolarProductID {
		// Product changed (tier upgrade/downgrade) — re-derive plan type and billing cycle
		product, err := pwh.polarService.GetProduct(subscription.ProductID)
		if err == nil {
			if pt, ok := product.Metadata["plan_type"].(string); ok && pt != "" {
				newPlanType = pt
			}
			if subscription.PriceID != nil {
				for _, price := range product.Prices {
					if price.ID == *subscription.PriceID {
						newBillingCycle = price.RecurringInterval
						break
					}
				}
			}
		}
	} else if subscription.PriceID != nil && (polarSub.PolarPriceID == nil || *subscription.PriceID != *polarSub.PolarPriceID) {
		// Same product, different price (monthly → yearly switch)
		product, err := pwh.polarService.GetProduct(subscription.ProductID)
		if err == nil {
			for _, price := range product.Prices {
				if price.ID == *subscription.PriceID {
					newBillingCycle = price.RecurringInterval
					break
				}
			}
		}
	}

	// Update fields
	polarSub.Status = subscription.Status
	polarSub.PolarProductID = subscription.ProductID
	polarSub.PolarPriceID = subscription.PriceID
	polarSub.PlanType = newPlanType
	polarSub.BillingCycle = newBillingCycle
	polarSub.CurrentPeriodStart = subscription.CurrentPeriodStart
	polarSub.CurrentPeriodEnd = subscription.CurrentPeriodEnd
	polarSub.CancelAtPeriodEnd = subscription.CancelAtPeriodEnd
	polarSub.EndedAt = subscription.EndedAt
	polarSub.UpdatedAt = time.Now()

	err = pwh.db.Save(&polarSub).Error
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Sync account_type based on new status
	switch subscription.Status {
	case "active", "trialing":
		pwh.updateUserAccountType(polarSub.UserID, polarSub.PlanType)
	case "cancelled", "revoked", "inactive", "past_due":
		// Only downgrade if not just flagged for future cancellation
		if !subscription.CancelAtPeriodEnd {
			pwh.updateUserAccountType(polarSub.UserID, "free")
		}
	}
	log.Printf("Subscription updated in database")
	return nil
}

func (pwh *PolarWebhookHandler) handleSubscriptionCancelled(data json.RawMessage) error {
	var subscription SubscriptionEventData
	if err := json.Unmarshal(data, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	log.Printf("Subscription cancelled: %s", subscription.ID)

	// Find existing subscription
	var polarSub models.PolarSubscription
	err := pwh.db.Where("polar_subscription_id = ?", subscription.ID).First(&polarSub).Error
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Update status to cancelled
	now := time.Now()
	polarSub.Status = "cancelled"
	polarSub.CancelledAt = &now
	polarSub.CancelAtPeriodEnd = subscription.CancelAtPeriodEnd
	polarSub.UpdatedAt = time.Now()

	err = pwh.db.Save(&polarSub).Error
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	// Only downgrade immediately if cancel_at_period_end is false (immediate cancellation)
	if !subscription.CancelAtPeriodEnd {
		pwh.updateUserAccountType(polarSub.UserID, "free")
	}
	log.Printf("Subscription marked as cancelled in database")
	return nil
}

func (pwh *PolarWebhookHandler) handleOrderCreated(data json.RawMessage) error {
	var order OrderEventData
	if err := json.Unmarshal(data, &order); err != nil {
		return fmt.Errorf("failed to parse order data: %w", err)
	}

	log.Printf("Order created: %s, amount: %d %s", order.ID, order.Amount, order.Currency)

	// If this is a subscription order, it will be handled by subscription.created event
	// For one-time purchases, you can add custom logic here

	return nil
}

// New Event Handlers

func (pwh *PolarWebhookHandler) handleCheckoutExpired(data json.RawMessage) error {
	var checkout CheckoutEventData
	if err := json.Unmarshal(data, &checkout); err != nil {
		return fmt.Errorf("failed to parse checkout data: %w", err)
	}

	log.Printf("Checkout expired: %s - abandoned by user", checkout.ID)

	// Track abandoned checkouts for analytics
	// You can send reminder emails here or track conversion funnel dropoff

	return nil
}

func (pwh *PolarWebhookHandler) handleSubscriptionActive(data json.RawMessage) error {
	var subscription SubscriptionEventData
	if err := json.Unmarshal(data, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	log.Printf("Subscription activated: %s, customer: %s", subscription.ID, subscription.CustomerID)

	// Find existing subscription
	var polarSub models.PolarSubscription
	err := pwh.db.Where("polar_subscription_id = ?", subscription.ID).First(&polarSub).Error
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Update status to active
	polarSub.Status = "active"
	polarSub.UpdatedAt = time.Now()

	err = pwh.db.Save(&polarSub).Error
	if err != nil {
		return fmt.Errorf("failed to activate subscription: %w", err)
	}

	pwh.updateUserAccountType(polarSub.UserID, polarSub.PlanType)
	log.Printf("Subscription marked as active - user features enabled")
	return nil
}

// handleSubscriptionCanceled - American spelling version (Polar uses this)
func (pwh *PolarWebhookHandler) handleSubscriptionCanceled(data json.RawMessage) error {
	// Delegate to existing handler with British spelling
	return pwh.handleSubscriptionCancelled(data)
}

func (pwh *PolarWebhookHandler) handleSubscriptionUncanceled(data json.RawMessage) error {
	var subscription SubscriptionEventData
	if err := json.Unmarshal(data, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	log.Printf("Subscription reactivated: %s", subscription.ID)

	// Find existing subscription
	var polarSub models.PolarSubscription
	err := pwh.db.Where("polar_subscription_id = ?", subscription.ID).First(&polarSub).Error
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Reactivate subscription
	polarSub.Status = subscription.Status
	polarSub.CancelAtPeriodEnd = false
	polarSub.CancelledAt = nil
	polarSub.UpdatedAt = time.Now()

	err = pwh.db.Save(&polarSub).Error
	if err != nil {
		return fmt.Errorf("failed to reactivate subscription: %w", err)
	}

	pwh.updateUserAccountType(polarSub.UserID, polarSub.PlanType)
	log.Printf("Subscription reactivated successfully")
	return nil
}

func (pwh *PolarWebhookHandler) handleSubscriptionRevoked(data json.RawMessage) error {
	var subscription SubscriptionEventData
	if err := json.Unmarshal(data, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	log.Printf("⚠️  Subscription revoked immediately: %s (fraud/abuse)", subscription.ID)

	// Find existing subscription
	var polarSub models.PolarSubscription
	err := pwh.db.Where("polar_subscription_id = ?", subscription.ID).First(&polarSub).Error
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Hard cancel - revoke access immediately
	now := time.Now()
	polarSub.Status = "revoked"
	polarSub.CancelledAt = &now
	polarSub.EndedAt = &now
	polarSub.UpdatedAt = time.Now()

	err = pwh.db.Save(&polarSub).Error
	if err != nil {
		return fmt.Errorf("failed to revoke subscription: %w", err)
	}

	pwh.updateUserAccountType(polarSub.UserID, "free")
	log.Printf("⛔ Subscription revoked - access terminated immediately")
	// TODO: Send notification to user about revocation
	return nil
}

func (pwh *PolarWebhookHandler) handleSubscriptionPastDue(data json.RawMessage) error {
	var subscription SubscriptionEventData
	if err := json.Unmarshal(data, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	log.Printf("⚠️  Subscription past due: %s - payment failed", subscription.ID)

	// Find existing subscription
	var polarSub models.PolarSubscription
	err := pwh.db.Where("polar_subscription_id = ?", subscription.ID).First(&polarSub).Error
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Update status to past_due
	polarSub.Status = "past_due"
	polarSub.UpdatedAt = time.Now()

	err = pwh.db.Save(&polarSub).Error
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	log.Printf("⚠️  Subscription marked as past_due - user should update payment method")
	// TODO: Send email notification to user about failed payment
	// TODO: Consider limiting feature access or showing payment reminder banner
	return nil
}

func (pwh *PolarWebhookHandler) handleOrderPaid(data json.RawMessage) error {
	var order OrderEventData
	if err := json.Unmarshal(data, &order); err != nil {
		return fmt.Errorf("failed to parse order data: %w", err)
	}

	log.Printf("✅ Order paid successfully: %s, amount: %d %s", order.ID, order.Amount, order.Currency)

	// Payment confirmed - subscription should be activated via subscription.active event
	// This is mainly for logging and analytics

	return nil
}

func (pwh *PolarWebhookHandler) handleOrderRefunded(data json.RawMessage) error {
	var order OrderEventData
	if err := json.Unmarshal(data, &order); err != nil {
		return fmt.Errorf("failed to parse order data: %w", err)
	}

	log.Printf("💰 Order refunded: %s, amount: %d %s", order.ID, order.Amount, order.Currency)

	// If this order has a subscription, handle subscription cancellation
	if order.SubscriptionID != nil {
		var polarSub models.PolarSubscription
		err := pwh.db.Where("polar_subscription_id = ?", *order.SubscriptionID).First(&polarSub).Error
		if err == nil {
			// Mark subscription as refunded/cancelled
			now := time.Now()
			polarSub.Status = "cancelled"
			polarSub.CancelledAt = &now
			polarSub.UpdatedAt = time.Now()
			pwh.db.Save(&polarSub)
			log.Printf("Subscription %s cancelled due to refund", *order.SubscriptionID)
		}
	}

	// TODO: Send refund confirmation email to user
	return nil
}

func (pwh *PolarWebhookHandler) handleCustomerCreated(data json.RawMessage) error {
	// Parse customer data
	var customerData map[string]interface{}
	if err := json.Unmarshal(data, &customerData); err != nil {
		return fmt.Errorf("failed to parse customer data: %w", err)
	}

	customerID, _ := customerData["id"].(string)
	email, _ := customerData["email"].(string)

	log.Printf("Customer created in Polar: %s (%s)", customerID, email)

	// Customer records are tracked via subscriptions
	// This is mainly for logging

	return nil
}

func (pwh *PolarWebhookHandler) handleCustomerUpdated(data json.RawMessage) error {
	// Parse customer data
	var customerData map[string]interface{}
	if err := json.Unmarshal(data, &customerData); err != nil {
		return fmt.Errorf("failed to parse customer data: %w", err)
	}

	customerID, _ := customerData["id"].(string)

	log.Printf("Customer updated in Polar: %s", customerID)

	// Sync any relevant customer changes to subscriptions
	// Email changes, metadata updates, etc.

	return nil
}

func (pwh *PolarWebhookHandler) handleBenefitGrantCreated(data json.RawMessage) error {
	// Parse benefit grant data
	var grantData map[string]interface{}
	if err := json.Unmarshal(data, &grantData); err != nil {
		return fmt.Errorf("failed to parse benefit grant data: %w", err)
	}

	subscriptionID, _ := grantData["subscription_id"].(string)
	benefitID, _ := grantData["benefit_id"].(string)

	log.Printf("✅ Benefit granted: %s for subscription %s", benefitID, subscriptionID)

	// Fetch and update subscription benefits
	// This ensures benefits are in sync

	return nil
}

func (pwh *PolarWebhookHandler) handleBenefitGrantRevoked(data json.RawMessage) error {
	// Parse benefit grant data
	var grantData map[string]interface{}
	if err := json.Unmarshal(data, &grantData); err != nil {
		return fmt.Errorf("failed to parse benefit grant data: %w", err)
	}

	subscriptionID, _ := grantData["subscription_id"].(string)
	benefitID, _ := grantData["benefit_id"].(string)

	log.Printf("❌ Benefit revoked: %s for subscription %s", benefitID, subscriptionID)

	// Remove benefit from subscription record
	// Disable associated features

	return nil
}
