package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/URLshorter/url-shortener/internal/utils"
	"github.com/URLshorter/url-shortener/internal/workers"
)

// BenchmarkBase62Encoding tests the performance of the optimized base62 encoding
func BenchmarkBase62Encoding(b *testing.B) {
	testIDs := []int64{
		1,
		1000,
		1000000,
		1000000000,
		1000000000000,
	}

	for _, id := range testIDs {
		b.Run(fmt.Sprintf("ID_%d", id), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = utils.EncodeBase62(id)
			}
		})
	}
}

// BenchmarkBase62EncodingParallel tests parallel encoding performance
func BenchmarkBase62EncodingParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		id := int64(1000000000)
		for pb.Next() {
			_ = utils.EncodeBase62(id)
		}
	})
}

// BenchmarkBloomFilterAdd tests Bloom filter insertion performance
func BenchmarkBloomFilterAdd(b *testing.B) {
	bf := utils.NewBloomFilter(1000000, 0.01)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Add(fmt.Sprintf("short_code_%d", i))
	}
}

// BenchmarkBloomFilterCheck tests Bloom filter lookup performance
func BenchmarkBloomFilterCheck(b *testing.B) {
	bf := utils.NewBloomFilter(1000000, 0.01)

	// Pre-populate with items
	for i := 0; i < 10000; i++ {
		bf.Add(fmt.Sprintf("short_code_%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bf.MightContain(fmt.Sprintf("short_code_%d", i%10000))
	}
}

// BenchmarkWorkerPool tests worker pool performance
func BenchmarkWorkerPool(b *testing.B) {
	pool := workers.NewWorkerPool(20, 1000)
	defer pool.Shutdown(5 * time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Submit(workers.Task{
			ID: fmt.Sprintf("task_%d", i),
			Execute: func(ctx context.Context) error {
				// Simulate work
				time.Sleep(1 * time.Millisecond)
				return nil
			},
		})
	}
}

// BenchmarkWorkerPoolParallel tests parallel task submission
func BenchmarkWorkerPoolParallel(b *testing.B) {
	pool := workers.NewWorkerPool(50, 10000)
	defer pool.Shutdown(10 * time.Second)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			pool.Submit(workers.Task{
				ID: fmt.Sprintf("task_%d", i),
				Execute: func(ctx context.Context) error {
					// Simulate work
					time.Sleep(100 * time.Microsecond)
					return nil
				},
			})
			i++
		}
	})
}

// TestWorkerPoolThroughput measures actual throughput
func TestWorkerPoolThroughput(t *testing.T) {
	tests := []struct {
		name       string
		workers    int
		queueSize  int
		tasks      int
		workTimeMs int
	}{
		{"Small_10workers_100tasks", 10, 100, 100, 1},
		{"Medium_20workers_1000tasks", 20, 1000, 1000, 1},
		{"Large_50workers_10000tasks", 50, 10000, 10000, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := workers.NewWorkerPool(tt.workers, tt.queueSize)
			defer pool.Shutdown(30 * time.Second)

			start := time.Now()
			var wg sync.WaitGroup

			completed := 0
			var mu sync.Mutex

			for i := 0; i < tt.tasks; i++ {
				wg.Add(1)
				err := pool.Submit(workers.Task{
					ID: fmt.Sprintf("task_%d", i),
					Execute: func(ctx context.Context) error {
						// Simulate work
						time.Sleep(time.Duration(tt.workTimeMs) * time.Millisecond)

						mu.Lock()
						completed++
						mu.Unlock()

						wg.Done()
						return nil
					},
				})

				if err != nil {
					t.Errorf("Failed to submit task: %v", err)
					wg.Done()
				}
			}

			wg.Wait()
			duration := time.Since(start)

			throughput := float64(completed) / duration.Seconds()
			t.Logf("Completed %d tasks in %v (%.2f tasks/sec)", completed, duration, throughput)

			metrics := pool.GetMetrics()
			t.Logf("Metrics: Processed=%d, Failed=%d, InQueue=%d",
				metrics.TasksProcessed, metrics.TasksFailed, metrics.TasksInQueue)
		})
	}
}

// BenchmarkBloomFilterVsMap compares Bloom filter with map lookups
func BenchmarkBloomFilterVsMap(b *testing.B) {
	const items = 100000

	// Setup Bloom filter
	bf := utils.NewBloomFilter(items, 0.01)
	for i := 0; i < items; i++ {
		bf.Add(fmt.Sprintf("key_%d", i))
	}

	// Setup map
	m := make(map[string]bool, items)
	for i := 0; i < items; i++ {
		m[fmt.Sprintf("key_%d", i)] = true
	}

	b.Run("BloomFilter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = bf.MightContain(fmt.Sprintf("key_%d", i%items))
		}
	})

	b.Run("Map", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = m[fmt.Sprintf("key_%d", i%items)]
		}
	})
}

// TestBloomFilterAccuracy tests the false positive rate
func TestBloomFilterAccuracy(t *testing.T) {
	const items = 10000
	const tests = 10000

	bf := utils.NewBloomFilter(items, 0.01) // 1% false positive rate

	// Add items
	for i := 0; i < items; i++ {
		bf.Add(fmt.Sprintf("exists_%d", i))
	}

	// Test false positives
	falsePositives := 0
	for i := 0; i < tests; i++ {
		if bf.MightContain(fmt.Sprintf("notexists_%d", i)) {
			falsePositives++
		}
	}

	actualFPR := float64(falsePositives) / float64(tests)
	t.Logf("False positive rate: %.4f (expected: 0.01)", actualFPR)

	// Should be close to 1% (allow some variance)
	if actualFPR > 0.02 {
		t.Errorf("False positive rate too high: %.4f > 0.02", actualFPR)
	}
}

// BenchmarkGoroutineSpawning compares goroutine spawning vs worker pool
func BenchmarkGoroutineSpawning(b *testing.B) {
	b.Run("DirectGoroutines", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var wg sync.WaitGroup
			for j := 0; j < 5; j++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					time.Sleep(1 * time.Microsecond)
				}()
			}
			wg.Wait()
		}
	})

	b.Run("WorkerPool", func(b *testing.B) {
		pool := workers.NewWorkerPool(20, 1000)
		defer pool.Shutdown(5 * time.Second)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var wg sync.WaitGroup
			for j := 0; j < 5; j++ {
				wg.Add(1)
				pool.Submit(workers.Task{
					ID: fmt.Sprintf("task_%d_%d", i, j),
					Execute: func(ctx context.Context) error {
						defer wg.Done()
						time.Sleep(1 * time.Microsecond)
						return nil
					},
				})
			}
			wg.Wait()
		}
	})
}

// LoadTestAnalyticsProcessing simulates high load analytics processing
func TestLoadTestAnalyticsProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	pool := workers.NewWorkerPool(50, 5000)
	defer pool.Shutdown(30 * time.Second)

	// Simulate 1000 RPS for 10 seconds = 10,000 clicks
	const rps = 1000
	const duration = 10 * time.Second

	start := time.Now()
	submitted := 0
	failed := 0

	ticker := time.NewTicker(time.Second / time.Duration(rps))
	defer ticker.Stop()

	timeout := time.After(duration)

	for {
		select {
		case <-timeout:
			goto done
		case <-ticker.C:
			err := pool.Submit(workers.Task{
				ID: fmt.Sprintf("analytics_%d", submitted),
				Execute: func(ctx context.Context) error {
					// Simulate analytics processing
					time.Sleep(5 * time.Millisecond)
					return nil
				},
			})
			if err != nil {
				failed++
			}
			submitted++
		}
	}

done:
	elapsed := time.Since(start)
	actualRPS := float64(submitted) / elapsed.Seconds()

	t.Logf("Load test results:")
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Submitted: %d requests", submitted)
	t.Logf("  Failed: %d requests", failed)
	t.Logf("  Actual RPS: %.2f", actualRPS)

	metrics := pool.GetMetrics()
	t.Logf("  Worker metrics: Processed=%d, Failed=%d, InQueue=%d",
		metrics.TasksProcessed, metrics.TasksFailed, metrics.TasksInQueue)

	if failed > submitted/10 {
		t.Errorf("Too many failed submissions: %d/%d (%.2f%%)",
			failed, submitted, float64(failed)/float64(submitted)*100)
	}
}
