package storage

// CacheError represents a cache-related error
type CacheError struct {
	Message string
}

func (e *CacheError) Error() string {
	return e.Message
}

// Common storage errors
var (
	// ErrCacheKeyNotFound is returned when a cache key is not found
	ErrCacheKeyNotFound = &CacheError{Message: "cache key not found"}

	// ErrURLExpired indicates the URL has expired
	ErrURLExpired = &CacheError{Message: "URL has expired"}

	// ErrURLNotFound indicates the URL mapping was not found
	ErrURLNotFound = &CacheError{Message: "URL not found"}

	// ErrUnauthorized indicates unauthorized access
	ErrUnauthorized = &CacheError{Message: "unauthorized access"}
)
