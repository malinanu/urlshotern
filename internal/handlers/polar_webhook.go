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
	case "checkout.created":
		processingErr = pwh.handleCheckoutCreated(payload.Data)
	case "checkout.updated":
		processingErr = pwh.handleCheckoutUpdated(payload.Data)
	case "subscription.created":
		processingErr = pwh.handleSubscriptionCreated(payload.Data)
	case "subscription.updated":
		processingErr = pwh.handleSubscriptionUpdated(payload.Data)
	case "subscription.cancelled":
		processingErr = pwh.handleSubscriptionCancelled(payload.Data)
	case "order.created":
		processingErr = pwh.handleOrderCreated(payload.Data)
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
		webhookEvent.ProcessedAt = &now
		log.Printf("Error processing webhook %s: %v", payload.Type, processingErr)
	} else {
		webhookEvent.ProcessingStatus = "processed"
		webhookEvent.ProcessedAt = &now
	}

	pwh.db.Save(&webhookEvent)

	if processingErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
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
	planType := "premium" // default
	productID := checkout.ProductID
	// You can add logic here to map product IDs to plan types

	// Get product to extract benefits
	product, err := pwh.polarService.GetProduct(productID)
	if err != nil {
		log.Printf("Warning: failed to fetch product: %v", err)
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

	log.Printf("Subscription created/updated for user %d: %s", userID, subscription.ID)
	return nil
}

func (pwh *PolarWebhookHandler) handleSubscriptionCreated(data json.RawMessage) error {
	var subscription SubscriptionEventData
	if err := json.Unmarshal(data, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	log.Printf("Subscription created: %s, customer: %s", subscription.ID, subscription.CustomerID)

	// Extract user ID from metadata
	userIDFloat, ok := subscription.Metadata["user_id"].(float64)
	if !ok {
		return fmt.Errorf("user_id not found in subscription metadata")
	}
	userID := int64(userIDFloat)

	// Get product to determine plan type and benefits
	planType := "premium"
	benefits := make([]models.PolarBenefit, 0)

	product, err := pwh.polarService.GetProduct(subscription.ProductID)
	if err == nil && product.Metadata != nil {
		if pt, ok := product.Metadata["plan_type"].(string); ok {
			planType = pt
		}

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

	// Update fields
	polarSub.Status = subscription.Status
	polarSub.CurrentPeriodStart = subscription.CurrentPeriodStart
	polarSub.CurrentPeriodEnd = subscription.CurrentPeriodEnd
	polarSub.CancelAtPeriodEnd = subscription.CancelAtPeriodEnd
	polarSub.EndedAt = subscription.EndedAt
	polarSub.UpdatedAt = time.Now()

	err = pwh.db.Save(&polarSub).Error
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
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
