package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/your-org/lang-portal/internal/metrics"
)

type MonitoredDB struct {
	*sql.DB
	metrics *metrics.Metrics
}

func NewMonitoredDB(db *sql.DB, m *metrics.Metrics) *MonitoredDB {
	return &MonitoredDB{
		DB:      db,
		metrics: m,
	}
}

func (mdb *MonitoredDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	timer := mdb.metrics.NewTimer("db.query.duration")
	defer timer.ObserveDuration()

	rows, err := mdb.DB.QueryContext(ctx, query, args...)
	if err != nil {
		mdb.metrics.IncCounter("db.query.error")
		return nil, err
	}

	mdb.metrics.IncCounter("db.query.success")
	return rows, nil
}

func (mdb *MonitoredDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	timer := mdb.metrics.NewTimer("db.exec.duration")
	defer timer.ObserveDuration()

	result, err := mdb.DB.ExecContext(ctx, query, args...)
	if err != nil {
		mdb.metrics.IncCounter("db.exec.error")
		return nil, err
	}

	mdb.metrics.IncCounter("db.exec.success")
	return result, nil
}

func (mdb *MonitoredDB) Begin() (*sql.Tx, error) {
	timer := mdb.metrics.NewTimer("db.transaction.duration")
	defer timer.ObserveDuration()

	tx, err := mdb.DB.Begin()
	if err != nil {
		mdb.metrics.IncCounter("db.transaction.error")
		return nil, err
	}

	mdb.metrics.IncCounter("db.transaction.begin")
	return tx, nil
} 