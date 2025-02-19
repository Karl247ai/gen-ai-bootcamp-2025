#!/bin/bash
set -e

# Configuration
GRAFANA_URL="http://localhost:3000"
GRAFANA_API_KEY="${GRAFANA_API_KEY:-}"
EXPORT_DIR="monitoring/dashboards"

# Check if API key is set
if [ -z "$GRAFANA_API_KEY" ]; then
    echo "GRAFANA_API_KEY environment variable is required"
    exit 1
fi

# Create export directory
mkdir -p "$EXPORT_DIR"

# Get list of dashboards
echo "Fetching dashboard list..."
DASHBOARDS=$(curl -s -H "Authorization: Bearer $GRAFANA_API_KEY" \
    "${GRAFANA_URL}/api/search?type=dash-db")

# Export each dashboard
echo "$DASHBOARDS" | jq -r '.[] | .uid' | while read -r uid; do
    if [ ! -z "$uid" ]; then
        echo "Exporting dashboard: $uid"
        
        # Get dashboard JSON
        DASHBOARD=$(curl -s -H "Authorization: Bearer $GRAFANA_API_KEY" \
            "${GRAFANA_URL}/api/dashboards/uid/${uid}")
        
        # Get dashboard title
        TITLE=$(echo "$DASHBOARD" | jq -r '.dashboard.title' | tr ' ' '_' | tr '[:upper:]' '[:lower:]')
        
        # Save dashboard JSON
        echo "$DASHBOARD" | jq '.dashboard' > "${EXPORT_DIR}/${TITLE}.json"
        echo "Saved to ${EXPORT_DIR}/${TITLE}.json"
    fi
done

echo "Dashboard export completed!" 