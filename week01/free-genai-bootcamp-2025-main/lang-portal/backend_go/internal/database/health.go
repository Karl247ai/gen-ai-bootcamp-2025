package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/your-org/lang-portal/internal/errors"
)

type HealthCheck struct {
	db           *sql.DB
	timeout      time.Duration
	maxIdleTime  time.Duration
	checkPeriod  time.Duration
}

func NewHealthCheck(db *sql.DB) *HealthCheck {
	return &HealthCheck{
		db:           db,
		timeout:      5 * time.Second,
		maxIdleTime:  time.Hour,
		checkPeriod:  time.Minute,
	}
}

func (h *HealthCheck) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		return errors.Wrap(err, errors.ErrDBConnection, "database health check failed")
	}

	// Check connection stats
	stats := h.db.Stats()
	if stats.MaxOpenConnections > 0 && stats.InUse == stats.MaxOpenConnections {
		return errors.New(errors.ErrDBConnection, "all connections are in use")
	}

	return nil
}

func (h *HealthCheck) StartPeriodicCheck(ctx context.Context) {
	ticker := time.NewTicker(h.checkPeriod)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if err := h.Check(ctx); err != nil {
					// Log the error using our logger
					logger.FromContext(ctx).Error("Database health check failed", 
						"error", err,
						"stats", h.db.Stats(),
					)
				}
			}
		}
	}()
} 