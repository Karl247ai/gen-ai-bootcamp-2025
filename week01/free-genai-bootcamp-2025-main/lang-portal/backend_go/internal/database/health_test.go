package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/your-org/lang-portal/internal/errors"
	_ "github.com/mattn/go-sqlite3"
)

func TestHealthCheck(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	healthCheck := NewHealthCheck(db)

	t.Run("successful health check", func(t *testing.T) {
		ctx := context.Background()
		if err := healthCheck.Check(ctx); err != nil {
			t.Errorf("Expected successful health check, got error: %v", err)
		}
	})

	t.Run("failed health check - closed connection", func(t *testing.T) {
		db.Close()
		ctx := context.Background()
		err := healthCheck.Check(ctx)
		if !errors.IsErrorCode(err, errors.ErrDBConnection) {
			t.Errorf("Expected DB connection error, got: %v", err)
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(1 * time.Millisecond) // Ensure timeout occurs

		err := healthCheck.Check(ctx)
		if !errors.IsErrorCode(err, errors.ErrDBConnection) {
			t.Errorf("Expected timeout error, got: %v", err)
		}
	})
} 