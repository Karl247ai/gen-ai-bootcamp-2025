package integration

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/your-org/lang-portal/internal/config"
	"github.com/your-org/lang-portal/internal/models"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestMonitoringEdgeCases(t *testing.T) {
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			Metrics: config.MetricsConfig{
				Enabled: true,
			},
			RateLimit: config.RateLimitConfig{
				Enabled:           true,
				RequestsPerMinute: 10,
				Window:           time.Second,
			},
		},
	}

	server := setupTestServer(t, cfg)
	ma := testutil.NewMetricsAssertion(t, server.Metrics, server.Metrics.Registry())

	t.Run("concurrent requests", func(t *testing.T) {
		var wg sync.WaitGroup
		numRequests := 20

		// Launch concurrent requests
		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				word := models.Word{
					Japanese: "テスト",
					Romaji:   "tesuto",
					English:  "test",
				}
				server.SendRequest("POST", "/api/v1/words", word)
			}(i)
		}

		wg.Wait()

		// Verify metrics
		ma.AssertCounterValue("handler_word_create_success", float64(10))
		ma.AssertCounterValue("ratelimit_exceeded", float64(10))
	})

	t.Run("slow database", func(t *testing.T) {
		// Inject delay into database operations
		db := server.DB.(*testutil.MockDB)
		db.SetDelay(500 * time.Millisecond)
		defer db.SetDelay(0)

		start := time.Now()
		w := server.SendRequest("GET", "/api/v1/words", nil)
		duration := time.Since(start)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Greater(t, duration, 500*time.Millisecond)

		ma.AssertHistogramBounds("handler_request_duration", 1.0)
		ma.AssertCounterValue("db_slow_query", 1)
	})

	t.Run("cache failure", func(t *testing.T) {
		// Simulate cache failure
		cache := server.Cache.(*testutil.MockCache)
		cache.SetError(context.DeadlineExceeded)
		defer cache.SetError(nil)

		w := server.SendRequest("GET", "/api/v1/words/1", nil)
		assert.Equal(t, http.StatusOK, w.Code) // Should still work, just slower

		ma.AssertCounterValue("cache_get_error", 1)
		ma.AssertCounterValue("cache_fallback_success", 1)
	})

	t.Run("memory pressure", func(t *testing.T) {
		// Allocate memory to trigger resource monitoring
		data := make([][]byte, 100)
		for i := range data {
			data[i] = make([]byte, 1024*1024) // 1MB each
		}

		time.Sleep(20 * time.Second) // Wait for resource metrics to update

		metrics, err := server.Metrics.Registry().Gather()
		assert.NoError(t, err)

		var memoryMetricFound bool
		for _, metric := range metrics {
			if metric.GetName() == "runtime_memory_alloc" {
				memoryMetricFound = true
				value := metric.GetMetric()[0].GetGauge().GetValue()
				assert.Greater(t, value, float64(50*1024*1024)) // At least 50MB
			}
		}
		assert.True(t, memoryMetricFound)
	})

	t.Run("metric cardinality", func(t *testing.T) {
		// Generate requests with many different paths
		for i := 0; i < 1000; i++ {
			path := fmt.Sprintf("/api/v1/custom/%d", i)
			server.SendRequest("GET", path, nil)
		}

		// Verify metric cardinality hasn't exploded
		metrics, err := server.Metrics.Registry().Gather()
		assert.NoError(t, err)

		pathLabels := make(map[string]struct{})
		for _, metric := range metrics {
			if metric.GetName() == "handler_request_count" {
				for _, m := range metric.GetMetric() {
					for _, label := range m.GetLabel() {
						if label.GetName() == "path" {
							pathLabels[label.GetValue()] = struct{}{}
						}
					}
				}
			}
		}

		assert.Less(t, len(pathLabels), 100, "Too many unique path labels")
	})

	t.Run("metric cardinality limits", func(t *testing.T) {
		ma := testutil.NewMetricsAssertion(t, server.Metrics)
		
		// Generate high cardinality metrics
		for i := 0; i < 1000; i++ {
			server.Router.Use(func(c *gin.Context) {
				c.Set("user_id", fmt.Sprintf("user_%d", i))
				c.Next()
			})
			testutil.GenerateTestTraffic(server.Router, 1)
		}

		// Verify cardinality limits
		metrics, err := server.Registry.Gather()
		assert.NoError(t, err)
		for _, mf := range metrics {
			assert.Less(t, len(mf.GetMetric()), 1000, 
				"Metric cardinality too high: %s", mf.GetName())
		}
	})

	t.Run("concurrent maintenance windows", func(t *testing.T) {
		mwt := testutil.NewMaintenanceWindowTester(t, server.Registry)

		// Start multiple maintenance windows
		windows := []string{"backup", "upgrade", "cleanup"}
		for _, w := range windows {
			mwt.StartWindow(testutil.MaintenanceWindow{
				Name: w,
				Labels: map[string]string{"type": w},
			})
		}

		// Verify concurrent windows
		for _, w := range windows {
			mwt.VerifyStatus(w, "active")
		}

		// End windows
		for _, w := range windows {
			mwt.EndWindow(w)
			mwt.VerifyStatus(w, "completed")
		}
	})

	t.Run("alert storm prevention", func(t *testing.T) {
		av := testutil.NewAlertVerifier(t, server.Registry)

		// Generate conditions for multiple alerts
		server.WithErrorRate(1.0)
		server.WithSlowResponses(500 * time.Millisecond)
		testutil.GenerateTestTraffic(server.Router, 100)

		// Verify alert grouping
		groups := av.GetAlertGroups()
		assert.Equal(t, 1, len(groups), "Alerts should be grouped")
	})
} 