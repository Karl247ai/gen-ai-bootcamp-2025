package integration

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/your-org/lang-portal/internal/config"
	"github.com/your-org/lang-portal/internal/metrics"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestMonitoringPerformance(t *testing.T) {
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			Metrics: config.MetricsConfig{
				Enabled: true,
			},
		},
	}

	server := setupTestServer(t, cfg)
	ma := testutil.NewMetricsAssertion(t, server.Metrics, server.Metrics.Registry())

	t.Run("metric collection performance", func(t *testing.T) {
		// Generate metrics load
		for i := 0; i < 1000; i++ {
			server.SendRequest("GET", "/api/v1/words", nil)
		}

		// Measure collection time
		start := time.Now()
		metrics, err := server.Metrics.Registry().Gather()
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Less(t, duration.Milliseconds(), int64(100), "Metric collection took too long")
		assert.Less(t, len(metrics), 5000, "Too many time series")
	})

	t.Run("metric cardinality", func(t *testing.T) {
		// Generate requests with different labels
		paths := []string{"/api/v1/words", "/api/v1/groups", "/api/v1/study-sessions"}
		methods := []string{"GET", "POST", "PUT", "DELETE"}

		for _, path := range paths {
			for _, method := range methods {
				server.SendRequest(method, path, nil)
			}
		}

		// Check cardinality
		metrics, err := server.Metrics.Registry().Gather()
		require.NoError(t, err)

		labelSets := make(map[string]int)
		for _, metric := range metrics {
			if metric.GetName() == "handler_request_count" {
				for _, m := range metric.GetMetric() {
					key := getLabelKey(m.GetLabel())
					labelSets[key]++
				}
			}
		}

		assert.Less(t, len(labelSets), 100, "Too many unique label combinations")
	})

	t.Run("query performance", func(t *testing.T) {
		// Generate some data
		for i := 0; i < 100; i++ {
			server.SendRequest("GET", "/api/v1/words", nil)
			time.Sleep(10 * time.Millisecond)
		}

		// Test common queries
		queries := []string{
			`rate(handler_request_count{job="lang-portal"}[5m])`,
			`histogram_quantile(0.95, rate(handler_request_duration_bucket{job="lang-portal"}[5m]))`,
			`sum by (path) (rate(handler_error_count{job="lang-portal"}[5m]))`,
		}

		for _, query := range queries {
			start := time.Now()
			result := testutil.QueryPrometheus(t, query)
			duration := time.Since(start)

			assert.NotNil(t, result)
			assert.Less(t, duration.Milliseconds(), int64(500), 
				"Query took too long: %s", query)
		}
	})

	t.Run("resource usage", func(t *testing.T) {
		// Generate constant load
		done := make(chan bool)
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					server.SendRequest("GET", "/api/v1/words", nil)
					time.Sleep(10 * time.Millisecond)
				}
			}
		}()

		// Monitor resource usage
		time.Sleep(30 * time.Second)
		metrics, err := server.Metrics.Registry().Gather()
		require.NoError(t, err)

		var memoryUsage, cpuUsage float64
		for _, metric := range metrics {
			switch metric.GetName() {
			case "process_resident_memory_bytes":
				memoryUsage = metric.GetMetric()[0].GetGauge().GetValue()
			case "process_cpu_seconds_total":
				cpuUsage = metric.GetMetric()[0].GetCounter().GetValue()
			}
		}

		close(done)

		assert.Less(t, memoryUsage/(1024*1024), float64(500), 
			"Memory usage too high: %f MB", memoryUsage/(1024*1024))
		assert.Less(t, cpuUsage, float64(30), 
			"CPU usage too high: %f seconds", cpuUsage)
	})
}

func getLabelKey(labels []*prometheus.LabelPair) string {
	key := ""
	for _, label := range labels {
		key += label.GetName() + "=" + label.GetValue() + ","
	}
	return key
} 