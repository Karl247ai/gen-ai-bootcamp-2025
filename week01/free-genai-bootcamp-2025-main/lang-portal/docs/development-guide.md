# Development Guide

## Setup Instructions

### Local Development Environment
1. Prerequisites
   ```bash
   # Required software versions
   Go: 1.21 or higher
   SQLite: 3.x
   Git: 2.x
   Make: 4.x (optional)
   ```

2. Initial Setup
   ```bash
   # Clone repository
   git clone https://github.com/your-org/lang-portal
   cd lang-portal

   # Install Go dependencies
   go mod download

   # Install development tools
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   go install github.com/golang/mock/mockgen@latest
   go install github.com/swaggo/swag/cmd/swag@latest
   ```

3. Database Setup
   ```bash
   # Initialize development database
   make initdb
   # or
   go run cmd/server/main.go initdb

   # Run migrations
   make migrate
   # or
   go run cmd/server/main.go migrate

   # Import seed data
   make seed
   # or
   go run cmd/server/main.go seed
   ```

## Monitoring Setup

### Local Development

1. **Prerequisites**
   ```bash
   # Install required tools
   brew install prometheus grafana alertmanager
   ```

2. **Configuration**
   ```bash
   # Set up environment variables
   export SLACK_WEBHOOK_URL="your-webhook-url"
   export ALERT_EMAIL="your-email@company.com"
   ```

3. **Start Monitoring**
   ```bash
   # Deploy monitoring stack
   ./scripts/deploy-monitoring.sh
   ```

4. **Access Monitoring**
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000
   - Alertmanager: http://localhost:9093

### Testing Monitoring

1. **Run Tests**
   ```bash
   # Run all monitoring tests
   go test ./test/integration/... -tags=monitoring
   
   # Run specific test
   go test ./test/integration/monitoring_test.go
   ```

2. **Generate Test Load**
   ```bash
   # Generate test traffic
   ./scripts/test-monitoring.sh
   ```

3. **Verify Alerts**
   ```bash
   # Check active alerts
   ./scripts/manage-alerts.sh list
   ```

### Development Workflow

1. **Adding New Metrics**
   ```go
   // Add metric to internal/metrics/metrics.go
   counter := prometheus.NewCounter(prometheus.CounterOpts{
       Name: "my_metric_name",
       Help: "Description of the metric",
   })
   ```

2. **Adding New Alerts**
   ```yaml
   # Add to monitoring/prometheus/alerts/application.yml
   - alert: MyNewAlert
     expr: my_metric > threshold
     for: 5m
     labels:
       severity: warning
     annotations:
       description: "Alert description"
   ```

3. **Adding New Dashboards**
   ```bash
   # Export dashboard from Grafana UI
   # Save to monitoring/dashboards/my_dashboard.json
   ```

### Best Practices

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

### Troubleshooting

1. **Common Issues**
   - Check service logs: `docker-compose logs`
   - Verify metrics endpoint: `curl localhost:8080/metrics`
   - Check alert rules: `./scripts/manage-alerts.sh rules`

2. **Performance Issues**
   - Monitor resource usage
   - Check database connections
   - Review cache hit rates

3. **Alert Storm Handling**
   - Group related alerts
   - Adjust thresholds
   - Update notification settings

## Coding Standards

### Code Organization
```
backend_go/
├── cmd/                    # Application entrypoints
├── internal/              # Private application code
│   ├── api/              # API handlers
│   ├── middleware/       # HTTP middleware
│   ├── models/           # Database models
│   ├── repository/       # Database operations
│   └── service/          # Business logic
└── pkg/                  # Public packages
```

### Code Style
1. Go Standards
   ```go
   // Package names should be short and clear
   package user

   // Use clear, descriptive names
   func CreateUser(ctx context.Context, user *User) error {
       // Indent with tabs
       if user == nil {
           return ErrInvalidUser
       }
       
       // Group similar declarations
       var (
           db  *sql.DB
           err error
       )
       
       // Return early for errors
       if err != nil {
           return fmt.Errorf("failed to create user: %w", err)
       }
   }
   ```

2. Error Handling
   ```go
   // Define custom errors
   var (
       ErrUserNotFound = errors.New("user not found")
       ErrInvalidInput = errors.New("invalid input")
   )

   // Wrap errors with context
   if err != nil {
       return fmt.Errorf("failed to fetch user %d: %w", id, err)
   }
   ```

## Testing Guidelines

### Unit Testing
```go
// user_test.go
func TestCreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   *User
        wantErr bool
    }{
        {
            name: "valid user",
            input: &User{
                Name:  "Test User",
                Email: "test@example.com",
            },
            wantErr: false,
        },
        {
            name:    "nil user",
            input:   nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := CreateUser(context.Background(), tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Testing
```go
// integration_test.go
func TestUserRepository(t *testing.T) {
    // Skip in short mode
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Setup test database
    db, cleanup := setupTestDB(t)
    defer cleanup()

    // Run tests
    t.Run("CreateAndFetchUser", func(t *testing.T) {
        repo := NewUserRepository(db)
        // Test implementation
    })
}
```

### Monitoring Tests

1. **Metric Tests**
```go
// Test metric collection
func TestMonitoringMetrics(t *testing.T) {
    server := setupTestServer(t)
    ma := testutil.NewMetricsAssertion(t, server.Metrics)

    // Generate test traffic
    testutil.PerformRequest(server.Router, "GET", "/api/v1/words", nil)

    // Verify metrics
    count := ma.GetCounterValue("handler_request_count")
    assert.Equal(t, float64(1), count)
}
```

2. **Alert Tests**
```go
// Test alert rules
func TestMonitoringAlerts(t *testing.T) {
    server := setupTestServer(t)
    ma := testutil.NewMetricsAssertion(t, server.Metrics)

    // Generate alert condition
    generateHighLatency(t, server.Router)

    // Verify alert would fire
    alertWouldFire := ma.CheckAlert("HighLatency")
    assert.True(t, alertWouldFire)
}
```

3. **Dashboard Tests**
```go
// Test dashboard metrics
func TestMonitoringDashboards(t *testing.T) {
    server := setupTestServer(t)
    ma := testutil.NewMetricsAssertion(t, server.Metrics)

    // Verify required metrics exist
    exists := ma.MetricExists("handler_request_count")
    assert.True(t, exists)
}
```

### Monitoring Test Scenarios

The following test scenarios are available for monitoring:

1. **Basic Scenarios**
   - High Latency Detection
   - Error Rate Spike
   - Cache Performance
   - Resource Exhaustion
   - Maintenance Impact

2. **Advanced Scenarios**
   - Concurrent Maintenance
   - Maintenance Error Handling
   - Performance Degradation
   - Maintenance Window
   - Alert Correlation

3. **System Scenarios**
   - Metric Cardinality
   - Alert Silencing
   - Metric Aggregation
   - Metric Persistence
   - Alert Recovery

Example usage:
```go
// Test metric cardinality
func TestMetricCardinality(t *testing.T) {
    scenario := testutil.CommonScenarios["metric_cardinality"]
    testutil.RunTestScenario(t, scenario)
}

// Test alert recovery
func TestAlertRecovery(t *testing.T) {
    scenario := testutil.CommonScenarios["alert_recovery"]
    testutil.RunTestScenario(t, scenario)
}
```

For custom scenarios, implement the TestScenario interface:
```go
type TestScenario struct {
    Name        string
    Description string
    Setup       func(*TestServer)
    Verify      func(*testing.T, *TestServer)
    Cleanup     func(*TestServer)
    Timeout     time.Duration
}
```

### Test Configuration

1. **Timeouts and Retries**
```go
scenario := TestScenario{
    Name: "Custom Test",
    Timeout: 30 * time.Second,
    Setup: func(s *TestServer) {
        // Setup with retry
        testutil.RetryWithTimeout(func() error {
            return setupFunction()
        }, 5, time.Second)
    },
}
```

2. **Cleanup Behavior**
```go
scenario := TestScenario{
    Name: "Cleanup Test",
    Cleanup: func(s *TestServer) {
        // Custom cleanup logic
        s.Metrics.Reset()
        s.Registry.Clear()
    },
}
```

## Git Workflow

### Branch Naming
```bash
# Feature branches
feature/add-user-authentication
feature/implement-study-session

# Bug fixes
bugfix/fix-session-timeout
bugfix/correct-word-count

# Releases
release/v1.0.0
release/v1.1.0
```

### Commit Messages
```bash
# Format: <type>(<scope>): <description>

# Examples:
feat(auth): add user authentication
fix(session): correct session timeout calculation
docs(api): update endpoint documentation
test(words): add tests for word creation
```

### Pull Request Process
1. Checklist
   ```markdown
   - [ ] Tests added/updated
   - [ ] Documentation updated
   - [ ] Changelog updated
   - [ ] Version bumped (if applicable)
   - [ ] Reviewed by team member
   ```

2. Review Guidelines
   ```markdown
   1. Code Quality
      - Follows coding standards
      - Proper error handling
      - Adequate test coverage

   2. Performance
      - No N+1 queries
      - Proper indexing
      - Resource efficient

   3. Security
      - Input validation
      - SQL injection prevention
      - XSS prevention
   ```

## Development Tools

### Makefile Commands
```makefile
# Common development tasks
.PHONY: build test lint run

build:
    go build -o bin/server cmd/server/main.go

test:
    go test -v ./...

lint:
    golangci-lint run

run:
    go run cmd/server/main.go
```

### Debug Configuration
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch API Server",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/server/main.go",
            "env": {
                "GO_ENV": "development",
                "API_PORT": "8080",
                "DATABASE_PATH": "./words.db"
            }
        }
    ]
}
``` 