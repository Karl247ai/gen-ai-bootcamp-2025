package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/metrics"
)

func Monitoring(m *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		method := c.Request.Method

		// Create timer
		timer := m.NewTimer("handler.request.duration",
			"path", path,
			"method", method,
		)
		defer timer.ObserveDuration()

		// Increment request counter
		m.IncCounter("handler.request.count",
			"path", path,
			"method", method,
		)

		// Process request
		c.Next()

		// Record status code
		status := c.Writer.Status()
		m.IncCounter("handler.response.status",
			"path", path,
			"method", method,
			"status", status,
		)

		// Record errors if any
		if len(c.Errors) > 0 {
			m.IncCounter("handler.error.count",
				"path", path,
				"method", method,
				"error_type", c.Errors[0].Type.String(),
			)
		}
	}
} 