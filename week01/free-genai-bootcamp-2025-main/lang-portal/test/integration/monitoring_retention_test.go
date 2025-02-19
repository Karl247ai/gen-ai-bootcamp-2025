package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestMonitoringRetention(t *testing.T) {
	server := testutil.SetupTestServer(t)
	defer server.Cleanup()

	t.Run("metric retention", func(t *testing.T) {
		ma := testutil.NewMetricsAssertion(t, server.Metrics)

		// Generate historical metrics
		testutil.GenerateHistoricalMetrics(server.Registry, 30*24*time.Hour)

		// Verify retention policies
		metrics := ma.GetAllMetrics()
		assert.Less(t, metrics.OldestTimestamp, time.Now().Add(-31*24*time.Hour))
		assert.Greater(t, metrics.OldestTimestamp, time.Now().Add(-32*24*time.Hour))
	})

	t.Run("alert history", func(t *testing.T) {
		av := testutil.NewAlertVerifier(t, server.Registry)

		// Generate historical alerts
		testutil.GenerateHistoricalAlerts(server.Registry, 90*24*time.Hour)

		// Verify alert retention
		alerts := av.GetAlertHistory()
		assert.Less(t, alerts.OldestAlert, time.Now().Add(-91*24*time.Hour))
		assert.Greater(t, alerts.OldestAlert, time.Now().Add(-92*24*time.Hour))
	})

	t.Run("log rotation", func(t *testing.T) {
		// Generate historical logs
		testutil.GenerateHistoricalLogs(server.LogDir, 7*24*time.Hour)

		// Verify log rotation
		logs := testutil.GetLogFiles(server.LogDir)
		assert.Less(t, logs.OldestLog, time.Now().Add(-8*24*time.Hour))
		assert.Greater(t, logs.OldestLog, time.Now().Add(-9*24*time.Hour))
	})
} 