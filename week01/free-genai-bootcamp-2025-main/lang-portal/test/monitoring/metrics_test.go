package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/your-org/lang-portal/internal/metrics"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestMetricsCollection(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)

	t.Run("request metrics", func(t *testing.T) {
		m.IncCounter("handler.word.create.success")
		m.ObserveHistogram("handler.word.create", 0.1)

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

	t.Run("database metrics", func(t *testing.T) {
		m.SetGauge("db.connections.in_use", 5)
		m.SetGauge("db.connections.idle", 3)

		metrics, err := reg.Gather()
		assert.NoError(t, err)

		for _, metric := range metrics {
			switch metric.GetName() {
			case "db_connections_in_use":
				assert.Equal(t, float64(5), metric.GetMetric()[0].GetGauge().GetValue())
			case "db_connections_idle":
				assert.Equal(t, float64(3), metric.GetMetric()[0].GetGauge().GetValue())
			}
		}
	})
}

func TestMetricsPerformance(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)

	perf := testutil.NewPerformanceTest(t)
	perf.ConcurrentUsers = 100
	perf.Duration = 5 * time.Second

	results := perf.Run("metrics_collection", func(ctx context.Context) error {
		m.IncCounter("test.counter")
		m.ObserveHistogram("test.latency", 0.001)
		m.SetGauge("test.gauge", float64(1))
		return nil
	})

	results.Assert(t, 1*time.Millisecond, 0.001)
} 