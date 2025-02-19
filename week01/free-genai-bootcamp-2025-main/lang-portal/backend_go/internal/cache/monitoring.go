package cache

import (
	"context"
	"time"

	"github.com/your-org/lang-portal/internal/metrics"
)

type MonitoredCache struct {
	Cache   Cache
	metrics *metrics.Metrics
}

func NewMonitoredCache(cache Cache, m *metrics.Metrics) *MonitoredCache {
	return &MonitoredCache{
		Cache:   cache,
		metrics: m,
	}
}

func (mc *MonitoredCache) Get(ctx context.Context, key string) (interface{}, error) {
	timer := mc.metrics.NewTimer("cache.get.duration")
	defer timer.ObserveDuration()

	value, err := mc.Cache.Get(ctx, key)
	if err != nil {
		if err == ErrNotFound {
			mc.metrics.IncCounter("cache.get.miss")
		} else {
			mc.metrics.IncCounter("cache.get.error")
		}
		return nil, err
	}

	mc.metrics.IncCounter("cache.get.hit")
	return value, nil
}

func (mc *MonitoredCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	timer := mc.metrics.NewTimer("cache.set.duration")
	defer timer.ObserveDuration()

	if err := mc.Cache.Set(ctx, key, value, ttl); err != nil {
		mc.metrics.IncCounter("cache.set.error")
		return err
	}

	mc.metrics.IncCounter("cache.set.success")
	return nil
}

func (mc *MonitoredCache) Delete(ctx context.Context, key string) error {
	timer := mc.metrics.NewTimer("cache.delete.duration")
	defer timer.ObserveDuration()

	if err := mc.Cache.Delete(ctx, key); err != nil {
		mc.metrics.IncCounter("cache.delete.error")
		return err
	}

	mc.metrics.IncCounter("cache.delete.success")
	return nil
} 