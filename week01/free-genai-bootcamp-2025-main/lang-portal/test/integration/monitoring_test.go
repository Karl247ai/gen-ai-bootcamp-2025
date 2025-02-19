package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/your-org/lang-portal/internal/metrics"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestMonitoringIntegration(t *testing.T) {
	// Setup test environment
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../migrations/0001_init.sql")

	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)
	server := setupTestAPI(t)

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
		json.NewDecoder(w.Body).Decode(&resp)

		assert.Equal(t, "healthy", resp.Status)
		assert.Contains(t, resp.Details, "database")
	})

	t.Run("request metrics", func(t *testing.T) {
		// Create a word and verify metrics
		word := models.Word{
			Japanese: "テスト",
			Romaji:   "tesuto",
			English:  "test",
		}
		w := server.SendRequest("POST", "/api/v1/words", word)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Wait for metrics to be updated
		time.Sleep(100 * time.Millisecond)

		metrics, err := reg.Gather()
		assert.NoError(t, err)

		found := false
		for _, metric := range metrics {
			if metric.GetName() == "handler_word_create_success" {
				found = true
				assert.Equal(t, float64(1), metric.GetMetric()[0].GetCounter().GetValue())
			}
		}
		assert.True(t, found, "Expected metric not found")
	})

	t.Run("error metrics", func(t *testing.T) {
		// Trigger an error and verify metrics
		w := server.SendRequest("GET", "/api/v1/words/999999", nil)
		assert.Equal(t, http.StatusNotFound, w.Code)

		time.Sleep(100 * time.Millisecond)

		metrics, err := reg.Gather()
		assert.NoError(t, err)

		found := false
		for _, metric := range metrics {
			if metric.GetName() == "handler_word_get_error" {
				found = true
				assert.Equal(t, float64(1), metric.GetMetric()[0].GetCounter().GetValue())
			}
		}
		assert.True(t, found, "Expected error metric not found")
	})
}

func TestMonitoringSystem(t *testing.T) {
	server := testutil.SetupTestServer(t)
	defer server.Cleanup()

	t.Run("basic metrics", func(t *testing.T) {
		ma := testutil.NewMetricsAssertion(t, server.Metrics)

		// Generate traffic
		testutil.GenerateTestTraffic(server.Router, 10)

		// Verify basic metrics
		assert.True(t, ma.MetricExists("request_total"))
		assert.True(t, ma.MetricExists("request_duration_seconds"))
	})

	t.Run("error metrics", func(t *testing.T) {
		ma := testutil.NewMetricsAssertion(t, server.Metrics)

		// Generate errors
		server.WithErrorRate(0.5)
		testutil.GenerateTestTraffic(server.Router, 20)

		// Verify error metrics
		errors := ma.GetCounterValue("error_total", nil)
		assert.Greater(t, errors, float64(5))
	})

	t.Run("performance metrics", func(t *testing.T) {
		ma := testutil.NewMetricsAssertion(t, server.Metrics)

		// Generate slow responses
		server.WithSlowResponses(200 * time.Millisecond)
		testutil.GenerateTestTraffic(server.Router, 10)

		// Verify latency metrics
		histogram := ma.GetHistogramValue("request_duration_seconds")
		assert.Greater(t, histogram.GetSampleSum()/float64(histogram.GetSampleCount()), 0.1)
	})

	t.Run("resource metrics", func(t *testing.T) {
		ma := testutil.NewMetricsAssertion(t, server.Metrics)

		// Verify resource metrics exist
		assert.True(t, ma.MetricExists("process_cpu_seconds_total"))
		assert.True(t, ma.MetricExists("process_resident_memory_bytes"))
		assert.True(t, ma.MetricExists("go_goroutines"))
	})
} 