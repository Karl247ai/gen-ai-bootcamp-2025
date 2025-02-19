#!/bin/bash
set -e

# Configuration
MONITORING_DIR="monitoring"
CONFIG_DIR="config"
TEST_DURATION=300 # 5 minutes

# Function to test Prometheus rules
test_prometheus_rules() {
    echo "Testing Prometheus alert rules..."
    
    # Create test data
    cat > "${MONITORING_DIR}/prometheus/test/test.yml" <<EOF
rule_files:
    - /alerts/*.yml

evaluation_interval: 1m

tests:
    - interval: 1m
      input_series:
        - series: 'http_requests_total{status="500"}'
          values: '0 1 2 3 4 5 6 7 8 9 10'
        - series: 'process_resident_memory_bytes'
          values: '1000000000+100x10'
        - series: 'go_goroutines'
          values: '100+1000x10'

      alert_rule_test:
        - eval_time: 5m
          alertname: HighErrorRate
          exp_alerts:
            - exp_labels:
                severity: critical
                
        - eval_time: 5m
          alertname: HighMemoryUsage
          exp_alerts:
            - exp_labels:
                severity: warning
EOF

    # Run tests
    docker run --rm -v "${MONITORING_DIR}/prometheus/config:/etc/prometheus" \
        -v "${MONITORING_DIR}/prometheus/test:/test" \
        prom/prometheus:v${PROMETHEUS_VERSION} \
        promtool test rules /test/test.yml
}

# Function to test Alertmanager configuration
test_alertmanager_config() {
    echo "Testing Alertmanager configuration..."
    
    # Test alert routing
    docker run --rm -v "${MONITORING_DIR}/alertmanager/config:/config" \
        prom/alertmanager:v${ALERTMANAGER_VERSION} \
        amtool --config.file=/config/alertmanager.yml config routes test \
        --verify.receivers=team-slack \
        --verify.labels=severity=critical

    # Test alert templates
    docker run --rm -v "${MONITORING_DIR}/alertmanager/config:/config" \
        prom/alertmanager:v${ALERTMANAGER_VERSION} \
        amtool --config.file=/config/alertmanager.yml config routes test \
        --verify.receivers=team-slack \
        --verify.labels=severity=warning
}

# Function to test Grafana dashboards
test_grafana_dashboards() {
    echo "Testing Grafana dashboards..."
    
    for dashboard in "${MONITORING_DIR}/grafana/dashboards"/*.json; do
        echo "Testing dashboard: ${dashboard}"
        
        # Verify dashboard queries
        queries=$(jq -r '.panels[].targets[].expr' "$dashboard")
        for query in $queries; do
            if [ ! -z "$query" ]; then
                echo "Testing query: $query"
                docker run --rm -v "${MONITORING_DIR}/prometheus/config:/etc/prometheus" \
                    prom/prometheus:v${PROMETHEUS_VERSION} \
                    promtool check metrics "$query"
            fi
        done
    done
}

# Function to test metric collection
test_metric_collection() {
    echo "Testing metric collection..."
    
    # Start test containers
    docker-compose -f "${MONITORING_DIR}/docker-compose.yml" up -d
    
    # Wait for services to be ready
    sleep 30
    
    # Check metric endpoints
    if ! curl -s "http://localhost:9090/api/v1/targets" | jq -e '.data.activeTargets[] | select(.health=="up")' > /dev/null; then
        echo "Error: Not all targets are up"
        docker-compose -f "${MONITORING_DIR}/docker-compose.yml" down
        return 1
    fi
    
    # Generate test load
    ./scripts/test-monitoring.sh
    
    # Verify metrics are being collected
    if ! curl -s "http://localhost:9090/api/v1/query?query=up" | jq -e '.data.result[] | select(.value[1]=="1")' > /dev/null; then
        echo "Error: Metrics are not being collected"
        docker-compose -f "${MONITORING_DIR}/docker-compose.yml" down
        return 1
    fi
    
    # Clean up
    docker-compose -f "${MONITORING_DIR}/docker-compose.yml" down
}

# Main execution
echo "Starting monitoring configuration tests..."

# Run tests
test_prometheus_rules || exit 1
test_alertmanager_config || exit 1
test_grafana_dashboards || exit 1
test_metric_collection || exit 1

echo "Monitoring configuration tests completed successfully!"
echo "Next steps:"
echo "1. Review test results"
echo "2. Deploy monitoring stack: ./scripts/deploy-monitoring.sh" 