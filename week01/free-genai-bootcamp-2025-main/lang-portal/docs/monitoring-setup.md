# Monitoring Setup Guide

## Overview

This guide describes how to set up and verify the monitoring stack for the Lang Portal application.

## Prerequisites

- Docker and Docker Compose
- jq command-line tool
- curl command-line tool

## Installation Steps

1. **Deploy Monitoring Stack**

```bash
# Deploy the monitoring stack
./scripts/deploy-monitoring.sh
```

2. **Verify Installation**

```bash
# Run monitoring setup test
./scripts/setup-monitoring-test.sh
```

3. **Configure Alert Channels**

```bash
# Set up Slack alerts
export SLACK_WEBHOOK_URL="your-webhook-url"

# Set up email alerts
export ALERT_EMAIL="oncall@your-company.com"
```

### Dashboard Setup

1. **Import Dashboards**
```bash
# Copy dashboard configurations
cp monitoring/grafana/dashboards/*.json /etc/grafana/provisioning/dashboards/

# Restart Grafana to load dashboards
docker-compose restart grafana
```

2. **Configure Data Sources**
```yaml
# /etc/grafana/provisioning/datasources/prometheus.yaml
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
```

3. **Dashboard Permissions**
```yaml
# /etc/grafana/provisioning/dashboards/provider.yaml
apiVersion: 1
providers:
  - name: 'Default'
    folder: 'General'
    type: file
    options:
      path: /etc/grafana/provisioning/dashboards
```

### Maintenance Configuration

1. **Initialize Tracking**
```bash
# Create maintenance directories
mkdir -p monitoring/maintenance
touch monitoring/maintenance_history.jsonl

# Set permissions
chmod 644 monitoring/maintenance_history.jsonl
```

2. **Configure Retention**
```bash
# Set retention period for maintenance records
export MAINTENANCE_RETENTION_DAYS=90

# Set backup retention
export BACKUP_RETENTION_DAYS=30
```

3. **Setup Notifications**
```bash
# Configure notification channels
export SLACK_WEBHOOK_URL="your-webhook-url"
export ALERT_EMAIL="oncall@your-company.com"
```

## Maintenance Tasks

### Backup and Restore

1. **Create Backup**
```bash
./scripts/backup-metrics.sh
```

2. **Restore from Backup**
```bash
./scripts/restore-metrics.sh <backup_timestamp>
```

### Data Rotation

1. **Rotate Old Data**
```bash
./scripts/rotate-metrics.sh
```

### Alert Management

1. **List Active Alerts**
```bash
./scripts/manage-alerts.sh list
```

2. **Silence an Alert**
```bash
./scripts/manage-alerts.sh silence 'alertname=HighErrorRate' 2h 'Investigating issue'
```

## Monitoring Tests

1. **Run Load Test**
```bash
./scripts/test-monitoring.sh
```

2. **Export Dashboards**
```bash
export GRAFANA_API_KEY="your-api-key"
./scripts/export-dashboards.sh
```

## Troubleshooting

### Common Issues

1. **Services Not Starting**
   - Check Docker logs: `docker-compose logs`
   - Verify configurations: `./scripts/setup-monitoring-test.sh`
   - Check disk space: `df -h`

2. **Missing Metrics**
   - Verify Prometheus targets: `http://localhost:9090/targets`
   - Check application metrics endpoint: `curl http://localhost:8080/metrics`
   - Review Prometheus configuration

3. **Alert Issues**
   - Check Alertmanager status: `http://localhost:9093/#/status`
   - Verify alert rules: `./scripts/manage-alerts.sh rules`
   - Check notification channels

### Health Checks

1. **Prometheus**
```bash
curl -s http://localhost:9090/-/healthy
```

2. **Alertmanager**
```bash
curl -s http://localhost:9093/-/healthy
```

3. **Grafana**
```bash
curl -s http://localhost:3000/api/health
```

## Best Practices

1. **Regular Maintenance**
   - Run backups daily
   - Rotate metrics weekly
   - Export dashboards monthly
   - Review alert rules quarterly

2. **Resource Management**
   - Monitor disk usage
   - Adjust retention periods
   - Clean up old data
   - Optimize storage

3. **Security**
   - Rotate API keys
   - Review access logs
   - Update passwords
   - Audit configurations

## Maintenance

### Data Management

1. **Backup**
```bash
# Backup all monitoring data
./scripts/backup-metrics.sh
```

2. **Cleanup**
```bash
# Clean specific service data
./scripts/clean-metrics.sh prometheus
./scripts/clean-metrics.sh grafana
./scripts/clean-metrics.sh alertmanager

# Clean all services
./scripts/clean-metrics.sh all
```

3. **Alert Management**
```bash
# List active alerts
./scripts/manage-alerts.sh list

# Show alert history
./scripts/manage-alerts.sh history 7

# Silence alerts during maintenance
./scripts/manage-alerts.sh silence 'severity=critical' '2h' 'Maintenance window'
```

### Maintenance Tracking

1. **Record Maintenance**
```bash
# Start maintenance
./scripts/track-maintenance.sh start "Upgrading Prometheus"

# End maintenance
./scripts/track-maintenance.sh end "Upgrade completed"
```

2. **View History**
```bash
# Show recent maintenance
./scripts/track-maintenance.sh show

# Generate report
./scripts/track-maintenance.sh report
```

3. **Visualize**
```bash
# Generate visualizations
./scripts/visualize-maintenance.sh timeline
./scripts/visualize-maintenance.sh heatmap
./scripts/visualize-maintenance.sh summary
``` 