package testutil

import (
	"regexp"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

// MetricValidator provides utilities for validating metrics
type MetricValidator struct {
	t        *testing.T
	registry *prometheus.Registry
	rules    ValidationRules
}

// ValidationRules defines metric validation rules
type ValidationRules struct {
	MaxLabelLength     int
	AllowedLabelChars  string
	MaxMetricValue     float64
	DisallowedPrefixes []string
}

// NewMetricValidator creates a new metric validator
func NewMetricValidator(t *testing.T, registry *prometheus.Registry) *MetricValidator {
	return &MetricValidator{
		t:        t,
		registry: registry,
		rules: ValidationRules{
			MaxLabelLength:     100,
			AllowedLabelChars: "^[a-zA-Z0-9_]+$",
			MaxMetricValue:    1e6,
			DisallowedPrefixes: []string{
				"test_",
				"temp_",
				"debug_",
			},
		},
	}
}

// ValidateMetrics validates all metrics against rules
func (mv *MetricValidator) ValidateMetrics() error {
	metrics, err := mv.registry.Gather()
	if err != nil {
		return err
	}

	for _, mf := range metrics {
		// Validate metric name
		mv.validateMetricName(mf.GetName())

		// Validate each metric instance
		for _, m := range mf.GetMetric() {
			mv.validateLabels(m.GetLabel())
			mv.validateValue(mf.GetName(), m)
		}
	}

	return nil
}

// validateMetricName validates a metric name
func (mv *MetricValidator) validateMetricName(name string) {
	for _, prefix := range mv.rules.DisallowedPrefixes {
		assert.NotContains(mv.t, name, prefix,
			"Metric name contains disallowed prefix: %s", name)
	}
}

// validateLabels validates metric labels
func (mv *MetricValidator) validateLabels(labels []*dto.LabelPair) {
	for _, label := range labels {
		// Check label name length
		assert.Less(mv.t, len(label.GetName()), mv.rules.MaxLabelLength,
			"Label name too long: %s", label.GetName())

		// Check label value length
		assert.Less(mv.t, len(label.GetValue()), mv.rules.MaxLabelLength,
			"Label value too long: %s", label.GetValue())

		// Check label characters
		matched, err := regexp.MatchString(mv.rules.AllowedLabelChars, label.GetValue())
		assert.NoError(mv.t, err)
		assert.True(mv.t, matched,
			"Label value contains invalid characters: %s", label.GetValue())
	}
}

// validateValue validates a metric value
func (mv *MetricValidator) validateValue(name string, m *dto.Metric) {
	switch {
	case m.Gauge != nil:
		assert.Less(mv.t, m.GetGauge().GetValue(), mv.rules.MaxMetricValue,
			"Gauge value too high for %s", name)
	case m.Counter != nil:
		assert.Less(mv.t, m.GetCounter().GetValue(), mv.rules.MaxMetricValue,
			"Counter value too high for %s", name)
	case m.Histogram != nil:
		assert.Less(mv.t, m.GetHistogram().GetSampleSum(), mv.rules.MaxMetricValue,
			"Histogram sum too high for %s", name)
	}
} 