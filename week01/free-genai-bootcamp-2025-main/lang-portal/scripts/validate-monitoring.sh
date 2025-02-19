#!/bin/bash
set -e

# Configuration
MONITORING_DIR="monitoring"
CONFIG_DIR="config"
TEMP_DIR="/tmp/monitoring-validate"

# Function to validate Prometheus config
validate_prometheus_config() {
    echo "Validating Prometheus configuration..."
    
    # Run configuration tests
    if ! ./scripts/test-monitoring-config.sh; then
        echo "Error: Configuration tests failed"
        return 1
    fi

    # Create temporary container to validate config
    docker run --rm -v "${MONITORING_DIR}/prometheus/config:/config" \
        prom/prometheus:v${PROMETHEUS_VERSION} \
        promtool check config /config/prometheus.yml
}

# Function to validate Alertmanager config
validate_alertmanager_config() {
    echo "Validating Alertmanager configuration..."
    
    docker run --rm -v "${MONITORING_DIR}/alertmanager/config:/config" \
        prom/alertmanager:v${ALERTMANAGER_VERSION} \
        amtool check-config /config/alertmanager.yml
}

# Function to validate Grafana dashboards
validate_grafana_dashboards() {
    echo "Validating Grafana dashboards..."
    
    mkdir -p "$TEMP_DIR"
    
    for dashboard in "${MONITORING_DIR}/grafana/dashboards"/*.json; do
        echo "Validating dashboard: ${dashboard}"
        
        # Test dashboard queries
        if ! ./scripts/test-monitoring-config.sh --test-dashboard "$dashboard"; then
            echo "Error: Dashboard tests failed for ${dashboard}"
            return 1
        fi
        
        # Basic JSON validation
        if ! jq empty "$dashboard" > /dev/null 2>&1; then
            echo "Error: Invalid JSON in ${dashboard}"
            return 1
        fi
    done
}

# Function to validate Docker Compose
validate_docker_compose() {
    echo "Validating Docker Compose configuration..."
    
    docker-compose -f "${MONITORING_DIR}/docker-compose.yml" config
}

# Function to validate environment variables
validate_environment() {
    echo "Validating environment variables..."
    
    local required_vars=(
        "SLACK_WEBHOOK_URL"
        "ALERT_EMAIL"
        "SMTP_USERNAME"
        "SMTP_PASSWORD"
    )
    
    local missing_vars=()
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            missing_vars+=("$var")
        fi
    done
    
    if [ ${#missing_vars[@]} -ne 0 ]; then
        echo "Error: Missing required environment variables:"
        printf '%s\n' "${missing_vars[@]}"
        return 1
    fi
}

# Function to validate network access
validate_network() {
    echo "Validating network access..."
    
    local ports=(
        9090  # Prometheus
        3000  # Grafana
        9093  # Alertmanager
    )
    
    for port in "${ports[@]}"; do
        if netstat -tuln | grep -q ":${port} "; then
            echo "Error: Port ${port} is already in use"
            return 1
        fi
    done
}

# Function to validate disk space
validate_disk_space() {
    echo "Validating disk space..."
    
    local required_space=10  # GB
    local available_space=$(df -BG "${MONITORING_DIR}" | awk 'NR==2 {print $4}' | sed 's/G//')
    
    if [ "$available_space" -lt "$required_space" ]; then
        echo "Error: Insufficient disk space. Required: ${required_space}GB, Available: ${available_space}GB"
        return 1
    fi
}

# Main execution
echo "Starting monitoring configuration validation..."

# Create temporary directory
mkdir -p "$TEMP_DIR"
trap 'rm -rf "$TEMP_DIR"' EXIT

# Check if test script exists
if [ ! -f "./scripts/test-monitoring-config.sh" ]; then
    echo "Error: test-monitoring-config.sh script not found"
    exit 1
fi

# Make test script executable
chmod +x ./scripts/test-monitoring-config.sh

# Run validations
validate_environment || exit 1
validate_network || exit 1
validate_disk_space || exit 1
validate_prometheus_config || exit 1
validate_alertmanager_config || exit 1
validate_grafana_dashboards || exit 1
validate_docker_compose || exit 1

echo "Monitoring configuration validation completed successfully!"
echo "Next steps:"
echo "1. Deploy monitoring stack: ./scripts/deploy-monitoring.sh"
echo "2. Verify deployment: ./scripts/test-monitoring.sh" 