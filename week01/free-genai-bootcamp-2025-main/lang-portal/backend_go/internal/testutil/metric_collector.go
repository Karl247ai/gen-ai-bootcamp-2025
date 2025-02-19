package testutil

import (
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// MetricCollector provides utilities for collecting and analyzing metrics
type MetricCollector struct {
	t        *testing.T
	registry *prometheus.Registry
	samples  map[string][]Sample
	mu       sync.RWMutex
}

// Sample represents a metric sample
type Sample struct {
	Value     float64
	Timestamp time.Time
	Labels    map[string]string
}

// NewMetricCollector creates a new metric collector
func NewMetricCollector(t *testing.T, registry *prometheus.Registry) *MetricCollector {
	return &MetricCollector{
		t:        t,
		registry: registry,
		samples:  make(map[string][]Sample),
	}
}

// Collect collects metric samples
func (mc *MetricCollector) Collect(metricName string) error {
	metrics, err := mc.registry.Gather()
	if err != nil {
		return err
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	for _, mf := range metrics {
		if mf.GetName() == metricName {
			for _, m := range mf.GetMetric() {
				sample := Sample{
					Timestamp: time.Now(),
					Labels:    make(map[string]string),
				}

				// Extract value based on metric type
				switch {
				case m.Gauge != nil:
					sample.Value = m.GetGauge().GetValue()
				case m.Counter != nil:
					sample.Value = m.GetCounter().GetValue()
				case m.Histogram != nil:
					sample.Value = float64(m.GetHistogram().GetSampleCount())
				}

				// Extract labels
				for _, l := range m.GetLabel() {
					sample.Labels[l.GetName()] = l.GetValue()
				}

				mc.samples[metricName] = append(mc.samples[metricName], sample)
			}
		}
	}

	return nil
}

// GetSamples returns collected samples for a metric
func (mc *MetricCollector) GetSamples(metricName string) []Sample {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.samples[metricName]
}

// CalculateRate calculates the rate of change for a metric
func (mc *MetricCollector) CalculateRate(metricName string, duration time.Duration) float64 {
	samples := mc.GetSamples(metricName)
	if len(samples) < 2 {
		return 0
	}

	first := samples[0]
	last := samples[len(samples)-1]
	timeDiff := last.Timestamp.Sub(first.Timestamp).Seconds()
	valueDiff := last.Value - first.Value

	return valueDiff / timeDiff
}

// Reset clears collected samples
func (mc *MetricCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.samples = make(map[string][]Sample)
} 