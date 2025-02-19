#!/bin/bash
set -e

# Configuration
MONITORING_DIR="monitoring"
CONFIG_DIR="config"
GRAFANA_VERSION="9.5.2"
PROMETHEUS_VERSION="v2.44.0"
ALERTMANAGER_VERSION="v0.25.0"
MAINTENANCE_DIR="${MONITORING_DIR}/maintenance"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check prerequisites
check_prerequisites() {
    echo "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        echo "Error: Docker is required but not installed"
        exit 1
    fi

    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        echo "Error: Docker Compose is required but not installed"
        exit 1
    }

    # Check required environment variables
    if [ -z "$SLACK_WEBHOOK_URL" ]; then
        echo "Warning: SLACK_WEBHOOK_URL is not set. Slack notifications will be disabled."
    fi
    
    if [ -z "$ALERT_EMAIL" ]; then
        echo "Warning: ALERT_EMAIL is not set. Email notifications will be disabled."
    }
}

# Function to create directories
create_directories() {
    echo "Creating directories..."
    mkdir -p "${MONITORING_DIR}/prometheus/data"
    mkdir -p "${MONITORING_DIR}/grafana/data"
    mkdir -p "${MONITORING_DIR}/alertmanager/data"
    
    # Set permissions
    chmod -R 777 "${MONITORING_DIR}/prometheus/data"
    chmod -R 777 "${MONITORING_DIR}/grafana/data"
    chmod -R 777 "${MONITORING_DIR}/alertmanager/data"
}

# Function to deploy monitoring stack
deploy_stack() {
    echo "Deploying monitoring stack..."
    
    # Pull images
    docker-compose -f "${MONITORING_DIR}/docker-compose.yml" pull
    
    # Start services
    docker-compose -f "${MONITORING_DIR}/docker-compose.yml" up -d
    
    # Wait for services to be ready
    echo "Waiting for services to be ready..."
    sleep 10
}

# Function to configure Grafana
configure_grafana() {
    echo "Configuring Grafana..."
    
    # Create provisioning directories
    mkdir -p "${MONITORING_DIR}/grafana/provisioning/datasources"
    mkdir -p "${MONITORING_DIR}/grafana/provisioning/dashboards"
    
    # Copy dashboard configurations
    cp monitoring/grafana/dashboards/*.json "${MONITORING_DIR}/grafana/provisioning/dashboards/"
    
    # Configure data sources
    cat > "${MONITORING_DIR}/grafana/provisioning/datasources/prometheus.yaml" <<EOF
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
EOF

    # Configure dashboard provider
    cat > "${MONITORING_DIR}/grafana/provisioning/dashboards/provider.yaml" <<EOF
apiVersion: 1
providers:
  - name: 'Default'
    folder: 'General'
    type: file
    options:
      path: /etc/grafana/provisioning/dashboards
EOF

    # Wait for Grafana to be ready
    until curl -s "http://localhost:3000/api/health" > /dev/null; do
        echo "Waiting for Grafana..."
        sleep 5
    done
}

# Function to verify deployment
verify_deployment() {
    echo "Verifying deployment..."
    
    # Check Prometheus
    if ! curl -s "http://localhost:9090/-/healthy" > /dev/null; then
        echo "Error: Prometheus is not healthy"
        exit 1
    fi
    
    # Check Alertmanager
    if ! curl -s "http://localhost:9093/-/healthy" > /dev/null; then
        echo "Error: Alertmanager is not healthy"
        exit 1
    fi
    
    # Check Grafana
    if ! curl -s "http://localhost:3000/api/health" > /dev/null; then
        echo "Error: Grafana is not healthy"
        exit 1
    fi
    
    # Run monitoring tests
    if ! ./scripts/test-monitoring.sh; then
        echo "Error: Monitoring tests failed"
        exit 1
    fi
}

# Main execution
echo "Starting monitoring deployment..."

# Check prerequisites
check_prerequisites

# Create directories
create_directories

# Initialize maintenance tracking
mkdir -p "$MAINTENANCE_DIR"
touch "${MAINTENANCE_DIR}/maintenance_history.jsonl"

# Record deployment start
./scripts/track-maintenance.sh start "Deploying monitoring stack"

# Deploy stack
deploy_stack

# Configure Grafana
configure_grafana

# Verify deployment
verify_deployment

# Record deployment completion
if [ $? -eq 0 ]; then
    ./scripts/track-maintenance.sh end "Monitoring stack deployed successfully"
    echo -e "${GREEN}Deployment completed successfully!${NC}"
else
    ./scripts/track-maintenance.sh end "Monitoring stack deployment failed"
    echo -e "${RED}Deployment failed!${NC}"
    exit 1
fi

# Print access information
echo "Access monitoring services at:"
echo "Grafana: http://localhost:3000"
echo "Prometheus: http://localhost:9090"
echo "AlertManager: http://localhost:9093"

# Deploy monitoring stack for testing
docker run -d \
    --name prometheus \
    -p 9090:9090 \
    -v $(pwd)/monitoring/prometheus:/etc/prometheus \
    prom/prometheus

docker run -d \
    --name grafana \
    -p 3000:3000 \
    -v $(pwd)/monitoring/grafana:/etc/grafana \
    grafana/grafana

docker run -d \
    --name alertmanager \
    -p 9093:9093 \
    -v $(pwd)/monitoring/alertmanager:/etc/alertmanager \
    prom/alertmanager

echo "Monitoring stack deployed:"
echo "Prometheus: http://localhost:9090"
echo "Grafana: http://localhost:3000"
echo "AlertManager: http://localhost:9093" 