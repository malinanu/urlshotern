# Goroutine Error Handling

This document explains the goroutine error handling system implemented in the URL Shortener service.

## Overview

The application uses a centralized goroutine management system to handle errors and panics in background tasks, providing proper error tracking, logging, and monitoring capabilities.

## Architecture

### GoroutineManager

The `GoroutineManager` in `internal/utils/goroutine.go` provides:

- **Panic Recovery**: Automatically catches and logs panics in goroutines
- **Error Logging**: Structured logging of goroutine errors
- **Error Tracking**: Maintains a rotating buffer of recent errors
- **Statistics**: Provides metrics on goroutine failures

## Usage

### Basic Usage

```go
import "github.com/URLshorter/url-shortener/internal/utils"

// Simple goroutine with error handling
utils.Go("operation_name", func() error {
    // Your code here
    return someOperation()
})
```

### With Context

```go
import (
    "context"
    "github.com/URLshorter/url-shortener/internal/utils"
)

utils.GoWithContext(ctx, "operation_name", func(ctx context.Context) error {
    // Your code here with context support
    return someOperationWithContext(ctx)
})
```

### Creating Custom Managers

For isolated error tracking:

```go
manager := utils.NewGoroutineManager(100) // Keep last 100 errors

manager.Go("custom_operation", func() error {
    return myOperation()
})

stats := manager.GetStats()
errors := manager.GetRecentErrors(10)
```

## Error Tracking

### Error Structure

Each error is tracked with:

```go
type GoroutineError struct {
    Timestamp   time.Time
    Operation   string
    Error       error
    StackTrace  string
    Recovered   bool  // true if panic was recovered
}
```

### Statistics

Access goroutine statistics via:

1. **Code**:
```go
stats := utils.GetGoroutineStats()
```

2. **HTTP Endpoint**:
```bash
curl http://localhost:8080/health/goroutines?limit=10
```

Response:
```json
{
  "status": "success",
  "data": {
    "stats": {
      "total_errors": 42,
      "recent_errors": 10,
      "recent_panics": 2,
      "recent_errors_only": 8
    },
    "recent_errors": [
      {
        "timestamp": "2025-12-22T10:30:45Z",
        "operation": "save_click_event",
        "error": "database connection timeout",
        "recovered": false
      }
    ]
  }
}
```

## Common Patterns

### Async Database Operations

```go
// Instead of:
go func() {
    if err := db.SaveEvent(event); err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}()

// Use:
utils.Go("save_event", func() error {
    return db.SaveEvent(event)
})
```

### Analytics Processing

```go
utils.Go("process_analytics", func() error {
    return analyticsService.ProcessEvent(event)
})
```

### Click Tracking

```go
utils.Go("record_click", func() error {
    return shortenerService.RecordClick(shortCode, ip, userAgent, referrer)
})
```

## Monitoring

### Production Monitoring

Set up alerts for:

1. **High Error Rates**:
```bash
# Alert if error count grows too fast
total_errors_threshold=100_per_hour
```

2. **Panic Detection**:
```bash
# Alert on any recovered panics
recent_panics > 0
```

3. **Specific Operation Failures**:
```bash
# Filter by operation name
operation == "save_click_event" && error_count > threshold
```

### Log Analysis

Goroutine errors are logged with the prefix:
- `[GOROUTINE ERROR]` - Regular errors
- `[GOROUTINE PANIC]` - Recovered panics

Example log entry:
```
2025-12-22T10:30:45Z [GOROUTINE ERROR] save_click_event: database connection timeout
```

## Best Practices

1. **Always Use utils.Go for Background Tasks**:
   ```go
   // ✅ Good
   utils.Go("operation", func() error {
       return doWork()
   })

   // ❌ Bad
   go func() {
       doWork()
   }()
   ```

2. **Use Descriptive Operation Names**:
   ```go
   // ✅ Good
   utils.Go("save_user_profile", ...)
   utils.Go("send_welcome_email", ...)

   // ❌ Bad
   utils.Go("op1", ...)
   utils.Go("task", ...)
   ```

3. **Return Errors Properly**:
   ```go
   // ✅ Good
   utils.Go("operation", func() error {
       err := doWork()
       if err != nil {
           return fmt.Errorf("failed to do work: %w", err)
       }
       return nil
   })

   // ❌ Bad
   utils.Go("operation", func() error {
       doWork() // ignoring error
       return nil
   })
   ```

4. **Use Context for Long-Running Operations**:
   ```go
   utils.GoWithContext(ctx, "long_operation", func(ctx context.Context) error {
       select {
       case <-ctx.Done():
           return ctx.Err()
       case <-time.After(time.Second):
           return doWork()
       }
   })
   ```

5. **Monitor in Production**:
   - Set up dashboards for goroutine stats
   - Alert on high error rates or panics
   - Regularly review error logs

## Integration with Existing Code

### Files Updated

The following files now use proper goroutine error handling:

1. `internal/services/shortener.go`:
   - `save_click_event`
   - `process_enhanced_click_analytics`
   - `increment_click_count`

2. `internal/handlers/handlers.go`:
   - `record_click`

3. `internal/handlers/cms_handlers.go`:
   - `record_page_visit`

### Migration Guide

To migrate existing fire-and-forget goroutines:

```go
// Before:
go func() {
    if err := operation(); err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}()

// After:
utils.Go("operation_name", func() error {
    return operation()
})
```

## Performance Considerations

- **Minimal Overhead**: The error tracking adds negligible overhead
- **Memory Bounded**: Error buffer is size-limited (default: 100 errors)
- **Lock-Free Counting**: Uses atomic operations for counters
- **Efficient Logging**: Only logs errors, not successes

## Troubleshooting

### High Error Counts

If you see high error counts:

1. Check `/health/goroutines` endpoint for details
2. Identify the failing operation
3. Review recent errors for patterns
4. Check database/Redis connectivity
5. Verify external service availability

### Memory Issues

If error tracking uses too much memory:

1. Reduce max errors in GoroutineManager
2. Implement error aggregation
3. Export errors to external monitoring

### Missing Errors

If errors aren't being tracked:

1. Verify `utils.Go()` is being used
2. Check that operations return errors
3. Ensure errors aren't being swallowed
4. Verify log output configuration

## Future Enhancements

Potential improvements:

1. **Metrics Export**: Prometheus metrics for goroutine errors
2. **Error Aggregation**: Group similar errors
3. **Rate Limiting**: Limit error logging rate
4. **External Alerting**: Integration with PagerDuty/Slack
5. **Error Recovery**: Automatic retry logic
6. **Performance Profiling**: Track goroutine execution time

## References

- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Error Handling in Go](https://go.dev/blog/error-handling-and-go)
- [Effective Go - Concurrency](https://go.dev/doc/effective_go#concurrency)
