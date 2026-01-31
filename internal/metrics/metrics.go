package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestsTotal counts total number of requests
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jellywolproxy_requests_total",
			Help: "Total number of requests processed",
		},
		[]string{"path", "method", "status"},
	)

	// WakeupAttemptsTotal counts wake-up attempts
	WakeupAttemptsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "jellywolproxy_wakeup_attempts_total",
			Help: "Total number of wake-up attempts",
		},
	)

	// WakeupSuccessTotal counts successful wake-ups
	WakeupSuccessTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "jellywolproxy_wakeup_success_total",
			Help: "Total number of successful wake-ups",
		},
	)

	// ServerStateGauge indicates if server is up (1) or down (0)
	ServerStateGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "jellywolproxy_server_state",
			Help: "Current server state (1 = up, 0 = down)",
		},
	)

	// RequestDurationHistogram measures request duration
	RequestDurationHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "jellywolproxy_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)

	// CacheHitsTotal counts cache hits
	CacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "jellywolproxy_cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	// CacheMissesTotal counts cache misses
	CacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "jellywolproxy_cache_misses_total",
			Help: "Total number of cache misses",
		},
	)
)
