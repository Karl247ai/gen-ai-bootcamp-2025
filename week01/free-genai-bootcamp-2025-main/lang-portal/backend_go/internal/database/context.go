package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/your-org/lang-portal/internal/errors"
)

// QueryContext executes a query with context and timeout
func QueryContext(ctx context.Context, db *sql.DB, timeout time.Duration, query string, args ...interface{}) (*sql.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDBQuery, "query execution failed")
	}
	return rows, nil
}

// ExecContext executes a statement with context and timeout
func ExecContext(ctx context.Context, db *sql.DB, timeout time.Duration, query string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDBQuery, "statement execution failed")
	}
	return result, nil
}

// TransactionContext executes a function within a transaction with context
func TransactionContext(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, errors.ErrDBTransaction, "failed to begin transaction")
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Wrap(rbErr, errors.ErrDBTransaction, "failed to rollback transaction")
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, errors.ErrDBTransaction, "failed to commit transaction")
	}

	return nil
} 