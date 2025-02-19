#!/bin/bash
set -e

# Configuration
MONITORING_DIR="monitoring"
ACTION="${1:-}"
MATCHER="${2:-}"
DURATION="${3:-}"
COMMENT="${4:-}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to list active alerts
list_alerts() {
    echo "Active alerts:"
    curl -s "http://localhost:9093/api/v2/alerts" | jq '.'
}

# Function to show alert history
show_history() {
    local days=${1:-7}
    echo "Alert history for the last ${days} days:"
    curl -s "http://localhost:9093/api/v2/alerts/groups" | \
        jq --arg days "$days" '.[] | select(.startsAt >= (now - ($days | tonumber * 86400)))'
}

# Function to silence alerts
silence_alerts() {
    local matcher="$1"
    local duration="$2"
    local comment="$3"
    
    # Create silence
    curl -X POST -H "Content-Type: application/json" \
        -d "{
            \"matchers\": [{\"name\": \"${matcher%%=*}\", \"value\": \"${matcher#*=}\"}],
            \"startsAt\": \"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\",
            \"endsAt\": \"$(date -u -d "+${duration}" +"%Y-%m-%dT%H:%M:%SZ")\",
            \"createdBy\": \"$(whoami)\",
            \"comment\": \"${comment}\"
        }" \
        "http://localhost:9093/api/v2/silences"
}

# Function to unsilence alerts
unsilence_alerts() {
    local silence_id="$1"
    
    # Delete silence
    curl -X DELETE "http://localhost:9093/api/v2/silence/${silence_id}"
}

# Main execution
case "$ACTION" in
    "list")
        list_alerts
        ;;
    "history")
        show_history "${MATCHER:-7}"
        ;;
    "silence")
        if [ -z "$MATCHER" ] || [ -z "$DURATION" ]; then
            echo "Usage: $0 silence <matcher> <duration> [comment]"
            echo "Example: $0 silence 'severity=critical' '2h' 'Maintenance window'"
            exit 1
        fi
        silence_alerts "$MATCHER" "$DURATION" "${COMMENT:-Silenced by script}"
        ;;
    "unsilence")
        if [ -z "$MATCHER" ]; then
            echo "Usage: $0 unsilence <silence_id>"
            exit 1
        fi
        unsilence_alerts "$MATCHER"
        ;;
    *)
        echo "Usage: $0 {list|history|silence|unsilence} [args...]"
        echo "  list                    - List active alerts"
        echo "  history [days]          - Show alert history"
        echo "  silence <matcher> <duration> [comment] - Silence alerts"
        echo "  unsilence <silence_id>  - Remove silence"
        exit 1
        ;;
esac 