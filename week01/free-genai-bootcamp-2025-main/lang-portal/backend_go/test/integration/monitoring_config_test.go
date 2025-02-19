package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/your-org/lang-portal/internal/config"
)

func TestMonitoringConfiguration(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		cfg := &config.Config{
			Monitoring: config.MonitoringConfig{
				Metrics: config.MetricsConfig{
					Enabled: true,
					Path:    "/metrics",
				},
			},
		}

		server := setupTestServer(t, cfg)
		assert.NotNil(t, server.Metrics)
		assert.NotNil(t, server.Metrics.Registry())
	})

	t.Run("custom configuration", func(t *testing.T) {
		cfg := &config.Config{
			Monitoring: config.MonitoringConfig{
				Metrics: config.MetricsConfig{
					Enabled:    true,
					Path:      "/custom/metrics",
					Namespace: "custom",
				},
				Resource: config.ResourceConfig{
					Enabled:            true,
					Interval:           15 * time.Second,
					MemoryThreshold:    85,
					GoroutineThreshold: 10000,
				},
				RateLimit: config.RateLimitConfig{
					Enabled:           true,
					RequestsPerMinute: 1000,
					Window:           time.Minute,
				},
			},
		}

		server := setupTestServer(t, cfg)
		require.NotNil(t, server.Metrics)

		// Test custom metrics path
		w := server.SendRequest("GET", "/custom/metrics", nil)
		assert.Equal(t, 200, w.Code)

		// Test rate limiting
		for i := 0; i < 1100; i++ {
			w = server.SendRequest("GET", "/api/v1/words", nil)
			if i < 1000 {
				assert.Equal(t, 200, w.Code)
			} else {
				assert.Equal(t, 429, w.Code)
			}
		}
	})

	t.Run("disabled monitoring", func(t *testing.T) {
		cfg := &config.Config{
			Monitoring: config.MonitoringConfig{
				Metrics: config.MetricsConfig{
					Enabled: false,
				},
			},
		}

		server := setupTestServer(t, cfg)

		// Metrics endpoint should not be available
		w := server.SendRequest("GET", "/metrics", nil)
		assert.Equal(t, 404, w.Code)

		// Health endpoint should still work
		w = server.SendRequest("GET", "/health", nil)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("alert configuration", func(t *testing.T) {
		cfg := &config.Config{
			Monitoring: config.MonitoringConfig{
				Metrics: config.MetricsConfig{
					Enabled: true,
				},
				Alerts: config.AlertConfig{
					ErrorRateThreshold:    0.05,
					LatencyThreshold:      500 * time.Millisecond,
					MemoryThreshold:       85,
					GoroutineThreshold:    1000,
					ConnectionThreshold:   80,
					NotificationChannels: []string{"slack", "email"},
				},
			},
		}

		server := setupTestServer(t, cfg)
		require.NotNil(t, server.Metrics)

		// Generate errors to trigger alert
		for i := 0; i < 100; i++ {
			server.SendRequest("GET", "/api/v1/words/999999", nil)
		}

		// Check alert metrics
		metrics, err := server.Metrics.Registry().Gather()
		require.NoError(t, err)

		var alertFound bool
		for _, metric := range metrics {
			if metric.GetName() == "alert_error_rate_threshold_exceeded" {
				alertFound = true
				break
			}
		}
		assert.True(t, alertFound, "Alert metric not found")
	})
} 