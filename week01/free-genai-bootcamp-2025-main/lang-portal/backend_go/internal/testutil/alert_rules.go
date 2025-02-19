package testutil

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

// AlertRule represents a Prometheus alert rule
type AlertRule struct {
	Name        string
	Expression  string
	Duration    time.Duration
	Labels      map[string]string
	Annotations map[string]string
}

// AlertRuleTester provides utilities for testing alert rules
type AlertRuleTester struct {
	t        *testing.T
	registry *prometheus.Registry
	rules    map[string]AlertRule
}

// NewAlertRuleTester creates a new alert rule tester
func NewAlertRuleTester(t *testing.T, registry *prometheus.Registry) *AlertRuleTester {
	return &AlertRuleTester{
		t:        t,
		registry: registry,
		rules:    make(map[string]AlertRule),
	}
}

// AddRule adds an alert rule for testing
func (art *AlertRuleTester) AddRule(rule AlertRule) {
	art.rules[rule.Name] = rule
}

// VerifyExpression verifies an alert rule expression
func (art *AlertRuleTester) VerifyExpression(name string, expectedResult bool, timeout time.Duration) bool {
	rule, exists := art.rules[name]
	assert.True(art.t, exists, "Alert rule not found: %s", name)

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		result := art.evaluateExpr(rule.Expression)
		if result == expectedResult {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// VerifyLabels verifies alert rule labels
func (art *AlertRuleTester) VerifyLabels(name string, expectedLabels map[string]string) {
	rule, exists := art.rules[name]
	assert.True(art.t, exists, "Alert rule not found: %s", name)

	for k, v := range expectedLabels {
		assert.Equal(art.t, v, rule.Labels[k],
			"Alert rule %s label mismatch: %s", name, k)
	}
}

// VerifyAnnotations verifies alert rule annotations
func (art *AlertRuleTester) VerifyAnnotations(name string, expectedAnnotations map[string]string) {
	rule, exists := art.rules[name]
	assert.True(art.t, exists, "Alert rule not found: %s", name)

	for k, v := range expectedAnnotations {
		assert.Equal(art.t, v, rule.Annotations[k],
			"Alert rule %s annotation mismatch: %s", name, k)
	}
}

// evaluateExpr evaluates a Prometheus expression
func (art *AlertRuleTester) evaluateExpr(expr string) bool {
	metrics, err := art.registry.Gather()
	if err != nil {
		return false
	}

	// Simple expression evaluation based on metric presence
	for _, mf := range metrics {
		for _, m := range mf.GetMetric() {
			switch {
			case m.Gauge != nil:
				if m.GetGauge().GetValue() > 0 {
					return true
				}
			case m.Counter != nil:
				if m.GetCounter().GetValue() > 0 {
					return true
				}
			case m.Histogram != nil:
				if m.GetHistogram().GetSampleCount() > 0 {
					return true
				}
			}
		}
	}
	return false
}

// GetRule gets an alert rule by name
func (art *AlertRuleTester) GetRule(name string) (AlertRule, bool) {
	rule, exists := art.rules[name]
	return rule, exists
}

// GetAllRules gets all alert rules
func (art *AlertRuleTester) GetAllRules() map[string]AlertRule {
	return art.rules
} 