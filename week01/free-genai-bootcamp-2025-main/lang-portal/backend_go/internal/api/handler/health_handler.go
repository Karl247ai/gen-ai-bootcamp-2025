package handler

import (
	"database/sql"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/api/response"
)

type HealthHandler struct {
	db        *sql.DB
	startTime time.Time
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		db:        db,
		startTime: time.Now(),
	}
}

type HealthResponse struct {
	Status    string         `json:"status"`
	Uptime    string         `json:"uptime"`
	DBStatus  string         `json:"db_status"`
	Memory    MemoryMetrics  `json:"memory"`
	GoRoutines int           `json:"goroutines"`
}

type MemoryMetrics struct {
	Alloc      uint64 `json:"alloc"`      // bytes allocated and still in use
	TotalAlloc uint64 `json:"totalAlloc"` // total bytes allocated (even if freed)
	Sys        uint64 `json:"sys"`        // total system memory obtained
	NumGC      uint32 `json:"numGC"`      // number of completed GC cycles
}

func (h *HealthHandler) Check(c *gin.Context) {
	// Check database connection
	dbStatus := "up"
	if err := h.db.Ping(); err != nil {
		dbStatus = "down"
	}

	// Get memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	healthResp := HealthResponse{
		Status:   "healthy",
		Uptime:   time.Since(h.startTime).String(),
		DBStatus: dbStatus,
		Memory: MemoryMetrics{
			Alloc:      memStats.Alloc,
			TotalAlloc: memStats.TotalAlloc,
			Sys:        memStats.Sys,
			NumGC:      memStats.NumGC,
		},
		GoRoutines: runtime.NumGoroutine(),
	}

	c.JSON(http.StatusOK, response.NewResponse(healthResp))
}

func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, response.NewResponse(map[string]string{
		"status": "alive",
	}))
}

func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// Check if database is accessible
	if err := h.db.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, response.NewErrorResponse(
			"SERVICE_UNAVAILABLE",
			"Database is not accessible",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, response.NewResponse(map[string]string{
		"status": "ready",
	}))
} 