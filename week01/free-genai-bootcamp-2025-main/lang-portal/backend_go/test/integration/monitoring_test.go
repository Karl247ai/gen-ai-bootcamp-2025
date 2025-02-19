package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/your-org/lang-portal/internal/config"
	"github.com/your-org/lang-portal/internal/metrics"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestMonitoringMetrics(t *testing.T) {
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			Metrics: config.MetricsConfig{
				Enabled: true,
			},
		},
	}

	server := setupTestServer(t, cfg)
	ma := testutil.NewMetricsAssertion(t, server.Metrics)

	t.Run("api metrics", func(t *testing.T) {
		// Generate test traffic
		for i := 0; i < 10; i++ {
			testutil.PerformRequest(server.Router, "GET", "/api/v1/words", nil)
		}

		// Verify request count
		count := ma.GetCounterValue("handler_request_count", map[string]string{
			"endpoint": "/api/v1/words",
			"method":   "GET",
		})
		assert.Equal(t, float64(10), count)

		// Verify latency metrics
		latency := ma.GetHistogramValue("handler_request_duration")
		assert.Greater(t, latency.Count, uint64(0))
	})

	t.Run("cache metrics", func(t *testing.T) {
		// Generate cache activity
		for i := 0; i < 5; i++ {
			testutil.PerformRequest(server.Router, "GET", "/api/v1/words/1", nil)
		}

		// Verify cache hits
		hits := ma.GetCounterValue("cache_get_hit_total", nil)
		misses := ma.GetCounterValue("cache_get_miss_total", nil)
		assert.Greater(t, hits+misses, float64(0))
	})

	t.Run("error metrics", func(t *testing.T) {
		// Generate error
		testutil.PerformRequest(server.Router, "GET", "/api/v1/words/999", nil)

		// Verify error count
		errors := ma.GetCounterValue("handler_error_count", map[string]string{
			"type": "not_found",
		})
		assert.Greater(t, errors, float64(0))
	})
}

func TestMonitoringAlerts(t *testing.T) {
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			Alerts: config.AlertConfig{
				Enabled: true,
			},
		},
	}

	server := setupTestServer(t, cfg)
	ma := testutil.NewMetricsAssertion(t, server.Metrics)

	t.Run("high latency alert", func(t *testing.T) {
		// Simulate slow requests
		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			testutil.PerformRequest(server.Router, "GET", "/api/v1/words", nil)
		}

		// Verify alert would fire
		alertWouldFire := ma.CheckAlert("HighLatency")
		assert.True(t, alertWouldFire)
	})

	t.Run("error rate alert", func(t *testing.T) {
		// Generate errors
		for i := 0; i < 10; i++ {
			testutil.PerformRequest(server.Router, "GET", "/api/v1/words/999", nil)
		}

		// Verify alert would fire
		alertWouldFire := ma.CheckAlert("HighErrorRate")
		assert.True(t, alertWouldFire)
	})
}

func TestMonitoringDashboards(t *testing.T) {
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			Dashboards: config.DashboardConfig{
				Enabled: true,
			},
		},
	}

	server := setupTestServer(t, cfg)
	ma := testutil.NewMetricsAssertion(t, server.Metrics)

	t.Run("dashboard metrics available", func(t *testing.T) {
		// Generate activity
		testutil.GenerateTestTraffic(t, server.Router)

		// Verify metrics for dashboards exist
		metrics := []string{
			"handler_request_count",
			"handler_request_duration_bucket",
			"handler_error_count",
			"cache_get_hit_total",
			"cache_get_miss_total",
		}

		for _, m := range metrics {
			exists := ma.MetricExists(m)
			assert.True(t, exists, "Metric %s should exist", m)
		}
	})
}

func TestAlertRules(t *testing.T) {
	server := testutil.SetupTestServer(t)
	defer server.Cleanup()

	// Create alert rule tester
	art := testutil.NewAlertRuleTester(t, server.Registry)

	// Add test rules
	art.AddRule(testutil.AlertRule{
		Name:       "HighLatency",
		Expression: "handler_latency_seconds > 0.5",
		Duration:   5 * time.Minute,
		Labels: map[string]string{
			"severity": "warning",
			"team":     "platform",
		},
		Annotations: map[string]string{
			"description": "High latency detected",
			"runbook":     "docs/runbooks/high_latency.md",
		},
	})

	// Generate high latency condition
	server.WithSlowResponses(600 * time.Millisecond)
	testutil.GenerateTestTraffic(server.Router, 10)

	// Verify alert rule
	assert.True(t, art.VerifyExpression("HighLatency", true, 5*time.Second))
	art.VerifyLabels("HighLatency", map[string]string{
		"severity": "warning",
	})
}

func TestMaintenanceWindows(t *testing.T) {
	server := testutil.SetupTestServer(t)
	defer server.Cleanup()

	// Create maintenance window tester
	mwt := testutil.NewMaintenanceWindowTester(t, server.Registry)

	// Start maintenance window
	mwt.StartWindow(testutil.MaintenanceWindow{
		Name: "database_upgrade",
		Labels: map[string]string{
			"type": "upgrade",
			"team": "database",
		},
		Annotations: map[string]string{
			"description": "Database version upgrade",
			"ticket":      "MAINT-123",
		},
	})

	// Perform maintenance operations
	time.Sleep(2 * time.Second)
	testutil.GenerateTestTraffic(server.Router, 50)

	// End maintenance window
	mwt.EndWindow("database_upgrade")

	// Verify maintenance window
	mwt.VerifyStatus("database_upgrade", "completed")
	mwt.VerifyDuration("database_upgrade", 2*time.Second)
	mwt.VerifyLabels("database_upgrade", map[string]string{
		"type": "upgrade",
	})
}

func TestMetricValidation(t *testing.T) {
	server := testutil.SetupTestServer(t)
	defer server.Cleanup()

	// Create metric validator
	validator := testutil.NewMetricValidator(t, server.Registry)

	// Generate test metrics
	testutil.GenerateTestTraffic(server.Router, 10)

	// Verify metrics
	err := validator.ValidateMetrics()
	assert.NoError(t, err)
}

func TestMetricCollection(t *testing.T) {
	server := testutil.SetupTestServer(t)
	defer server.Cleanup()

	// Create metric collector
	collector := testutil.NewMetricCollector(t, server.Registry)

	// Generate test traffic
	testutil.GenerateTestTraffic(server.Router, 50)

	// Collect samples
	err := collector.Collect("handler_request_count")
	assert.NoError(t, err)

	// Calculate request rate
	rate := collector.CalculateRate("handler_request_count", time.Second)
	assert.Greater(t, rate, float64(0))
}

func TestRecordedTests(t *testing.T) {
	server := testutil.SetupTestServer(t)
	defer server.Cleanup()

	// Create test recorder
	recorder := testutil.NewTestRecorder(t, "test_results.json")

	// Start test recording
	record := recorder.StartTest("performance_test", map[string]string{
		"type": "latency",
		"env":  "test",
	})

	// Run test
	server.WithSlowResponses(100 * time.Millisecond)
	testutil.GenerateTestTraffic(server.Router, 20)

	// Record metrics
	ma := testutil.NewMetricsAssertion(t, server.Metrics)
	latency := ma.GetHistogramValue("handler_request_duration")
	recorder.RecordMetric(record, "avg_latency", latency.GetSampleSum()/float64(latency.GetSampleCount()))

	// End test recording
	recorder.EndTest(record, "completed")
} 