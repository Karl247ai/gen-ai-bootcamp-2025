package testutil

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

// AlertRouter provides utilities for testing alert routing
type AlertRouter struct {
	t        *testing.T
	registry *prometheus.Registry
	routes   map[string][]string
}

// AlertRoute represents an alert routing rule
type AlertRoute struct {
	Team        string
	AlertName   string
	Severity    string
	Description string
}

// NewAlertRouter creates a new alert router
func NewAlertRouter(t *testing.T, registry *prometheus.Registry) *AlertRouter {
	return &AlertRouter{
		t:        t,
		registry: registry,
		routes:   make(map[string][]string),
	}
}

// AddRoute adds a routing rule
func (ar *AlertRouter) AddRoute(team, alertName string) {
	if ar.routes[team] == nil {
		ar.routes[team] = make([]string, 0)
	}
	ar.routes[team] = append(ar.routes[team], alertName)
}

// VerifyRouting verifies alert routing
func (ar *AlertRouter) VerifyRouting(alert string, expectedTeam string) bool {
	alerts := ar.routes[expectedTeam]
	for _, a := range alerts {
		if a == alert {
			return true
		}
	}
	return false
}

// WaitForRoute waits for an alert to be routed
func (ar *AlertRouter) WaitForRoute(alert, team string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ar.VerifyRouting(alert, team) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// GetTeamAlerts gets all alerts for a team
func (ar *AlertRouter) GetTeamAlerts(team string) []string {
	return ar.routes[team]
}

// GetAllRoutes gets all routing rules
func (ar *AlertRouter) GetAllRoutes() map[string][]string {
	return ar.routes
} 