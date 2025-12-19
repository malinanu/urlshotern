package resilience

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Bulkhead implements the bulkhead pattern for resource isolation
type Bulkhead struct {
	name           string
	maxConcurrent  int
	maxQueueDepth  int
	sem            chan struct{}
	queue          chan *bulkheadRequest
	activeCount    int
	queuedCount    int
	totalRequests  uint64
	totalSuccesses uint64
	totalFailures  uint64
	totalTimeouts  uint64
	mutex          sync.RWMutex
	onReject       func(name string)
}

// bulkheadRequest represents a queued request
type bulkheadRequest struct {
	fn     func() (interface{}, error)
	result chan *bulkheadResult
}

// bulkheadResult represents the result of a bulkhead request
type bulkheadResult struct {
	value interface{}
	err   error
}

// BulkheadConfig holds bulkhead configuration
type BulkheadConfig struct {
	Name          string
	MaxConcurrent int           // Maximum concurrent executions
	MaxQueueDepth int           // Maximum queue depth
	Timeout       time.Duration // Request timeout
	OnReject      func(name string)
}

// NewBulkhead creates a new bulkhead
func NewBulkhead(config BulkheadConfig) *Bulkhead {
	if config.MaxConcurrent <= 0 {
		config.MaxConcurrent = 10 // Default
	}

	if config.MaxQueueDepth < 0 {
		config.MaxQueueDepth = 0
	}

	b := &Bulkhead{
		name:          config.Name,
		maxConcurrent: config.MaxConcurrent,
		maxQueueDepth: config.MaxQueueDepth,
		sem:           make(chan struct{}, config.MaxConcurrent),
		queue:         make(chan *bulkheadRequest, config.MaxQueueDepth),
		onReject:      config.OnReject,
	}

	if b.onReject == nil {
		b.onReject = func(name string) {
			fmt.Printf("[Bulkhead:%s] Request rejected\n", name)
		}
	}

	return b
}

// Execute runs the given request with bulkhead protection
func (b *Bulkhead) Execute(fn func() (interface{}, error)) (interface{}, error) {
	b.mutex.Lock()
	b.totalRequests++
	b.mutex.Unlock()

	// Try to acquire semaphore immediately
	select {
	case b.sem <- struct{}{}:
		defer func() { <-b.sem }()
		return b.executeRequest(fn)
	default:
		// Semaphore full, try to queue
		if b.maxQueueDepth > 0 {
			req := &bulkheadRequest{
				fn:     fn,
				result: make(chan *bulkheadResult, 1),
			}

			select {
			case b.queue <- req:
				b.mutex.Lock()
				b.queuedCount++
				b.mutex.Unlock()

				// Wait for execution
				result := <-req.result

				b.mutex.Lock()
				b.queuedCount--
				b.mutex.Unlock()

				return result.value, result.err
			default:
				// Queue full, reject request
				b.mutex.Lock()
				b.totalFailures++
				b.mutex.Unlock()
				b.onReject(b.name)
				return nil, ErrBulkheadFull
			}
		}

		// No queue configured, reject immediately
		b.mutex.Lock()
		b.totalFailures++
		b.mutex.Unlock()
		b.onReject(b.name)
		return nil, ErrBulkheadFull
	}
}

// ExecuteWithTimeout runs the request with a timeout
func (b *Bulkhead) ExecuteWithTimeout(timeout time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	resultChan := make(chan *bulkheadResult, 1)

	go func() {
		value, err := b.Execute(fn)
		resultChan <- &bulkheadResult{value: value, err: err}
	}()

	select {
	case result := <-resultChan:
		return result.value, result.err
	case <-time.After(timeout):
		b.mutex.Lock()
		b.totalTimeouts++
		b.mutex.Unlock()
		return nil, ErrBulkheadTimeout
	}
}

// ExecuteContext runs the request with context support
func (b *Bulkhead) ExecuteContext(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	b.mutex.Lock()
	b.totalRequests++
	b.mutex.Unlock()

	// Try to acquire semaphore with context cancellation support
	select {
	case b.sem <- struct{}{}:
		defer func() { <-b.sem }()

		// Check if context is already cancelled
		select {
		case <-ctx.Done():
			b.mutex.Lock()
			b.totalTimeouts++
			b.mutex.Unlock()
			return nil, ctx.Err()
		default:
			return b.executeRequestContext(ctx, fn)
		}
	case <-ctx.Done():
		b.mutex.Lock()
		b.totalTimeouts++
		b.mutex.Unlock()
		return nil, ctx.Err()
	default:
		// Semaphore full, reject request
		b.mutex.Lock()
		b.totalFailures++
		b.mutex.Unlock()
		b.onReject(b.name)
		return nil, ErrBulkheadFull
	}
}

// executeRequest runs the function and tracks statistics
func (b *Bulkhead) executeRequest(fn func() (interface{}, error)) (interface{}, error) {
	b.mutex.Lock()
	b.activeCount++
	b.mutex.Unlock()

	defer func() {
		b.mutex.Lock()
		b.activeCount--
		b.mutex.Unlock()
	}()

	value, err := fn()

	b.mutex.Lock()
	if err != nil {
		b.totalFailures++
	} else {
		b.totalSuccesses++
	}
	b.mutex.Unlock()

	return value, err
}

// executeRequestContext runs the function with context
func (b *Bulkhead) executeRequestContext(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	b.mutex.Lock()
	b.activeCount++
	b.mutex.Unlock()

	defer func() {
		b.mutex.Lock()
		b.activeCount--
		b.mutex.Unlock()
	}()

	// Execute with context cancellation support
	resultChan := make(chan *bulkheadResult, 1)
	go func() {
		value, err := fn(ctx)
		resultChan <- &bulkheadResult{value: value, err: err}
	}()

	select {
	case result := <-resultChan:
		b.mutex.Lock()
		if result.err != nil {
			b.totalFailures++
		} else {
			b.totalSuccesses++
		}
		b.mutex.Unlock()
		return result.value, result.err
	case <-ctx.Done():
		b.mutex.Lock()
		b.totalTimeouts++
		b.mutex.Unlock()
		return nil, ctx.Err()
	}
}

// Call is a convenient wrapper for Execute
func (b *Bulkhead) Call(fn func() error) error {
	_, err := b.Execute(func() (interface{}, error) {
		return nil, fn()
	})
	return err
}

// Stats returns bulkhead statistics
func (b *Bulkhead) Stats() BulkheadStats {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return BulkheadStats{
		Name:           b.name,
		MaxConcurrent:  b.maxConcurrent,
		MaxQueueDepth:  b.maxQueueDepth,
		ActiveCount:    b.activeCount,
		QueuedCount:    b.queuedCount,
		TotalRequests:  b.totalRequests,
		TotalSuccesses: b.totalSuccesses,
		TotalFailures:  b.totalFailures,
		TotalTimeouts:  b.totalTimeouts,
	}
}

// BulkheadStats holds bulkhead statistics
type BulkheadStats struct {
	Name           string
	MaxConcurrent  int
	MaxQueueDepth  int
	ActiveCount    int
	QueuedCount    int
	TotalRequests  uint64
	TotalSuccesses uint64
	TotalFailures  uint64
	TotalTimeouts  uint64
}

// SuccessRate returns the success rate as a percentage
func (s BulkheadStats) SuccessRate() float64 {
	if s.TotalRequests == 0 {
		return 0
	}
	return float64(s.TotalSuccesses) / float64(s.TotalRequests) * 100
}

// RejectionRate returns the rejection rate as a percentage
func (s BulkheadStats) RejectionRate() float64 {
	if s.TotalRequests == 0 {
		return 0
	}
	return float64(s.TotalFailures) / float64(s.TotalRequests) * 100
}

// Reset resets all statistics
func (b *Bulkhead) Reset() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.totalRequests = 0
	b.totalSuccesses = 0
	b.totalFailures = 0
	b.totalTimeouts = 0
}

// Errors
var (
	// ErrBulkheadFull is returned when the bulkhead is at capacity
	ErrBulkheadFull = errors.New("bulkhead is full")
	// ErrBulkheadTimeout is returned when the request times out
	ErrBulkheadTimeout = errors.New("bulkhead request timeout")
)

// IsBulkheadError checks if error is a bulkhead error
func IsBulkheadError(err error) bool {
	return errors.Is(err, ErrBulkheadFull) || errors.Is(err, ErrBulkheadTimeout)
}
