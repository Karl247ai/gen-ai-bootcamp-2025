#!/bin/bash
set -e

# Configuration
MONITORING_DIR="monitoring"
CONFIG_DIR="config"
TEST_DURATION=120 # 2 minutes

# Function to verify configuration files
verify_configs() {
    echo "Verifying configuration files..."
    
    # Check Prometheus config
    if ! docker run --rm -v "${MONITORING_DIR}/prometheus/config:/config" \
        prom/prometheus:v2.32.1 promtool check config /config/prometheus.yml; then
        echo "Error: Invalid Prometheus configuration"
        return 1
    fi

    # Check Alertmanager config
    if ! docker run --rm -v "${MONITORING_DIR}/alertmanager/config:/config" \
        prom/alertmanager:v0.23.0 amtool check-config /config/alertmanager.yml; then
        echo "Error: Invalid Alertmanager configuration"
        return 1
    fi

    return 0
}

# Function to test service integration
test_integration() {
    echo "Testing service integration..."
    
    # Test Prometheus-Alertmanager integration
    local alert_managers=$(curl -s "http://localhost:9090/api/v1/alertmanagers" | \
        jq -r '.data.activeAlertmanagers[].url')
    if [[ ! "$alert_managers" =~ "alertmanager:9093" ]]; then
        echo "Error: Alertmanager not connected to Prometheus"
        return 1
    fi

    # Test Grafana-Prometheus integration
    local datasources=$(curl -s -H "Authorization: Bearer ${GRAFANA_API_KEY}" \
        "http://localhost:3000/api/datasources")
    if [[ ! "$datasources" =~ "prometheus" ]]; then
        echo "Error: Prometheus datasource not configured in Grafana"
        return 1
    fi

    return 0
}

# Function to test alert rules
test_alert_rules() {
    echo "Testing alert rules..."
    
    # Create test alert condition
    docker-compose exec -T prometheus promtool test rules /etc/prometheus/alerts/*.yml

    return $?
}

# Main execution
echo "Starting monitoring setup test..."

# Verify configurations
if ! verify_configs; then
    echo "Configuration verification failed!"
    exit 1
fi

# Start monitoring stack
echo "Starting monitoring stack..."
cd "$MONITORING_DIR"
docker-compose up -d

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 30

# Test service integration
if ! test_integration; then
    echo "Service integration test failed!"
    docker-compose down
    exit 1
fi

# Test alert rules
if ! test_alert_rules; then
    echo "Alert rules test failed!"
    docker-compose down
    exit 1
fi

# Run monitoring test script
if ! ../scripts/test-monitoring.sh; then
    echo "Monitoring test failed!"
    docker-compose down
    exit 1
fi

echo "Monitoring setup test completed successfully!" 