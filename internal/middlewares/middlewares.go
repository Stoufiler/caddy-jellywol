package middlewares

import (
	"bufio"
	"net"
	"net/http"

	"github.com/Stoufiler/JellyWolProxy/internal/dashboard"
	logger2 "github.com/Stoufiler/JellyWolProxy/internal/logger"
	"github.com/sirupsen/logrus"
)

// responseWriterWrapper wraps http.ResponseWriter to track bytes written
type responseWriterWrapper struct {
	http.ResponseWriter
	bytesWritten int64
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += int64(n)
	return n, err
}

// Hijack implements http.Hijacker interface for WebSocket support
func (w *responseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func RequestLoggerMiddleware(logger *logrus.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger2.LogRequest(logger, r)
		next.ServeHTTP(w, r)
	})
}

// NetworkStatsMiddleware tracks network bytes in/out
func NetworkStatsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Track request body size (bytes in)
		bytesIn := r.ContentLength
		if bytesIn < 0 {
			bytesIn = 0
		}

		// Wrap response writer to track bytes out
		wrapper := &responseWriterWrapper{ResponseWriter: w}
		next.ServeHTTP(wrapper, r)

		// Record bytes
		dashboard.GetStats().RecordBytes(bytesIn, wrapper.bytesWritten)
	})
}
