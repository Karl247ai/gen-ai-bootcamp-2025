# Deployment Guide

## Monitoring Setup

### Prerequisites

1. **Environment Variables**
```bash
# Required
export SLACK_WEBHOOK_URL="your-webhook-url"
export ALERT_EMAIL="oncall@company.com"
export SMTP_USERNAME="alerts"
export SMTP_PASSWORD="your-smtp-password"

# Optional
export GRAFANA_ADMIN_PASSWORD="your-secure-password"
export PAGERDUTY_KEY="your-pagerduty-key"
```

2. **Network Requirements**
```
Ports:
- 9090: Prometheus
- 3000: Grafana
- 9093: Alertmanager
```

### Configuration

1. **Generate Configuration**
```bash
# Generate monitoring configuration
./scripts/generate-monitoring-config.sh
```

2. **Validate Configuration**
```bash
# Validate monitoring configuration
./scripts/validate-monitoring.sh

# Common validation errors:
# - Invalid alert rules
# - Port conflicts
# - Insufficient disk space
# - Missing environment variables
```

3. **Review Configuration**
```bash
# Review Prometheus config
cat monitoring/prometheus/config/prometheus.yml

# Review Alertmanager config
cat monitoring/alertmanager/config/alertmanager.yml

# Review Docker Compose
cat monitoring/docker-compose.yml
```

4. **Customize Alerts**
```yaml
# Edit monitoring/prometheus/alerts/application.yml
groups:
  - name: application
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          description: "High error rate detected"
```

### Deployment

1. **Deploy Stack**
```bash
# Validate configuration before deployment
./scripts/validate-monitoring.sh

# Deploy monitoring stack
./scripts/deploy-monitoring.sh
```

2. **Verify Deployment**
```bash
# Check service status
docker-compose -f monitoring/docker-compose.yml ps

# Check service logs
docker-compose -f monitoring/docker-compose.yml logs -f
```

3. **Access Services**
```
Prometheus: http://localhost:9090
Grafana: http://localhost:3000
Alertmanager: http://localhost:9093
```

### Maintenance

1. **Backup Data**
```bash
# Backup monitoring data
./scripts/backup-metrics.sh
```

2. **Rotate Data**
```bash
# Rotate old metrics data
./scripts/rotate-metrics.sh
```

3. **Update Configuration**
```bash
# Reload Prometheus config
curl -X POST http://localhost:9090/-/reload

# Reload Alertmanager config
curl -X POST http://localhost:9093/-/reload
```

### Troubleshooting

1. **Check Service Health**
```bash
# Check Prometheus
curl -s http://localhost:9090/-/healthy

# Check Alertmanager
curl -s http://localhost:9093/-/healthy

# Check Grafana
curl -s http://localhost:3000/api/health
```

2. **View Service Logs**
```bash
# View Prometheus logs
docker-compose -f monitoring/docker-compose.yml logs prometheus

# View Alertmanager logs
docker-compose -f monitoring/docker-compose.yml logs alertmanager

# View Grafana logs
docker-compose -f monitoring/docker-compose.yml logs grafana
```

3. **Common Issues**

- **No Data in Grafana**
  - Check Prometheus targets
  - Verify metrics endpoint
  - Check network connectivity

- **No Alerts**
  - Check alert rules
  - Verify Alertmanager config
  - Check notification channels

- **High Resource Usage**
  - Check data retention settings
  - Monitor disk usage
  - Review query performance 