package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware creates structured logging middleware
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\" %s\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
			GetRequestIDFromParam(param),
		)
	})
}

// GetRequestIDFromParam extracts request ID from log parameters
func GetRequestIDFromParam(param gin.LogFormatterParams) string {
	if requestID := param.Request.Header.Get(RequestIDHeader); requestID != "" {
		return fmt.Sprintf("request_id=%s", requestID)
	}
	return ""
}

// LogError logs errors with context
func LogError(c *gin.Context, err error, message string) {
	requestID := GetRequestID(c)
	userID, _ := GetUserID(c)
	
	log.Printf("ERROR [%s] user_id=%d message=%s error=%v path=%s method=%s", 
		requestID, userID, message, err, c.Request.URL.Path, c.Request.Method)
}

// LogInfo logs info messages with context
func LogInfo(c *gin.Context, message string) {
	requestID := GetRequestID(c)
	userID, _ := GetUserID(c)
	
	log.Printf("INFO [%s] user_id=%d message=%s path=%s method=%s", 
		requestID, userID, message, c.Request.URL.Path, c.Request.Method)
}