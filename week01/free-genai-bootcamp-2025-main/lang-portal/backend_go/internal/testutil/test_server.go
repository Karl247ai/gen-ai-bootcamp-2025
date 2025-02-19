package testutil

import (
	"math/rand"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/your-org/lang-portal/internal/config"
	"github.com/your-org/lang-portal/internal/metrics"
)

// TestServer represents a test server instance
type TestServer struct {
	Router   *gin.Engine
	Metrics  *metrics.Metrics
	Config   *config.Config
	Registry *prometheus.Registry
	Cleanup  func()
}

// SetupTestServer creates a new test server
func SetupTestServer(t *testing.T) *TestServer {
	// Use test mode
	gin.SetMode(gin.TestMode)

	// Load test configuration
	cfg := config.NewTestConfig()

	// Create registry
	registry := prometheus.NewRegistry()

	// Initialize metrics
	m := metrics.NewMetrics(cfg.Monitoring.Metrics, registry)

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())

	// Add metrics middleware
	router.Use(metrics.MetricsMiddleware(m))

	// Setup routes
	setupTestRoutes(router)

	// Add metrics endpoint
	router.GET("/metrics", metrics.PrometheusHandler(registry))

	return &TestServer{
		Router:   router,
		Metrics:  m,
		Config:   cfg,
		Registry: registry,
		Cleanup: func() {
			// Cleanup test resources
			registry.Unregister(m)
		},
	}
}

// setupTestRoutes adds test routes
func setupTestRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		// Word endpoints
		v1.GET("/words", func(c *gin.Context) {
			c.JSON(200, []string{"test"})
		})
		v1.GET("/words/:id", func(c *gin.Context) {
			id := c.Param("id")
			if id == "999" {
				c.JSON(404, gin.H{"error": "not found"})
				return
			}
			c.JSON(200, "test")
		})

		// Group endpoints
		v1.GET("/groups", func(c *gin.Context) {
			c.JSON(200, []string{"test"})
		})
		v1.GET("/groups/:id", func(c *gin.Context) {
			id := c.Param("id")
			if id == "999" {
				c.JSON(404, gin.H{"error": "not found"})
				return
			}
			c.JSON(200, "test")
		})

		// Health check
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "healthy"})
		})
	}
}

// WithSlowResponses adds artificial latency to responses
func (s *TestServer) WithSlowResponses(latency time.Duration) {
	s.Router.Use(func(c *gin.Context) {
		time.Sleep(latency)
		c.Next()
	})
}

// WithErrorRate adds artificial errors
func (s *TestServer) WithErrorRate(rate float64) {
	s.Router.Use(func(c *gin.Context) {
		if rand.Float64() < rate {
			c.AbortWithStatus(500)
			return
		}
		c.Next()
	})
} 