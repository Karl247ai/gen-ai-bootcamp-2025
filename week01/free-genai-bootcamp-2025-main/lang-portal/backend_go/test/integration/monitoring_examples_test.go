package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/your-org/lang-portal/internal/testutil"
)

func ExamplePerformanceTest(t *testing.T) {
	server := testutil.SetupTestServer(t)
	defer server.Cleanup()

	// Create performance test
	pt := testutil.NewPerformanceTest(t, server.Router).
		WithConcurrency(5).
		WithDuration(2 * time.Second)

	// Run test
	results := pt.Run("example_test", func(ctx context.Context) error {
		resp := testutil.PerformRequest(server.Router, "GET", "/api/v1/words", nil)
		if resp.StatusCode != 200 {
			return fmt.Errorf("unexpected status: %d", resp.StatusCode)
		}
		return nil
	})

	// Assert results
	results.Assert(t, 100*time.Millisecond, 0.95)
}

func ExampleAlertVerification(t *testing.T) {
	server := testutil.SetupTestServer(t)
	defer server.Cleanup()

	// Create alert verifier
	av := testutil.NewAlertVerifier(t, server.Registry)

	// Add latency to trigger alert
	server.WithSlowResponses(200 * time.Millisecond)

	// Generate traffic
	testutil.GenerateTestTraffic(t, server.Router)

	// Verify alert fires
	alertFired := av.VerifyAlert("HighLatency", 5*time.Second)
	assert.True(t, alertFired, "Expected HighLatency alert to fire")

	// Verify no other alerts
	av.AssertNoAlerts(t)
}

func ExampleMetricsAssertion(t *testing.T) {
	server := testutil.SetupTestServer(t)
	defer server.Cleanup()

	// Create metrics assertion
	ma := testutil.NewMetricsAssertion(t, server.Metrics)

	// Generate traffic
	testutil.PerformRequest(server.Router, "GET", "/api/v1/words", nil)

	// Verify metrics
	count := ma.GetCounterValue("handler_request_count", map[string]string{
		"endpoint": "/api/v1/words",
		"method":   "GET",
	})
	assert.Equal(t, float64(1), count)

	// Verify histogram
	histogram := ma.GetHistogramValue("handler_request_duration")
	assert.Greater(t, histogram.GetSampleCount(), uint64(0))
} 