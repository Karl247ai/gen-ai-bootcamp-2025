#!/bin/bash
set -e

# Configuration
RETENTION_DAYS=30
PROMETHEUS_DATA="/var/lib/prometheus"
GRAFANA_DATA="/var/lib/grafana"
MIN_DISK_SPACE=20 # Minimum disk space in GB

# Function to check disk space
check_disk_space() {
    local path="$1"
    local available_space=$(df -BG "$path" | awk 'NR==2 {print $4}' | sed 's/G//')
    
    if [ "$available_space" -lt "$MIN_DISK_SPACE" ]; then
        echo "Warning: Low disk space on $path ($available_space GB available)"
        return 1
    fi
    return 0
}

# Function to clean old data
clean_old_data() {
    local data_dir="$1"
    local days="$2"
    
    find "$data_dir" -type f -mtime +"$days" -delete
    find "$data_dir" -type d -empty -delete
}

# Function to compact Prometheus data
compact_prometheus() {
    echo "Compacting Prometheus data..."
    docker exec prometheus promtool tsdb compact /prometheus
}

# Function to optimize Grafana database
optimize_grafana() {
    echo "Optimizing Grafana database..."
    docker exec grafana grafana-cli admin database optimize
}

# Main execution
echo "Starting metrics rotation..."

# Check disk space
if ! check_disk_space "$PROMETHEUS_DATA" || ! check_disk_space "$GRAFANA_DATA"; then
    echo "Low disk space detected. Starting emergency cleanup..."
    RETENTION_DAYS=7
fi

# Stop services
echo "Stopping monitoring services..."
cd monitoring
docker-compose stop prometheus grafana

# Clean old data
echo "Cleaning old Prometheus data..."
clean_old_data "$PROMETHEUS_DATA" "$RETENTION_DAYS"

echo "Cleaning old Grafana data..."
clean_old_data "$GRAFANA_DATA/png" "$RETENTION_DAYS" # Clean rendered images
clean_old_data "$GRAFANA_DATA/csv" "$RETENTION_DAYS" # Clean exported data

# Compact and optimize
compact_prometheus
optimize_grafana

# Start services
echo "Starting monitoring services..."
docker-compose start prometheus grafana

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 10

# Verify services
echo "Verifying services..."
if curl -s "http://localhost:9090/-/healthy" > /dev/null && \
   curl -s "http://localhost:3000/api/health" > /dev/null; then
    echo "Services are healthy"
else
    echo "Warning: Services may not be healthy. Please check manually."
fi

echo "Metrics rotation completed!" 