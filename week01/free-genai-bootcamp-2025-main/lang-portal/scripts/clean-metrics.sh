#!/bin/bash
set -e

# Configuration
MONITORING_DIR="monitoring"
SERVICE="${1:-}"
RETENTION_DAYS=15

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to clean Prometheus data
clean_prometheus() {
    echo "Cleaning Prometheus data..."
    
    # Stop Prometheus
    docker-compose -f "${MONITORING_DIR}/docker-compose.yml" stop prometheus
    
    # Clean old data
    find "${MONITORING_DIR}/prometheus/data" -mtime +${RETENTION_DAYS} -delete
    
    # Start Prometheus
    docker-compose -f "${MONITORING_DIR}/docker-compose.yml" start prometheus
}

# Function to clean Grafana data
clean_grafana() {
    echo "Cleaning Grafana data..."
    
    # Stop Grafana
    docker-compose -f "${MONITORING_DIR}/docker-compose.yml" stop grafana
    
    # Clean session data
    find "${MONITORING_DIR}/grafana/data/sessions" -mtime +1 -delete
    
    # Start Grafana
    docker-compose -f "${MONITORING_DIR}/docker-compose.yml" start grafana
}

# Function to clean AlertManager data
clean_alertmanager() {
    echo "Cleaning AlertManager data..."
    
    # Stop AlertManager
    docker-compose -f "${MONITORING_DIR}/docker-compose.yml" stop alertmanager
    
    # Clean notification history
    find "${MONITORING_DIR}/alertmanager/data" -name "*.db" -mtime +${RETENTION_DAYS} -delete
    
    # Start AlertManager
    docker-compose -f "${MONITORING_DIR}/docker-compose.yml" start alertmanager
}

# Main execution
case "$SERVICE" in
    "prometheus")
        clean_prometheus
        ;;
    "grafana")
        clean_grafana
        ;;
    "alertmanager")
        clean_alertmanager
        ;;
    "all")
        clean_prometheus
        clean_grafana
        clean_alertmanager
        ;;
    *)
        echo "Usage: $0 {prometheus|grafana|alertmanager|all}"
        exit 1
        ;;
esac

echo -e "${GREEN}Cleanup completed successfully!${NC}" 