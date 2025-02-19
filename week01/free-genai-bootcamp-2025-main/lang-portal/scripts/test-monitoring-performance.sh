#!/bin/bash
set -e

# Configuration
MONITORING_DIR="monitoring"
TEST_DURATION=600 # 10 minutes
LOAD_LEVELS=(100 500 1000) # Requests per second
METRICS_THRESHOLD=5000 # Maximum number of time series

# Function to generate load
generate_load() {
    local rps=$1
    local duration=$2
    
    echo "Generating load at ${rps} requests/second for ${duration} seconds..."
    
    # Use vegeta for load testing
    echo "GET http://localhost:8080/api/v1/words" | \
        vegeta attack -rate=$rps -duration=${duration}s | \
        vegeta report
}

# Function to check Prometheus performance
check_prometheus_performance() {
    echo "Checking Prometheus performance..."
    
    # Check number of time series
    local series_count=$(curl -s "http://localhost:9090/api/v1/query?query=prometheus_tsdb_head_series" | \
        jq -r '.data.result[0].value[1]')
    
    if [ "$series_count" -gt "$METRICS_THRESHOLD" ]; then
        echo "Warning: High number of time series: $series_count"
        return 1
    fi
    
    # Check query performance
    local query_duration=$(curl -s "http://localhost:9090/api/v1/query?query=rate(prometheus_engine_query_duration_seconds_sum[5m])" | \
        jq -r '.data.result[0].value[1]')
    
    if (( $(echo "$query_duration > 1" | bc -l) )); then
        echo "Warning: Slow query performance: ${query_duration}s average"
        return 1
    fi
    
    return 0
}

# Function to check Grafana performance
check_grafana_performance() {
    echo "Checking Grafana performance..."
    
    # Check response time
    local response_time=$(curl -s -w "%{time_total}" -o /dev/null "http://localhost:3000/api/health")
    
    if (( $(echo "$response_time > 1" | bc -l) )); then
        echo "Warning: Slow Grafana response time: ${response_time}s"
        return 1
    fi
    
    return 0
}

# Function to check resource usage
check_resource_usage() {
    echo "Checking resource usage..."
    
    # Check container stats
    local containers=("prometheus" "grafana" "alertmanager")
    
    for container in "${containers[@]}"; do
        local stats=$(docker stats --no-stream --format "{{.CPUPerc}},{{.MemPerc}}" "$container")
        local cpu_usage=$(echo "$stats" | cut -d',' -f1 | sed 's/%//')
        local mem_usage=$(echo "$stats" | cut -d',' -f2 | sed 's/%//')
        
        if (( $(echo "$cpu_usage > 80" | bc -l) )); then
            echo "Warning: High CPU usage in $container: ${cpu_usage}%"
            return 1
        fi
        
        if (( $(echo "$mem_usage > 80" | bc -l) )); then
            echo "Warning: High memory usage in $container: ${mem_usage}%"
            return 1
        fi
    done
    
    return 0
}

# Function to run performance test
run_performance_test() {
    local level=$1
    
    echo "Running performance test at ${level} RPS..."
    
    # Generate load
    generate_load "$level" "$TEST_DURATION"
    
    # Check monitoring stack performance
    check_prometheus_performance || return 1
    check_grafana_performance || return 1
    check_resource_usage || return 1
    
    return 0
}

# Main execution
echo "Starting monitoring performance tests..."

# Ensure monitoring stack is running
if ! curl -s "http://localhost:9090/-/healthy" > /dev/null; then
    echo "Error: Monitoring stack is not running"
    exit 1
fi

# Run tests for each load level
for level in "${LOAD_LEVELS[@]}"; do
    if ! run_performance_test "$level"; then
        echo "Performance test failed at ${level} RPS"
        exit 1
    fi
    
    # Wait between tests
    sleep 30
done

echo "Performance tests completed successfully!"
echo "Next steps:"
echo "1. Review performance metrics in Grafana"
echo "2. Adjust resource limits if needed" 