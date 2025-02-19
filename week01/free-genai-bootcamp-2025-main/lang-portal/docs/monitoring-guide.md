# Monitoring Guide

This guide describes the monitoring setup for the Lang Portal application.

## Overview

For detailed operational procedures, see the [Operations Manual](operations-manual.md).

The monitoring stack consists of:
- Prometheus for metrics collection
- Grafana for visualization
- AlertManager for alerts

## Metrics

### Maintenance Metrics

1. **Activity Tracking**
```text
# Total maintenance activities
maintenance_total{type="upgrade"} 5
maintenance_total{type="backup"} 12
maintenance_total{type="cleanup"} 3

# Maintenance duration
maintenance_duration_seconds_bucket{type="upgrade",le="60"} 0
maintenance_duration_seconds_bucket{type="upgrade",le="120"} 2

# Maintenance errors
maintenance_errors_total{type="upgrade"} 1
```

2. **Performance Impact**
```text
# Response time during maintenance
http_request_duration_seconds{maintenance="true"}

# Error rate during maintenance
http_errors_total{maintenance="true"}
```

### Application Metrics

Key metrics exposed by the application:

- `handler_request_count` - Request count by endpoint
- `handler_request_duration` - Request duration histogram
- `handler_error_count` - Error count by type
- `cache_get_hit/miss` - Cache performance metrics
- `db_connections_in_use` - Active database connections

### System Metrics

- `runtime_memory_alloc` - Memory allocation
- `runtime_goroutines` - Number of goroutines
- `process_cpu_seconds_total` - CPU usage

## Dashboards

### Maintenance Dashboard

The maintenance dashboard provides visibility into maintenance activities:

1. **Activity Overview**
   - Total maintenance activities by type
   - Duration distribution
   - Error count

2. **Performance Impact**
   - Response time during maintenance
   - Error rate changes
   - Resource utilization

3. **Trends**
   - Historical maintenance patterns
   - Common maintenance windows
   - Duration trends

### System Health Dashboard

The system health dashboard provides real-time visibility into system performance:

1. **Resource Utilization**
   - CPU usage gauge
   - Memory usage gauge
   - Goroutine count

2. **Service Health**
   - Service uptime history
   - Health check status
   - Component availability

3. **Component Performance**
   - Database connection pools
   - Cache hit/miss rates
   - System load trends

### API Performance Dashboard

1. **API Performance Dashboard**
   - Request rates by endpoint
   - Response time percentiles
   - Error rates
   - Cache performance

2. **Service Health Dashboard**
   - Memory usage
   - Goroutine count
   - Database connection pool
   - Overall service health

## Alerts

### Application Alerts

- High error rate (>5% for 5m)
- Slow responses (p95 >500ms)
- High request rate (>1000 req/s)
- High memory usage (>85%)
- High goroutine count (>10000)

### Database Alerts

- High connection usage (>80%)
- Slow queries detected
- Low cache hit rate (<80%)

## Setup

1. Deploy monitoring stack:
```bash
./scripts/deploy-monitoring.sh
```

2. Access dashboards:
- Grafana: http://localhost:3000
- Prometheus: http://localhost:9090
- AlertManager: http://localhost:9093

## Testing

### Performance Testing

1. **Load Testing**
```bash
# Run performance tests
./scripts/test-monitoring-performance.sh
```

2. **Go Performance Tests**
```bash
# Run Go performance tests
go test ./test/integration -run TestMonitoringPerformance
```

Performance test coverage:
- Metric collection performance
- Metric cardinality
- Query performance
- Resource usage

2. **Performance Metrics**
- Time series count: < 5000
- Query latency: < 1s
- Resource usage: < 80%

3. **Scaling Guidelines**
- Increase retention period when disk usage > 70%
- Add Prometheus replicas when query latency > 1s
- Adjust scrape interval when CPU usage > 80%

4. **Performance Optimization**
- Use recording rules for expensive queries
- Limit label cardinality
- Adjust scrape intervals based on load
- Monitor memory and CPU usage

### Advanced Test Scenarios

The monitoring system includes several advanced test scenarios:

1. **Data Quality**
   - Metric Staleness: Verifies stale metric detection
   - Metric Relabeling: Tests label transformation rules
   - Metric Compaction: Validates metric compaction behavior

2. **Alert Behavior**
   - Alert Grouping: Tests alert correlation and grouping
   - Alert Inhibition: Verifies alert inhibition rules
   - Alert Recovery: Tests alert recovery behavior

Example usage:
```go
func TestAdvancedMonitoring(t *testing.T) {
    scenarios := []string{
        "metric_staleness",
        "alert_grouping",
        "metric_relabeling",
        "alert_inhibition",
        "metric_compaction",
    }

    for _, name := range scenarios {
        t.Run(name, func(t *testing.T) {
            scenario := testutil.CommonScenarios[name]
            testutil.RunTestScenario(t, scenario)
        })
    }
}
```

3. **Data Management**
   - Metric Deletion: Tests metric cleanup
   - Metric Validation: Verifies metric format rules
   - Alert Routing: Tests alert team routing

Example usage:
```go
func TestDataManagement(t *testing.T) {
    // Test metric deletion
    testutil.RunTestScenario(t, testutil.CommonScenarios["metric_deletion"])

    // Test metric validation
    testutil.RunTestScenario(t, testutil.CommonScenarios["metric_validation"])

    // Test alert routing
    testutil.RunTestScenario(t, testutil.CommonScenarios["alert_routing"])
}
```

### Test Configuration

Configure test behavior:
```go
type TestConfig struct {
    Timeout        time.Duration // Maximum test duration
    RetryInterval  time.Duration // Time between retries
    RetryAttempts  int          // Number of retry attempts
    CleanupEnabled bool         // Enable test cleanup
}
```

### Test Utilities

1. **Alert Rule Testing**
```go
// Create alert rule tester
art := testutil.NewAlertRuleTester(t, registry)

// Add rule
art.AddRule(testutil.AlertRule{
    Name:       "HighLatency",
    Expression: "handler_latency_seconds > 0.5",
    Duration:   5 * time.Minute,
    Labels: map[string]string{
        "severity": "warning",
        "team":     "platform",
    },
})

// Verify rule
art.VerifyExpression("HighLatency", true)
art.VerifyDuration("HighLatency", 5 * time.Minute)
```

2. **Maintenance Window Testing**
```go
// Create maintenance window tester
mwt := testutil.NewMaintenanceWindowTester(t, registry)

// Start window
mwt.StartWindow(testutil.MaintenanceWindow{
    Name: "planned_upgrade",
    Labels: map[string]string{
        "type": "upgrade",
        "team": "platform",
    },
})

// Verify window
mwt.VerifyLabels("planned_upgrade", map[string]string{
    "type": "upgrade",
})
```

### Test Recording

The monitoring system includes test recording capabilities:

```go
// Create test recorder
recorder := testutil.NewTestRecorder(t, "test_results.json")

// Start test
record := recorder.StartTest("performance_test", labels)

// Record metrics
recorder.RecordMetric(record, "latency", value)

// Add metadata
recorder.AddMetadata(record, "version", "1.0.0")

// End test
recorder.EndTest(record, "completed")
```

### Metric Collection

Collect and analyze metrics during tests:

```go
// Create collector
collector := testutil.NewMetricCollector(t, registry)

// Collect samples
collector.Collect("request_count")

// Calculate rates
rate := collector.CalculateRate("request_count", time.Second)
```

## Best Practices

1. **Metric Naming**
   - Use snake_case for metric names
   - Include relevant labels
   - Follow Prometheus naming conventions

2. **Alert Configuration**
   - Set appropriate thresholds
   - Include clear descriptions
   - Configure proper notification channels

3. **Dashboard Organization**
   - Group related metrics
   - Use consistent time ranges
   - Include documentation links

4. **Performance Optimization**
   - Use recording rules for complex queries
   - Limit cardinality of labels
   - Configure appropriate retention periods
   - Monitor resource usage trends

## Troubleshooting

Common issues and solutions:

1. **High Memory Usage**
   - Check for memory leaks
   - Review cache size configuration
   - Consider scaling vertically

2. **High Error Rate**
   - Check application logs
   - Review recent deployments
   - Check external dependencies

3. **Slow Responses**
   - Check database performance
   - Review cache hit rates
   - Monitor system resources

4. **High Resource Usage**
   - Check container stats
   - Adjust resource limits if needed
   - Monitor resource usage trends 