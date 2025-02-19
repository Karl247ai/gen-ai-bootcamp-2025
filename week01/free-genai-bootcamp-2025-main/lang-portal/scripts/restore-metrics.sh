#!/bin/bash
set -e

# Configuration
BACKUP_DIR="/backup/monitoring"
BACKUP_TIMESTAMP=$1

if [ -z "$BACKUP_TIMESTAMP" ]; then
    echo "Usage: ./restore-metrics.sh <backup_timestamp>"
    echo "Available backups:"
    ls -1 "$BACKUP_DIR"
    exit 1
fi

BACKUP_PATH="${BACKUP_DIR}/${BACKUP_TIMESTAMP}"

if [ ! -d "$BACKUP_PATH" ]; then
    echo "Backup not found: $BACKUP_PATH"
    exit 1
fi

# Verify backup manifest
if [ ! -f "${BACKUP_PATH}/manifest.json" ]; then
    echo "Invalid backup: manifest.json not found"
    exit 1
fi

# Stop monitoring stack
echo "Stopping monitoring stack..."
cd monitoring
docker-compose down

# Restore Prometheus data
echo "Restoring Prometheus data..."
docker run --rm \
  -v prometheus_data:/prometheus \
  -v "${BACKUP_PATH}:/backup" \
  ubuntu bash -c "cd /prometheus && tar xzf /backup/prometheus.tar.gz --strip-components=1"

# Restore Grafana data
echo "Restoring Grafana data..."
docker run --rm \
  -v grafana_data:/grafana \
  -v "${BACKUP_PATH}:/backup" \
  ubuntu bash -c "cd /grafana && tar xzf /backup/grafana.tar.gz --strip-components=1"

# Restore Alertmanager data
echo "Restoring Alertmanager data..."
docker run --rm \
  -v alertmanager_data:/alertmanager \
  -v "${BACKUP_PATH}:/backup" \
  ubuntu bash -c "cd /alertmanager && tar xzf /backup/alertmanager.tar.gz --strip-components=1"

# Restore configurations
echo "Restoring configurations..."
cp -r "${BACKUP_PATH}/prometheus_config/"* monitoring/prometheus/config/
cp -r "${BACKUP_PATH}/grafana_config/"* monitoring/grafana/config/
cp -r "${BACKUP_PATH}/alertmanager_config/"* monitoring/alertmanager/config/

# Start monitoring stack
echo "Starting monitoring stack..."
docker-compose up -d

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 10

# Verify restoration
echo "Verifying restoration..."
PROMETHEUS_HEALTH=$(curl -s http://localhost:9090/-/healthy)
ALERTMANAGER_HEALTH=$(curl -s http://localhost:9093/-/healthy)
GRAFANA_HEALTH=$(curl -s http://localhost:3000/api/health)

if [[ "$PROMETHEUS_HEALTH" == "Prometheus is Healthy." && \
      "$ALERTMANAGER_HEALTH" == "Alertmanager is Healthy." && \
      "$GRAFANA_HEALTH" == *"\"database\":\"ok\""* ]]; then
    echo "Restoration completed successfully!"
else
    echo "Warning: One or more services may not be healthy. Please check the monitoring stack."
fi 