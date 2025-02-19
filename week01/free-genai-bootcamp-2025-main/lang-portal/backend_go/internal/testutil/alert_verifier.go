package testutil

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

// AlertVerifier provides utilities for verifying alert conditions
type AlertVerifier struct {
	t        *testing.T
	registry *prometheus.Registry
}

// NewAlertVerifier creates a new alert verifier
func NewAlertVerifier(t *testing.T, registry *prometheus.Registry) *AlertVerifier {
	return &AlertVerifier{
		t:        t,
		registry: registry,
	}
}

// VerifyAlert checks if an alert would fire
func (av *AlertVerifier) VerifyAlert(name string, duration time.Duration) bool {
	deadline := time.Now().Add(duration)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		<-ticker.C
		if av.isAlertFiring(name) {
			return true
		}
	}
	return false
}

// isAlertFiring checks if an alert is currently firing
func (av *AlertVerifier) isAlertFiring(name string) bool {
	metrics, err := av.registry.Gather()
	if err != nil {
		av.t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, mf := range metrics {
		if mf.GetName() == "alert_"+name+"_firing" {
			for _, m := range mf.GetMetric() {
				if m.GetGauge().GetValue() > 0 {
					return true
				}
			}
		}
	}
	return false
}

// AssertNoAlerts verifies no alerts are firing
func (av *AlertVerifier) AssertNoAlerts(t *testing.T) {
	metrics, err := av.registry.Gather()
	assert.NoError(t, err)

	for _, mf := range metrics {
		if len(mf.GetName()) >= 6 && mf.GetName()[:6] == "alert_" {
			for _, m := range mf.GetMetric() {
				assert.Equal(t, float64(0), m.GetGauge().GetValue(),
					"Alert %s is firing", mf.GetName())
			}
		}
	}
}

// GetAlertMetric retrieves the metric for an alert
func (av *AlertVerifier) GetAlertMetric(name string) *dto.Metric {
	metrics, err := av.registry.Gather()
	if err != nil {
		av.t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, mf := range metrics {
		if mf.GetName() == "alert_"+name+"_firing" {
			if len(mf.GetMetric()) > 0 {
				return mf.GetMetric()[0]
			}
		}
	}
	return nil
} 