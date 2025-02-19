package integration

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/your-org/lang-portal/internal/database"
	"github.com/your-org/lang-portal/internal/metrics"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestDatabaseMonitoring(t *testing.T) {
	// Setup
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../migrations/0001_init.sql")

	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)
	monitoredDB := database.NewMonitoredDB(db, m)

	t.Run("query metrics", func(t *testing.T) {
		ctx := context.Background()
		rows, err := monitoredDB.QueryContext(ctx, "SELECT * FROM words LIMIT 5")
		assert.NoError(t, err)
		rows.Close()

		metrics, err := reg.Gather()
		assert.NoError(t, err)

		found := false
		for _, metric := range metrics {
			if metric.GetName() == "db_query_success" {
				found = true
				assert.Equal(t, float64(1), metric.GetMetric()[0].GetCounter().GetValue())
			}
		}
		assert.True(t, found, "Expected metric not found")
	})

	t.Run("transaction metrics", func(t *testing.T) {
		tx, err := monitoredDB.Begin()
		assert.NoError(t, err)

		_, err = tx.Exec("INSERT INTO words (japanese, romaji, english) VALUES (?, ?, ?)",
			"テスト", "tesuto", "test")
		assert.NoError(t, err)

		err = tx.Commit()
		assert.NoError(t, err)

		metrics, err := reg.Gather()
		assert.NoError(t, err)

		found := false
		for _, metric := range metrics {
			if metric.GetName() == "db_transaction_begin" {
				found = true
				assert.Equal(t, float64(1), metric.GetMetric()[0].GetCounter().GetValue())
			}
		}
		assert.True(t, found, "Expected metric not found")
	})
} 