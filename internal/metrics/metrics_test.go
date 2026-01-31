package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestMetricsRegistration(t *testing.T) {
	t.Run("RequestsTotal is registered", func(t *testing.T) {
		if RequestsTotal == nil {
			t.Error("RequestsTotal should not be nil")
		}
		RequestsTotal.WithLabelValues("/test", "GET", "200").Inc()
	})

	t.Run("WakeupAttemptsTotal is registered", func(t *testing.T) {
		if WakeupAttemptsTotal == nil {
			t.Error("WakeupAttemptsTotal should not be nil")
		}
		WakeupAttemptsTotal.Inc()
	})

	t.Run("WakeupSuccessTotal is registered", func(t *testing.T) {
		if WakeupSuccessTotal == nil {
			t.Error("WakeupSuccessTotal should not be nil")
		}
		WakeupSuccessTotal.Inc()
	})

	t.Run("ServerStateGauge is registered", func(t *testing.T) {
		if ServerStateGauge == nil {
			t.Error("ServerStateGauge should not be nil")
		}
		ServerStateGauge.Set(1)
		ServerStateGauge.Set(0)
	})

	t.Run("RequestDurationHistogram is registered", func(t *testing.T) {
		if RequestDurationHistogram == nil {
			t.Error("RequestDurationHistogram should not be nil")
		}
		RequestDurationHistogram.WithLabelValues("/test", "GET").Observe(0.5)
	})

	t.Run("CacheHitsTotal is registered", func(t *testing.T) {
		if CacheHitsTotal == nil {
			t.Error("CacheHitsTotal should not be nil")
		}
		CacheHitsTotal.Inc()
	})

	t.Run("CacheMissesTotal is registered", func(t *testing.T) {
		if CacheMissesTotal == nil {
			t.Error("CacheMissesTotal should not be nil")
		}
		CacheMissesTotal.Inc()
	})
}

func TestMetricsLabels(t *testing.T) {
	t.Run("RequestsTotal accepts correct labels", func(t *testing.T) {
		testCases := []struct {
			path   string
			method string
			status string
		}{
			{"/api/test", "GET", "200"},
			{"/api/test", "POST", "201"},
			{"/health", "GET", "200"},
			{"/videos/stream", "GET", "503"},
		}

		for _, tc := range testCases {
			// Should not panic
			RequestsTotal.WithLabelValues(tc.path, tc.method, tc.status).Inc()
		}
	})

	t.Run("RequestDurationHistogram accepts correct labels", func(t *testing.T) {
		testCases := []struct {
			path     string
			method   string
			duration float64
		}{
			{"/api/test", "GET", 0.001},
			{"/videos/stream", "GET", 1.5},
			{"/health", "GET", 0.0001},
		}

		for _, tc := range testCases {
			// Should not panic
			RequestDurationHistogram.WithLabelValues(tc.path, tc.method).Observe(tc.duration)
		}
	})
}

func TestMetricDescriptions(t *testing.T) {
	t.Run("RequestsTotal has description", func(t *testing.T) {
		desc := make(chan *prometheus.Desc, 1)
		RequestsTotal.Describe(desc)
		d := <-desc
		if d == nil {
			t.Error("RequestsTotal should have a description")
		}
	})

	t.Run("WakeupAttemptsTotal has description", func(t *testing.T) {
		desc := make(chan *prometheus.Desc, 1)
		WakeupAttemptsTotal.Describe(desc)
		d := <-desc
		if d == nil {
			t.Error("WakeupAttemptsTotal should have a description")
		}
	})

	t.Run("ServerStateGauge has description", func(t *testing.T) {
		desc := make(chan *prometheus.Desc, 1)
		ServerStateGauge.Describe(desc)
		d := <-desc
		if d == nil {
			t.Error("ServerStateGauge should have a description")
		}
	})
}
