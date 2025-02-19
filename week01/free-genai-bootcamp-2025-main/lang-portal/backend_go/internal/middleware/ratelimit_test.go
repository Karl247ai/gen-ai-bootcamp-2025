package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/your-org/lang-portal/internal/metrics"
)

func TestRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)

	t.Run("basic rate limiting", func(t *testing.T) {
		limiter := NewRateLimiter(2, time.Minute, m)
		router := gin.New()
		router.Use(limiter.Middleware())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// First request should succeed
		w := performRequest(router, "GET", "/test", "127.0.0.1")
		assert.Equal(t, http.StatusOK, w.Code)

		// Second request should succeed
		w = performRequest(router, "GET", "/test", "127.0.0.1")
		assert.Equal(t, http.StatusOK, w.Code)

		// Third request should be rate limited
		w = performRequest(router, "GET", "/test", "127.0.0.1")
		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		// Verify metrics
		metrics, err := reg.Gather()
		assert.NoError(t, err)
		found := false
		for _, metric := range metrics {
			if metric.GetName() == "ratelimit_exceeded" {
				found = true
				assert.Equal(t, float64(1), metric.GetMetric()[0].GetCounter().GetValue())
			}
		}
		assert.True(t, found, "Expected metric not found")
	})

	t.Run("window reset", func(t *testing.T) {
		limiter := NewRateLimiter(1, 100*time.Millisecond, m)
		router := gin.New()
		router.Use(limiter.Middleware())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// First request should succeed
		w := performRequest(router, "GET", "/test", "127.0.0.1")
		assert.Equal(t, http.StatusOK, w.Code)

		// Second request should be rate limited
		w = performRequest(router, "GET", "/test", "127.0.0.1")
		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		// Wait for window to reset
		time.Sleep(150 * time.Millisecond)

		// Request should succeed after window reset
		w = performRequest(router, "GET", "/test", "127.0.0.1")
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func performRequest(r http.Handler, method, path, ip string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("X-Real-IP", ip)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
} 