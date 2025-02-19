package testutil

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/your-org/lang-portal/internal/metrics"
)

// MetricsAssertion provides utilities for testing metrics
type MetricsAssertion struct {
	t       *testing.T
	metrics *metrics.Metrics
}

// NewMetricsAssertion creates a new MetricsAssertion
func NewMetricsAssertion(t *testing.T, m *metrics.Metrics) *MetricsAssertion {
	return &MetricsAssertion{
		t:       t,
		metrics: m,
	}
}

// GetCounterValue returns the value of a counter metric
func (ma *MetricsAssertion) GetCounterValue(name string, labels map[string]string) float64 {
	metric := ma.getMetric(name, labels)
	if metric == nil {
		ma.t.Fatalf("Metric %s not found", name)
	}
	return metric.GetCounter().GetValue()
}

// GetHistogramValue returns histogram data
func (ma *MetricsAssertion) GetHistogramValue(name string) *dto.Histogram {
	metric := ma.getMetric(name, nil)
	if metric == nil {
		ma.t.Fatalf("Metric %s not found", name)
	}
	return metric.GetHistogram()
}

// MetricExists checks if a metric exists
func (ma *MetricsAssertion) MetricExists(name string) bool {
	return ma.getMetric(name, nil) != nil
}

// CheckAlert verifies if an alert would fire
func (ma *MetricsAssertion) CheckAlert(name string) bool {
	// Wait for metrics to be collected
	time.Sleep(100 * time.Millisecond)

	switch name {
	case "HighLatency":
		value := ma.GetHistogramValue("handler_request_duration")
		return value.GetSampleSum()/float64(value.GetSampleCount()) > 0.5

	case "HighErrorRate":
		errors := ma.GetCounterValue("handler_error_count", nil)
		requests := ma.GetCounterValue("handler_request_count", nil)
		return errors/requests > 0.05

	default:
		ma.t.Fatalf("Unknown alert: %s", name)
		return false
	}
}

// getMetric retrieves a metric by name and labels
func (ma *MetricsAssertion) getMetric(name string, labels map[string]string) *dto.Metric {
	mf, err := ma.metrics.Registry().Gather()
	if err != nil {
		ma.t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, f := range mf {
		if f.GetName() == name {
			for _, m := range f.GetMetric() {
				if ma.matchLabels(m, labels) {
					return m
				}
			}
		}
	}
	return nil
}

// matchLabels checks if a metric's labels match the expected labels
func (ma *MetricsAssertion) matchLabels(metric *dto.Metric, expected map[string]string) bool {
	if expected == nil {
		return true
	}

	actual := make(map[string]string)
	for _, l := range metric.GetLabel() {
		actual[l.GetName()] = l.GetValue()
	}

	for k, v := range expected {
		if actual[k] != v {
			return false
		}
	}
	return true
}

func (ma *MetricsAssertion) AssertCounterValue(name string, expected float64) {
	metrics, err := ma.metrics.Registry().Gather()
	assert.NoError(ma.t, err)

	found := false
	for _, metric := range metrics {
		if metric.GetName() == name {
			found = true
			assert.Equal(ma.t, expected, metric.GetMetric()[0].GetCounter().GetValue())
			break
		}
	}
	assert.True(ma.t, found, "Metric %s not found", name)
}

func (ma *MetricsAssertion) AssertCounterRatio(numerator, denominator string, minRatio float64) {
	metrics, err := ma.metrics.Registry().Gather()
	assert.NoError(ma.t, err)

	var num, denom float64
	for _, metric := range metrics {
		if metric.GetName() == numerator {
			num = metric.GetMetric()[0].GetCounter().GetValue()
		}
		if metric.GetName() == denominator {
			denom = metric.GetMetric()[0].GetCounter().GetValue()
		}
	}

	assert.Greater(ma.t, denom, float64(0), "Denominator metric %s is zero", denominator)
	ratio := num / denom
	assert.GreaterOrEqual(ma.t, ratio, minRatio, "Ratio %f is below minimum %f", ratio, minRatio)
}

func (ma *MetricsAssertion) AssertAlertFiring(alertName string, check func(*Alert) bool) {
	const maxWait = 30 * time.Second
	const interval = 100 * time.Millisecond

	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		metrics, err := ma.metrics.Registry().Gather()
		assert.NoError(ma.t, err)

		for _, metric := range metrics {
			if metric.GetName() == "alert_"+alertName {
				alert := &Alert{
					Name:   alertName,
					Labels: make(map[string]string),
					Value:  metric.GetMetric()[0].GetGauge().GetValue(),
				}
				for _, label := range metric.GetMetric()[0].GetLabel() {
					alert.Labels[label.GetName()] = label.GetValue()
				}
				if check(alert) {
					return
				}
			}
		}
		time.Sleep(interval)
	}
	ma.t.Errorf("Alert %s not firing within %v", alertName, maxWait)
}

type Alert struct {
	Name   string
	Labels map[string]string
	Value  float64
} 