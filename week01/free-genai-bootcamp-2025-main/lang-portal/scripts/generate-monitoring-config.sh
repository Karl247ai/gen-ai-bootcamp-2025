#!/bin/bash
set -e

# Configuration
MONITORING_DIR="monitoring"
CONFIG_DIR="config"

# Function to generate Prometheus config
generate_prometheus_config() {
    echo "Generating Prometheus configuration..."
    
    cat > "${MONITORING_DIR}/prometheus/config/prometheus.yml" <<EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['alertmanager:9093']

rule_files:
  - "alerts/*.yml"

scrape_configs:
  - job_name: 'lang-portal'
    static_configs:
      - targets: ['app:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s
    scrape_timeout: 4s
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: lang-portal
EOF
}

# Function to generate Alertmanager config
generate_alertmanager_config() {
    echo "Generating Alertmanager configuration..."
    
    cat > "${MONITORING_DIR}/alertmanager/config/alertmanager.yml" <<EOF
global:
  resolve_timeout: 5m
  slack_api_url: '${SLACK_WEBHOOK_URL}'
  smtp_smarthost: '${SMTP_HOST:-smtp.company.com}:587'
  smtp_from: '${ALERT_EMAIL:-alerts@lang-portal.com}'
  smtp_auth_username: '${SMTP_USERNAME}'
  smtp_auth_password: '${SMTP_PASSWORD}'

route:
  group_by: ['alertname', 'severity']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'team-slack'
  routes:
    - match:
        severity: critical
      receiver: 'team-pager'
      repeat_interval: 5m
    - match:
        severity: warning
      receiver: 'team-slack'
      repeat_interval: 15m

receivers:
  - name: 'team-slack'
    slack_configs:
      - channel: '#alerts'
        send_resolved: true
        title: '{{ template "slack.title" . }}'
        text: '{{ template "slack.text" . }}'

  - name: 'team-pager'
    pagerduty_configs:
      - service_key: '${PAGERDUTY_KEY}'
        send_resolved: true

templates:
  - '/etc/alertmanager/templates/*.tmpl'
EOF
}

# Function to generate Docker Compose file
generate_docker_compose() {
    echo "Generating Docker Compose configuration..."
    
    cat > "${MONITORING_DIR}/docker-compose.yml" <<EOF
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:v${PROMETHEUS_VERSION}
    volumes:
      - ./prometheus/config:/etc/prometheus
      - ./prometheus/data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/usr/share/prometheus/console_libraries'
      - '--web.console.templates=/usr/share/prometheus/consoles'
      - '--storage.tsdb.retention.time=15d'
      - '--web.enable-lifecycle'
    ports:
      - "9090:9090"
    restart: unless-stopped
    networks:
      - monitoring

  grafana:
    image: grafana/grafana:${GRAFANA_VERSION}
    volumes:
      - ./grafana/data:/var/lib/grafana
      - ./grafana/config:/etc/grafana/provisioning
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD:-admin}
      - GF_USERS_ALLOW_SIGN_UP=false
      - GF_AUTH_ANONYMOUS_ENABLED=false
      - GF_INSTALL_PLUGINS=grafana-piechart-panel,grafana-worldmap-panel
    ports:
      - "3000:3000"
    depends_on:
      - prometheus
    restart: unless-stopped
    networks:
      - monitoring

  alertmanager:
    image: prom/alertmanager:v${ALERTMANAGER_VERSION}
    volumes:
      - ./alertmanager/config:/etc/alertmanager
      - ./alertmanager/data:/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
      - '--web.external-url=http://localhost:9093'
    ports:
      - "9093:9093"
    restart: unless-stopped
    networks:
      - monitoring

networks:
  monitoring:
    driver: bridge
EOF
}

# Function to generate alert templates
generate_alert_templates() {
    echo "Generating alert templates..."
    
    mkdir -p "${MONITORING_DIR}/alertmanager/config/templates"
    
    cat > "${MONITORING_DIR}/alertmanager/config/templates/slack.tmpl" <<EOF
{{ define "slack.title" }}
[{{ .Status | toUpper }}{{ if eq .Status "firing" }}:{{ .Alerts.Firing | len }}{{ end }}] {{ .CommonLabels.alertname }}
{{ end }}

{{ define "slack.text" }}
{{ range .Alerts }}
*Alert:* {{ .Labels.alertname }}{{ if .Labels.severity }} - {{ .Labels.severity }}{{ end }}
*Description:* {{ .Annotations.description }}
*Details:*
  {{ range .Labels.SortedPairs }}• {{ .Name }}: {{ .Value }}
  {{ end }}
{{ end }}
{{ end }}
EOF
}

# Main execution
echo "Starting monitoring configuration generation..."

# Create directory structure
mkdir -p "${MONITORING_DIR}"/{prometheus,grafana,alertmanager}/{config,data}

# Generate configurations
generate_prometheus_config
generate_alertmanager_config
generate_docker_compose
generate_alert_templates

echo "Monitoring configuration generated successfully!"
echo "Next steps:"
echo "1. Review generated configurations in ${MONITORING_DIR}"
echo "2. Set required environment variables"
echo "3. Run ./scripts/deploy-monitoring.sh" 