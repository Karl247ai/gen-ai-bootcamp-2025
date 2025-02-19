#!/bin/bash
set -e

# Configuration
MONITORING_DIR="monitoring"
SLACK_WEBHOOK_URL="${SLACK_WEBHOOK_URL:-}"
ALERT_EMAIL="${ALERT_EMAIL:-}"
MAINTENANCE_WINDOW="${1:-30}"  # minutes
MAINTENANCE_TYPE="${2:-planned}"  # planned or emergency

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to send Slack notification
send_slack_notification() {
    local message=$1
    local color=$2

    if [ -z "$SLACK_WEBHOOK_URL" ]; then
        echo "Warning: SLACK_WEBHOOK_URL not set, skipping Slack notification"
        return 0
    fi

    curl -s -X POST -H 'Content-type: application/json' \
        --data "{
            \"attachments\": [
                {
                    \"color\": \"${color}\",
                    \"title\": \"Monitoring Maintenance Notice\",
                    \"text\": \"${message}\",
                    \"fields\": [
                        {
                            \"title\": \"Type\",
                            \"value\": \"${MAINTENANCE_TYPE}\",
                            \"short\": true
                        },
                        {
                            \"title\": \"Duration\",
                            \"value\": \"${MAINTENANCE_WINDOW} minutes\",
                            \"short\": true
                        }
                    ]
                }
            ]
        }" \
        "${SLACK_WEBHOOK_URL}"
}

# Function to send email notification
send_email_notification() {
    local message=$1

    if [ -z "$ALERT_EMAIL" ]; then
        echo "Warning: ALERT_EMAIL not set, skipping email notification"
        return 0
    }

    echo "$message" | mail -s "Monitoring Maintenance Notice" "$ALERT_EMAIL"
}

# Function to silence alerts during maintenance
silence_alerts() {
    local duration=$1

    echo "Silencing alerts for ${duration} minutes..."
    
    ./scripts/manage-alerts.sh silence \
        'alertname=~".+"' \
        "${duration}m" \
        "Maintenance window: ${MAINTENANCE_TYPE}"
}

# Function to create maintenance record
create_maintenance_record() {
    local start_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local end_time=$(date -u -d "+${MAINTENANCE_WINDOW} minutes" +"%Y-%m-%dT%H:%M:%SZ")
    
    cat > "${MONITORING_DIR}/maintenance.json" <<EOF
{
    "type": "${MAINTENANCE_TYPE}",
    "start_time": "${start_time}",
    "end_time": "${end_time}",
    "duration": "${MAINTENANCE_WINDOW}",
    "services": ["prometheus", "grafana", "alertmanager"]
}
EOF
}

# Main execution
echo "Preparing maintenance notifications..."

# Validate input
if [ "$MAINTENANCE_TYPE" != "planned" ] && [ "$MAINTENANCE_TYPE" != "emergency" ]; then
    echo -e "${RED}Error: Invalid maintenance type. Use 'planned' or 'emergency'${NC}"
    exit 1
fi

# Prepare notification message
if [ "$MAINTENANCE_TYPE" = "planned" ]; then
    color="good"
    message="Scheduled maintenance will begin in 15 minutes and last for ${MAINTENANCE_WINDOW} minutes. Monitoring services may be unavailable during this time."
else
    color="danger"
    message="Emergency maintenance is required and will begin immediately. Expected duration: ${MAINTENANCE_WINDOW} minutes. Monitoring services will be unavailable."
fi

# Send notifications
echo "Sending notifications..."
send_slack_notification "$message" "$color"
send_email_notification "$message"

# Create maintenance record
create_maintenance_record

# Silence alerts if this is planned maintenance
if [ "$MAINTENANCE_TYPE" = "planned" ]; then
    silence_alerts "$MAINTENANCE_WINDOW"
fi

echo -e "${GREEN}Maintenance notifications sent successfully!${NC}"
echo "Maintenance window: ${MAINTENANCE_WINDOW} minutes"
echo "Type: ${MAINTENANCE_TYPE}"
echo "Record created: ${MONITORING_DIR}/maintenance.json" 