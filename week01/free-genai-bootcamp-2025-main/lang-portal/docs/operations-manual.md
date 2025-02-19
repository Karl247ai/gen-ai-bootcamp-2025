# Operations Manual

## Deployment Procedures

### Environment Setup
1. System Requirements
   ```bash
   # Minimum server requirements
   CPU: 1 core
   RAM: 512MB
   Disk: 1GB
   OS: Linux (Ubuntu 20.04 or newer)
   ```

2. Required Software
   ```bash
   # Install required packages
   apt-get update
   apt-get install -y \
       sqlite3 \
       golang \
       nginx \
       supervisor
   ```

### Deployment Steps
1. Application Deployment
   ```bash
   # Create application directory
   mkdir -p /opt/lang-portal
   cd /opt/lang-portal
   
   # Deploy application binary
   cp ./bin/server /opt/lang-portal/
   chmod +x /opt/lang-portal/server
   
   # Configure environment
   cp .env.production /opt/lang-portal/.env
   ```

2. Database Setup
   ```bash
   # Initialize database
   ./server initdb
   
   # Run migrations
   ./server migrate
   
   # Import seed data (if needed)
   ./server seed
   ```

## Monitoring Setup

### System Monitoring
1. Metrics Collection
   ```yaml
   # Prometheus configuration
   scrape_configs:
     - job_name: 'lang-portal'
       scrape_interval: 15s
       static_configs:
         - targets: ['localhost:8080']
   ```

2. Alert Configuration
   ```yaml
   # Alert rules
   groups:
   - name: lang-portal
     rules:
     - alert: HighErrorRate
       expr: error_rate > 0.05
       for: 5m
       labels:
         severity: critical
     - alert: HighLatency
       expr: http_request_duration_seconds > 0.5
       for: 5m
       labels:
         severity: warning
   ```

### Application Monitoring
1. Log Management
   ```bash
   # Log locations
   Application: /var/log/lang-portal/app.log
   Access: /var/log/lang-portal/access.log
   Error: /var/log/lang-portal/error.log
   
   # Log rotation configuration
   /etc/logrotate.d/lang-portal:
   /var/log/lang-portal/*.log {
       daily
       rotate 14
       compress
       delaycompress
       notifempty
       create 0640 lang-portal lang-portal
   }
   ```

2. Health Checks
   ```bash
   # Health check endpoint
   curl -f http://localhost:8080/health
   
   # Expected response
   {
       "status": "healthy",
       "version": "1.0.0",
       "checks": {
           "database": "up",
           "api": "up"
       }
   }
   ```

## Backup & Recovery

### Backup Procedures
1. Database Backup
   ```bash
   #!/bin/bash
   # backup-db.sh
   
   BACKUP_DIR="/backup/lang-portal"
   TIMESTAMP=$(date +%Y%m%d_%H%M%S)
   
   # Create backup
   sqlite3 /opt/lang-portal/words.db ".backup '${BACKUP_DIR}/words_${TIMESTAMP}.db'"
   
   # Compress backup
   gzip "${BACKUP_DIR}/words_${TIMESTAMP}.db"
   
   # Cleanup old backups (keep last 30 days)
   find ${BACKUP_DIR} -name "words_*.db.gz" -mtime +30 -delete
   ```

2. Application State Backup
   ```bash
   #!/bin/bash
   # backup-app.sh
   
   BACKUP_DIR="/backup/lang-portal"
   TIMESTAMP=$(date +%Y%m%d_%H%M%S)
   
   # Backup configuration
   tar -czf "${BACKUP_DIR}/config_${TIMESTAMP}.tar.gz" /opt/lang-portal/config/
   
   # Backup logs
   tar -czf "${BACKUP_DIR}/logs_${TIMESTAMP}.tar.gz" /var/log/lang-portal/
   ```

### Recovery Procedures
1. Database Recovery
   ```bash
   #!/bin/bash
   # restore-db.sh
   
   BACKUP_FILE=$1
   
   # Stop application
   supervisorctl stop lang-portal
   
   # Restore database
   gunzip -c ${BACKUP_FILE} > /tmp/words.db
   sqlite3 /opt/lang-portal/words.db ".restore '/tmp/words.db'"
   
   # Start application
   supervisorctl start lang-portal
   ```

2. Application Recovery
   ```bash
   #!/bin/bash
   # disaster-recovery.sh
   
   # 1. Stop services
   supervisorctl stop all
   
   # 2. Restore configuration
   tar -xzf /backup/lang-portal/config_latest.tar.gz -C /
   
   # 3. Restore database
   ./restore-db.sh /backup/lang-portal/words_latest.db.gz
   
   # 4. Start services
   supervisorctl start all
   ```

## Security Measures

### Access Control
1. Firewall Configuration
   ```bash
   # Allow only necessary ports
   ufw allow 80/tcp
   ufw allow 443/tcp
   ufw allow 22/tcp
   
   # Enable firewall
   ufw enable
   ```

2. Rate Limiting (Nginx)
   ```nginx
   # /etc/nginx/conf.d/rate-limiting.conf
   limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
   
   location /api/ {
       limit_req zone=api_limit burst=20 nodelay;
       proxy_pass http://localhost:8080;
   }
   ```

### Security Headers
```nginx
# /etc/nginx/conf.d/security-headers.conf
add_header X-Content-Type-Options nosniff;
add_header X-Frame-Options DENY;
add_header X-XSS-Protection "1; mode=block";
add_header Content-Security-Policy "default-src 'self'";
add_header Strict-Transport-Security "max-age=31536000";
```

### SSL Configuration
```nginx
# /etc/nginx/conf.d/ssl.conf
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
ssl_prefer_server_ciphers on;
ssl_session_cache shared:SSL:10m;
ssl_session_timeout 10m;
```

# Monitoring Operations Manual

## Daily Operations

### Health Checks

1. **Service Status**
```bash
# Check all monitoring services
./scripts/check-monitoring-health.sh

# Health check output includes:
# - Service availability
# - Metrics collection
# - Alert manager status
# - Grafana dashboards
# - Disk space usage
# - Container status

# Individual checks
curl -s http://localhost:9090/-/healthy  # Prometheus
curl -s http://localhost:3000/api/health # Grafana
curl -s http://localhost:9093/-/healthy  # Alertmanager
```

**Interpreting Health Check Results**
- Green: Service is healthy
- Yellow: Warning condition
- Red: Critical condition

**Common Health Check Issues**
1. Service Unavailable
   - Check container logs
   - Verify network connectivity
   - Check resource usage

2. Metrics Collection Failed
   - Verify scrape configs
   - Check target endpoints
   - Review Prometheus logs

3. Disk Space Warning
   - Review retention settings
   - Clean old data
   - Consider scaling storage

2. **Data Collection**
- Verify metrics are being collected
- Check scrape targets in Prometheus
- Review error logs

3. **Alert Status**
```bash
# Check active alerts
./scripts/manage-alerts.sh list

# Review alert history
./scripts/manage-alerts.sh history
```

## Weekly Tasks

### Data Management

1. **Metric Rotation**
```bash
# Rotate old metrics
./scripts/rotate-metrics.sh

# Verify disk usage
df -h /var/lib/prometheus
```

2. **Performance Check**
```bash
# Run performance tests
./scripts/test-monitoring-performance.sh

# Review results in Grafana
```

3. **Backup**
```bash
# Create backup
./scripts/backup-metrics.sh

# Verify backup
ls -l /backup/monitoring/$(date +%Y%m%d)*
```

### Configuration Review

1. **Alert Rules**
- Review alert thresholds
- Check for noisy alerts
- Update based on patterns

2. **Dashboard Updates**
- Export dashboard changes
- Review dashboard usage
- Clean up unused panels

## Monthly Tasks

### Capacity Planning

1. **Resource Usage**
- Review storage growth
- Check query performance
- Monitor container resources

2. **Scaling Decisions**
- Evaluate retention periods
- Consider horizontal scaling
- Review resource limits

### Security

1. **Access Review**
- Audit Grafana users
- Review API keys
- Check access logs

2. **Updates**
```bash
# Update monitoring stack
./scripts/update-monitoring.sh

# Verify after update
./scripts/test-monitoring.sh
```

## Incident Response

### Alert Response

1. **High Error Rate**
```bash
# Check application logs
docker logs lang-portal-app

# Review error metrics
curl -s 'http://localhost:9090/api/v1/query?query=rate(handler_error_count[5m])'
```

2. **Resource Issues**
```bash
# Check resource usage
docker stats

# Review container logs
docker-compose logs prometheus grafana alertmanager
```

3. **Data Collection Issues**
```bash
# Check Prometheus targets
curl -s http://localhost:9090/api/v1/targets

# Verify scrape configs
cat monitoring/prometheus/config/prometheus.yml
```

### Recovery Procedures

1. **Service Recovery**
```bash
# Restart services
docker-compose -f monitoring/docker-compose.yml restart

# Verify recovery
./scripts/check-monitoring-health.sh
```

2. **Data Recovery**
```bash
# List available backups
ls -l /backup/monitoring/

# Restore from backup
./scripts/restore-metrics.sh <backup_timestamp>
```

3. **Configuration Recovery**
```bash
# Restore configs
git checkout monitoring/
./scripts/deploy-monitoring.sh

# Verify configuration
./scripts/validate-monitoring.sh
```

## Performance Tuning

### Query Optimization

1. **Recording Rules**
```yaml
# Add to prometheus/rules/recording.yml
groups:
  - name: api_metrics
    rules:
      - record: job:request_duration_seconds:p95
        expr: histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, job))
```

2. **Query Guidelines**
- Use rate() for counters
- Limit label combinations
- Use recording rules for complex queries

### Resource Optimization

1. **Storage**
```yaml
# Update retention in prometheus.yml
storage:
  tsdb:
    retention.time: 15d
    retention.size: 50GB
```

2. **Memory**
- Monitor heap usage
- Adjust query timeout
- Review container limits

## Maintenance Windows

### Maintenance Tracking

1. **Recording Maintenance**
```bash
# Start maintenance
./scripts/track-maintenance.sh start "Upgrading Prometheus to v2.44.0"

# End maintenance
./scripts/track-maintenance.sh end "Upgrade completed successfully"
```

2. **Viewing History**
```bash
# Show last 7 days
./scripts/track-maintenance.sh show

# Show specific period
./scripts/track-maintenance.sh show 30  # last 30 days
```

3. **Monthly Reports**
```bash
# Generate current month's report
./scripts/track-maintenance.sh report

# Generate specific month's report
./scripts/track-maintenance.sh report "2024-03"
```

### Planned Maintenance

1. **Pre-maintenance**
```bash
# Schedule maintenance window (duration in minutes)
./scripts/notify-maintenance.sh 60 planned

# Backup data
./scripts/backup-metrics.sh
```

**Maintenance Schedule Guidelines**
- Schedule during low-traffic periods
- Minimum 24-hour advance notice
- Maximum 2-hour window
- Coordinate with application maintenance
- Document all maintenance activities

2. **During Maintenance**
```bash
# Record maintenance start
./scripts/track-maintenance.sh start "Performing system updates"
echo "$(date -u +"%Y-%m-%dT%H:%M:%SZ") START" >> "${MONITORING_DIR}/maintenance.log"

# Stop services
docker-compose -f monitoring/docker-compose.yml down

# Perform updates
./scripts/update-monitoring.sh

# Record steps in log
echo "Performing: $ACTION" >> "${MONITORING_DIR}/maintenance.log"
```

3. **Post-maintenance**
```bash
# Start services
docker-compose -f monitoring/docker-compose.yml up -d

# Verify deployment
./scripts/test-monitoring.sh

# Record maintenance end
./scripts/track-maintenance.sh end "System updates completed"
echo "$(date -u +"%Y-%m-%dT%H:%M:%SZ") END" >> "${MONITORING_DIR}/maintenance.log"

# Send completion notification
./scripts/notify-maintenance.sh 0 completed
```

### Emergency Maintenance

1. **Quick Recovery**
```bash
# Notify emergency maintenance
./scripts/notify-maintenance.sh 30 emergency

# Stop affected service
docker-compose -f monitoring/docker-compose.yml stop <service>

# Clear corrupted data
./scripts/clean-metrics.sh <service>

# Restart service
docker-compose -f monitoring/docker-compose.yml up -d <service>
```

**Emergency Response Guidelines**
1. Assessment
   - Identify affected services
   - Estimate recovery time
   - Document incident details

2. Communication
   - Notify stakeholders immediately
   - Provide status updates every 15 minutes
   - Document all actions taken

3. Recovery
   - Follow recovery procedures
   - Verify service health
   - Update documentation 

## Monitoring System

### Components
- Prometheus: Metrics collection and storage
- Grafana: Visualization and dashboards
- AlertManager: Alert handling and routing

### Deployment
1. Deploy monitoring stack:
```bash
./scripts/deploy-monitoring.sh
```

2. Verify deployment:
- Check Prometheus: http://localhost:9090/targets
- Check Grafana: http://localhost:3000/dashboards
- Check AlertManager: http://localhost:9093/#/alerts

### Maintenance
1. Updating configurations:
```bash
# Update Prometheus config
vim monitoring/prometheus/prometheus.yml
docker-compose restart prometheus

# Update alert rules
vim monitoring/prometheus/rules/application.yml
docker-compose restart prometheus

# Update AlertManager config
vim monitoring/alertmanager/alertmanager.yml
docker-compose restart alertmanager
```

2. Backup procedures:
```bash
# Backup data volumes
./scripts/backup-monitoring.sh

# Restore from backup
./scripts/restore-monitoring.sh <backup-file>
```

### Troubleshooting
1. Check component health:
```bash
# Check component status
docker-compose ps

# View component logs
docker-compose logs prometheus
docker-compose logs grafana
docker-compose logs alertmanager
```

2. Common issues:
- High memory usage: Check retention periods
- Missing metrics: Verify scrape configs
- Alert delays: Check evaluation intervals 