package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Stoufiler/JellyWolProxy/internal/metrics"
)

// MetricsMiddleware tracks request metrics
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture the status code
		rw := &responseWriter{w, http.StatusOK}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		metrics.RequestsTotal.WithLabelValues(r.URL.Path, r.Method, strconv.Itoa(rw.statusCode)).Inc()
		metrics.RequestDurationHistogram.WithLabelValues(r.URL.Path, r.Method).Observe(duration)
	})
}

// responseWriter is a custom response writer that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
