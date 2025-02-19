#!/bin/bash
set -e

# Configuration
MONITORING_DIR="monitoring"
HISTORY_FILE="${MONITORING_DIR}/maintenance_history.jsonl"
OUTPUT_DIR="${MONITORING_DIR}/reports/visualizations"
CHART_TYPE="${1:-timeline}"  # timeline, heatmap, or summary

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to generate timeline HTML
generate_timeline() {
    local output_file="${OUTPUT_DIR}/maintenance_timeline.html"
    mkdir -p "$(dirname "$output_file")"
    
    # Create HTML header
    cat > "$output_file" <<EOF
<!DOCTYPE html>
<html>
<head>
    <title>Maintenance Timeline</title>
    <script src="https://cdn.plot.ly/plotly-latest.min.js"></script>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        #timeline { width: 100%; height: 600px; }
    </style>
</head>
<body>
    <h1>Maintenance Timeline</h1>
    <div id="timeline"></div>
    <script>
        const data = {
            type: 'scatter',
            mode: 'markers',
            x: [],
            y: [],
            text: [],
            marker: { size: 10, color: [] }
        };
        
EOF
    
    # Add data points
    echo "data.x = [" >> "$output_file"
    echo "data.y = [" >> "$output_file"
    echo "data.text = [" >> "$output_file"
    echo "data.marker.color = [" >> "$output_file"
    
    while IFS= read -r line; do
        local timestamp=$(echo "$line" | jq -r '.timestamp')
        local action=$(echo "$line" | jq -r '.action')
        local details=$(echo "$line" | jq -r '.details')
        
        echo "\"$timestamp\"," >> "$output_file"
        echo "\"$action\"," >> "$output_file"
        echo "\"$details\"," >> "$output_file"
        if [[ $details == *"emergency"* ]]; then
            echo "'red'," >> "$output_file"
        else
            echo "'blue'," >> "$output_file"
        fi
    done < "$HISTORY_FILE"
    
    # Close arrays
    echo "];" >> "$output_file"
    echo "];" >> "$output_file"
    echo "];" >> "$output_file"
    echo "];" >> "$output_file"
    
    # Add plot configuration
    cat >> "$output_file" <<EOF
        const layout = {
            title: 'Maintenance Activities',
            xaxis: { title: 'Date' },
            yaxis: { title: 'Action' },
            hovermode: 'closest'
        };
        
        Plotly.newPlot('timeline', [data], layout);
    </script>
</body>
</html>
EOF
}

# Function to generate heatmap
generate_heatmap() {
    local output_file="${OUTPUT_DIR}/maintenance_heatmap.html"
    mkdir -p "$(dirname "$output_file")"
    
    # Initialize heatmap data
    declare -A heatmap
    for hour in {0..23}; do
        for day in {0..6}; do
            heatmap[$hour,$day]=0
        done
    done
    
    # Collect data
    while IFS= read -r line; do
        local timestamp=$(echo "$line" | jq -r '.timestamp')
        local hour=$(date -d "$timestamp" +%H)
        local day=$(date -d "$timestamp" +%w)  # 0-6, Sunday is 0
        heatmap[$hour,$day]=$((${heatmap[$hour,$day]} + 1))
    done < "$HISTORY_FILE"
    
    # Create HTML file
    cat > "$output_file" <<EOF
<!DOCTYPE html>
<html>
<head>
    <title>Maintenance Heatmap</title>
    <script src="https://cdn.plot.ly/plotly-latest.min.js"></script>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        #heatmap { width: 100%; height: 600px; }
    </style>
</head>
<body>
    <h1>Maintenance Activity Heatmap</h1>
    <div id="heatmap"></div>
    <script>
        const data = [{
            type: 'heatmap',
            z: [
EOF
    
    # Add heatmap data
    for hour in {0..23}; do
        echo -n "[" >> "$output_file"
        for day in {0..6}; do
            echo -n "${heatmap[$hour,$day]}," >> "$output_file"
        done
        echo "]," >> "$output_file"
    done
    
    # Add layout configuration
    cat >> "$output_file" <<EOF
            ],
            x: ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'],
            y: $(printf '["%02d:00",' {0..23})]
        }];
        
        const layout = {
            title: 'Maintenance Activity by Day and Hour',
            xaxis: { title: 'Day of Week' },
            yaxis: { title: 'Hour of Day' }
        };
        
        Plotly.newPlot('heatmap', data, layout);
    </script>
</body>
</html>
EOF
}

# Function to generate summary visualization
generate_summary() {
    local output_file="${OUTPUT_DIR}/maintenance_summary.html"
    mkdir -p "$(dirname "$output_file")"
    
    # Collect summary data
    local total=0
    local emergency=0
    local planned=0
    declare -A monthly_counts
    
    while IFS= read -r line; do
        total=$((total + 1))
        local timestamp=$(echo "$line" | jq -r '.timestamp')
        local month=$(date -d "$timestamp" +%Y-%m)
        local details=$(echo "$line" | jq -r '.details')
        
        monthly_counts[$month]=$((${monthly_counts[$month]:-0} + 1))
        
        if [[ $details == *"emergency"* ]]; then
            emergency=$((emergency + 1))
        else
            planned=$((planned + 1))
        fi
    done < "$HISTORY_FILE"
    
    # Create HTML file
    cat > "$output_file" <<EOF
<!DOCTYPE html>
<html>
<head>
    <title>Maintenance Summary</title>
    <script src="https://cdn.plot.ly/plotly-latest.min.js"></script>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .chart { width: 100%; height: 400px; margin-bottom: 20px; }
    </style>
</head>
<body>
    <h1>Maintenance Activity Summary</h1>
    <div id="pie" class="chart"></div>
    <div id="trend" class="chart"></div>
    <script>
        // Pie chart for maintenance types
        const pieData = [{
            type: 'pie',
            values: [$planned, $emergency],
            labels: ['Planned', 'Emergency'],
            hole: 0.4
        }];
        
        const pieLayout = {
            title: 'Maintenance Type Distribution'
        };
        
        Plotly.newPlot('pie', pieData, pieLayout);
        
        // Trend chart for monthly activities
        const trendData = [{
            type: 'scatter',
            mode: 'lines+markers',
            x: [$(for month in "${!monthly_counts[@]}"; do echo "\"$month\","; done)],
            y: [$(for count in "${monthly_counts[@]}"; do echo "$count,"; done)]
        }];
        
        const trendLayout = {
            title: 'Monthly Maintenance Trend',
            xaxis: { title: 'Month' },
            yaxis: { title: 'Number of Activities' }
        };
        
        Plotly.newPlot('trend', trendData, trendLayout);
    </script>
</body>
</html>
EOF
}

# Main execution
case "$CHART_TYPE" in
    "timeline")
        generate_timeline
        echo -e "${GREEN}Timeline visualization generated: ${OUTPUT_DIR}/maintenance_timeline.html${NC}"
        ;;
    "heatmap")
        generate_heatmap
        echo -e "${GREEN}Heatmap visualization generated: ${OUTPUT_DIR}/maintenance_heatmap.html${NC}"
        ;;
    "summary")
        generate_summary
        echo -e "${GREEN}Summary visualization generated: ${OUTPUT_DIR}/maintenance_summary.html${NC}"
        ;;
    *)
        echo "Usage: $0 {timeline|heatmap|summary}"
        echo "  timeline - Generate maintenance timeline"
        echo "  heatmap  - Generate activity heatmap"
        echo "  summary  - Generate summary charts"
        exit 1
        ;;
esac 