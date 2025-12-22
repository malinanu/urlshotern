package utils

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

// GoroutineManager manages goroutine execution with proper error handling
type GoroutineManager struct {
	errorCount int64
	mu         sync.RWMutex
	errors     []GoroutineError
	maxErrors  int
}

// GoroutineError represents an error that occurred in a goroutine
type GoroutineError struct {
	Timestamp   time.Time
	Operation   string
	Error       error
	StackTrace  string
	Recovered   bool
}

// NewGoroutineManager creates a new goroutine manager
func NewGoroutineManager(maxErrors int) *GoroutineManager {
	if maxErrors <= 0 {
		maxErrors = 100
	}
	return &GoroutineManager{
		maxErrors: maxErrors,
		errors:    make([]GoroutineError, 0, maxErrors),
	}
}

// Go executes a function in a goroutine with panic recovery and error logging
func (gm *GoroutineManager) Go(operation string, fn func() error) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic: %v", r)
				gm.recordError(GoroutineError{
					Timestamp:   time.Now(),
					Operation:   operation,
					Error:       err,
					StackTrace:  string(debug.Stack()),
					Recovered:   true,
				})
				log.Printf("[GOROUTINE PANIC] %s: %v\nStack: %s", operation, r, debug.Stack())
			}
		}()

		if err := fn(); err != nil {
			gm.recordError(GoroutineError{
				Timestamp:  time.Now(),
				Operation:  operation,
				Error:      err,
				StackTrace: string(debug.Stack()),
				Recovered:  false,
			})
			log.Printf("[GOROUTINE ERROR] %s: %v", operation, err)
		}
	}()
}

// GoWithContext executes a function in a goroutine with context support
func (gm *GoroutineManager) GoWithContext(ctx context.Context, operation string, fn func(context.Context) error) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic: %v", r)
				gm.recordError(GoroutineError{
					Timestamp:   time.Now(),
					Operation:   operation,
					Error:       err,
					StackTrace:  string(debug.Stack()),
					Recovered:   true,
				})
				log.Printf("[GOROUTINE PANIC] %s: %v\nStack: %s", operation, r, debug.Stack())
			}
		}()

		if err := fn(ctx); err != nil && err != context.Canceled {
			gm.recordError(GoroutineError{
				Timestamp:  time.Now(),
				Operation:  operation,
				Error:      err,
				StackTrace: string(debug.Stack()),
				Recovered:  false,
			})
			log.Printf("[GOROUTINE ERROR] %s: %v", operation, err)
		}
	}()
}

// recordError records an error with rotation of old errors
func (gm *GoroutineManager) recordError(err GoroutineError) {
	atomic.AddInt64(&gm.errorCount, 1)

	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Add error to the list
	gm.errors = append(gm.errors, err)

	// Rotate old errors if we exceed max
	if len(gm.errors) > gm.maxErrors {
		gm.errors = gm.errors[1:]
	}
}

// GetErrorCount returns the total number of errors encountered
func (gm *GoroutineManager) GetErrorCount() int64 {
	return atomic.LoadInt64(&gm.errorCount)
}

// GetRecentErrors returns the most recent errors
func (gm *GoroutineManager) GetRecentErrors(limit int) []GoroutineError {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	if limit <= 0 || limit > len(gm.errors) {
		limit = len(gm.errors)
	}

	start := len(gm.errors) - limit
	if start < 0 {
		start = 0
	}

	errors := make([]GoroutineError, limit)
	copy(errors, gm.errors[start:])
	return errors
}

// GetStats returns goroutine error statistics
func (gm *GoroutineManager) GetStats() map[string]interface{} {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	panicCount := 0
	errorCount := 0
	for _, err := range gm.errors {
		if err.Recovered {
			panicCount++
		} else {
			errorCount++
		}
	}

	return map[string]interface{}{
		"total_errors":       gm.GetErrorCount(),
		"recent_errors":      len(gm.errors),
		"recent_panics":      panicCount,
		"recent_errors_only": errorCount,
	}
}

// Default global goroutine manager
var defaultManager = NewGoroutineManager(100)

// Go executes a function in a goroutine with panic recovery using the default manager
func Go(operation string, fn func() error) {
	defaultManager.Go(operation, fn)
}

// GoWithContext executes a function in a goroutine with context support using the default manager
func GoWithContext(ctx context.Context, operation string, fn func(context.Context) error) {
	defaultManager.GoWithContext(ctx, operation, fn)
}

// GetGoroutineStats returns statistics from the default manager
func GetGoroutineStats() map[string]interface{} {
	return defaultManager.GetStats()
}

// GetRecentGoroutineErrors returns recent errors from the default manager
func GetRecentGoroutineErrors(limit int) []GoroutineError {
	return defaultManager.GetRecentErrors(limit)
}
