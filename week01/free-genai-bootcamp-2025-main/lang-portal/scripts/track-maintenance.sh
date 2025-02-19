#!/bin/bash
set -e

# Configuration
MONITORING_DIR="monitoring"
HISTORY_FILE="${MONITORING_DIR}/maintenance_history.jsonl"
ACTION="${1:-}"
DETAILS="${2:-}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to add maintenance record
add_maintenance_record() {
    local action=$1
    local details=$2
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local user=$(whoami)
    
    # Create history file if it doesn't exist
    if [ ! -f "$HISTORY_FILE" ]; then
        mkdir -p "$(dirname "$HISTORY_FILE")"
        touch "$HISTORY_FILE"
    }
    
    # Add record
    cat >> "$HISTORY_FILE" <<EOF
{"timestamp":"${timestamp}","action":"${action}","details":"${details}","user":"${user}"}
EOF
}

# Function to show maintenance history
show_maintenance_history() {
    local days=${1:-7}  # Default to last 7 days
    local since=$(date -d "-${days} days" +%s)
    
    echo "Maintenance history for the last ${days} days:"
    echo "--------------------------------------------"
    
    while IFS= read -r line; do
        local timestamp=$(echo "$line" | jq -r '.timestamp')
        local action=$(echo "$line" | jq -r '.action')
        local details=$(echo "$line" | jq -r '.details')
        local user=$(echo "$line" | jq -r '.user')
        
        # Convert timestamp to epoch for comparison
        local record_time=$(date -d "$timestamp" +%s)
        
        if [ $record_time -ge $since ]; then
            echo -e "${YELLOW}${timestamp}${NC}"
            echo "Action: $action"
            echo "Details: $details"
            echo "User: $user"
            echo "--------------------------------------------"
        fi
    done < "$HISTORY_FILE"
}

# Function to generate maintenance report
generate_maintenance_report() {
    local month=${1:-$(date +%Y-%m)}
    local report_file="${MONITORING_DIR}/reports/maintenance_${month}.md"
    local template_file="templates/maintenance_report.md.template"
    
    mkdir -p "$(dirname "$report_file")"
    
    # Read template
    if [ ! -f "$template_file" ]; then
        echo -e "${RED}Error: Template file not found: $template_file${NC}"
        exit 1
    fi
    
    # Calculate statistics
    local total_activities=0
    local planned_count=0
    local emergency_count=0
    local total_duration=0
    
    while IFS= read -r line; do
        local timestamp=$(echo "$line" | jq -r '.timestamp')
        if [[ $timestamp == ${month}* ]]; then
            total_activities=$((total_activities + 1))
            
            local details=$(echo "$line" | jq -r '.details')
            if [[ $details == *"emergency"* ]]; then
                emergency_count=$((emergency_count + 1))
            else
                planned_count=$((planned_count + 1))
            fi
            
            # Calculate duration if end record exists
            if [[ $(echo "$line" | jq -r '.action') == "end" ]]; then
                local start_time=$(echo "$line" | jq -r '.start_time')
                local end_time=$(echo "$line" | jq -r '.timestamp')
                local duration=$(($(date -d "$end_time" +%s) - $(date -d "$start_time" +%s)))
                total_duration=$((total_duration + duration))
            fi
        fi
    done < "$HISTORY_FILE"
    
    # Calculate averages
    local avg_duration=0
    if [ $total_activities -gt 0 ]; then
        avg_duration=$((total_duration / total_activities))
    fi
    
    # Generate report from template
    sed -e "s/{{MONTH}}/${month}/g" \
        -e "s/{{TIMESTAMP}}/$(date -u +"%Y-%m-%d %H:%M:%S UTC")/g" \
        -e "s/{{TOTAL_ACTIVITIES}}/${total_activities}/g" \
        -e "s/{{PLANNED_COUNT}}/${planned_count}/g" \
        -e "s/{{EMERGENCY_COUNT}}/${emergency_count}/g" \
        -e "s/{{AVG_DURATION}}/${avg_duration}/g" \
        -e "s/{{TOTAL_DOWNTIME}}/${total_duration}/g" \
        "$template_file" > "$report_file"
    
    # Add maintenance activities
    echo -e "\n## Detailed Activities\n" >> "$report_file"
    while IFS= read -r line; do
        local timestamp=$(echo "$line" | jq -r '.timestamp')
        if [[ $timestamp == ${month}* ]]; then
            local action=$(echo "$line" | jq -r '.action')
            local details=$(echo "$line" | jq -r '.details')
            local user=$(echo "$line" | jq -r '.user')
            
            cat >> "$report_file" <<EOF

### ${timestamp}
- Action: ${action}
- Details: ${details}
- Performed by: ${user}
EOF
        fi
    done < "$HISTORY_FILE"
    
    # Add performance comparison
    echo -e "\n## Performance Impact\n" >> "$report_file"
    if command -v curl &> /dev/null; then
        local query='rate(handler_request_duration_seconds_sum[1h])'
        local result=$(curl -s "http://localhost:9090/api/v1/query?query=${query}")
        echo "### Response Time Trend" >> "$report_file"
        echo "\`\`\`" >> "$report_file"
        echo "$result" | jq -r '.data.result[] | .metric.job + ": " + .value[1]' >> "$report_file"
        echo "\`\`\`" >> "$report_file"
    fi

    echo "Report generated: $report_file"
}

# Function to analyze maintenance patterns
analyze_maintenance_patterns() {
    local days=${1:-30}  # Default to last 30 days
    local since=$(date -d "-${days} days" +%s)
    
    echo "Analyzing maintenance patterns for the last ${days} days..."
    echo "--------------------------------------------"
    
    local total=0
    local emergency=0
    local planned=0
    local peak_hours=0
    declare -A weekday_count
    
    while IFS= read -r line; do
        local timestamp=$(echo "$line" | jq -r '.timestamp')
        local record_time=$(date -d "$timestamp" +%s)
        
        if [ $record_time -ge $since ]; then
            total=$((total + 1))
            
            # Count emergency vs planned
            local details=$(echo "$line" | jq -r '.details')
            if [[ $details == *"emergency"* ]]; then
                emergency=$((emergency + 1))
            else
                planned=$((planned + 1))
            fi
            
            # Track weekday distribution
            local weekday=$(date -d "$timestamp" +%u)  # 1-7, Monday is 1
            weekday_count[$weekday]=$((${weekday_count[$weekday]:-0} + 1))
            
            # Track hour distribution
            local hour=$(date -d "$timestamp" +%H)
            if [ $hour -ge 9 ] && [ $hour -le 17 ]; then
                peak_hours=$((peak_hours + 1))
            fi
        fi
    done < "$HISTORY_FILE"
    
    # Print analysis
    echo "Total maintenance activities: $total"
    echo "Emergency maintenance: $emergency ($(( emergency * 100 / total ))%)"
    echo "Planned maintenance: $planned ($(( planned * 100 / total ))%)"
    echo "Peak hours maintenance: $peak_hours ($(( peak_hours * 100 / total ))%)"
    echo
    echo "Weekday distribution:"
    for day in {1..7}; do
        echo "$(date -d "Monday +$((day-1)) days" +%A): ${weekday_count[$day]:-0}"
    done
}

# Main execution
case "$ACTION" in
    "start")
        if [ -z "$DETAILS" ]; then
            echo -e "${RED}Error: Maintenance details required${NC}"
            exit 1
        fi
        add_maintenance_record "start" "$DETAILS"
        echo -e "${GREEN}Maintenance start recorded${NC}"
        ;;
        
    "end")
        if [ -z "$DETAILS" ]; then
            DETAILS="Maintenance completed successfully"
        fi
        add_maintenance_record "end" "$DETAILS"
        echo -e "${GREEN}Maintenance end recorded${NC}"
        ;;
        
    "show")
        show_maintenance_history "${DETAILS:-7}"
        ;;
        
    "report")
        generate_maintenance_report "${DETAILS:-$(date +%Y-%m)}"
        ;;
        
    "analyze")
        analyze_maintenance_patterns "${DETAILS:-30}"
        ;;
        
    *)
        echo "Usage: $0 {start|end|show|report|analyze} [details]"
        echo "  start <details>  - Record maintenance start"
        echo "  end [details]    - Record maintenance end"
        echo "  show [days]      - Show maintenance history"
        echo "  report [month]   - Generate monthly report"
        echo "  analyze [days]   - Analyze maintenance patterns"
        exit 1
        ;;
esac 