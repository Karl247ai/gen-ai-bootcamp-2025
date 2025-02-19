package middleware

import (
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/metrics"
)

func ResourceMonitoring(m *metrics.Metrics) gin.HandlerFunc {
	// Update resource metrics every 15 seconds
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		for range ticker.C {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			m.SetGauge("runtime.memory.alloc", float64(memStats.Alloc))
			m.SetGauge("runtime.memory.total_alloc", float64(memStats.TotalAlloc))
			m.SetGauge("runtime.memory.sys", float64(memStats.Sys))
			m.SetGauge("runtime.memory.heap_alloc", float64(memStats.HeapAlloc))
			m.SetGauge("runtime.memory.heap_sys", float64(memStats.HeapSys))
			m.SetGauge("runtime.memory.heap_idle", float64(memStats.HeapIdle))
			m.SetGauge("runtime.memory.heap_inuse", float64(memStats.HeapInuse))

			m.SetGauge("runtime.goroutines", float64(runtime.NumGoroutine()))
			m.SetGauge("runtime.num_gc", float64(memStats.NumGC))
			m.SetGauge("runtime.gc_pause_total", float64(memStats.PauseTotalNs))
		}
	}()

	return func(c *gin.Context) {
		c.Next()
	}
} 