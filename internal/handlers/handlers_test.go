package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StephanGR/JellyWolProxy/internal/config"
	"github.com/StephanGR/JellyWolProxy/internal/server_state"
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

type mockServerWaiter struct {
	willSucceed bool
}

func (m *mockServerWaiter) WaitServerOnline(logger *logrus.Logger, serverAddress string, config *config.Config) bool {
	return m.willSucceed
}

func TestHandler(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Don't show logs during tests

	baseConfig := config.Config{
		WakeUpEndpoints: []string{"/api/wakeup"},
		ForwardIp:       "127.0.0.1",
		ForwardPort:     8080, // A dummy port for the test
	}

	t.Run("non-wakeup endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/other", nil)
		rr := httptest.NewRecorder()

		checker := &mockServerStateChecker{isUp: true} // Doesn't matter for this test
		waker := &mockWaker{}
		waiter := &mockServerWaiter{}
		serverState := &server_state.ServerState{}

		Handler(rr, req, logger, baseConfig, serverState, checker, waker, waiter)

		if waker.called {
			t.Error("WakeServer should not be called for non-wakeup endpoint")
		}
		// We expect the proxy to be called, but testing httputil.ReverseProxy is complex.
		// For this test, we focus on the wake-up logic.
	})

	t.Run("wakeup endpoint when server is up", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/wakeup", nil)
		rr := httptest.NewRecorder()

		checker := &mockServerStateChecker{isUp: true}
		waker := &mockWaker{}
		waiter := &mockServerWaiter{}
		serverState := &server_state.ServerState{}

		Handler(rr, req, logger, baseConfig, serverState, checker, waker, waiter)

		if waker.called {
			t.Error("WakeServer should not be called when server is already up")
		}
	})

	t.Run("wakeup endpoint, server down, successful wake", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/wakeup", nil)
		rr := httptest.NewRecorder()

		checker := &mockServerStateChecker{isUp: false}
		waker := &mockWaker{}
		waiter := &mockServerWaiter{willSucceed: true}
		serverState := &server_state.ServerState{}

		Handler(rr, req, logger, baseConfig, serverState, checker, waker, waiter)

		if !waker.called {
			t.Error("WakeServer should be called when server is down")
		}
		// Again, not testing the proxy itself, but we expect a successful status
		// if the proxying was successful. Since the test proxy will fail to connect,
		// we can't assert on status 200. We focus on the wake-up logic.
	})

	t.Run("wakeup endpoint, server down, failed wake", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/wakeup", nil)
		rr := httptest.NewRecorder()

		checker := &mockServerStateChecker{isUp: false}
		waker := &mockWaker{}
		waiter := &mockServerWaiter{willSucceed: false}
		serverState := &server_state.ServerState{}

		Handler(rr, req, logger, baseConfig, serverState, checker, waker, waiter)

		if !waker.called {
			t.Error("WakeServer should have been called")
		}

		if rr.Code != http.StatusGatewayTimeout {
			t.Errorf("expected status %d, got %d", http.StatusGatewayTimeout, rr.Code)
		}
	})
}
