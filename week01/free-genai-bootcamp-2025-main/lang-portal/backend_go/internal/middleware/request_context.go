package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/logger"
	"github.com/google/uuid"
)

// RequestContext adds request-specific context values and logging
func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()

		// Create request-scoped logger
		log := logger.FromContext(c.Request.Context()).With(
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"client_ip", c.ClientIP(),
		)

		// Create new context with request ID and logger
		ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
		ctx = log.WithContext(ctx)
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()
	}
} 