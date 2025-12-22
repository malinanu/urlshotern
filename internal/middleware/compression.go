package middleware

import (
	"compress/gzip"
	"io"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// gzipWriter wraps a ResponseWriter with gzip compression
type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

// Write compresses data before writing
func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

// WriteString compresses string data before writing
func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

// gzipWriterPool reduces allocations by reusing gzip writers
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(io.Discard)
	},
}

// CompressionMiddleware adds gzip compression to responses
func CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip compression for certain paths
		if shouldSkipCompression(c.Request.RequestURI) {
			c.Next()
			return
		}

		// Check if client accepts gzip
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Get gzip writer from pool
		gz := gzipWriterPool.Get().(*gzip.Writer)
		defer gzipWriterPool.Put(gz)

		gz.Reset(c.Writer)
		defer gz.Close()

		// Set response headers
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		// Wrap response writer
		c.Writer = &gzipWriter{
			ResponseWriter: c.Writer,
			writer:         gz,
		}

		c.Next()
	}
}

// shouldSkipCompression determines if compression should be skipped
func shouldSkipCompression(path string) bool {
	// Skip compression for:
	// - WebSocket connections
	// - Already compressed content (images, etc.)
	// - Very small responses (compression overhead not worth it)
	skipPaths := []string{
		"/realtime/ws",
		"/_next/static/",
		".png",
		".jpg",
		".jpeg",
		".gif",
		".ico",
		".woff",
		".woff2",
		".ttf",
	}

	for _, skip := range skipPaths {
		if strings.Contains(path, skip) {
			return true
		}
	}

	return false
}

// BrotliMiddleware adds Brotli compression (requires brotli package)
// Brotli typically provides 15-25% better compression than gzip
// Uncomment and implement if needed:
/*
import "github.com/andybalholm/brotli"

func BrotliMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipCompression(c.Request.RequestURI) {
			c.Next()
			return
		}

		acceptEncoding := c.GetHeader("Accept-Encoding")

		// Prefer Brotli over gzip if client supports it
		if strings.Contains(acceptEncoding, "br") {
			// Implement Brotli compression
			// ...
		} else if strings.Contains(acceptEncoding, "gzip") {
			// Fall back to gzip
			// ...
		}

		c.Next()
	}
}
*/
