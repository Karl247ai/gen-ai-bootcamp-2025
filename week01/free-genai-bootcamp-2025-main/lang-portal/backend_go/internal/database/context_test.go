package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/your-org/lang-portal/internal/errors"
)

func TestDatabaseContext(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create test table
	_, err = db.Exec(`CREATE TABLE test (id INTEGER PRIMARY KEY, value TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	t.Run("QueryContext success", func(t *testing.T) {
		ctx := context.Background()
		rows, err := QueryContext(ctx, db, time.Second, "SELECT * FROM test")
		if err != nil {
			t.Errorf("Expected successful query, got error: %v", err)
		}
		defer rows.Close()
	})

	t.Run("ExecContext success", func(t *testing.T) {
		ctx := context.Background()
		result, err := ExecContext(ctx, db, time.Second,
			"INSERT INTO test (value) VALUES (?)", "test")
		if err != nil {
			t.Errorf("Expected successful exec, got error: %v", err)
		}

		affected, err := result.RowsAffected()
		if err != nil || affected != 1 {
			t.Errorf("Expected 1 row affected, got %d", affected)
		}
	})

	t.Run("TransactionContext success", func(t *testing.T) {
		ctx := context.Background()
		err := TransactionContext(ctx, db, func(tx *sql.Tx) error {
			_, err := tx.Exec("INSERT INTO test (value) VALUES (?)", "tx-test")
			return err
		})
		if err != nil {
			t.Errorf("Expected successful transaction, got error: %v", err)
		}
	})

	t.Run("TransactionContext rollback", func(t *testing.T) {
		ctx := context.Background()
		err := TransactionContext(ctx, db, func(tx *sql.Tx) error {
			return errors.New(errors.ErrDBQuery, "test error")
		})
		if !errors.IsErrorCode(err, errors.ErrDBQuery) {
			t.Errorf("Expected query error, got: %v", err)
		}
	})
} 