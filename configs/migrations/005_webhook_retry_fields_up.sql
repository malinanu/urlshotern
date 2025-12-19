-- Migration: Add webhook retry fields
-- Date: 2025-12-19
-- Description: Adds retry mechanism fields to polar_webhook_events table

-- Add new columns for webhook retry mechanism
ALTER TABLE polar_webhook_events
ADD COLUMN IF NOT EXISTS last_retry_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS dead_letter_reason TEXT;

-- Update processing_status index to include new status
DROP INDEX IF EXISTS idx_polar_webhook_events_processing_status;
CREATE INDEX idx_polar_webhook_events_processing_status ON polar_webhook_events(processing_status);

-- Add index for retry queries (failed events eligible for retry)
CREATE INDEX IF NOT EXISTS idx_webhook_retry
ON polar_webhook_events(processing_status, retry_count, last_retry_at)
WHERE processing_status IN ('failed', 'dead_letter');

-- Add index for last_retry_at
CREATE INDEX IF NOT EXISTS idx_polar_webhook_events_last_retry_at
ON polar_webhook_events(last_retry_at)
WHERE last_retry_at IS NOT NULL;

-- Add comment to table
COMMENT ON COLUMN polar_webhook_events.retry_count IS 'Number of retry attempts (max 5)';
COMMENT ON COLUMN polar_webhook_events.last_retry_at IS 'Timestamp of last retry attempt';
COMMENT ON COLUMN polar_webhook_events.dead_letter_reason IS 'Reason for entering dead letter queue';
COMMENT ON COLUMN polar_webhook_events.processing_status IS 'Status: pending, processed, failed, dead_letter';

-- Update existing failed events to have last_retry_at = updated_at
UPDATE polar_webhook_events
SET last_retry_at = updated_at
WHERE processing_status = 'failed' AND last_retry_at IS NULL;
