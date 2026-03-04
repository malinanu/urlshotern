package workers

import (
	"context"
	"log"
	"time"

	"github.com/URLshorter/url-shortener/internal/storage"
)

// DataRetainer defines the interface for running the data retention job
type DataRetainer interface {
	RunRetentionJob() error
}

// RetentionWorker handles nightly cleanup of raw IPs and old click data
type RetentionWorker struct {
	db       DataRetainer
	stopChan chan struct{}
	lastRun  time.Time
	runs     int64
	failures int64
}

// NewRetentionWorker creates a new retention worker
func NewRetentionWorker(db *storage.PostgresStorage) *RetentionWorker {
	return &RetentionWorker{
		db:       db,
		stopChan: make(chan struct{}),
	}
}

// Start runs the retention worker. It checks every hour and executes at 2 AM.
func (w *RetentionWorker) Start(ctx context.Context) {
	log.Println("🗑️  Starting data retention worker (nightly at 02:00)")

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if time.Now().Hour() == 2 {
				w.runRetention()
			}
		case <-w.stopChan:
			log.Println("⏹️  Data retention worker stopped")
			return
		case <-ctx.Done():
			log.Println("⏹️  Data retention worker stopped (context cancelled)")
			return
		}
	}
}

func (w *RetentionWorker) runRetention() {
	log.Println("🗑️  Running data retention job...")
	start := time.Now()

	if err := w.db.RunRetentionJob(); err != nil {
		w.failures++
		log.Printf("❌ Data retention job failed: %v", err)
		return
	}

	w.runs++
	w.lastRun = time.Now()
	log.Printf("✅ Data retention job completed in %v", time.Since(start))
}

// Stop gracefully stops the retention worker
func (w *RetentionWorker) Stop() {
	close(w.stopChan)
}

// GetStats returns worker statistics
func (w *RetentionWorker) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"last_run": w.lastRun,
		"runs":     w.runs,
		"failures": w.failures,
	}
}
