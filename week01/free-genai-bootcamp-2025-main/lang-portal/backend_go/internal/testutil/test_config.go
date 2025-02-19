package testutil

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// TestConfig provides configuration for monitoring tests
type TestConfig struct {
	// Test timeouts
	DefaultTimeout    time.Duration
	AlertTimeout     time.Duration
	MetricTimeout    time.Duration
	ShutdownTimeout  time.Duration

	// Retry settings
	RetryAttempts    int
	RetryInterval    time.Duration

	// Metric settings
	MetricPrefix     string
	MaxLabelCount    int
	MaxMetricValue   float64

	// Alert settings
	AlertNamePrefix  string
	MinAlertDuration time.Duration

	// Maintenance settings
	MaintenancePrefix string
	MaxWindowDuration time.Duration

	// Cleanup settings
	CleanupEnabled   bool
	CleanupDelay     time.Duration
}

// DefaultTestConfig returns default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		DefaultTimeout:    10 * time.Second,
		AlertTimeout:     5 * time.Second,
		MetricTimeout:    2 * time.Second,
		ShutdownTimeout:  1 * time.Second,

		RetryAttempts:    3,
		RetryInterval:    time.Second,

		MetricPrefix:     "test_",
		MaxLabelCount:    10,
		MaxMetricValue:   1e6,

		AlertNamePrefix:  "Test_",
		MinAlertDuration: time.Minute,

		MaintenancePrefix: "test_maintenance_",
		MaxWindowDuration: time.Hour,

		CleanupEnabled:   true,
		CleanupDelay:     100 * time.Millisecond,
	}
}

// WithTimeouts sets test timeouts
func (c *TestConfig) WithTimeouts(timeouts map[string]time.Duration) *TestConfig {
	for name, duration := range timeouts {
		switch name {
		case "default":
			c.DefaultTimeout = duration
		case "alert":
			c.AlertTimeout = duration
		case "metric":
			c.MetricTimeout = duration
		case "shutdown":
			c.ShutdownTimeout = duration
		}
	}
	return c
}

// WithRetry sets retry configuration
func (c *TestConfig) WithRetry(attempts int, interval time.Duration) *TestConfig {
	c.RetryAttempts = attempts
	c.RetryInterval = interval
	return c
}

// WithMetricSettings sets metric configuration
func (c *TestConfig) WithMetricSettings(prefix string, maxLabels int, maxValue float64) *TestConfig {
	c.MetricPrefix = prefix
	c.MaxLabelCount = maxLabels
	c.MaxMetricValue = maxValue
	return c
} 