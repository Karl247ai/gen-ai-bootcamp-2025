package database

import (
	"database/sql"
	"fmt"
	"time"
)

type RetryConfig struct {
	MaxAttempts int
	WaitTime    time.Duration
	MaxWaitTime time.Duration
}

// DefaultRetryConfig provides sensible defaults for retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 5,
		WaitTime:    time.Second,
		MaxWaitTime: 30 * time.Second,
	}
}

// WithRetry attempts to execute a database operation with retries
func WithRetry(db *sql.DB, retryConfig RetryConfig, operation func() error) error {
	var err error
	waitTime := retryConfig.WaitTime

	for attempt := 1; attempt <= retryConfig.MaxAttempts; attempt++ {
		err = operation()
		if err == nil {
			return nil
		}

		if attempt == retryConfig.MaxAttempts {
			break
		}

		// Wait before retrying
		time.Sleep(waitTime)

		// Exponential backoff with max wait time
		waitTime *= 2
		if waitTime > retryConfig.MaxWaitTime {
			waitTime = retryConfig.MaxWaitTime
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", retryConfig.MaxAttempts, err)
} 