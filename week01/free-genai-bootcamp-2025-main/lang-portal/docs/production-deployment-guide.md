# Production Deployment Guide

## Environment Setup

### Infrastructure Requirements
1. Server Specifications
   ```yaml
   Minimum Requirements:
     CPU: 2 cores
     RAM: 2GB
     Disk: 20GB SSD
     OS: Ubuntu 20.04 LTS
   
   Recommended:
     CPU: 4 cores
     RAM: 4GB
     Disk: 50GB SSD
     OS: Ubuntu 22.04 LTS
   ```

2. Network Configuration
   ```nginx
   # /etc/nginx/sites-available/lang-portal
   server {
       listen 80;
       server_name api.lang-portal.com;
       
       # Redirect HTTP to HTTPS
       return 301 https://$server_name$request_uri;
   }

   server {
       listen 443 ssl http2;
       server_name api.lang-portal.com;

       # SSL Configuration
       ssl_certificate /etc/letsencrypt/live/api.lang-portal.com/fullchain.pem;
       ssl_certificate_key /etc/letsencrypt/live/api.lang-portal.com/privkey.pem;
       
       # Proxy Configuration
       location / {
           proxy_pass http://localhost:8080;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
           proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
           proxy_set_header X-Forwarded-Proto $scheme;
       }
   }
   ```

## Configuration Management

### Environment Variables
```bash
# /opt/lang-portal/.env.production
# Server Configuration
GO_ENV=production
API_PORT=8080
API_HOST=0.0.0.0
ALLOWED_ORIGINS=https://app.lang-portal.com

# Database Configuration
DATABASE_PATH=/data/lang-portal/words.db
MAX_CONNECTIONS=50
QUERY_TIMEOUT=5s

# Cache Configuration
CACHE_SIZE=100MB
CACHE_TTL=15m

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
LOG_PATH=/var/log/lang-portal

# Security
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
MAX_REQUEST_SIZE=10MB
```

### Application Configuration
```yaml
# /opt/lang-portal/config/production.yaml
server:
  graceful_timeout: 30s
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s

database:
  max_idle_conns: 10
  max_open_conns: 50
  conn_max_lifetime: 1h

monitoring:
  metrics_enabled: true
  tracing_enabled: true
  health_check_interval: 30s
```

## Performance Monitoring Setup

### Metrics Collection
1. Install Prometheus:
```bash
helm install prometheus prometheus-community/prometheus
```

2. Configure scrape targets in `prometheus.yml`:
```yaml
scrape_configs:
  - job_name: 'lang-portal-api'
    scrape_interval: 15s
    static_configs:
      - targets: ['lang-portal-api:8080']
```

3. Install Grafana:
```bash
helm install grafana grafana/grafana
```

4. Import dashboards from `monitoring/dashboards/`.

### Alert Configuration
1. Create alert rules in Prometheus:
```yaml
groups:
  - name: lang-portal
    rules:
      - alert: HighErrorRate
        expr: rate(handler_error_count[5m]) > 0.01
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: High error rate detected
```

2. Configure alert manager:
```yaml
receivers:
  - name: 'team-alerts'
    slack_configs:
      - channel: '#alerts'
        send_resolved: true
```

## Performance Tuning

### System Tuning
```bash
# /etc/sysctl.conf
# Network tuning
net.core.somaxconn = 4096
net.ipv4.tcp_max_syn_backlog = 4096
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_time = 300

# File system tuning
fs.file-max = 100000
fs.inotify.max_user_watches = 524288

# Apply changes
sysctl -p
```

### Database Optimization
```sql
-- SQLite Performance Tuning
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = -2000; -- 2MB cache
PRAGMA temp_store = MEMORY;
PRAGMA mmap_size = 30000000000;
```

### Database Configuration
```ini
max_connections = 100
shared_buffers = 256MB
work_mem = 16MB
maintenance_work_mem = 64MB
effective_cache_size = 768MB
```

### Application Settings
```yaml
server:
  read_timeout: 5s
  write_timeout: 10s
  max_header_bytes: 1048576

database:
  max_open_conns: 50
  max_idle_conns: 10
  conn_max_lifetime: 1h

cache:
  size: 1000
  expiration: 5m
```

### Load Balancer Configuration
```nginx
upstream api_servers {
    least_conn;
    server api1:8080;
    server api2:8080;
    keepalive 32;
}

server {
    listen 80;
    location / {
        proxy_pass http://api_servers;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Request-ID $request_id;
    }
}
```

## Scaling Strategies

### Horizontal Scaling
```yaml
# Docker Compose configuration for multiple instances
version: '3.8'
services:
  api:
    image: lang-portal-api:1.0.0
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '1'
          memory: 1G
    environment:
      - GO_ENV=production
    volumes:
      - /data/lang-portal:/data
```

### Load Balancing
```nginx
# /etc/nginx/conf.d/upstream.conf
upstream api_servers {
    least_conn; # Load balancing method
    server 127.0.0.1:8081;
    server 127.0.0.1:8082;
    server 127.0.0.1:8083;
    
    keepalive 32;
}

# Health checks
match api_health {
    status 200;
    header Content-Type = application/json;
    body ~ '"status":"healthy"';
}
```

## Monitoring Setup

### Metrics Collection
```yaml
# /etc/prometheus/prometheus.yml
scrape_configs:
  - job_name: 'lang-portal'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scheme: 'http'
```

### Alert Configuration
```yaml
# /etc/prometheus/alerts.yml
groups:
- name: lang-portal-alerts
  rules:
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: High error rate detected
      description: Error rate is above 10% for 5 minutes
```

## Deployment Process

### Deployment Script
```bash
#!/bin/bash
# deploy.sh

set -e

# Configuration
APP_NAME="lang-portal"
DEPLOY_DIR="/opt/lang-portal"
BACKUP_DIR="/backup/lang-portal"
VERSION=$1

# Validate version
if [ -z "$VERSION" ]; then
    echo "Usage: ./deploy.sh <version>"
    exit 1
fi

# Create backup
echo "Creating backup..."
./backup.sh

# Stop service
echo "Stopping service..."
systemctl stop $APP_NAME

# Deploy new version
echo "Deploying version $VERSION..."
cp "bin/server-$VERSION" "$DEPLOY_DIR/server"
cp "config/production.yaml" "$DEPLOY_DIR/config/"

# Run migrations
echo "Running migrations..."
$DEPLOY_DIR/server migrate

# Start service
echo "Starting service..."
systemctl start $APP_NAME

# Health check
echo "Checking health..."
for i in {1..12}; do
    if curl -sf http://localhost:8080/health; then
        echo "Deployment successful!"
        exit 0
    fi
    sleep 5
done

echo "Deployment failed! Rolling back..."
./rollback.sh
exit 1
```

### Rollback Procedure
```bash
#!/bin/bash
# rollback.sh

set -e

# Configuration
APP_NAME="lang-portal"
DEPLOY_DIR="/opt/lang-portal"
BACKUP_DIR="/backup/lang-portal"

echo "Rolling back to previous version..."

# Stop service
systemctl stop $APP_NAME

# Restore from backup
cp "$BACKUP_DIR/server.backup" "$DEPLOY_DIR/server"
cp "$BACKUP_DIR/config.backup" "$DEPLOY_DIR/config/"

# Restore database if needed
sqlite3 "$DEPLOY_DIR/words.db" ".restore '$BACKUP_DIR/words.db.backup'"

# Start service
systemctl start $APP_NAME

echo "Rollback complete!"
```

## Security Hardening

### System Security
```bash
# /etc/security/limits.conf
# File descriptor limits
*       soft    nofile      65535
*       hard    nofile      65535

# Process limits
*       soft    nproc       32768
*       hard    nproc       32768
```

### Firewall Rules
```bash
# UFW Configuration
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable

# Rate limiting SSH
ufw limit ssh/tcp
```

### Security Scanning
```yaml
# Security scan schedule
security_checks:
  - name: "Dependencies audit"
    command: "go list -json -m all | nancy sleuth"
    frequency: "daily"
  
  - name: "Static code analysis"
    command: "gosec ./..."
    frequency: "on-deploy"
  
  - name: "Container scan"
    command: "trivy image lang-portal-api:latest"
    frequency: "weekly"
```

## Disaster Recovery Plan

### Backup Verification
```bash
#!/bin/bash
# verify-backup.sh

# Test environment setup
TEST_DIR="/tmp/backup-test"
mkdir -p $TEST_DIR

# Restore backup to test environment
sqlite3 "$TEST_DIR/words.db" ".restore '$BACKUP_DIR/words.db.backup'"

# Verify data integrity
echo "Verifying record counts..."
PROD_COUNT=$(sqlite3 /data/lang-portal/words.db "SELECT COUNT(*) FROM words;")
TEST_COUNT=$(sqlite3 $TEST_DIR/words.db "SELECT COUNT(*) FROM words;")

if [ "$PROD_COUNT" != "$TEST_COUNT" ]; then
    echo "Backup verification failed!"
    exit 1
fi
```

### Recovery Time Objectives
```yaml
# RTO and RPO targets
recovery_objectives:
  database:
    rto: 1h    # Recovery Time Objective
    rpo: 6h    # Recovery Point Objective
  application:
    rto: 30m
    rpo: 1h
  full_system:
    rto: 2h
    rpo: 12h
```

## Capacity Planning

### Resource Monitoring
```yaml
# Resource thresholds and alerts
thresholds:
  cpu_usage: 80%
  memory_usage: 85%
  disk_usage: 75%
  connection_pool: 80%
  
scaling_triggers:
  - metric: "cpu_usage"
    threshold: 75%
    duration: "5m"
    action: "scale_up"
    
  - metric: "response_time"
    threshold: 500ms
    duration: "10m"
    action: "scale_up"
```

### Growth Planning
```yaml
# Capacity growth estimates
growth_projections:
  users_monthly_growth: 20%
  data_growth_monthly: 5GB
  traffic_growth_quarterly: 30%

resource_planning:
  storage:
    current: 20GB
    buffer: 40%
    review_threshold: 70%
  
  memory:
    current: 4GB
    buffer: 50%
    review_threshold: 75%
```

## Service Level Agreements (SLA)

### Performance SLAs
```yaml
sla_targets:
  availability:
    target: 99.9%
    measurement_window: 30d
    
  response_time:
    p95: 200ms
    p99: 500ms
    measurement_window: 1h
    
  error_rate:
    target: < 0.1%
    measurement_window: 1h
```

### Maintenance Windows
```yaml
maintenance_windows:
  scheduled:
    frequency: "monthly"
    day: "last Sunday"
    time: "02:00 UTC"
    duration: "2h"
    notification_lead_time: "7d"
  
  emergency:
    max_duration: "4h"
    notification_lead_time: "1h"
```

## Documentation Requirements

### Operational Documentation
```yaml
required_documentation:
  runbooks:
    - incident_response
    - backup_recovery
    - deployment_procedures
    - scaling_procedures
    
  monitoring:
    - metrics_description
    - alert_responses
    - dashboard_guide
    
  maintenance:
    - update_procedures
    - cleanup_procedures
    - optimization_guide
```

### Incident Response
```yaml
incident_severity_levels:
  sev1:
    description: "Complete service outage"
    response_time: "15m"
    update_frequency: "30m"
    
  sev2:
    description: "Partial service degradation"
    response_time: "30m"
    update_frequency: "2h"
    
  sev3:
    description: "Minor issue, no service impact"
    response_time: "4h"
    update_frequency: "24h"
```

## Performance Testing

### Load Test Scenarios
1. Run performance tests:
```bash
go test -v ./test/integration -run=TestPerformance
```

2. Monitor results in Grafana dashboard.

3. Analyze metrics:
- Response times (p95, p99)
- Error rates
- Resource utilization
- Database connection pool stats

### Performance Baselines
- API Response Time: < 200ms (p95)
- Error Rate: < 0.1%
- Throughput: > 100 req/s per instance
- CPU Usage: < 70%
- Memory Usage: < 80% 

## Prerequisites

- Docker and Docker Compose
- Access to production environment
- Required environment variables

## Deployment Steps

1. **Application Deployment**
```bash
# Deploy application
./scripts/deploy.sh
```

2. **Monitoring Stack**
```bash
# Deploy monitoring services
./scripts/deploy-monitoring.sh

# Verify deployment
./scripts/check-monitoring-health.sh
```

3. **Configure Alerts**
```bash
# Review and update alert rules
vim monitoring/prometheus/config/alerts/

# Validate configuration
./scripts/validate-monitoring.sh
```

4. **Setup Maintenance**
```bash
# Initialize maintenance tracking
mkdir -p monitoring/maintenance
touch monitoring/maintenance_history.jsonl

# Configure backup retention
export RETENTION_DAYS=30
```

## Monitoring Configuration

1. **Resource Limits**
```yaml
monitoring:
  resource:
    memory_threshold: 85
    goroutine_threshold: 10000
  rate_limit:
    requests_per_minute: 1000
  cache:
    hit_rate_threshold: 0.8
```

2. **Alert Thresholds**
```yaml
alerts:
  error_rate: 0.05
  response_time: 500ms
  connection_usage: 0.8
```

## Health Checks

1. **Application Health**
```bash
curl http://your-app/health
```

2. **Monitoring Health**
```bash
curl http://your-app/metrics
```

## Scaling Guidelines

1. **Vertical Scaling**
- Monitor memory usage
- Adjust resource limits
- Update connection pools

2. **Horizontal Scaling**
- Deploy multiple instances
- Configure load balancer
- Update monitoring targets

## Backup and Recovery

1. **Metrics Data**
```bash
# Backup Prometheus data
./scripts/backup-metrics.sh

# Restore from backup
./scripts/restore-metrics.sh
```

2. **Dashboard Configuration**
```bash
# Export dashboards
./scripts/export-dashboards.sh

# Import dashboards
./scripts/import-dashboards.sh
```

## Troubleshooting

1. **Common Issues**
- Check application logs
- Verify monitoring endpoints
- Review alert history

2. **Performance Issues**
- Review monitoring dashboards
- Check resource usage
- Analyze slow queries

3. **Alert Storm Handling**
- Group related alerts
- Adjust thresholds
- Update notification settings 

## Maintenance

### Monitoring Maintenance

1. **Backup**
```bash
# Backup monitoring data
./scripts/backup-metrics.sh
```

2. **Cleanup**
```bash
# Clean old metrics
./scripts/clean-metrics.sh all
```

3. **Updates**
```bash
# Record maintenance
./scripts/track-maintenance.sh start "Updating monitoring stack"

# Perform updates
docker-compose -f monitoring/docker-compose.yml pull
docker-compose -f monitoring/docker-compose.yml up -d

# Record completion
./scripts/track-maintenance.sh end "Update completed"
``` 