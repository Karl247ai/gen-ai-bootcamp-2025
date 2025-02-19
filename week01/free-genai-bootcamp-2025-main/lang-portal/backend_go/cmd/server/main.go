package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/your-org/lang-portal/internal/api/handler"
	"github.com/your-org/lang-portal/internal/api/router"
	"github.com/your-org/lang-portal/internal/config"
	"github.com/your-org/lang-portal/internal/database"
	"github.com/your-org/lang-portal/internal/logger"
	"github.com/your-org/lang-portal/internal/middleware"
	"github.com/your-org/lang-portal/internal/metrics"
	"github.com/your-org/lang-portal/internal/repository/sqlite"
	"github.com/your-org/lang-portal/internal/service"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	migrationsPath := flag.String("migrations", "migrations", "path to migrations directory")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize monitoring
	metrics.Init(cfg.Monitoring)
	
	// Register maintenance metrics
	registerMaintenanceMetrics()

	// Initialize database with retry
	db, err := database.SetupDBWithRetry(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}
	defer database.CloseDB(db)

	// Initialize health checker
	healthChecker := database.NewHealthCheck(db)
	
	// Create root context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start periodic health checks
	healthChecker.StartPeriodicCheck(ctx)

	// Initialize logger
	log := logger.New(os.Stdout, logger.InfoLevel)
	ctx = log.WithContext(ctx)

	// Run migrations
	if err := database.Migrate(db, *migrationsPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	wordRepo := sqlite.NewWordRepository(db)
	groupRepo := sqlite.NewGroupRepository(db)
	wordGroupRepo := sqlite.NewWordGroupRepository(db)

	// Initialize services
	wordService := service.NewWordService(wordRepo)
	groupService := service.NewGroupService(groupRepo, wordGroupRepo, db)

	// Initialize handlers
	wordHandler := handler.NewWordHandler(wordService)
	groupHandler := handler.NewGroupHandler(groupService)

	// Initialize router
	r := router.NewRouter(wordHandler, groupHandler)

	// Set up Gin
	if cfg.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.Default()

	// Initialize monitoring components
	reg := prometheus.NewRegistry()
	metrics := metrics.NewMetrics(reg)

	// Initialize rate limiter with configuration
	rateLimiter := middleware.NewRateLimiter(
		cfg.Monitoring.RateLimit.RequestsPerMinute,
		cfg.Monitoring.RateLimit.Window,
		metrics,
	)

	// Initialize monitored database
	monitoredDB := database.NewMonitoredDB(db, metrics)

	// Initialize cache with monitoring
	var cache Cache
	if cfg.Monitoring.Cache.Enabled {
		baseCache := cache.NewInMemoryCache(cfg.Cache.Size)
		cache = cache.NewMonitoredCache(baseCache, metrics)
	}

	// Set up Gin middleware with monitoring
	engine.Use(middleware.RequestContext())
	engine.Use(middleware.Recovery())
	engine.Use(middleware.Tracing())
	engine.Use(middleware.Monitoring(metrics))
	engine.Use(middleware.ResourceMonitoring(metrics))
	engine.Use(rateLimiter.Middleware())

	// Setup monitoring endpoints with health checks
	if err := setupMonitoring(engine, cfg, metrics, monitoredDB, cache); err != nil {
		log.Fatalf("Failed to setup monitoring: %v", err)
	}

	// Initialize handlers with monitored components
	setupHandlers(engine, monitoredDB, cache, metrics)

	// Create server with monitoring-aware shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		BaseContext: func(l net.Listener) context.Context {
			ctx := context.Background()
			ctx = metrics.WithContext(ctx)
			return ctx
		},
	}

	// Graceful shutdown with monitoring cleanup
	go func() {
		<-quit
		log.Println("Shutting down server...")

		// Create shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Stop accepting new requests
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}

		// Cleanup monitoring resources
		if err := metrics.Shutdown(ctx); err != nil {
			log.Printf("Failed to cleanup metrics: %v", err)
		}

		// Close cache if enabled
		if cache != nil {
			if err := cache.Close(); err != nil {
				log.Printf("Failed to close cache: %v", err)
			}
		}

		close(done)
	}()
}

func setupMonitoring(engine *gin.Engine, cfg *config.Config, metrics *metrics.Metrics, db *database.MonitoredDB, cache Cache) error {
	// Prometheus metrics endpoint
	if cfg.Monitoring.Metrics.Enabled {
		engine.GET(cfg.Monitoring.Metrics.Path, gin.WrapH(promhttp.HandlerFor(
			metrics.Registry(),
			promhttp.HandlerOpts{
				ErrorHandling: promhttp.ContinueOnError,
			},
		)))
	}

	// Health check endpoint
	if cfg.Monitoring.Health.Enabled {
		engine.GET(cfg.Monitoring.Health.Path, func(c *gin.Context) {
			status := "healthy"
			details := make(map[string]string)
			errors := make([]string, 0)

			// Check database health
			if err := checkDatabaseHealth(c.Request.Context(), db); err != nil {
				status = "unhealthy"
				details["database"] = "error"
				errors = append(errors, fmt.Sprintf("database: %v", err))
			} else {
				details["database"] = "connected"
			}

			// Check cache if configured
			if cache != nil {
				if err := cache.Ping(c.Request.Context()); err != nil {
					status = "degraded"
					details["cache"] = "error"
					errors = append(errors, fmt.Sprintf("cache: %v", err))
				} else {
					details["cache"] = "available"
				}
			}

			// Check resource usage
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			if float64(memStats.Alloc)/float64(memStats.Sys) > 0.85 {
				status = "degraded"
				details["memory"] = "high"
				errors = append(errors, "memory usage above 85%")
			} else {
				details["memory"] = "normal"
			}

			response := gin.H{
				"status":    status,
				"details":   details,
				"timestamp": time.Now().UTC(),
			}
			if len(errors) > 0 {
				response["errors"] = errors
			}

			c.JSON(http.StatusOK, response)
		})
	}

	// Debug endpoints
	if cfg.Monitoring.Debug.Enabled {
		debug := engine.Group("/debug", middleware.AdminOnly())
		{
			// Database stats
			debug.GET("/db/stats", func(c *gin.Context) {
				stats := db.Stats()
				c.JSON(http.StatusOK, stats)
			})

			// Cache stats
			if cache != nil {
				debug.GET("/cache/stats", func(c *gin.Context) {
					stats := cache.Stats()
					c.JSON(http.StatusOK, stats)
				})
			}

			// Metrics reset
			debug.POST("/metrics/reset", func(c *gin.Context) {
				metrics.Reset()
				c.Status(http.StatusOK)
			})

			// Goroutine dump
			debug.GET("/goroutines", func(c *gin.Context) {
				buf := make([]byte, 2<<20)
				n := runtime.Stack(buf, true)
				c.Data(http.StatusOK, "text/plain", buf[:n])
			})
		}
	}

	return nil
}

func checkDatabaseHealth(ctx context.Context, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}

func setupHandlers(engine *gin.Engine, db *sql.DB, cache Cache, metrics *metrics.Metrics) {
	// Initialize repositories
	wordRepo := sqlite.NewWordRepository(db)
	groupRepo := sqlite.NewGroupRepository(db)
	wordGroupRepo := sqlite.NewWordGroupRepository(db)

	// Initialize services
	wordService := service.NewWordService(wordRepo)
	groupService := service.NewGroupService(groupRepo, wordGroupRepo, db)

	// Initialize handlers
	wordHandler := handler.NewWordHandler(wordService, metrics)
	groupHandler := handler.NewGroupHandler(groupService, metrics)

	// Set up API routes
	v1 := engine.Group("/api/v1")
	{
		// Word routes
		v1.POST("/words", wordHandler.CreateWord)
		v1.GET("/words/:id", wordHandler.GetWord)
		v1.GET("/words", wordHandler.ListWords)

		// Group routes
		v1.POST("/groups", groupHandler.CreateGroup)
		v1.GET("/groups/:id", groupHandler.GetGroup)
		v1.GET("/groups", groupHandler.ListGroups)
		v1.POST("/groups/:id/words", groupHandler.AddWordsToGroup)
	}

	// Metrics endpoint
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

func registerMaintenanceMetrics() {
	// Register maintenance metrics
	metrics.RegisterCounter("maintenance_total", "Total number of maintenance activities",
		[]string{"type"})
	
	metrics.RegisterHistogram("maintenance_duration_seconds", 
		"Duration of maintenance activities",
		[]string{"type"},
		prometheus.ExponentialBuckets(60, 2, 10))
	
	metrics.RegisterCounter("maintenance_errors_total", 
		"Total number of maintenance errors",
		[]string{"type"})
}

func recordMaintenance(maintenanceType string, duration float64, err error) {
	metrics.IncrementCounter("maintenance_total", maintenanceType)
	metrics.ObserveHistogram("maintenance_duration_seconds", duration, maintenanceType)
	
	if err != nil {
		metrics.IncrementCounter("maintenance_errors_total", maintenanceType)
	}
} 