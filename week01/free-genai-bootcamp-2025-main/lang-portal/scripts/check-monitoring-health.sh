#!/bin/bash
set -e

# Configuration
MONITORING_DIR="monitoring"
HEALTH_CHECK_TIMEOUT=5  # seconds
RETRY_ATTEMPTS=3
RETRY_DELAY=5  # seconds

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check service health
check_service_health() {
    local service=$1
    local url=$2
    local attempt=1

    echo -n "Checking ${service} health... "
    
    while [ $attempt -le $RETRY_ATTEMPTS ]; do
        if curl -s --max-time $HEALTH_CHECK_TIMEOUT "$url" > /dev/null; then
            echo -e "${GREEN}OK${NC}"
            return 0
        fi
        
        if [ $attempt -lt $RETRY_ATTEMPTS ]; then
            echo -n "Retry $attempt/$RETRY_ATTEMPTS... "
            sleep $RETRY_DELAY
        fi
        attempt=$((attempt + 1))
    done
    
    echo -e "${RED}FAILED${NC}"
    return 1
}

# Function to check metrics collection
check_metrics_collection() {
    echo -n "Checking metrics collection... "
    
    local result=$(curl -s "http://localhost:9090/api/v1/query?query=up")
    if [[ "$result" =~ "success" ]]; then
        echo -e "${GREEN}OK${NC}"
        return 0
    else
        echo -e "${RED}FAILED${NC}"
        return 1
    fi
}

# Function to check alert manager status
check_alertmanager_status() {
    echo -n "Checking alert manager status... "
    
    local result=$(curl -s "http://localhost:9093/api/v2/status")
    if [[ "$result" =~ "cluster" ]]; then
        echo -e "${GREEN}OK${NC}"
        return 0
    else
        echo -e "${RED}FAILED${NC}"
        return 1
    fi
}

# Function to check Grafana dashboards
check_grafana_dashboards() {
    echo -n "Checking Grafana dashboards... "
    
    local result=$(curl -s -H "Authorization: Bearer ${GRAFANA_API_KEY}" \
        "http://localhost:3000/api/search?type=dash-db")
    if [[ "$result" =~ "id" ]]; then
        echo -e "${GREEN}OK${NC}"
        return 0
    else
        echo -e "${RED}FAILED${NC}"
        return 1
    fi
}

# Function to check disk space
check_disk_space() {
    echo -n "Checking disk space... "
    
    local threshold=90  # percentage
    local usage=$(df -h "${MONITORING_DIR}" | awk 'NR==2 {print $5}' | sed 's/%//')
    
    if [ "$usage" -gt "$threshold" ]; then
        echo -e "${RED}CRITICAL: ${usage}% used${NC}"
        return 1
    else
        echo -e "${GREEN}OK (${usage}% used)${NC}"
        return 0
    fi
}

# Function to check container status
check_container_status() {
    echo -n "Checking container status... "
    
    local containers=("prometheus" "grafana" "alertmanager")
    local failed_containers=()
    
    for container in "${containers[@]}"; do
        if ! docker ps --format '{{.Names}}' | grep -q "$container"; then
            failed_containers+=("$container")
        fi
    done
    
    if [ ${#failed_containers[@]} -eq 0 ]; then
        echo -e "${GREEN}OK${NC}"
        return 0
    else
        echo -e "${RED}FAILED: ${failed_containers[*]} not running${NC}"
        return 1
    fi
}

# Main execution
echo "Starting monitoring health check..."

# Track overall status
status=0

# Run health checks
check_service_health "Prometheus" "http://localhost:9090/-/healthy" || status=1
check_service_health "Grafana" "http://localhost:3000/api/health" || status=1
check_service_health "Alertmanager" "http://localhost:9093/-/healthy" || status=1
check_metrics_collection || status=1
check_alertmanager_status || status=1
check_grafana_dashboards || status=1
check_disk_space || status=1
check_container_status || status=1

# Print summary
echo
if [ $status -eq 0 ]; then
    echo -e "${GREEN}All health checks passed!${NC}"
else
    echo -e "${RED}Some health checks failed!${NC}"
    echo "Please check the logs for more details:"
    echo "docker-compose -f monitoring/docker-compose.yml logs"
fi

exit $status 