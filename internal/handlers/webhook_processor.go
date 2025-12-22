package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/URLshorter/url-shortener/internal/models"
)

// WebhookEventProcessor implements webhook processing logic
type WebhookEventProcessor struct {
	handler *PolarWebhookHandler
}

// NewWebhookEventProcessor creates a new webhook event processor
func NewWebhookEventProcessor(handler *PolarWebhookHandler) *WebhookEventProcessor {
	return &WebhookEventProcessor{
		handler: handler,
	}
}

// ProcessWebhookEvent processes a webhook event (implements workers.WebhookProcessor interface)
func (p *WebhookEventProcessor) ProcessWebhookEvent(event *models.PolarWebhookEvent) error {
	// Extract the data from the payload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var payload struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Route to the appropriate handler based on event type
	switch event.EventType {
	case "checkout.created":
		return p.handler.handleCheckoutCreated(payload.Data)
	case "checkout.updated":
		return p.handler.handleCheckoutUpdated(payload.Data)
	case "subscription.created":
		return p.handler.handleSubscriptionCreated(payload.Data)
	case "subscription.updated":
		return p.handler.handleSubscriptionUpdated(payload.Data)
	case "subscription.cancelled":
		return p.handler.handleSubscriptionCancelled(payload.Data)
	case "order.created":
		return p.handler.handleOrderCreated(payload.Data)
	default:
		// Unhandled event type - not an error
		return nil
	}
}
