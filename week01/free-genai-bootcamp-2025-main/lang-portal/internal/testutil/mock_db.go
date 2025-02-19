package testutil

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

type MockDB struct {
	*sql.DB
	mu    sync.RWMutex
	delay time.Duration
	err   error
}

func NewMockDB(realDB *sql.DB) *MockDB {
	return &MockDB{
		DB: realDB,
	}
}

func (m *MockDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	if m.err != nil {
		return nil, m.err
	}

	return m.DB.QueryContext(ctx, query, args...)
}

func (m *MockDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	if m.err != nil {
		return nil, m.err
	}

	return m.DB.ExecContext(ctx, query, args...)
}

func (m *MockDB) SetDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delay = delay
}

func (m *MockDB) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
} 