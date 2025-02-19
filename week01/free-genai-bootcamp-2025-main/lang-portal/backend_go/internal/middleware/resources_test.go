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

func TestResourceMonitoring(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)

	router := gin.New()
	router.Use(ResourceMonitoring(m))
	router.GET("/test", func(c *gin.Context) {
		// Allocate some memory to test monitoring
		data := make([]byte, 1024*1024)
		for i := range data {
			data[i] = byte(i)
		}
		c.Status(http.StatusOK)
	})

	// Wait for metrics to be collected
	time.Sleep(20 * time.Second)

	// Make a request to trigger memory allocation
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify metrics
	metrics, err := reg.Gather()
	assert.NoError(t, err)

	expectedMetrics := []string{
		"runtime_memory_alloc",
		"runtime_memory_heap_alloc",
		"runtime_goroutines",
		"runtime_num_gc",
	}

	for _, name := range expectedMetrics {
		found := false
		for _, metric := range metrics {
			if metric.GetName() == name {
				found = true
				assert.Greater(t, metric.GetMetric()[0].GetGauge().GetValue(), float64(0))
				break
			}
		}
		assert.True(t, found, "Expected metric %s not found", name)
	}
} 