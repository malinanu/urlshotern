package workers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"gorm.io/gorm"
)

const (
	MaxRetries           = 5
	InitialRetryInterval = 15 * time.Minute
)

// WebhookRetryWorker handles automatic retry of failed webhook events
type WebhookRetryWorker struct {
	db             *gorm.DB
	webhookHandler WebhookProcessor
	retryInterval  time.Duration
	stopChan       chan struct{}
}

// WebhookProcessor interface for processing webhooks
type WebhookProcessor interface {
	ProcessWebhookEvent(event *models.PolarWebhookEvent) error
}

// NewWebhookRetryWorker creates a new webhook retry worker
func NewWebhookRetryWorker(db *gorm.DB, webhookHandler WebhookProcessor) *WebhookRetryWorker {
	return &WebhookRetryWorker{
		db:             db,
		webhookHandler: webhookHandler,
		retryInterval:  5 * time.Minute, // Check every 5 minutes
		stopChan:       make(chan struct{}),
	}
}

// Start begins the retry worker loop
func (w *WebhookRetryWorker) Start(ctx context.Context) {
	log.Println("🔄 Starting webhook retry worker...")
	ticker := time.NewTicker(w.retryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.retryFailedWebhooks(ctx)
		case <-ctx.Done():
			log.Println("⏹️  Webhook retry worker stopped")
			return
		case <-w.stopChan:
			log.Println("⏹️  Webhook retry worker stopped")
			return
		}
	}
}

// Stop stops the retry worker
func (w *WebhookRetryWorker) Stop() {
	close(w.stopChan)
}

// retryFailedWebhooks finds and retries failed webhook events
func (w *WebhookRetryWorker) retryFailedWebhooks(ctx context.Context) {
	var failedEvents []models.PolarWebhookEvent

	// Find events eligible for retry
	// - Status is "failed"
	// - Retry count < max retries
	// - Last retry was more than exponential backoff time ago
	result := w.db.Where("processing_status = ? AND retry_count < ?", "failed", MaxRetries).
		Find(&failedEvents)

	if result.Error != nil {
		log.Printf("❌ Error querying failed webhooks: %v", result.Error)
		return
	}

	if len(failedEvents) == 0 {
		return
	}

	log.Printf("🔍 Found %d failed webhooks eligible for retry", len(failedEvents))

	for _, event := range failedEvents {
		// Check if enough time has passed for retry (exponential backoff)
		if !w.shouldRetryNow(&event) {
			continue
		}

		log.Printf("♻️  Retrying webhook event %s (attempt %d/%d)", event.PolarEventID, event.RetryCount+1, MaxRetries)

		// Attempt to process the webhook again
		err := w.webhookHandler.ProcessWebhookEvent(&event)

		// Update retry metadata
		now := time.Now()
		event.LastRetryAt = &now
		event.RetryCount++

		if err != nil {
			// Still failed
			errMsg := err.Error()
			event.ErrorMessage = &errMsg
			log.Printf("❌ Retry failed for event %s: %v", event.PolarEventID, err)

			if event.RetryCount >= MaxRetries {
				// Max retries reached, move to dead letter queue
				event.ProcessingStatus = "dead_letter"
				deadLetterReason := fmt.Sprintf("Max retries (%d) exceeded. Last error: %s", MaxRetries, err.Error())
				event.DeadLetterReason = &deadLetterReason
				log.Printf("⚠️  Event %s moved to dead letter queue after %d retries", event.PolarEventID, MaxRetries)

				// Alert ops team (implement notification system)
				w.alertDeadLetter(&event)
			}
		} else {
			// Success!
			event.ProcessingStatus = "processed"
			event.ProcessedAt = &now
			event.ErrorMessage = nil
			log.Printf("✅ Successfully processed event %s on retry #%d", event.PolarEventID, event.RetryCount)
		}

		// Save the updated event
		if err := w.db.Save(&event).Error; err != nil {
			log.Printf("❌ Error updating webhook event %s: %v", event.PolarEventID, err)
		}
	}
}

// shouldRetryNow determines if enough time has passed for exponential backoff
func (w *WebhookRetryWorker) shouldRetryNow(event *models.PolarWebhookEvent) bool {
	if event.LastRetryAt == nil {
		// Never retried, allow retry
		return true
	}

	// Exponential backoff: 15min, 30min, 1hr, 2hr, 4hr
	backoffMinutes := []int{15, 30, 60, 120, 240}
	if event.RetryCount >= len(backoffMinutes) {
		// Use last backoff interval for any additional retries
		return time.Since(*event.LastRetryAt) >= time.Duration(backoffMinutes[len(backoffMinutes)-1])*time.Minute
	}

	backoffDuration := time.Duration(backoffMinutes[event.RetryCount]) * time.Minute
	return time.Since(*event.LastRetryAt) >= backoffDuration
}

// alertDeadLetter sends an alert when an event enters the dead letter queue
func (w *WebhookRetryWorker) alertDeadLetter(event *models.PolarWebhookEvent) {
	// TODO: Implement alerting (Slack, email, PagerDuty, etc.)
	log.Printf("🚨 ALERT: Webhook event %s (type: %s) entered dead letter queue", event.PolarEventID, event.EventType)
	log.Printf("   Reason: %s", *event.DeadLetterReason)

	// You can integrate with:
	// - Slack webhook
	// - Email service
	// - PagerDuty
	// - Prometheus metrics + Alertmanager
	// - CloudWatch Alarms
	// - Datadog
}

// RetryWebhookManually manually retries a specific webhook event (for admin use)
func (w *WebhookRetryWorker) RetryWebhookManually(ctx context.Context, eventID string) error {
	var event models.PolarWebhookEvent
	if err := w.db.Where("polar_event_id = ?", eventID).First(&event).Error; err != nil {
		return fmt.Errorf("webhook event not found: %w", err)
	}

	// Process the webhook
	err := w.webhookHandler.ProcessWebhookEvent(&event)

	// Update status
	now := time.Now()
	if err != nil {
		errMsg := err.Error()
		event.ErrorMessage = &errMsg
		event.ProcessingStatus = "failed"
	} else {
		event.ProcessingStatus = "processed"
		event.ProcessedAt = &now
		event.ErrorMessage = nil
	}

	event.LastRetryAt = &now
	event.UpdatedAt = now

	if saveErr := w.db.Save(&event).Error; saveErr != nil {
		return fmt.Errorf("failed to update event: %w", saveErr)
	}

	return err
}

// GetDeadLetterEvents retrieves all events in the dead letter queue
func (w *WebhookRetryWorker) GetDeadLetterEvents(ctx context.Context, limit int) ([]models.PolarWebhookEvent, error) {
	var events []models.PolarWebhookEvent

	result := w.db.Where("processing_status = ?", "dead_letter").
		Order("created_at DESC").
		Limit(limit).
		Find(&events)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query dead letter events: %w", result.Error)
	}

	return events, nil
}

// GetRetryMetrics returns statistics about webhook retries
func (w *WebhookRetryWorker) GetRetryMetrics(ctx context.Context) (map[string]interface{}, error) {
	var (
		totalFailed     int64
		totalDeadLetter int64
		totalProcessed  int64
		totalPending    int64
	)

	w.db.Model(&models.PolarWebhookEvent{}).Where("processing_status = ?", "failed").Count(&totalFailed)
	w.db.Model(&models.PolarWebhookEvent{}).Where("processing_status = ?", "dead_letter").Count(&totalDeadLetter)
	w.db.Model(&models.PolarWebhookEvent{}).Where("processing_status = ?", "processed").Count(&totalProcessed)
	w.db.Model(&models.PolarWebhookEvent{}).Where("processing_status = ?", "pending").Count(&totalPending)

	metrics := map[string]interface{}{
		"failed":      totalFailed,
		"dead_letter": totalDeadLetter,
		"processed":   totalProcessed,
		"pending":     totalPending,
		"total":       totalFailed + totalDeadLetter + totalProcessed + totalPending,
	}

	return metrics, nil
}
