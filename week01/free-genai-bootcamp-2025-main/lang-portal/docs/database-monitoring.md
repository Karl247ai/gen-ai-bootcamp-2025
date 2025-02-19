# Database Monitoring Guide

## Overview
This guide describes the database monitoring setup for the Lang Portal application.

### Metrics
1. Connection Pool Metrics
   - `db_connections_total`: Total connections in pool
   - `db_connections_in_use`: Active connections
   - `db_connection_wait_duration_seconds`: Connection wait time
   - `db_connection_timeouts_total`: Connection timeout count

2. Query Performance Metrics
   - `db_query_duration_seconds`: Query execution time
   - `db_query_errors_total`: Failed query count
   - `db_slow_queries_total`: Slow query count
   - `db_queries_total`: Query count by type

### Alerts
1. High Connection Usage
   - Triggers when pool usage > 80%
   - Check for connection leaks
   - Consider increasing pool size
   - Runbook: [High Connection Usage](runbooks/high_connection_usage.md)

2. Slow Queries
   - Triggers when p95 latency > 100ms
   - Review query plans
   - Check for missing indexes
   - Runbook: [Slow Queries](runbooks/slow_queries.md)

3. Query Errors
   - Triggers on any query errors
   - Check error logs
   - Verify database health
   - Runbook: [Query Errors](runbooks/query_errors.md)

4. Connection Timeouts
   - Triggers on connection timeouts
   - Check pool configuration
   - Monitor system resources
   - Runbook: [Connection Timeouts](runbooks/connection_timeouts.md)

5. High Slow Query Rate
   - Triggers when slow queries > 1/s
   - Analyze query patterns
   - Review indexing strategy
   - Runbook: [Slow Query Rate](runbooks/slow_query_rate.md)

6. Unbalanced Query Types
   - Triggers on unusual write/read ratio
   - Review application behavior
   - Check for query optimization
   - Runbook: [Query Ratio](runbooks/query_ratio.md)

### Dashboards
1. Connection Pool Dashboard
   - Pool usage gauge
   - Wait time trends
   - Error rates
   - Timeout tracking

2. Query Performance Dashboard
   - Latency percentiles
   - Error counts
   - Slow query tracking
   - Query type distribution

### Best Practices
1. Connection Management
   - Monitor pool usage
   - Set appropriate timeouts
   - Handle connection errors
   - Implement retry logic

2. Query Optimization
   - Use prepared statements
   - Maintain indexes
   - Monitor query plans
   - Batch operations

3. Error Handling
   - Log query errors
   - Set up alerts
   - Document recovery procedures
   - Implement circuit breakers

### Troubleshooting
1. High Connection Usage
   ```sql
   -- Check active connections
   SELECT count(*) FROM sqlite_master WHERE type='table';
   
   -- Identify long-running queries
   SELECT * FROM sqlite_master WHERE type='table' AND name LIKE 'sqlite_%';
   ```

2. Slow Queries
   ```sql
   -- Analyze query plan
   EXPLAIN QUERY PLAN SELECT * FROM words WHERE id = ?;
   
   -- Check index usage
   SELECT * FROM sqlite_master WHERE type='index';
   ```

3. Connection Timeouts
   ```bash
   # Check system resources
   top -b -n 1
   
   # Monitor file descriptors
   lsof -p $(pgrep lang-portal)
   ```

### Maintenance
1. Regular Tasks
   ```sql
   -- Analyze tables
   ANALYZE words;
   
   -- Update statistics
   ANALYZE sqlite_master;
   ```

2. Index Maintenance
   ```sql
   -- Rebuild indexes
   REINDEX words_idx;
   
   -- Check index fragmentation
   PRAGMA integrity_check;
   ```

3. Performance Tuning
   ```sql
   -- Set cache size
   PRAGMA cache_size = -2000; -- 2MB cache
   
   -- Enable WAL mode
   PRAGMA journal_mode = WAL;
   ```

### Recovery Procedures
1. Connection Pool Reset
   ```go
   // Graceful reset
   db.SetMaxOpenConns(0)
   time.Sleep(5 * time.Second)
   db.SetMaxOpenConns(10)
   ```

2. Query Optimization
   ```sql
   -- Add missing indexes
   CREATE INDEX IF NOT EXISTS words_status_idx ON words(status);
   
   -- Update statistics
   ANALYZE words;
   ```

3. Emergency Procedures
   ```bash
   # Backup database
   sqlite3 words.db ".backup 'words_backup.db'"
   
   # Restore from backup
   sqlite3 words.db ".restore 'words_backup.db'"
   ``` 