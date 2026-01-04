package workers

import (
	"context"
	"log"
	"time"
)

// ProductSyncer defines the interface for syncing products from Polar
type ProductSyncer interface {
	SyncProductsFromPolar() error
}

// ProductSyncWorker handles periodic synchronization of products from Polar API
type ProductSyncWorker struct {
	polarService  ProductSyncer
	syncInterval  time.Duration
	stopChan      chan struct{}
	lastSyncTime  time.Time
	syncSuccesses int64
	syncFailures  int64
}

// NewProductSyncWorker creates a new product sync worker
func NewProductSyncWorker(polarService ProductSyncer, syncInterval time.Duration) *ProductSyncWorker {
	if syncInterval < 1*time.Minute {
		log.Printf("⚠️  Product sync interval too short (%v), setting to 1 hour", syncInterval)
		syncInterval = 1 * time.Hour
	}

	return &ProductSyncWorker{
		polarService: polarService,
		syncInterval: syncInterval,
		stopChan:     make(chan struct{}),
	}
}

// Start begins the product sync worker
func (w *ProductSyncWorker) Start(ctx context.Context) {
	log.Printf("🔄 Starting product sync worker (interval: %v)", w.syncInterval)

	// Run initial sync immediately
	w.runSync()

	ticker := time.NewTicker(w.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.runSync()
		case <-w.stopChan:
			log.Println("⏹️  Product sync worker stopped")
			return
		case <-ctx.Done():
			log.Println("⏹️  Product sync worker stopped (context cancelled)")
			return
		}
	}
}

// runSync executes the product sync operation
func (w *ProductSyncWorker) runSync() {
	log.Println("🔄 Running scheduled product sync from Polar API...")
	startTime := time.Now()

	err := w.polarService.SyncProductsFromPolar()
	duration := time.Since(startTime)

	if err != nil {
		w.syncFailures++
		log.Printf("❌ Product sync failed after %v: %v", duration, err)
		log.Printf("📊 Sync stats: %d successes, %d failures", w.syncSuccesses, w.syncFailures)
	} else {
		w.syncSuccesses++
		w.lastSyncTime = time.Now()
		log.Printf("✅ Product sync completed successfully in %v", duration)
		log.Printf("📊 Sync stats: %d successes, %d failures (last sync: %v)",
			w.syncSuccesses, w.syncFailures, w.lastSyncTime.Format(time.RFC3339))
	}
}

// Stop gracefully stops the product sync worker
func (w *ProductSyncWorker) Stop() {
	close(w.stopChan)
}

// GetStats returns sync worker statistics
func (w *ProductSyncWorker) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"sync_interval":     w.syncInterval.String(),
		"last_sync_time":    w.lastSyncTime,
		"sync_successes":    w.syncSuccesses,
		"sync_failures":     w.syncFailures,
		"uptime_since_last": time.Since(w.lastSyncTime).String(),
	}
}
