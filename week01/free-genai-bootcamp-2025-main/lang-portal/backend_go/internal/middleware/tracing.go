package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/logger"
)

type RequestTrace struct {
	TraceID    string
	Path       string
	Method     string
	StartTime  time.Time
	Duration   time.Duration
	StatusCode int
	Error      error
}

func Tracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		trace := &RequestTrace{
			TraceID:   c.GetHeader("X-Request-ID"),
			Path:      c.FullPath(),
			Method:    c.Request.Method,
			StartTime: time.Now(),
		}

		if trace.TraceID == "" {
			trace.TraceID = fmt.Sprintf("trace-%d", time.Now().UnixNano())
		}

		c.Set("trace", trace)
		c.Header("X-Request-ID", trace.TraceID)

		// Process request
		c.Next()

		// Update trace info
		trace.Duration = time.Since(trace.StartTime)
		trace.StatusCode = c.Writer.Status()
		if len(c.Errors) > 0 {
			trace.Error = c.Errors[0].Err
		}

		// Log trace info
		log := logger.FromContext(c.Request.Context())
		log.WithFields(map[string]interface{}{
			"trace_id":    trace.TraceID,
			"path":        trace.Path,
			"method":      trace.Method,
			"duration_ms": trace.Duration.Milliseconds(),
			"status":      trace.StatusCode,
			"error":       trace.Error,
		}).Info("Request completed")
	}
} 