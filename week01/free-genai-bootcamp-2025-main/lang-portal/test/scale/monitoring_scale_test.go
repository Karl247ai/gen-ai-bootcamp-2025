package scale

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestMonitoringScalability(t *testing.T) {
	server := testutil.SetupTestServer(t)
	defer server.Cleanup()

	t.Run("metric cardinality scaling", func(t *testing.T) {
		scaleTest := testutil.NewScaleTest(t, server)
		
		// Generate high cardinality
		scaleTest.GenerateHighCardinality(1000000)

		// Verify system behavior
		assert.Less(t, scaleTest.GetMemoryUsage(), float64(2*1024*1024*1024)) // 2GB
		assert.True(t, scaleTest.VerifyQueryPerformance(100*time.Millisecond))
	})

	t.Run("concurrent scrapes", func(t *testing.T) {
		scaleTest := testutil.NewScaleTest(t, server)
		
		// Run concurrent scrapes
		results := scaleTest.RunConcurrentScrapes(100)

		// Verify performance
		assert.Less(t, results.MaxDuration, time.Second)
		assert.Equal(t, results.ErrorCount, 0)
	})

	t.Run("alert scaling", func(t *testing.T) {
		scaleTest := testutil.NewScaleTest(t, server)
		
		// Generate many alerts
		scaleTest.GenerateAlerts(1000)

		// Verify alert manager performance
		assert.True(t, scaleTest.VerifyAlertManagerPerformance())
		assert.True(t, scaleTest.VerifyNotificationDelivery())
	})
} 