package integration

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/your-org/lang-portal/internal/config"
	"github.com/your-org/lang-portal/internal/metrics"
)

func TestAlertRules(t *testing.T) {
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			Metrics: config.MetricsConfig{
				Enabled: true,
			},
			Alerts: config.AlertConfig{
				ErrorRateThreshold:   0.05,
				LatencyThreshold:    500 * time.Millisecond,
				MemoryThreshold:     85,
				GoroutineThreshold:  1000,
				ConnectionThreshold: 80,
			},
		},
	}

	server := setupTestServer(t, cfg)
	ma := testutil.NewMetricsAssertion(t, server.Metrics, server.Metrics.Registry())

	t.Run("error rate alert", func(t *testing.T) {
		// Generate errors
		for i := 0; i < 100; i++ {
			server.SendRequest("GET", "/api/v1/words/999999", nil)
		}

		// Make some successful requests
		for i := 0; i < 1000; i++ {
			server.SendRequest("GET", "/api/v1/words", nil)
		}

		// Verify error rate alert
		ma.AssertAlertFiring("HighErrorRate", func(alert *testutil.Alert) bool {
			return alert.Labels["severity"] == "warning" &&
				alert.Value > 0.05
		})
	})

	t.Run("latency alert", func(t *testing.T) {
		// Inject delay into database operations
		db := server.DB.(*testutil.MockDB)
		db.SetDelay(600 * time.Millisecond)
		defer db.SetDelay(0)

		// Make requests
		for i := 0; i < 10; i++ {
			server.SendRequest("GET", "/api/v1/words", nil)
		}

		// Verify latency alert
		ma.AssertAlertFiring("SlowResponses", func(alert *testutil.Alert) bool {
			return alert.Labels["severity"] == "warning" &&
				alert.Value > 0.5
		})
	})

	t.Run("memory alert", func(t *testing.T) {
		// Allocate memory
		data := make([][]byte, 100)
		for i := range data {
			data[i] = make([]byte, 1024*1024) // 1MB each
		}

		// Verify memory alert
		ma.AssertAlertFiring("HighMemoryUsage", func(alert *testutil.Alert) bool {
			return alert.Labels["severity"] == "warning" &&
				alert.Value > 85
		})
	})

	t.Run("goroutine alert", func(t *testing.T) {
		// Create many goroutines
		for i := 0; i < 1500; i++ {
			go func() {
				time.Sleep(100 * time.Millisecond)
			}()
		}

		// Verify goroutine alert
		ma.AssertAlertFiring("HighGoroutineCount", func(alert *testutil.Alert) bool {
			return alert.Labels["severity"] == "warning" &&
				alert.Value > 1000
		})
	})

	t.Run("connection pool alert", func(t *testing.T) {
		// Create many database connections
		for i := 0; i < 90; i++ {
			go func() {
				server.SendRequest("GET", "/api/v1/words", nil)
			}()
		}

		// Verify connection pool alert
		ma.AssertAlertFiring("HighConnectionUsage", func(alert *testutil.Alert) bool {
			return alert.Labels["severity"] == "warning" &&
				alert.Value > 80
		})
	})

	t.Run("alert notification", func(t *testing.T) {
		notifier := testutil.NewMockNotifier()
		server.SetNotifier(notifier)

		// Trigger alert
		for i := 0; i < 100; i++ {
			server.SendRequest("GET", "/api/v1/words/999999", nil)
		}

		// Verify notification
		notifications := notifier.GetNotifications()
		assert.Greater(t, len(notifications), 0)
		assert.Contains(t, notifications[0].Message, "High error rate detected")
	})
}

func TestAlertThrottling(t *testing.T) {
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			Alerts: config.AlertConfig{
				ErrorRateThreshold: 0.05,
				Levels: map[string]config.AlertLevel{
					"critical": {MinInterval: time.Minute},
					"warning":  {MinInterval: 5 * time.Minute},
				},
			},
		},
	}

	server := setupTestServer(t, cfg)
	notifier := testutil.NewMockNotifier()
	server.SetNotifier(notifier)

	// Generate errors multiple times
	for i := 0; i < 3; i++ {
		for j := 0; j < 100; j++ {
			server.SendRequest("GET", "/api/v1/words/999999", nil)
		}
		time.Sleep(30 * time.Second)
	}

	// Verify notification throttling
	notifications := notifier.GetNotifications()
	assert.Equal(t, 1, len(notifications), "Expected only one notification due to throttling")
} 