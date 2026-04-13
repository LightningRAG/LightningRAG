package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "X-Request-ID"

// RequestID generates or propagates a unique request ID for every HTTP request.
// The ID is set in both the context (for structured logging) and the response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(RequestIDKey)
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Set(RequestIDKey, rid)
		c.Header(RequestIDKey, rid)
		c.Next()
	}
}
