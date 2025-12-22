package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

type ClickBatchProcessor struct {
	db           *storage.PostgresStorage
	redis        *storage.RedisStorage
	batchSize    int
	batchTimeout time.Duration
	eventQueue   chan *models.ClickEvent
	stopChan     chan struct{}
	wg           sync.WaitGroup
	metrics      *ClickMetrics
	mu           sync.Mutex
}

type ClickMetrics struct {
	TotalClicks   int64
	BatchesAdded  int64
	FailedClicks  int64
	QueueSize     int
	LastBatchTime time.Time
	mu            sync.RWMutex
}

var (
	ErrProcessorStopped = fmt.Errorf("batch processor is stopped")
	ErrQueueFull        = fmt.Errorf("event queue is full")
)

// NewClickBatchProcessor creates a new click batch processor
// batchSize: number of clicks to accumulate before flushing (default: 100)
// batchTimeout: maximum time to wait before flushing partial batch (default: 5 seconds)
func NewClickBatchProcessor(
	db *storage.PostgresStorage,
	redis *storage.RedisStorage,
	batchSize int,
	batchTimeout time.Duration,
) *ClickBatchProcessor {
	if batchSize <= 0 {
		batchSize = 100
	}
	if batchTimeout <= 0 {
		batchTimeout = 5 * time.Second
	}

	return &ClickBatchProcessor{
		db:           db,
		redis:        redis,
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
		eventQueue:   make(chan *models.ClickEvent, batchSize*10),
		stopChan:     make(chan struct{}),
		metrics:      &ClickMetrics{},
	}
}

// Start starts the batch processor
func (cbp *ClickBatchProcessor) Start(ctx context.Context) {
	cbp.wg.Add(1)
	go cbp.processBatches(ctx)
}

// Stop gracefully stops the batch processor
func (cbp *ClickBatchProcessor) Stop() {
	close(cbp.stopChan)
	cbp.wg.Wait()
}

// AddClick adds a click to the batch queue for processing
func (cbp *ClickBatchProcessor) AddClick(event *models.ClickEvent) error {
	select {
	case cbp.eventQueue <- event:
		cbp.metrics.mu.Lock()
		cbp.metrics.TotalClicks++
		cbp.metrics.QueueSize = len(cbp.eventQueue)
		cbp.metrics.mu.Unlock()
		return nil
	case <-cbp.stopChan:
		return ErrProcessorStopped
	default:
		// Queue full, drop click (acceptable for analytics - we're sampling)
		cbp.metrics.mu.Lock()
		cbp.metrics.FailedClicks++
		cbp.metrics.mu.Unlock()
		return ErrQueueFull
	}
}

// processBatches processes accumulated click events in batches
func (cbp *ClickBatchProcessor) processBatches(ctx context.Context) {
	defer cbp.wg.Done()

	batch := make([]*models.ClickEvent, 0, cbp.batchSize)
	timer := time.NewTimer(cbp.batchTimeout)
	defer timer.Stop()

	for {
		select {
		case event := <-cbp.eventQueue:
			if event == nil {
				continue
			}

			batch = append(batch, event)

			// Process when batch is full
			if len(batch) >= cbp.batchSize {
				if err := cbp.flushBatch(ctx, batch); err != nil {
					// Log error but continue processing
					for range batch {
						cbp.metrics.mu.Lock()
						cbp.metrics.FailedClicks++
						cbp.metrics.mu.Unlock()
					}
				}
				batch = make([]*models.ClickEvent, 0, cbp.batchSize)
				timer.Reset(cbp.batchTimeout)
			}

		case <-timer.C:
			// Timeout - flush partial batch
			if len(batch) > 0 {
				if err := cbp.flushBatch(ctx, batch); err != nil {
					// Log error
				}
				batch = make([]*models.ClickEvent, 0, cbp.batchSize)
			}
			timer.Reset(cbp.batchTimeout)

		case <-cbp.stopChan:
			// Flush remaining events before shutdown
			if len(batch) > 0 {
				cbp.flushBatch(ctx, batch)
			}
			return

		case <-ctx.Done():
			// Context cancelled
			if len(batch) > 0 {
				cbp.flushBatch(ctx, batch)
			}
			return
		}
	}
}

// flushBatch inserts a batch of click events
func (cbp *ClickBatchProcessor) flushBatch(ctx context.Context, batch []*models.ClickEvent) error {
	if len(batch) == 0 {
		return nil
	}

	// Use COPY for maximum performance (10-100x faster than individual INSERTs)
	err := cbp.db.BatchInsertClickEvents(ctx, batch)
	if err != nil {
		return fmt.Errorf("failed to insert batch: %w", err)
	}

	cbp.metrics.mu.Lock()
	cbp.metrics.BatchesAdded++
	cbp.metrics.LastBatchTime = time.Now()
	cbp.metrics.mu.Unlock()

	return nil
}

// GetMetrics returns current processor metrics
func (cbp *ClickBatchProcessor) GetMetrics() ClickMetrics {
	cbp.metrics.mu.RLock()
	defer cbp.metrics.mu.RUnlock()
	// Return a copy without the mutex to avoid copying lock value
	return ClickMetrics{
		TotalClicks:   cbp.metrics.TotalClicks,
		BatchesAdded:  cbp.metrics.BatchesAdded,
		FailedClicks:  cbp.metrics.FailedClicks,
		QueueSize:     cbp.metrics.QueueSize,
		LastBatchTime: cbp.metrics.LastBatchTime,
	}
}
