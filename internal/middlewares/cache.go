package middlewares

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Stoufiler/JellyWolProxy/internal/cache"
	"github.com/Stoufiler/JellyWolProxy/internal/dashboard"
	"github.com/sirupsen/logrus"
)

// cacheResponseWriter wrapper to capture response
type cacheResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func newCacheResponseWriter(w http.ResponseWriter) *cacheResponseWriter {
	return &cacheResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

func (rw *cacheResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *cacheResponseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

// CacheMiddleware provides HTTP response caching
func CacheMiddleware(logger *logrus.Logger, cache *cache.ResponseCache, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only cache GET requests
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		// Skip caching for certain paths
		if shouldSkipCache(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Generate cache key
		cacheKey := generateCacheKey(r)

		// Try to get from cache
		if entry, found := cache.Get(cacheKey); found {
			logger.Debugf("Cache hit for %s", r.URL.Path)
			dashboard.GetStats().RecordCacheHit()

			// Write cached headers
			for key, values := range entry.Headers {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}
			w.Header().Set("X-Cache", "HIT")

			// Write cached response
			_, _ = w.Write(entry.Data)
			return
		}

		logger.Debugf("Cache miss for %s", r.URL.Path)
		dashboard.GetStats().RecordCacheMiss()

		// Capture response
		rw := newCacheResponseWriter(w)
		next.ServeHTTP(rw, r)

		// Cache successful responses
		if rw.statusCode >= 200 && rw.statusCode < 300 {
			headers := make(map[string][]string)
			for key, values := range rw.Header() {
				headers[key] = values
			}
			cache.Set(cacheKey, rw.body.Bytes(), headers)
		}

		w.Header().Set("X-Cache", "MISS")
	})
}

// generateCacheKey creates a unique key for the request
func generateCacheKey(r *http.Request) string {
	hash := sha256.New()
	_, _ = io.WriteString(hash, r.Method)
	_, _ = io.WriteString(hash, r.URL.String())

	// Include certain headers in cache key
	if auth := r.Header.Get("Authorization"); auth != "" {
		_, _ = io.WriteString(hash, auth)
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

// shouldSkipCache determines if a path should bypass caching
func shouldSkipCache(path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/ping",
		"/websocket",
		"/socket.io",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	// Skip caching for streaming endpoints
	if strings.Contains(path, "/stream") || strings.Contains(path, ".m3u8") || strings.Contains(path, ".ts") {
		return true
	}

	return false
}
