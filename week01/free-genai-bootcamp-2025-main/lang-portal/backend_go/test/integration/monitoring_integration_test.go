package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/your-org/lang-portal/internal/config"
	"github.com/your-org/lang-portal/internal/metrics"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestMonitoringEndpoints(t *testing.T) {
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			Metrics: config.MetricsConfig{
				Enabled: true,
				Path:    "/metrics",
			},
		},
	}

	server := setupTestServer(t, cfg)

	t.Run("metrics endpoint", func(t *testing.T) {
		w := server.SendRequest("GET", "/metrics", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
	})

	t.Run("health check", func(t *testing.T) {
		w := server.SendRequest("GET", "/health", nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Status  string            `json:"status"`
			Details map[string]string `json:"details"`
		}
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, "healthy", resp.Status)
		assert.Contains(t, resp.Details, "database")
	})

	t.Run("debug endpoints", func(t *testing.T) {
		endpoints := []string{
			"/debug/db/stats",
			"/debug/cache/stats",
			"/debug/metrics/reset",
			"/debug/goroutines",
		}

		for _, endpoint := range endpoints {
			w := server.SendRequest("GET", endpoint, nil)
			assert.Equal(t, http.StatusOK, w.Code, "Endpoint %s failed", endpoint)
		}
	})
}

func TestMetricsCollection(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)
	server := setupTestServer(t, &config.Config{})

	t.Run("request metrics", func(t *testing.T) {
		// Make some requests
		endpoints := []string{"/api/v1/words", "/api/v1/groups"}
		for _, endpoint := range endpoints {
			server.SendRequest("GET", endpoint, nil)
		}

		// Verify metrics
		metrics, err := reg.Gather()
		require.NoError(t, err)

		found := false
		for _, metric := range metrics {
			if metric.GetName() == "handler_request_count" {
				found = true
				assert.Greater(t, len(metric.GetMetric()), 0)
			}
		}
		assert.True(t, found, "Request count metric not found")
	})

	t.Run("error metrics", func(t *testing.T) {
		// Trigger some errors
		server.SendRequest("GET", "/api/v1/words/999999", nil)
		server.SendRequest("POST", "/api/v1/words", "invalid json")

		metrics, err := reg.Gather()
		require.NoError(t, err)

		found := false
		for _, metric := range metrics {
			if metric.GetName() == "handler_error_count" {
				found = true
				assert.Greater(t, len(metric.GetMetric()), 0)
			}
		}
		assert.True(t, found, "Error count metric not found")
	})

	t.Run("database metrics", func(t *testing.T) {
		// Make database operations
		for i := 0; i < 5; i++ {
			server.SendRequest("GET", "/api/v1/words", nil)
		}

		metrics, err := reg.Gather()
		require.NoError(t, err)

		expectedMetrics := []string{
			"db_connections_in_use",
			"db_query_duration_seconds",
		}

		for _, expected := range expectedMetrics {
			found := false
			for _, metric := range metrics {
				if metric.GetName() == expected {
					found = true
					break
				}
			}
			assert.True(t, found, "Metric %s not found", expected)
		}
	})
}

func TestResourceMonitoring(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)
	server := setupTestServer(t, &config.Config{})

	t.Run("memory metrics", func(t *testing.T) {
		// Allocate some memory
		data := make([][]byte, 10)
		for i := range data {
			data[i] = make([]byte, 1024*1024) // 1MB each
		}

		metrics, err := reg.Gather()
		require.NoError(t, err)

		memoryMetrics := []string{
			"runtime_memory_alloc_bytes",
			"runtime_memory_heap_bytes",
		}

		for _, expected := range memoryMetrics {
			found := false
			for _, metric := range metrics {
				if metric.GetName() == expected {
					found = true
					value := metric.GetMetric()[0].GetGauge().GetValue()
					assert.Greater(t, value, float64(0))
					break
				}
			}
			assert.True(t, found, "Metric %s not found", expected)
		}
	})

	t.Run("goroutine metrics", func(t *testing.T) {
		// Create some goroutines
		for i := 0; i < 10; i++ {
			go func() {
				time.Sleep(100 * time.Millisecond)
			}()
		}

		metrics, err := reg.Gather()
		require.NoError(t, err)

		found := false
		for _, metric := range metrics {
			if metric.GetName() == "runtime_goroutines" {
				found = true
				value := metric.GetMetric()[0].GetGauge().GetValue()
				assert.Greater(t, value, float64(10))
				break
			}
		}
		assert.True(t, found, "Goroutine metric not found")
	})
} 