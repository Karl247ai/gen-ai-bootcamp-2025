package testutil

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// PerformanceTest provides utilities for performance testing
type PerformanceTest struct {
	t               *testing.T
	router         *gin.Engine
	concurrency    int
	duration       time.Duration
	requestTimeout time.Duration
}

// NewPerformanceTest creates a new performance test
func NewPerformanceTest(t *testing.T, router *gin.Engine) *PerformanceTest {
	return &PerformanceTest{
		t:               t,
		router:         router,
		concurrency:    10,
		duration:       5 * time.Second,
		requestTimeout: time.Second,
	}
}

// WithConcurrency sets the concurrency level
func (pt *PerformanceTest) WithConcurrency(n int) *PerformanceTest {
	pt.concurrency = n
	return pt
}

// WithDuration sets the test duration
func (pt *PerformanceTest) WithDuration(d time.Duration) *PerformanceTest {
	pt.duration = d
	return pt
}

// Run executes the performance test
func (pt *PerformanceTest) Run(name string, fn func(context.Context) error) *PerformanceResults {
	results := &PerformanceResults{
		Name:       name,
		StartTime:  time.Now(),
		Durations:  make([]time.Duration, 0),
		ErrorCount: 0,
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), pt.duration)
	defer cancel()

	// Start worker goroutines
	for i := 0; i < pt.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()
					err := fn(ctx)
					duration := time.Since(start)
					
					results.mu.Lock()
					results.Durations = append(results.Durations, duration)
					if err != nil {
						results.ErrorCount++
					}
					results.mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()
	results.EndTime = time.Now()
	return results
}

// PerformanceResults holds test results
type PerformanceResults struct {
	Name       string
	StartTime  time.Time
	EndTime    time.Time
	Durations  []time.Duration
	ErrorCount int
	mu         sync.Mutex
}

// Assert verifies performance requirements
func (r *PerformanceResults) Assert(t *testing.T, maxLatency time.Duration, successRate float64) {
	// Calculate statistics
	var totalDuration time.Duration
	for _, d := range r.Durations {
		totalDuration += d
	}

	avgLatency := totalDuration / time.Duration(len(r.Durations))
	actualSuccessRate := 1 - (float64(r.ErrorCount) / float64(len(r.Durations)))

	// Assert requirements
	assert.Less(t, avgLatency, maxLatency, "Average latency exceeds maximum")
	assert.GreaterOrEqual(t, actualSuccessRate, successRate, "Success rate below requirement")
} 