package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/metrics"
)

type RateLimiter struct {
	requests map[string]*requestCounter
	mu       sync.RWMutex
	limit    int
	window   time.Duration
	metrics  *metrics.Metrics
}

type requestCounter struct {
	count    int
	start    time.Time
	lastSeen time.Time
}

func NewRateLimiter(limit int, window time.Duration, m *metrics.Metrics) *RateLimiter {
	limiter := &RateLimiter{
		requests: make(map[string]*requestCounter),
		limit:    limit,
		window:   window,
		metrics:  m,
	}

	// Clean up old entries periodically
	go limiter.cleanup()

	return limiter
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		counter, exists := rl.requests[ip]
		now := time.Now()

		if !exists {
			counter = &requestCounter{
				start:    now,
				lastSeen: now,
			}
			rl.requests[ip] = counter
		} else if now.Sub(counter.start) > rl.window {
			counter.count = 0
			counter.start = now
		}

		counter.count++
		counter.lastSeen = now
		count := counter.count
		rl.mu.Unlock()

		rl.metrics.SetGauge("ratelimit.requests",
			float64(count),
			"ip", ip,
		)

		if count > rl.limit {
			rl.metrics.IncCounter("ratelimit.exceeded",
				"ip", ip,
			)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, counter := range rl.requests {
			if now.Sub(counter.lastSeen) > rl.window*2 {
				delete(rl.requests, ip)
				rl.metrics.IncCounter("ratelimit.cleanup")
			}
		}
		rl.mu.Unlock()
	}
} 