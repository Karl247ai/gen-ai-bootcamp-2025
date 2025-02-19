package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/your-org/lang-portal/internal/config"
)

// SetupDB initializes and configures the database connection
func SetupDB(cfg *config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(time.Hour)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return db, nil
}

// CloseDB gracefully closes the database connection
func CloseDB(db *sql.DB) error {
	if err := db.Close(); err != nil {
		return fmt.Errorf("error closing database connection: %w", err)
	}
	return nil
}

func SetupDBWithRetry(cfg *config.DatabaseConfig) (*sql.DB, error) {
	var db *sql.DB
	retryConfig := DefaultRetryConfig()

	err := WithRetry(nil, retryConfig, func() error {
		var err error
		db, err = SetupDB(cfg)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to setup database with retry: %w", err)
	}

	return db, nil
} 