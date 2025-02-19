package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/your-org/lang-portal/internal/config"
	"github.com/your-org/lang-portal/internal/models"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestPerformance(t *testing.T) {
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			Metrics: config.MetricsConfig{
				Enabled: true,
			},
		},
	}

	server := setupTestServer(t, cfg)
	ma := testutil.NewMetricsAssertion(t, server.Metrics, server.Metrics.Registry())

	t.Run("write performance", func(t *testing.T) {
		perf := testutil.NewPerformanceTest(t)
		perf.ConcurrentUsers = 10
		perf.Duration = 5 * time.Second

		results := perf.Run("write_test", func(ctx context.Context) error {
			word := models.Word{
				Japanese: "テスト",
				Romaji:   "tesuto",
				English:  "test",
			}
			w := server.SendRequest("POST", "/api/v1/words", word)
			if w.Code != 201 {
				return fmt.Errorf("unexpected status code: %d", w.Code)
			}
			return nil
		})

		// Assert performance metrics
		results.Assert(t, 50*time.Millisecond, 0.95)
		ma.AssertHistogramQuantile("handler_request_duration", 0.95, 0.05)
	})

	t.Run("read performance", func(t *testing.T) {
		perf := testutil.NewPerformanceTest(t)
		perf.ConcurrentUsers = 50
		perf.Duration = 10 * time.Second

		results := perf.Run("read_test", func(ctx context.Context) error {
			w := server.SendRequest("GET", "/api/v1/words", nil)
			if w.Code != 200 {
				return fmt.Errorf("unexpected status code: %d", w.Code)
			}
			return nil
		})

		// Assert performance metrics
		results.Assert(t, 20*time.Millisecond, 0.95)
		ma.AssertHistogramQuantile("handler_request_duration", 0.95, 0.02)
	})

	t.Run("cache performance", func(t *testing.T) {
		// Create test data
		word := models.Word{
			Japanese: "キャッシュ",
			Romaji:   "kyasshu",
			English:  "cache",
		}
		w := server.SendRequest("POST", "/api/v1/words", word)
		assert.Equal(t, 201, w.Code)

		var response struct {
			ID int `json:"id"`
		}
		json.NewDecoder(w.Body).Decode(&response)

		perf := testutil.NewPerformanceTest(t)
		perf.ConcurrentUsers = 100
		perf.Duration = 5 * time.Second

		results := perf.Run("cache_test", func(ctx context.Context) error {
			w := server.SendRequest("GET", fmt.Sprintf("/api/v1/words/%d", response.ID), nil)
			if w.Code != 200 {
				return fmt.Errorf("unexpected status code: %d", w.Code)
			}
			return nil
		})

		// Assert cache metrics
		results.Assert(t, 5*time.Millisecond, 0.95)
		ma.AssertCounterRatio("cache_get_hit", "cache_get_miss", 0.9)
	})

	t.Run("database connection pool", func(t *testing.T) {
		perf := testutil.NewPerformanceTest(t)
		perf.ConcurrentUsers = 200
		perf.Duration = 10 * time.Second

		results := perf.Run("db_pool_test", func(ctx context.Context) error {
			w := server.SendRequest("GET", "/api/v1/words", nil)
			if w.Code != 200 {
				return fmt.Errorf("unexpected status code: %d", w.Code)
			}
			return nil
		})

		// Get database stats
		w := server.SendRequest("GET", "/debug/db/stats", nil)
		assert.Equal(t, 200, w.Code)

		var stats struct {
			MaxOpenConnections int `json:"max_open_connections"`
			OpenConnections   int `json:"open_connections"`
			InUse            int `json:"in_use"`
			Idle             int `json:"idle"`
		}
		json.NewDecoder(w.Body).Decode(&stats)

		// Assert connection pool metrics
		assert.Less(t, stats.InUse, stats.MaxOpenConnections)
		assert.Greater(t, stats.Idle, 0)
		results.Assert(t, 100*time.Millisecond, 0.95)
	})

	t.Run("maintenance metrics", func(t *testing.T) {
		// Setup test server
		server := setupTestServer(t)
		defer server.Close()
		
		// Generate maintenance activity
		start := time.Now()
		recordMaintenance("upgrade", 3600, nil)  // 1 hour maintenance
		recordMaintenance("backup", 300, nil)    // 5 minute backup
		recordMaintenance("cleanup", 600, errors.New("test error"))
		
		// Query metrics
		metrics, err := server.Metrics.Registry().Gather()
		require.NoError(t, err)
		
		// Verify maintenance counters
		var total, errors float64
		for _, m := range metrics {
			switch m.GetName() {
			case "maintenance_total":
				total = m.GetMetric()[0].GetCounter().GetValue()
			case "maintenance_errors_total":
				errors = m.GetMetric()[0].GetCounter().GetValue()
			}
		}
		
		assert.Equal(t, float64(3), total, "Expected 3 maintenance activities")
		assert.Equal(t, float64(1), errors, "Expected 1 maintenance error")
	})
}

func BenchmarkAPIEndpoints(b *testing.B) {
	// Setup test environment
	db := testutil.NewTestDB(b)
	testutil.ExecuteSQL(b, db, "../../migrations/0001_init.sql")

	metrics := testutil.NewTestMetrics(b)
	router := setupTestRouter(b, db, metrics)

	// Create test data
	createTestData(b, router)

	benchmarks := []struct {
		name    string
		method  string
		path    string
		body    interface{}
		setup   func() // Optional setup for each iteration
		cleanup func() // Optional cleanup for each iteration
	}{
		{
			name:   "CreateWord",
			method: "POST",
			path:   "/api/v1/words",
			body: models.Word{
				Japanese: "テスト",
				Romaji:   "tesuto",
				English:  "test",
			},
		},
		{
			name:   "ListWords",
			method: "GET",
			path:   "/api/v1/words?limit=50",
		},
		{
			name:   "GetWord",
			method: "GET",
			path:   "/api/v1/words/1",
		},
		{
			name:   "AddWordsToGroup",
			method: "POST",
			path:   "/api/v1/groups/1/words",
			body: struct {
				WordIDs []int `json:"word_ids"`
			}{
				WordIDs: []int{1, 2, 3, 4, 5},
			},
		},
		{
			name:   "SearchWords",
			method: "GET",
			path:   "/api/v1/words/search?q=test",
			setup: func() {
				// Create words for search
				for i := 0; i < 10; i++ {
					word := models.Word{
						Japanese: fmt.Sprintf("テスト%d", i),
						Romaji:   fmt.Sprintf("tesuto%d", i),
						English:  fmt.Sprintf("test%d", i),
					}
					req := testutil.NewRequest(b, "POST", "/api/v1/words", word)
					testutil.PerformRequest(router, req)
				}
			},
		},
		{
			name:   "BulkWordAddition",
			method: "POST",
			path:   "/api/v1/groups/1/words",
			setup: func() {
				// Create multiple words
				wordIDs := make([]int, 100)
				for i := range wordIDs {
					word := models.Word{
						Japanese: fmt.Sprintf("単語%d", i),
						Romaji:   fmt.Sprintf("tango%d", i),
						English:  fmt.Sprintf("word%d", i),
					}
					req := testutil.NewRequest(b, "POST", "/api/v1/words", word)
					w := testutil.PerformRequest(router, req)
					var resp struct {
						Data struct {
							ID int `json:"id"`
						} `json:"data"`
					}
					json.NewDecoder(w.Body).Decode(&resp)
					wordIDs[i] = resp.Data.ID
				}
			},
			body: struct {
				WordIDs []int `json:"word_ids"`
			}{
				WordIDs: make([]int, 100), // Will be filled in setup
			},
		},
		{
			name:   "ComplexQuery",
			method: "GET",
			path:   "/api/v1/words?sort=created_at&order=desc&limit=50&include=groups",
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if bm.setup != nil {
					b.StopTimer()
					bm.setup()
					b.StartTimer()
				}

				req := testutil.NewRequest(b, bm.method, bm.path, bm.body)
				w := testutil.PerformRequest(router, req)

				if w.Code != http.StatusOK && w.Code != http.StatusCreated {
					b.Fatalf("Request failed with status %d", w.Code)
				}

				if bm.cleanup != nil {
					b.StopTimer()
					bm.cleanup()
					b.StartTimer()
				}
			}
		})
	}
}

func BenchmarkConcurrentRequests(b *testing.B) {
	// Setup test environment
	db := testutil.NewTestDB(b)
	testutil.ExecuteSQL(b, db, "../../migrations/0001_init.sql")

	metrics := testutil.NewTestMetrics(b)
	router := setupTestRouter(b, db, metrics)

	// Create test data
	createTestData(b, router)

	concurrencyLevels := []int{1, 10, 50, 100}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency-%d", concurrency), func(b *testing.B) {
			var wg sync.WaitGroup
			requestCh := make(chan bool, concurrency)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					requestCh <- true
					defer func() { <-requestCh }()

					req := testutil.NewRequest(b, "GET", "/api/v1/words", nil)
					w := testutil.PerformRequest(router, req)

					if w.Code != http.StatusOK {
						b.Errorf("Request failed with status %d", w.Code)
					}
				}()
			}
			wg.Wait()
		})
	}
}

func createTestData(t testing.TB, router http.Handler) {
	// Create test words
	words := []models.Word{
		{Japanese: "一", Romaji: "ichi", English: "one"},
		{Japanese: "二", Romaji: "ni", English: "two"},
		{Japanese: "三", Romaji: "san", English: "three"},
		{Japanese: "四", Romaji: "yon", English: "four"},
		{Japanese: "五", Romaji: "go", English: "five"},
	}

	for _, word := range words {
		req := testutil.NewRequest(t, "POST", "/api/v1/words", word)
		w := testutil.PerformRequest(router, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("Failed to create test word: %d", w.Code)
		}
	}

	// Create test group
	group := models.Group{Name: "Test Group"}
	req := testutil.NewRequest(t, "POST", "/api/v1/groups", group)
	w := testutil.PerformRequest(router, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create test group: %d", w.Code)
	}
} 