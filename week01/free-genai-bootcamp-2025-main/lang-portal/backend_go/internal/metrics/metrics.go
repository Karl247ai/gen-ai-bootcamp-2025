package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	counters map[string]prometheus.Counter
	timers   map[string]prometheus.Histogram
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		counters: make(map[string]prometheus.Counter),
		timers:   make(map[string]prometheus.Histogram),
	}

	// Register counters
	counters := []string{
		"handler.group.create.success",
		"handler.group.create.error",
		"handler.group.add_words.success",
		"handler.group.add_words.error",
		"db.transaction.success",
		"db.transaction.error",
		"handler.word.create.success",
		"handler.word.create.error",
		"handler.word.get.success",
		"handler.word.get.error",
		"handler.word.list.success",
		"handler.word.list.error",
	}

	for _, name := range counters {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: "Counter for " + name,
		})
		reg.MustRegister(counter)
		m.counters[name] = counter
	}

	// Register timers
	timers := []string{
		"handler.group.create",
		"handler.group.add_words",
		"db.transaction.duration",
		"handler.word.create",
		"handler.word.get",
		"handler.word.list",
	}

	for _, name := range timers {
		timer := prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    name,
			Help:    "Histogram for " + name,
			Buckets: prometheus.DefBuckets,
		})
		reg.MustRegister(timer)
		m.timers[name] = timer
	}

	return m
}

func (m *Metrics) IncCounter(name string) {
	if counter, ok := m.counters[name]; ok {
		counter.Inc()
	}
}

type Timer struct {
	histogram prometheus.Histogram
	start     time.Time
}

func (m *Metrics) NewTimer(name string) *Timer {
	if histogram, ok := m.timers[name]; ok {
		return &Timer{
			histogram: histogram,
			start:     time.Now(),
		}
	}
	return nil
}

func (t *Timer) ObserveDuration() {
	if t != nil {
		t.histogram.Observe(time.Since(t.start).Seconds())
	}
} 