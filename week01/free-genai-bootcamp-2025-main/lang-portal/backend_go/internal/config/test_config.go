package config

// TestConfig provides test configuration
type TestConfig struct {
	Monitoring MonitoringConfig
}

// NewTestConfig creates a new test configuration
func NewTestConfig() *TestConfig {
	return &TestConfig{
		Monitoring: MonitoringConfig{
			Metrics: MetricsConfig{
				Enabled: true,
			},
			Alerts: AlertConfig{
				Enabled: true,
			},
			Dashboards: DashboardConfig{
				Enabled: true,
			},
		},
	}
}

// MonitoringConfig for tests
type MonitoringConfig struct {
	Metrics    MetricsConfig
	Alerts     AlertConfig
	Dashboards DashboardConfig
}

// MetricsConfig for tests
type MetricsConfig struct {
	Enabled bool
}

// AlertConfig for tests
type AlertConfig struct {
	Enabled bool
}

// DashboardConfig for tests
type DashboardConfig struct {
	Enabled bool
} 