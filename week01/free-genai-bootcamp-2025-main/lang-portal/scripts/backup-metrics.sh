#!/bin/bash
set -e

# Configuration
MONITORING_DIR="monitoring"
BACKUP_DIR="/backup/monitoring"
RETENTION_DAYS=30
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup Prometheus data
echo "Backing up Prometheus data..."
tar -czf "${BACKUP_DIR}/prometheus_${TIMESTAMP}.tar.gz" \
    "${MONITORING_DIR}/prometheus/data"

# Backup Grafana data
echo "Backing up Grafana data..."
tar -czf "${BACKUP_DIR}/grafana_${TIMESTAMP}.tar.gz" \
    "${MONITORING_DIR}/grafana/data"

# Backup configuration
echo "Backing up configuration..."
tar -czf "${BACKUP_DIR}/config_${TIMESTAMP}.tar.gz" \
    "${MONITORING_DIR}/prometheus/config" \
    "${MONITORING_DIR}/grafana/provisioning" \
    "${MONITORING_DIR}/alertmanager/config"

# Clean old backups
echo "Cleaning old backups..."
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +${RETENTION_DAYS} -delete

echo -e "${GREEN}Backup completed successfully!${NC}"
echo "Backup location: ${BACKUP_DIR}"
echo "Timestamp: ${TIMESTAMP}" 