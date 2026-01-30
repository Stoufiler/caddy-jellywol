package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Stoufiler/JellyWolProxy/internal/cache"
	"github.com/sirupsen/logrus"
)

func TestRequestLoggerMiddleware(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Don't show logs during tests

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestLoggerMiddleware(logger, handler)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestMetricsMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := MetricsMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestMetricsMiddlewareWithStatusCode(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedStatus int
	}{
		{"status 200", http.StatusOK, http.StatusOK},
		{"status 404", http.StatusNotFound, http.StatusNotFound},
		{"status 500", http.StatusInternalServerError, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			middleware := MetricsMiddleware(handler)

			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			middleware.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestCacheMiddleware(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	responseCache := cache.NewResponseCache(5 * time.Minute)

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"test"}`))
	})

	middleware := CacheMiddleware(logger, responseCache, handler)

	// First request - should call handler and cache response
	req1 := httptest.NewRequest("GET", "/api/test", nil)
	rr1 := httptest.NewRecorder()
	middleware.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr1.Code)
	}
	if rr1.Header().Get("X-Cache") != "MISS" {
		t.Errorf("expected X-Cache header to be MISS, got %s", rr1.Header().Get("X-Cache"))
	}
	if callCount != 1 {
		t.Errorf("expected handler to be called once, was called %d times", callCount)
	}

	// Second request - should use cached response
	req2 := httptest.NewRequest("GET", "/api/test", nil)
	rr2 := httptest.NewRecorder()
	middleware.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr2.Code)
	}
	if rr2.Header().Get("X-Cache") != "HIT" {
		t.Errorf("expected X-Cache header to be HIT, got %s", rr2.Header().Get("X-Cache"))
	}
	if callCount != 1 {
		t.Errorf("expected handler to still be called once, was called %d times", callCount)
	}
}

func TestCacheMiddleware_POSTNotCached(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	responseCache := cache.NewResponseCache(5 * time.Minute)

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})

	middleware := CacheMiddleware(logger, responseCache, handler)

	// POST request should not be cached
	req1 := httptest.NewRequest("POST", "/api/test", nil)
	rr1 := httptest.NewRecorder()
	middleware.ServeHTTP(rr1, req1)

	req2 := httptest.NewRequest("POST", "/api/test", nil)
	rr2 := httptest.NewRecorder()
	middleware.ServeHTTP(rr2, req2)

	if callCount != 2 {
		t.Errorf("expected handler to be called twice for POST requests, was called %d times", callCount)
	}
}

func TestCacheMiddleware_SkipHealthEndpoints(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	responseCache := cache.NewResponseCache(5 * time.Minute)

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	middleware := CacheMiddleware(logger, responseCache, handler)

	skipPaths := []string{"/health", "/metrics", "/ping"}

	for _, path := range skipPaths {
		req1 := httptest.NewRequest("GET", path, nil)
		rr1 := httptest.NewRecorder()
		middleware.ServeHTTP(rr1, req1)

		req2 := httptest.NewRequest("GET", path, nil)
		rr2 := httptest.NewRecorder()
		middleware.ServeHTTP(rr2, req2)
	}

	// Each path should be called twice (not cached)
	expectedCalls := len(skipPaths) * 2
	if callCount != expectedCalls {
		t.Errorf("expected handler to be called %d times for skip paths, was called %d times", expectedCalls, callCount)
	}
}

func TestCacheMiddleware_SkipStreamingEndpoints(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	responseCache := cache.NewResponseCache(5 * time.Minute)

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	middleware := CacheMiddleware(logger, responseCache, handler)

	streamPaths := []string{"/videos/stream", "/content/main.m3u8"}

	for _, path := range streamPaths {
		req1 := httptest.NewRequest("GET", path, nil)
		rr1 := httptest.NewRecorder()
		middleware.ServeHTTP(rr1, req1)

		req2 := httptest.NewRequest("GET", path, nil)
		rr2 := httptest.NewRecorder()
		middleware.ServeHTTP(rr2, req2)
	}

	// Each streaming path should be called twice (not cached)
	expectedCalls := len(streamPaths) * 2
	if callCount != expectedCalls {
		t.Errorf("expected handler to be called %d times for streaming paths, was called %d times", expectedCalls, callCount)
	}
}

func TestCacheMiddleware_ErrorStatusNotCached(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	responseCache := cache.NewResponseCache(5 * time.Minute)

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error"))
	})

	middleware := CacheMiddleware(logger, responseCache, handler)

	// First request with error
	req1 := httptest.NewRequest("GET", "/api/test", nil)
	rr1 := httptest.NewRecorder()
	middleware.ServeHTTP(rr1, req1)

	// Second request - error responses should not be cached
	req2 := httptest.NewRequest("GET", "/api/test", nil)
	rr2 := httptest.NewRecorder()
	middleware.ServeHTTP(rr2, req2)

	if callCount != 2 {
		t.Errorf("expected handler to be called twice for error responses, was called %d times", callCount)
	}
}

func TestGenerateCacheKey(t *testing.T) {
	req1 := httptest.NewRequest("GET", "/api/test?foo=bar", nil)
	req2 := httptest.NewRequest("GET", "/api/test?foo=bar", nil)
	req3 := httptest.NewRequest("GET", "/api/test?foo=baz", nil)

	key1 := generateCacheKey(req1)
	key2 := generateCacheKey(req2)
	key3 := generateCacheKey(req3)

	if key1 != key2 {
		t.Error("same requests should generate same cache key")
	}

	if key1 == key3 {
		t.Error("different requests should generate different cache keys")
	}
}

func TestGenerateCacheKey_WithAuth(t *testing.T) {
	req1 := httptest.NewRequest("GET", "/api/test", nil)
	req1.Header.Set("Authorization", "Bearer token1")

	req2 := httptest.NewRequest("GET", "/api/test", nil)
	req2.Header.Set("Authorization", "Bearer token2")

	key1 := generateCacheKey(req1)
	key2 := generateCacheKey(req2)

	if key1 == key2 {
		t.Error("requests with different auth headers should generate different cache keys")
	}
}

func TestShouldSkipCache(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/health", true},
		{"/health/ready", true},
		{"/metrics", true},
		{"/ping", true},
		{"/stream/video", true},
		{"/videos/stream", true},
		{"/content/main.m3u8", true},
		{"/video/segment.ts", true},
		{"/api/users", false},
		{"/videos/list", false},
		{"/content", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := shouldSkipCache(tt.path)
			if result != tt.expected {
				t.Errorf("shouldSkipCache(%s) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestCacheResponseWriter(t *testing.T) {
	rr := httptest.NewRecorder()
	cw := newCacheResponseWriter(rr)

	// Test WriteHeader
	cw.WriteHeader(http.StatusCreated)
	if cw.statusCode != http.StatusCreated {
		t.Errorf("expected statusCode %d, got %d", http.StatusCreated, cw.statusCode)
	}

	// Test Write
	data := []byte("test response")
	n, err := cw.Write(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected to write %d bytes, wrote %d", len(data), n)
	}
	if cw.body.String() != string(data) {
		t.Errorf("expected body %s, got %s", data, cw.body.String())
	}
}

func TestMetricsResponseWriter(t *testing.T) {
	rr := httptest.NewRecorder()
	mw := &metricsResponseWriter{rr, http.StatusOK}

	// Test default status code
	if mw.statusCode != http.StatusOK {
		t.Errorf("expected default statusCode %d, got %d", http.StatusOK, mw.statusCode)
	}

	// Test WriteHeader
	mw.WriteHeader(http.StatusNotFound)
	if mw.statusCode != http.StatusNotFound {
		t.Errorf("expected statusCode %d, got %d", http.StatusNotFound, mw.statusCode)
	}
}
