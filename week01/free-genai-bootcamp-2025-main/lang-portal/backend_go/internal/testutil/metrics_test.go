package testutil

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/your-org/lang-portal/internal/metrics"
)

func NewTestMetrics(t testing.TB) *metrics.Metrics {
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)
	return m
}

func AssertMetricRange(t testing.TB, m *metrics.Metrics, name string, min, max float64) {
	metrics, err := m.Registry().Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, metric := range metrics {
		if metric.GetName() == name {
			value := metric.GetMetric()[0].GetGauge().GetValue()
			if value < min || value > max {
				t.Errorf("Metric %s value %f outside range [%f, %f]", name, value, min, max)
			}
			return
		}
	}
	t.Errorf("Metric %s not found", name)
}

func AssertMetricLatency(t testing.TB, m *metrics.Metrics, name string, maxLatency time.Duration) {
	metrics, err := m.Registry().Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, metric := range metrics {
		if metric.GetName() == name {
			hist := metric.GetMetric()[0].GetHistogram()
			for _, bucket := range hist.GetBucket() {
				if bucket.GetUpperBound() > float64(maxLatency.Seconds()) {
					t.Errorf("Latency %s exceeds maximum %v", name, maxLatency)
					return
				}
			}
			return
		}
	}
	t.Errorf("Latency metric %s not found", name)
} 