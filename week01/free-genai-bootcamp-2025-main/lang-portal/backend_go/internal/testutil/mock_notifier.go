package testutil

import (
	"sync"
	"time"
)

type Notification struct {
	AlertName string
	Message   string
	Level     string
	Timestamp time.Time
}

type MockNotifier struct {
	notifications []Notification
	mu           sync.RWMutex
}

func NewMockNotifier() *MockNotifier {
	return &MockNotifier{
		notifications: make([]Notification, 0),
	}
}

func (n *MockNotifier) Notify(alertName, message, level string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.notifications = append(n.notifications, Notification{
		AlertName: alertName,
		Message:   message,
		Level:     level,
		Timestamp: time.Now(),
	})
	return nil
}

func (n *MockNotifier) GetNotifications() []Notification {
	n.mu.RLock()
	defer n.mu.RUnlock()

	result := make([]Notification, len(n.notifications))
	copy(result, n.notifications)
	return result
}

func (n *MockNotifier) Clear() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.notifications = make([]Notification, 0)
} 