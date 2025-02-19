package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// RequestLogger logs request and response details
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Read request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create buffer for response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Format request body for logging
		var requestJSON interface{}
		if len(requestBody) > 0 {
			json.Unmarshal(requestBody, &requestJSON)
		}

		// Format response body for logging
		var responseJSON interface{}
		if blw.body.Len() > 0 {
			json.Unmarshal(blw.body.Bytes(), &responseJSON)
		}

		// Log request/response details
		logEntry := map[string]interface{}{
			"timestamp":     time.Now().Format(time.RFC3339),
			"duration_ms":   duration.Milliseconds(),
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"status":        c.Writer.Status(),
			"client_ip":     c.ClientIP(),
			"request_body":  requestJSON,
			"response_body": responseJSON,
		}

		logJSON, _ := json.Marshal(logEntry)
		c.Request.Context().Value("logger").(interface{ Info(string) }).Info(string(logJSON))
	}
} 