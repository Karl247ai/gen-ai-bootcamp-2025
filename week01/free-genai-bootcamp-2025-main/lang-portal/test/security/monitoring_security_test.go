package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestSecurityMonitoring(t *testing.T) {
	server := testutil.SetupTestServer(t)
	defer server.Cleanup()

	t.Run("authentication metrics", func(t *testing.T) {
		ma := testutil.NewMetricsAssertion(t, server.Metrics)

		// Generate auth events
		testutil.GenerateAuthEvents(server.Router, 10)

		// Verify auth metrics
		assert.True(t, ma.MetricExists("auth_attempts_total"))
		assert.True(t, ma.MetricExists("auth_failures_total"))
		assert.True(t, ma.MetricExists("auth_success_total"))
	})

	t.Run("rate limiting", func(t *testing.T) {
		ma := testutil.NewMetricsAssertion(t, server.Metrics)

		// Generate rapid requests
		for i := 0; i < 100; i++ {
			testutil.GenerateTestTraffic(server.Router, 1)
		}

		// Verify rate limiting metrics
		blocked := ma.GetCounterValue("rate_limit_blocked_total", nil)
		assert.Greater(t, blocked, float64(0))
	})

	t.Run("suspicious activity", func(t *testing.T) {
		av := testutil.NewAlertVerifier(t, server.Registry)

		// Generate suspicious patterns
		testutil.GenerateSuspiciousTraffic(server.Router, 50)

		// Verify security alerts
		assert.True(t, av.VerifyAlert("SuspiciousActivity", 5*time.Second))
	})
} 