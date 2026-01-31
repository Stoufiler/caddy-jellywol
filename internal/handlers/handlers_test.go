package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/Stoufiler/JellyWolProxy/internal/config"
	"github.com/Stoufiler/JellyWolProxy/internal/server_state"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Mock services for testing
type mockServerStateChecker struct {
	isUp bool
}

func (m *mockServerStateChecker) IsServerUp(logger *logrus.Logger, address string) bool {
	return m.isUp
}

type mockWaker struct {
	called bool
}

func (m *mockWaker) WakeServer(logger *logrus.Logger, macAddress string, broadcastAddress string, config config.Config, serverState *server_state.ServerState) bool {
	m.called = true
	// In a real scenario, you might want to use channels to coordinate goroutines
	// For this test, we assume it returns true to indicate it started the process.
	return serverState.StartWakingUp()
}

func TestHandleDomainProxy(t *testing.T) {
	// Create a test server to act as the backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("Hello from backend")); err != nil {
			t.Fatal(err)
		}
	}))
	defer backend.Close()

	// Parse the backend URL
	backendURL, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatal(err)
	}

	// Create a config with the backend's host and port
	port, err := strconv.Atoi(backendURL.Port())
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		ForwardIp:   backendURL.Hostname(),
		ForwardPort: port,
	}

	// Create a logger for the test
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	// Create a request to proxy
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	// Call the handler
	handleDomainProxy(rr, req, cfg, logger)

	// Check the status code
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check the response body
	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "Hello from backend" {
		t.Errorf("expected body \"Hello from backend\", got \"%s\"", string(body))
	}
}

func TestPingHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/ping", nil)
	rr := httptest.NewRecorder()

	PingHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestHandler(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Don't show logs during tests

	// Create a test server to act as the backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	backendURL, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatal(err)
	}

	port, err := strconv.Atoi(backendURL.Port())
	if err != nil {
		t.Fatal(err)
	}

	baseConfig := config.Config{
		WakeUpEndpoints: []string{"/api/wakeup"},
		ForwardIp:       backendURL.Hostname(),
		ForwardPort:     port,
	}

	t.Run("non-wakeup endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/other", nil)
		rr := httptest.NewRecorder()

		checker := &mockServerStateChecker{isUp: true} // Doesn't matter for this test
		waker := &mockWaker{}
		serverState := &server_state.ServerState{}

		router := mux.NewRouter()
		router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Handler(w, r, logger, baseConfig, serverState, checker, waker)
		})
		router.ServeHTTP(rr, req)

		if waker.called {
			t.Error("WakeServer should not be called for non-wakeup endpoint")
		}

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
	})

	t.Run("wakeup endpoint when server is up", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/wakeup", nil)
		rr := httptest.NewRecorder()

		checker := &mockServerStateChecker{isUp: true}
		waker := &mockWaker{}
		serverState := &server_state.ServerState{}

		router := mux.NewRouter()
		router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Handler(w, r, logger, baseConfig, serverState, checker, waker)
		})
		router.ServeHTTP(rr, req)

		if waker.called {
			t.Error("WakeServer should not be called when server is already up")
		}

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
	})

	t.Run("wakeup endpoint, server down, successful wake", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/wakeup", nil)
		rr := httptest.NewRecorder()

		checker := &mockServerStateChecker{isUp: false}
		waker := &mockWaker{}
		serverState := &server_state.ServerState{}

		router := mux.NewRouter()
		router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Handler(w, r, logger, baseConfig, serverState, checker, waker)
		})
		router.ServeHTTP(rr, req)

		if !waker.called {
			t.Error("WakeServer should be called when server is down")
		}

		if rr.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, rr.Code)
		}

		if retryAfter := rr.Header().Get("Retry-After"); retryAfter != "30" {
			t.Errorf("expected Retry-After: 30, got: %s", retryAfter)
		}
	})

	t.Run("wakeup endpoint, server down, failed wake", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/wakeup", nil)
		rr := httptest.NewRecorder()

		checker := &mockServerStateChecker{isUp: false}
		waker := &mockWaker{}
		serverState := &server_state.ServerState{}

		router := mux.NewRouter()
		router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Handler(w, r, logger, baseConfig, serverState, checker, waker)
		})
		router.ServeHTTP(rr, req)

		if !waker.called {
			t.Error("WakeServer should have been called")
		}

		if rr.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, rr.Code)
		}

		if retryAfter := rr.Header().Get("Retry-After"); retryAfter != "30" {
			t.Errorf("expected Retry-After: 30, got: %s", retryAfter)
		}
	})
}
