package testutil

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

// MaintenanceWindow represents a maintenance window
type MaintenanceWindow struct {
	Name        string
	StartTime   time.Time
	Duration    time.Duration
	Labels      map[string]string
	Annotations map[string]string
	Status      string
}

// MaintenanceWindowTester provides utilities for testing maintenance windows
type MaintenanceWindowTester struct {
	t        *testing.T
	registry *prometheus.Registry
	windows  map[string]MaintenanceWindow
}

// NewMaintenanceWindowTester creates a new maintenance window tester
func NewMaintenanceWindowTester(t *testing.T, registry *prometheus.Registry) *MaintenanceWindowTester {
	return &MaintenanceWindowTester{
		t:        t,
		registry: registry,
		windows:  make(map[string]MaintenanceWindow),
	}
}

// StartWindow starts a maintenance window
func (mwt *MaintenanceWindowTester) StartWindow(window MaintenanceWindow) {
	window.StartTime = time.Now()
	window.Status = "active"
	mwt.windows[window.Name] = window

	// Record start metric
	mwt.recordMetric(window, 1)
}

// EndWindow ends a maintenance window
func (mwt *MaintenanceWindowTester) EndWindow(name string) {
	window, exists := mwt.windows[name]
	assert.True(mwt.t, exists, "Maintenance window not found: %s", name)

	window.Duration = time.Since(window.StartTime)
	window.Status = "completed"
	mwt.windows[name] = window

	// Record end metric
	mwt.recordMetric(window, 0)
}

// VerifyStatus verifies maintenance window status
func (mwt *MaintenanceWindowTester) VerifyStatus(name, expectedStatus string) {
	window, exists := mwt.windows[name]
	assert.True(mwt.t, exists, "Maintenance window not found: %s", name)
	assert.Equal(mwt.t, expectedStatus, window.Status,
		"Maintenance window status mismatch: %s", name)
}

// VerifyDuration verifies maintenance window duration
func (mwt *MaintenanceWindowTester) VerifyDuration(name string, expectedDuration time.Duration) {
	window, exists := mwt.windows[name]
	assert.True(mwt.t, exists, "Maintenance window not found: %s", name)

	if window.Status == "completed" {
		assert.InDelta(mwt.t, expectedDuration.Seconds(), window.Duration.Seconds(), 1,
			"Maintenance window duration mismatch: %s", name)
	}
}

// VerifyLabels verifies maintenance window labels
func (mwt *MaintenanceWindowTester) VerifyLabels(name string, expectedLabels map[string]string) {
	window, exists := mwt.windows[name]
	assert.True(mwt.t, exists, "Maintenance window not found: %s", name)

	for k, v := range expectedLabels {
		assert.Equal(mwt.t, v, window.Labels[k],
			"Maintenance window %s label mismatch: %s", name, k)
	}
}

// recordMetric records a maintenance window metric
func (mwt *MaintenanceWindowTester) recordMetric(window MaintenanceWindow, value float64) {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "maintenance_window_active",
		Help: "Indicates if a maintenance window is active",
	})
	mwt.registry.MustRegister(gauge)
	gauge.Set(value)
}

// GetWindow gets a maintenance window by name
func (mwt *MaintenanceWindowTester) GetWindow(name string) (MaintenanceWindow, bool) {
	window, exists := mwt.windows[name]
	return window, exists
}

// GetAllWindows gets all maintenance windows
func (mwt *MaintenanceWindowTester) GetAllWindows() map[string]MaintenanceWindow {
	return mwt.windows
} 