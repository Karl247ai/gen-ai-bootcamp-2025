package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/your-org/lang-portal/internal/cache"
	"github.com/your-org/lang-portal/internal/config"
	"github.com/your-org/lang-portal/internal/metrics"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestMonitoringIntegration(t *testing.T) {
	server := testutil.SetupTestServer(t)
	defer server.Cleanup()

	// Test recorder for tracking test execution
	recorder := testutil.NewTestRecorder(t, "test_results.json")
	record := recorder.StartTest("monitoring_integration", nil)
	defer func() {
		recorder.EndTest(record, "completed")
	}()

	t.Run("system metrics", func(t *testing.T) {
		ma := testutil.NewMetricsAssertion(t, server.Metrics)

		metrics := []string{
			"process_cpu_seconds_total",
			"process_resident_memory_bytes",
			"go_goroutines",
			"go_threads",
		}

		for _, m := range metrics {
			assert.True(t, ma.MetricExists(m))
		}
	})

	t.Run("application metrics", func(t *testing.T) {
		ma := testutil.NewMetricsAssertion(t, server.Metrics)

		metrics := []string{
			"request_total",
			"request_duration_seconds",
			"error_total",
			"cache_hits_total",
		}

		for _, m := range metrics {
			assert.True(t, ma.MetricExists(m))
		}
	})

	t.Run("alert integration", func(t *testing.T) {
		av := testutil.NewAlertVerifier(t, server.Registry)

		// Generate alert conditions
		server.WithErrorRate(0.5)
		testutil.GenerateTestTraffic(server.Router, 50)

		// Verify alert firing
		assert.True(t, av.VerifyAlert("HighErrorRate", 5*time.Second))

		// Fix conditions
		server.WithErrorRate(0)
		testutil.GenerateTestTraffic(server.Router, 50)

		// Verify alert recovery
		assert.True(t, av.VerifyAlertRecovery("HighErrorRate", 5*time.Second))
	})
}

func setupTestServer(t *testing.T, cfg *config.Config) *testutil.APITestServer {
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../migrations/0001_init.sql")

	metrics := testutil.NewTestMetrics(t)
	cache := cache.NewMonitoredCache(cache.NewInMemoryCache(1000), metrics)

	server := testutil.NewAPITestServer(t)
	setupTestHandlers(server.Engine, db, cache, metrics, cfg)

	return server
} 