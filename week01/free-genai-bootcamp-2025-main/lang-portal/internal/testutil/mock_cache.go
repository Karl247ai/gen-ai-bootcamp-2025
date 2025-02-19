package testutil

import (
	"context"
	"sync"
	"time"
)

type MockCache struct {
	data  map[string]interface{}
	mu    sync.RWMutex
	err   error
	delay time.Duration
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]interface{}),
	}
}

func (m *MockCache) Get(ctx context.Context, key string) (interface{}, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	if m.err != nil {
		return nil, m.err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, ErrNotFound
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	if m.err != nil {
		return m.err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = value
	return nil
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	if m.err != nil {
		return m.err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}

func (m *MockCache) SetError(err error) {
	m.err = err
}

func (m *MockCache) SetDelay(delay time.Duration) {
	m.delay = delay
}

func (m *MockCache) Stats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"items":  len(m.data),
		"errors": m.err != nil,
		"delay":  m.delay.String(),
	}
} 