package testutil

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/your-org/lang-portal/internal/metrics"
)

type MetricsAssertion struct {
	t       testing.TB
	metrics *metrics.Metrics
	reg     *prometheus.Registry
}

func NewMetricsAssertion(t testing.TB, m *metrics.Metrics, reg *prometheus.Registry) *MetricsAssertion {
	return &MetricsAssertion{
		t:       t,
		metrics: m,
		reg:     reg,
	}
}

func (ma *MetricsAssertion) AssertCounterValue(name string, expected float64) {
	metrics, err := ma.reg.Gather()
	if err != nil {
		ma.t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, m := range metrics {
		if m.GetName() == name {
			value := m.GetMetric()[0].GetCounter().GetValue()
			if value != expected {
				ma.t.Errorf("Expected counter %s to be %f, got %f", name, expected, value)
			}
			return
		}
	}
	ma.t.Errorf("Counter %s not found", name)
}

func (ma *MetricsAssertion) AssertHistogramCount(name string, expectedCount uint64) {
	metrics, err := ma.reg.Gather()
	if err != nil {
		ma.t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, m := range metrics {
		if m.GetName() == name {
			count := m.GetMetric()[0].GetHistogram().GetSampleCount()
			if count != expectedCount {
				ma.t.Errorf("Expected histogram %s to have %d samples, got %d", name, expectedCount, count)
			}
			return
		}
	}
	ma.t.Errorf("Histogram %s not found", name)
}

func (ma *MetricsAssertion) AssertHistogramBounds(name string, maxValue float64) {
	metrics, err := ma.reg.Gather()
	if err != nil {
		ma.t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, m := range metrics {
		if m.GetName() == name {
			hist := m.GetMetric()[0].GetHistogram()
			for _, bucket := range hist.GetBucket() {
				if bucket.GetUpperBound() > maxValue {
					ma.t.Errorf("Histogram %s has value greater than %f", name, maxValue)
					return
				}
			}
			return
		}
	}
	ma.t.Errorf("Histogram %s not found", name)
} 