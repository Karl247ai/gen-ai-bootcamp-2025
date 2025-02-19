package testutil

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestScenario represents a test scenario
type TestScenario struct {
	Name        string
	Description string
	Setup       func(*TestServer)
	Verify      func(*testing.T, *TestServer)
	Cleanup     func(*TestServer)
	Timeout     time.Duration
}

// RunTestScenario executes a test scenario
func RunTestScenario(t *testing.T, scenario TestScenario) {
	server := SetupTestServer(t)
	defer func() {
		if scenario.Cleanup != nil {
			scenario.Cleanup(server)
		}
		server.Cleanup()
	}()

	// Set default timeout
	if scenario.Timeout == 0 {
		scenario.Timeout = 10 * time.Second
	}

	// Run setup
	if scenario.Setup != nil {
		scenario.Setup(server)
	}

	// Run verification with timeout
	done := make(chan bool)
	go func() {
		if scenario.Verify != nil {
			scenario.Verify(t, server)
		}
		done <- true
	}()

	select {
	case <-done:
		// Test completed successfully
	case <-time.After(scenario.Timeout):
		t.Fatalf("Test scenario %s timed out after %v", scenario.Name, scenario.Timeout)
	}
}

// CommonScenarios provides predefined test scenarios
var CommonScenarios = map[string]TestScenario{
	"high_latency": {
		Name:        "High Latency Detection",
		Description: "Verifies that high latency triggers appropriate alerts",
		Setup: func(s *TestServer) {
			s.WithSlowResponses(200 * time.Millisecond)
			GenerateTestTraffic(s.Router, 50)
		},
		Verify: func(t *testing.T, s *TestServer) {
			av := NewAlertVerifier(t, s.Registry)
			assert.True(t, av.VerifyAlert("HighLatency", 5*time.Second))
		},
	},
	"error_spike": {
		Name:        "Error Rate Spike",
		Description: "Verifies that increased error rates trigger alerts",
		Setup: func(s *TestServer) {
			s.WithErrorRate(0.2) // 20% error rate
			GenerateTestTraffic(s.Router, 100)
		},
		Verify: func(t *testing.T, s *TestServer) {
			av := NewAlertVerifier(t, s.Registry)
			assert.True(t, av.VerifyAlert("HighErrorRate", 5*time.Second))
		},
	},
	"cache_performance": {
		Name:        "Cache Performance",
		Description: "Verifies cache hit rate monitoring",
		Setup: func(s *TestServer) {
			// Generate repeated requests to build cache
			for i := 0; i < 10; i++ {
				GenerateTestTraffic(s.Router, 10)
				time.Sleep(100 * time.Millisecond)
			}
		},
		Verify: func(t *testing.T, s *TestServer) {
			ma := NewMetricsAssertion(t, s.Metrics)
			hitRate := ma.GetCounterValue("cache_hit_rate", nil)
			assert.Greater(t, hitRate, 0.7, "Cache hit rate should be above 70%")
		},
	},
	"resource_exhaustion": {
		Name:        "Resource Exhaustion",
		Description: "Verifies resource usage alerts",
		Setup: func(s *TestServer) {
			// Generate high load
			for i := 0; i < 5; i++ {
				go GenerateTestTraffic(s.Router, 1000)
			}
		},
		Verify: func(t *testing.T, s *TestServer) {
			av := NewAlertVerifier(t, s.Registry)
			assert.True(t, av.VerifyAlert("HighCPUUsage", 10*time.Second))
		},
		Cleanup: func(s *TestServer) {
			time.Sleep(1 * time.Second) // Allow resources to recover
		},
	},
	"maintenance_impact": {
		Name:        "Maintenance Impact",
		Description: "Verifies impact tracking during maintenance",
		Setup: func(s *TestServer) {
			// Start maintenance
			s.Metrics.RecordMaintenance("test_maintenance", func() error {
				s.WithSlowResponses(50 * time.Millisecond)
				GenerateTestTraffic(s.Router, 100)
				return nil
			})
		},
		Verify: func(t *testing.T, s *TestServer) {
			ma := NewMetricsAssertion(t, s.Metrics)
			maintenanceCount := ma.GetCounterValue("maintenance_total", map[string]string{
				"type": "test_maintenance",
			})
			assert.Equal(t, float64(1), maintenanceCount)
		},
	},
	"concurrent_maintenance": {
		Name:        "Concurrent Maintenance",
		Description: "Verifies handling of concurrent maintenance operations",
		Setup: func(s *TestServer) {
			// Start multiple maintenance operations
			go s.Metrics.RecordMaintenance("backup", func() error {
				time.Sleep(2 * time.Second)
				return nil
			})
			go s.Metrics.RecordMaintenance("upgrade", func() error {
				time.Sleep(1 * time.Second)
				return nil
			})
			GenerateTestTraffic(s.Router, 50)
		},
		Verify: func(t *testing.T, s *TestServer) {
			ma := NewMetricsAssertion(t, s.Metrics)
			time.Sleep(3 * time.Second) // Wait for maintenance to complete
			
			total := ma.GetCounterValue("maintenance_total", nil)
			assert.Equal(t, float64(2), total, "Expected two maintenance operations")
		},
	},
	"maintenance_error_handling": {
		Name:        "Maintenance Error Handling",
		Description: "Verifies error handling during maintenance",
		Setup: func(s *TestServer) {
			s.Metrics.RecordMaintenance("failed_operation", func() error {
				return fmt.Errorf("simulated failure")
			})
		},
		Verify: func(t *testing.T, s *TestServer) {
			ma := NewMetricsAssertion(t, s.Metrics)
			errors := ma.GetCounterValue("maintenance_errors_total", map[string]string{
				"type": "failed_operation",
			})
			assert.Equal(t, float64(1), errors)
		},
	},
	"performance_degradation": {
		Name:        "Performance Degradation",
		Description: "Verifies detection of gradual performance degradation",
		Setup: func(s *TestServer) {
			// Gradually increase latency
			go func() {
				for i := 0; i < 5; i++ {
					s.WithSlowResponses(time.Duration(50*(i+1)) * time.Millisecond)
					GenerateTestTraffic(s.Router, 20)
					time.Sleep(500 * time.Millisecond)
				}
			}()
		},
		Verify: func(t *testing.T, s *TestServer) {
			av := NewAlertVerifier(t, s.Registry)
			assert.True(t, av.VerifyAlert("PerformanceDegradation", 10*time.Second))
		},
	},
	"maintenance_window": {
		Name:        "Maintenance Window",
		Description: "Verifies maintenance window tracking",
		Setup: func(s *TestServer) {
			s.Metrics.StartMaintenanceWindow("scheduled_maintenance", map[string]string{
				"type": "upgrade",
				"team": "platform",
			})
			GenerateTestTraffic(s.Router, 100)
		},
		Verify: func(t *testing.T, s *TestServer) {
			ma := NewMetricsAssertion(t, s.Metrics)
			
			// Verify maintenance window metrics
			active := ma.GetGaugeValue("maintenance_window_active", map[string]string{
				"type": "upgrade",
			})
			assert.Equal(t, float64(1), active)
			
			// End maintenance window
			s.Metrics.EndMaintenanceWindow("scheduled_maintenance")
			
			// Verify window ended
			active = ma.GetGaugeValue("maintenance_window_active", map[string]string{
				"type": "upgrade",
			})
			assert.Equal(t, float64(0), active)
		},
	},
	"alert_correlation": {
		Name:        "Alert Correlation",
		Description: "Verifies correlation between different types of alerts",
		Setup: func(s *TestServer) {
			// Generate conditions that should trigger multiple related alerts
			s.WithErrorRate(0.1)
			s.WithSlowResponses(150 * time.Millisecond)
			for i := 0; i < 3; i++ {
				go GenerateTestTraffic(s.Router, 100)
			}
		},
		Verify: func(t *testing.T, s *TestServer) {
			av := NewAlertVerifier(t, s.Registry)
			
			// Verify multiple alerts fire in sequence
			assert.True(t, av.VerifyAlert("HighErrorRate", 5*time.Second))
			assert.True(t, av.VerifyAlert("HighLatency", 5*time.Second))
			assert.True(t, av.VerifyAlert("HighLoad", 5*time.Second))
		},
		Cleanup: func(s *TestServer) {
			time.Sleep(2 * time.Second) // Allow system to stabilize
		},
	},
	"metric_cardinality": {
		Name:        "Metric Cardinality",
		Description: "Verifies metric cardinality remains within limits",
		Setup: func(s *TestServer) {
			// Generate requests with many different labels
			for i := 0; i < 100; i++ {
				s.Router.Use(func(c *gin.Context) {
					c.Set("custom_label", fmt.Sprintf("value_%d", i))
					c.Next()
				})
				GenerateTestTraffic(s.Router, 1)
			}
		},
		Verify: func(t *testing.T, s *TestServer) {
			metrics, err := s.Registry.Gather()
			assert.NoError(t, err)
			
			// Check cardinality for each metric
			for _, mf := range metrics {
				assert.Less(t, len(mf.GetMetric()), 1000,
					"Metric %s has too high cardinality", mf.GetName())
			}
		},
	},
	"alert_silencing": {
		Name:        "Alert Silencing",
		Description: "Verifies alert silencing during maintenance",
		Setup: func(s *TestServer) {
			// Start maintenance window
			s.Metrics.StartMaintenanceWindow("upgrade", map[string]string{
				"type": "planned",
			})
			
			// Generate conditions that would normally trigger alerts
			s.WithErrorRate(0.5)
			GenerateTestTraffic(s.Router, 100)
		},
		Verify: func(t *testing.T, s *TestServer) {
			av := NewAlertVerifier(t, s.Registry)
			
			// Verify alerts are silenced
			assert.False(t, av.VerifyAlert("HighErrorRate", 5*time.Second),
				"Alert should be silenced during maintenance")
			
			// End maintenance window
			s.Metrics.EndMaintenanceWindow("upgrade")
			
			// Verify alerts start firing again
			assert.True(t, av.VerifyAlert("HighErrorRate", 5*time.Second),
				"Alert should fire after maintenance")
		},
	},
	"metric_aggregation": {
		Name:        "Metric Aggregation",
		Description: "Verifies metric aggregation functions",
		Setup: func(s *TestServer) {
			// Generate traffic with varying response times
			for i := 0; i < 5; i++ {
				s.WithSlowResponses(time.Duration(50*(i+1)) * time.Millisecond)
				GenerateTestTraffic(s.Router, 20)
			}
		},
		Verify: func(t *testing.T, s *TestServer) {
			ma := NewMetricsAssertion(t, s.Metrics)
			
			// Verify histogram buckets
			histogram := ma.GetHistogramValue("handler_request_duration")
			assert.Greater(t, histogram.GetSampleCount(), uint64(0))
			assert.Greater(t, histogram.GetSampleSum(), float64(0))
			
			// Verify quantiles
			p99 := ma.GetHistogramQuantile(0.99, "handler_request_duration")
			p50 := ma.GetHistogramQuantile(0.50, "handler_request_duration")
			assert.Greater(t, p99, p50, "99th percentile should be higher than median")
		},
	},
	"metric_persistence": {
		Name:        "Metric Persistence",
		Description: "Verifies metrics survive server restarts",
		Setup: func(s *TestServer) {
			// Generate initial metrics
			GenerateTestTraffic(s.Router, 50)
			
			// Record initial values
			ma := NewMetricsAssertion(t, s.Metrics)
			s.initialCount = ma.GetCounterValue("handler_request_count", nil)
			
			// Simulate server restart
			s.Restart()
		},
		Verify: func(t *testing.T, s *TestServer) {
			ma := NewMetricsAssertion(t, s.Metrics)
			
			// Verify metrics persisted
			currentCount := ma.GetCounterValue("handler_request_count", nil)
			assert.Equal(t, s.initialCount, currentCount,
				"Metrics should persist across restarts")
			
			// Generate more traffic
			GenerateTestTraffic(s.Router, 50)
			
			// Verify counters continue incrementing
			newCount := ma.GetCounterValue("handler_request_count", nil)
			assert.Greater(t, newCount, currentCount,
				"Counters should continue incrementing after restart")
		},
	},
	"alert_recovery": {
		Name:        "Alert Recovery",
		Description: "Verifies alert recovery behavior",
		Setup: func(s *TestServer) {
			// Generate error condition
			s.WithErrorRate(0.5)
			GenerateTestTraffic(s.Router, 50)
			
			// Wait for alert to fire
			av := NewAlertVerifier(t, s.Registry)
			assert.True(t, av.VerifyAlert("HighErrorRate", 5*time.Second))
			
			// Remove error condition
			s.WithErrorRate(0)
			GenerateTestTraffic(s.Router, 50)
		},
		Verify: func(t *testing.T, s *TestServer) {
			av := NewAlertVerifier(t, s.Registry)
			
			// Verify alert recovers
			assert.True(t, av.VerifyAlertRecovery("HighErrorRate", 5*time.Second))
			
			// Verify no other alerts
			av.AssertNoAlerts(t)
		},
	},
	"metric_staleness": {
		Name:        "Metric Staleness",
		Description: "Verifies stale metric detection",
		Setup: func(s *TestServer) {
			// Generate initial metrics
			GenerateTestTraffic(s.Router, 50)
			
			// Stop metric collection
			s.Metrics.Pause()
			time.Sleep(2 * time.Second)
		},
		Verify: func(t *testing.T, s *TestServer) {
			av := NewAlertVerifier(t, s.Registry)
			assert.True(t, av.VerifyAlert("StaleMetrics", 5*time.Second))
			
			// Resume metric collection
			s.Metrics.Resume()
			GenerateTestTraffic(s.Router, 10)
			assert.True(t, av.VerifyAlertRecovery("StaleMetrics", 5*time.Second))
		},
	},
	"alert_grouping": {
		Name:        "Alert Grouping",
		Description: "Verifies alert grouping behavior",
		Setup: func(s *TestServer) {
			// Generate multiple related alerts
			s.WithSlowResponses(200 * time.Millisecond)
			s.WithErrorRate(0.3)
			for i := 0; i < 3; i++ {
				go GenerateTestTraffic(s.Router, 100)
			}
		},
		Verify: func(t *testing.T, s *TestServer) {
			av := NewAlertVerifier(t, s.Registry)
			
			// Verify alerts are grouped
			groups := av.GetAlertGroups()
			assert.Equal(t, 1, len(groups), "Related alerts should be grouped")
			assert.GreaterOrEqual(t, len(groups[0].Alerts), 2, "Group should contain multiple alerts")
		},
	},
	"metric_relabeling": {
		Name:        "Metric Relabeling",
		Description: "Verifies metric relabeling rules",
		Setup: func(s *TestServer) {
			// Add custom labels
			s.Router.Use(func(c *gin.Context) {
				c.Set("environment", "test")
				c.Set("region", "us-west")
				c.Next()
			})
			GenerateTestTraffic(s.Router, 50)
		},
		Verify: func(t *testing.T, s *TestServer) {
			ma := NewMetricsAssertion(t, s.Metrics)
			
			// Verify labels are properly transformed
			metrics, err := s.Registry.Gather()
			assert.NoError(t, err)
			
			for _, mf := range metrics {
				for _, m := range mf.GetMetric() {
					// Check label transformations
					for _, l := range m.GetLabel() {
						assert.NotContains(t, l.GetName(), "test_",
							"Labels should not have test_ prefix")
					}
				}
			}
		},
	},
	"alert_inhibition": {
		Name:        "Alert Inhibition",
		Description: "Verifies alert inhibition rules",
		Setup: func(s *TestServer) {
			// Generate parent alert
			s.WithErrorRate(1.0)
			GenerateTestTraffic(s.Router, 20)
			
			// Generate child alerts
			s.WithSlowResponses(100 * time.Millisecond)
			GenerateTestTraffic(s.Router, 20)
		},
		Verify: func(t *testing.T, s *TestServer) {
			av := NewAlertVerifier(t, s.Registry)
			
			// Verify parent alert is firing
			assert.True(t, av.VerifyAlert("ServiceDown", 5*time.Second))
			
			// Verify child alerts are inhibited
			assert.False(t, av.VerifyAlert("HighLatency", 5*time.Second),
				"Child alert should be inhibited by parent")
		},
	},
	"metric_compaction": {
		Name:        "Metric Compaction",
		Description: "Verifies metric compaction behavior",
		Setup: func(s *TestServer) {
			// Generate high-cardinality metrics
			for i := 0; i < 1000; i++ {
				s.Router.Use(func(c *gin.Context) {
					c.Set("id", fmt.Sprintf("user_%d", i))
					c.Next()
				})
				GenerateTestTraffic(s.Router, 1)
			}
		},
		Verify: func(t *testing.T, s *TestServer) {
			// Wait for compaction
			time.Sleep(5 * time.Second)
			
			metrics, err := s.Registry.Gather()
			assert.NoError(t, err)
			
			// Verify metrics were compacted
			for _, mf := range metrics {
				if strings.Contains(mf.GetName(), "request") {
					assert.Less(t, len(mf.GetMetric()), 100,
						"Metrics should be compacted")
				}
			}
		},
	},
	"metric_deletion": {
		Name:        "Metric Deletion",
		Description: "Verifies metric deletion behavior",
		Setup: func(s *TestServer) {
			// Generate metrics to be deleted
			for i := 0; i < 10; i++ {
				s.Router.Use(func(c *gin.Context) {
					c.Set("temp_label", fmt.Sprintf("temp_%d", i))
					c.Next()
				})
				GenerateTestTraffic(s.Router, 5)
			}
			
			// Mark metrics for deletion
			s.Metrics.MarkForDeletion("temp_metric")
		},
		Verify: func(t *testing.T, s *TestServer) {
			// Wait for deletion
			time.Sleep(2 * time.Second)
			
			metrics, err := s.Registry.Gather()
			assert.NoError(t, err)
			
			// Verify metrics were deleted
			for _, mf := range metrics {
				assert.NotContains(t, mf.GetName(), "temp_",
					"Temporary metrics should be deleted")
			}
		},
	},
	"metric_validation": {
		Name:        "Metric Validation",
		Description: "Verifies metric validation rules",
		Setup: func(s *TestServer) {
			// Create validator
			validator := NewMetricValidator(t, s.Registry)

			// Generate test metrics
			GenerateTestTraffic(s.Router, 10)

			// Verify validation
			err := validator.ValidateMetrics()
			assert.NoError(t, err)
		},
		Verify: func(t *testing.T, s *TestServer) {
			metrics, err := s.Registry.Gather()
			assert.NoError(t, err)
			
			// Verify metric validation
			for _, mf := range metrics {
				for _, m := range mf.GetMetric() {
					for _, l := range m.GetLabel() {
						// Check label length
						assert.Less(t, len(l.GetValue()), 100,
							"Label value too long: %s", l.GetValue())
						
						// Check label characters
						assert.Regexp(t, "^[a-zA-Z0-9_]+$", l.GetValue(),
							"Invalid characters in label value: %s", l.GetValue())
					}
				}
			}
		},
	},
	"alert_routing": {
		Name:        "Alert Routing",
		Description: "Verifies alert routing rules",
		Setup: func(s *TestServer) {
			// Create router
			router := NewAlertRouter(t, s.Registry)

			// Add routes
			router.AddRoute("platform", "HighErrorRate")
			router.AddRoute("api", "HighLatency")

			// Generate alerts
			s.WithErrorRate(0.5)
			GenerateTestTraffic(s.Router, 50)

			// Verify routing
			assert.True(t, router.WaitForRoute("HighErrorRate", "platform", 5*time.Second))
			assert.True(t, router.WaitForRoute("HighLatency", "api", 5*time.Second))
		},
		Verify: func(t *testing.T, s *TestServer) {
			av := NewAlertVerifier(t, s.Registry)
			
			// Verify alerts are routed to correct teams
			routes := av.GetAlertRoutes()
			assert.Contains(t, routes["platform"], "HighErrorRate")
			assert.Contains(t, routes["api"], "HighLatency")
		},
	},
}

// GenerateTestTraffic generates n requests
func GenerateTestTraffic(router *gin.Engine, n int) {
	for i := 0; i < n; i++ {
		PerformRequest(router, "GET", "/api/v1/words", nil)
		PerformRequest(router, "GET", "/api/v1/groups", nil)
		if i%5 == 0 {
			PerformRequest(router, "GET", "/api/v1/words/1", nil)
		}
		time.Sleep(10 * time.Millisecond)
	}
} 