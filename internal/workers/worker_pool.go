package workers

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Task represents a unit of work to be processed by the worker pool
type Task struct {
	ID      string
	Execute func(ctx context.Context) error
	OnError func(error)
}

// WorkerPool manages a pool of goroutines for async task processing
type WorkerPool struct {
	workers    int
	taskQueue  chan Task
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
	stopped    bool
	metrics    *PoolMetrics
}

// PoolMetrics tracks worker pool statistics
type PoolMetrics struct {
	TasksProcessed   int64
	TasksFailed      int64
	TasksInQueue     int64
	ActiveWorkers    int32
	mu               sync.RWMutex
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(workers int, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workers:   workers,
		taskQueue: make(chan Task, queueSize),
		ctx:       ctx,
		cancel:    cancel,
		metrics:   &PoolMetrics{},
	}

	pool.start()
	return pool
}

// start initializes and starts all workers
func (p *WorkerPool) start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker processes tasks from the queue
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case task, ok := <-p.taskQueue:
			if !ok {
				return
			}

			// Increment active workers
			p.incrementActiveWorkers()

			// Create a timeout context for the task
			taskCtx, cancel := context.WithTimeout(p.ctx, 30*time.Second)

			// Execute the task with error recovery
			err := p.executeTaskSafely(taskCtx, task)

			// Cleanup
			cancel()
			p.decrementActiveWorkers()

			// Update metrics
			if err != nil {
				p.metrics.mu.Lock()
				p.metrics.TasksFailed++
				p.metrics.mu.Unlock()

				// Call error handler if provided
				if task.OnError != nil {
					task.OnError(err)
				}
			} else {
				p.metrics.mu.Lock()
				p.metrics.TasksProcessed++
				p.metrics.mu.Unlock()
			}
		}
	}
}

// executeTaskSafely executes a task with panic recovery
func (p *WorkerPool) executeTaskSafely(ctx context.Context, task Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("task panicked: %v", r)
		}
	}()

	return task.Execute(ctx)
}

// Submit adds a task to the queue (non-blocking with timeout)
func (p *WorkerPool) Submit(task Task) error {
	p.mu.RLock()
	if p.stopped {
		p.mu.RUnlock()
		return fmt.Errorf("worker pool is stopped")
	}
	p.mu.RUnlock()

	// Use a timeout to prevent blocking forever
	select {
	case p.taskQueue <- task:
		p.metrics.mu.Lock()
		p.metrics.TasksInQueue++
		p.metrics.mu.Unlock()
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("task queue is full, timed out after 5s")
	case <-p.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	}
}

// SubmitWithContext adds a task to the queue with a custom context
func (p *WorkerPool) SubmitWithContext(ctx context.Context, task Task) error {
	p.mu.RLock()
	if p.stopped {
		p.mu.RUnlock()
		return fmt.Errorf("worker pool is stopped")
	}
	p.mu.RUnlock()

	select {
	case p.taskQueue <- task:
		p.metrics.mu.Lock()
		p.metrics.TasksInQueue++
		p.metrics.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-p.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	}
}

// Shutdown gracefully shuts down the worker pool
func (p *WorkerPool) Shutdown(timeout time.Duration) error {
	p.mu.Lock()
	if p.stopped {
		p.mu.Unlock()
		return nil
	}
	p.stopped = true
	p.mu.Unlock()

	// Close task queue to signal workers
	close(p.taskQueue)

	// Wait for all workers to finish with timeout
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.cancel()
		return nil
	case <-time.After(timeout):
		p.cancel()
		return fmt.Errorf("shutdown timed out after %v", timeout)
	}
}

// GetMetrics returns a copy of the current metrics
func (p *WorkerPool) GetMetrics() PoolMetrics {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	return PoolMetrics{
		TasksProcessed: p.metrics.TasksProcessed,
		TasksFailed:    p.metrics.TasksFailed,
		TasksInQueue:   p.metrics.TasksInQueue,
		ActiveWorkers:  p.metrics.ActiveWorkers,
	}
}

// QueueSize returns the current number of tasks in the queue
func (p *WorkerPool) QueueSize() int {
	return len(p.taskQueue)
}

// incrementActiveWorkers increments the active worker count
func (p *WorkerPool) incrementActiveWorkers() {
	p.metrics.mu.Lock()
	p.metrics.ActiveWorkers++
	p.metrics.TasksInQueue--
	p.metrics.mu.Unlock()
}

// decrementActiveWorkers decrements the active worker count
func (p *WorkerPool) decrementActiveWorkers() {
	p.metrics.mu.Lock()
	p.metrics.ActiveWorkers--
	p.metrics.mu.Unlock()
}

// AnalyticsWorkerPool is a singleton worker pool for analytics operations
var (
	analyticsPool     *WorkerPool
	analyticsPoolOnce sync.Once
)

// GetAnalyticsPool returns the singleton analytics worker pool
func GetAnalyticsPool() *WorkerPool {
	analyticsPoolOnce.Do(func() {
		// Create a pool with 20 workers and queue size of 1000
		analyticsPool = NewWorkerPool(20, 1000)
	})
	return analyticsPool
}

// ShutdownAnalyticsPool gracefully shuts down the analytics pool
func ShutdownAnalyticsPool(timeout time.Duration) error {
	if analyticsPool != nil {
		return analyticsPool.Shutdown(timeout)
	}
	return nil
}
