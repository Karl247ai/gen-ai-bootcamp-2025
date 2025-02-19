package testutil

import (
	"context"
	"sync"
	"testing"
	"time"
)

type PerformanceTest struct {
	T               testing.TB
	ConcurrentUsers int
	Duration        time.Duration
	RampUpTime      time.Duration
	ThinkTime       time.Duration
}

func NewPerformanceTest(t testing.TB) *PerformanceTest {
	return &PerformanceTest{
		T:               t,
		ConcurrentUsers: 10,
		Duration:        5 * time.Second,
		RampUpTime:      1 * time.Second,
		ThinkTime:       100 * time.Millisecond,
	}
}

type TestScenario func(ctx context.Context) error

func (p *PerformanceTest) Run(name string, scenario TestScenario) *PerformanceResults {
	results := &PerformanceResults{
		Name:      name,
		StartTime: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.Duration)
	defer cancel()

	var wg sync.WaitGroup
	errors := make(chan error, p.ConcurrentUsers)
	latencies := make(chan time.Duration, 1000)

	// Start users gradually during ramp-up
	userDelay := p.RampUpTime / time.Duration(p.ConcurrentUsers)
	for i := 0; i < p.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			time.Sleep(userDelay * time.Duration(userID))

			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()
					if err := scenario(ctx); err != nil {
						errors <- err
						continue
					}
					latencies <- time.Since(start)
					time.Sleep(p.ThinkTime)
				}
			}
		}(i)
	}

	// Wait for completion
	wg.Wait()
	close(errors)
	close(latencies)

	// Collect results
	results.EndTime = time.Now()
	results.TotalRequests = len(latencies)
	results.ErrorCount = len(errors)

	var totalLatency time.Duration
	for lat := range latencies {
		totalLatency += lat
		if lat > results.MaxLatency {
			results.MaxLatency = lat
		}
	}
	if results.TotalRequests > 0 {
		results.AverageLatency = totalLatency / time.Duration(results.TotalRequests)
	}

	return results
}

type PerformanceResults struct {
	Name           string
	StartTime      time.Time
	EndTime        time.Time
	TotalRequests  int
	ErrorCount     int
	AverageLatency time.Duration
	MaxLatency     time.Duration
}

func (r *PerformanceResults) Assert(t testing.TB, maxAvgLatency, maxErrorRate float64) {
	duration := r.EndTime.Sub(r.StartTime)
	requestRate := float64(r.TotalRequests) / duration.Seconds()
	errorRate := float64(r.ErrorCount) / float64(r.TotalRequests)

	t.Logf("Performance Results for %s:", r.Name)
	t.Logf("  Total Requests: %d", r.TotalRequests)
	t.Logf("  Request Rate: %.2f/s", requestRate)
	t.Logf("  Average Latency: %v", r.AverageLatency)
	t.Logf("  Max Latency: %v", r.MaxLatency)
	t.Logf("  Error Rate: %.2f%%", errorRate*100)

	if r.AverageLatency > maxAvgLatency {
		t.Errorf("Average latency %v exceeds maximum %v", r.AverageLatency, maxAvgLatency)
	}

	if errorRate > maxErrorRate {
		t.Errorf("Error rate %.2f%% exceeds maximum %.2f%%", errorRate*100, maxErrorRate*100)
	}
} 