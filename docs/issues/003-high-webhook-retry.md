# [HIGH] Implement Webhook Retry Mechanism

**Labels:** `high`, `backend`, `enhancement`, `medium`

## Description
Failed Polar webhooks are logged but never retried. The `retry_count` field exists in the database schema but is never incremented, meaning payment events can be lost permanently.

## Severity: HIGH
This is high priority because:
- Payment webhook failures result in data loss
- Subscription state can become out of sync with Polar
- Users may be incorrectly billed or have access issues
- No automated recovery from transient failures
- Financial reconciliation problems

## Affected Files
- [ ] `internal/handlers/polar_webhook.go` (lines 154-167)
- [ ] Database schema: `polar_webhook_events.retry_count` (unused field)

## Current Behavior
```go
if err := h.processWebhook(event); err != nil {
    log.Error("Failed to process webhook", "error", err)
    event.Status = "failed"
    event.Error = err.Error()
    h.db.Save(event)  // Just logs failure, no retry
    c.JSON(500, gin.H{"error": "webhook processing failed"})
    return
}
```

Failed webhooks are marked as "failed" in the database but:
- No retry mechanism exists
- `retry_count` field remains at 0
- Manual intervention required to recover
- No exponential backoff
- No dead letter queue

## Expected Behavior
- Automatic retry of failed webhooks
- Exponential backoff strategy
- Maximum retry attempts (e.g., 5 times)
- Dead letter queue for permanent failures
- Monitoring and alerting for webhook failures
- Manual replay capability

## Proposed Solution

### 1. Add Webhook Retry Worker
```go
// internal/workers/webhook_retry.go
type WebhookRetryWorker struct {
    db     *gorm.DB
    client *PolarClient
}

func (w *WebhookRetryWorker) Start(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            w.retryFailedWebhooks(ctx)
        case <-ctx.Done():
            return
        }
    }
}

func (w *WebhookRetryWorker) retryFailedWebhooks(ctx context.Context) {
    var failedEvents []PolarWebhookEvent

    // Find events eligible for retry
    w.db.Where("status = ? AND retry_count < ? AND last_retry_at < ?",
        "failed", 5, time.Now().Add(-15*time.Minute)).
        Find(&failedEvents)

    for _, event := range failedEvents {
        if err := w.retryWebhook(&event); err != nil {
            // Exponential backoff: 15min, 30min, 1hr, 2hr, 4hr
            event.LastRetryAt = time.Now()
            event.RetryCount++
            event.Error = err.Error()

            if event.RetryCount >= 5 {
                event.Status = "dead_letter"
                // Alert ops team
                w.alertDeadLetter(&event)
            }

            w.db.Save(&event)
        } else {
            event.Status = "processed"
            event.ProcessedAt = time.Now()
            w.db.Save(&event)
        }
    }
}
```

### 2. Update Webhook Handler
```go
func (h *PolarWebhookHandler) HandleWebhook(c *gin.Context) {
    // ... existing validation ...

    event := &PolarWebhookEvent{
        EventID:     webhookEvent.ID,
        EventType:   webhookEvent.Type,
        Status:      "pending",
        RetryCount:  0,
        ReceivedAt:  time.Now(),
    }

    if err := h.processWebhook(event); err != nil {
        log.Error("Webhook processing failed", "error", err)
        event.Status = "failed"
        event.Error = err.Error()
        event.LastRetryAt = time.Now() // Will be retried by worker
        h.db.Save(event)

        // Return 200 to Polar to prevent their retry
        // We'll handle retries ourselves
        c.JSON(200, gin.H{"message": "queued for retry"})
        return
    }

    event.Status = "processed"
    event.ProcessedAt = time.Now()
    h.db.Save(event)
    c.JSON(200, gin.H{"message": "processed"})
}
```

### 3. Add Migration for New Fields
```sql
ALTER TABLE polar_webhook_events
ADD COLUMN last_retry_at TIMESTAMP,
ADD COLUMN retry_count INTEGER DEFAULT 0,
ADD COLUMN dead_letter_reason TEXT;

CREATE INDEX idx_webhook_retry ON polar_webhook_events(status, retry_count, last_retry_at)
WHERE status IN ('failed', 'dead_letter');
```

### 4. Add Manual Replay Endpoint
```go
// POST /api/v1/admin/webhooks/:id/replay
func (h *PolarWebhookHandler) ReplayWebhook(c *gin.Context) {
    var event PolarWebhookEvent
    if err := h.db.First(&event, c.Param("id")).Error; err != nil {
        c.JSON(404, gin.H{"error": "event not found"})
        return
    }

    if err := h.retryWebhook(&event); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"message": "webhook replayed successfully"})
}
```

## Acceptance Criteria
- [ ] Retry worker implemented and running
- [ ] Exponential backoff working (15min, 30min, 1hr, 2hr, 4hr)
- [ ] Max 5 retry attempts before dead letter
- [ ] Database fields updated (`last_retry_at`, `retry_count`)
- [ ] Dead letter queue monitoring
- [ ] Manual replay endpoint functional
- [ ] Metrics/logging for retry attempts
- [ ] Integration tests for retry logic
- [ ] Documentation updated

## Implementation Steps
1. Create database migration for new fields
2. Implement WebhookRetryWorker
3. Update webhook handler to queue for retry
4. Add manual replay endpoint
5. Add monitoring and alerting
6. Write comprehensive tests
7. Deploy to staging and test with simulated failures
8. Deploy to production

## Monitoring & Alerts
- Alert when webhook enters dead letter queue
- Dashboard showing retry metrics
- Weekly report of failed webhooks
- Prometheus metrics for retry attempts

## Related Issues
- #4 (Mock Invoice Data) - proper invoice sync depends on reliable webhooks
- Observability system integration

## References
- [Webhook Best Practices](https://stripe.com/docs/webhooks/best-practices)
- [Idempotent Webhook Processing](https://docs.polar.sh/webhooks)
