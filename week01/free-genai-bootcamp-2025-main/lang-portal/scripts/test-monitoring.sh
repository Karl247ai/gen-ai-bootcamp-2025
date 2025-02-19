#!/bin/bash
set -e

# Configuration
API_URL="http://localhost:8080"
PROMETHEUS_URL="http://localhost:9090"
GRAFANA_URL="http://localhost:3000"
TEST_DURATION=300 # 5 minutes

# Function to generate test load
generate_load() {
    local duration="$1"
    local end_time=$((SECONDS + duration))

    echo "Generating test load for $duration seconds..."
    while [ $SECONDS -lt $end_time ]; do
        # Make API requests
        curl -s "$API_URL/api/v1/words" > /dev/null
        curl -s "$API_URL/api/v1/words/1" > /dev/null
        curl -s "$API_URL/health" > /dev/null
        
        # Introduce some errors
        curl -s "$API_URL/api/v1/words/999999" > /dev/null
        
        sleep 0.1
    done
}

# Function to verify metrics
verify_metrics() {
    echo "Verifying metrics..."
    
    # Check request count
    local request_count=$(curl -s "$PROMETHEUS_URL/api/v1/query" --data-urlencode 'query=sum(handler_request_count)' | jq -r '.data.result[0].value[1]')
    if [ -z "$request_count" ] || [ "$request_count" = "0" ]; then
        echo "Error: No requests recorded"
        return 1
    fi

    # Check error rate
    local error_rate=$(curl -s "$PROMETHEUS_URL/api/v1/query" --data-urlencode 'query=sum(rate(handler_error_count[5m]))' | jq -r '.data.result[0].value[1]')
    if [ -z "$error_rate" ]; then
        echo "Error: No error metrics recorded"
        return 1
    fi

    echo "Metrics verification passed"
    return 0
}

# Main execution
echo "Starting monitoring test..."

# Check if services are running
if ! curl -s "$PROMETHEUS_URL/-/healthy" > /dev/null || \
   ! curl -s "$GRAFANA_URL/api/health" > /dev/null || \
   ! curl -s "$API_URL/health" > /dev/null; then
    echo "Error: One or more required services are not running"
    exit 1
fi

# Generate test load
generate_load "$TEST_DURATION"

# Wait for metrics to be collected
echo "Waiting for metrics collection..."
sleep 30

# Verify metrics
if verify_metrics; then
    echo "Monitoring test passed successfully!"
else
    echo "Monitoring test failed!"
    exit 1
fi 