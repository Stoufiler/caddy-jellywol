package handlers

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Stoufiler/JellyWolProxy/internal/config"
	"github.com/Stoufiler/JellyWolProxy/internal/server_state"
	"github.com/sirupsen/logrus"
)

// Integration test mocks
type integrationServerStateChecker struct {
	serverUp bool
}

func (m *integrationServerStateChecker) IsServerUp(logger *logrus.Logger, address string) bool {
	return m.serverUp
}

type integrationWaker struct {
	wakeCalled bool
}

func (m *integrationWaker) WakeServer(logger *logrus.Logger, macAddress string, broadcastAddress string, config config.Config, serverState *server_state.ServerState) bool {
	m.wakeCalled = true
	return true
}

// TestIntegrationServerAlreadyOnline tests the flow when the server is already online
// Note: This test will attempt to proxy to an unreachable host, so we just verify wake logic
func TestIntegrationServerAlreadyOnline(t *testing.T) {
	t.Skip("Skipping integration test that requires real network connection")

	cfg := config.Config{
		ForwardIp:       "192.168.1.100",
		ForwardPort:     8096,
		WakeUpIp:        "192.168.1.100",
		WakeUpPort:      80,
		WakeUpEndpoints: []string{"/videos/*"},
	}

	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // Suppress all output in tests

	serverState := &server_state.ServerState{}
	checker := &integrationServerStateChecker{serverUp: true}
	waker := &integrationWaker{}

	// Test request
	req := httptest.NewRequest("GET", "/videos/test.mp4", nil)
	rr := httptest.NewRecorder()

	Handler(rr, req, log, cfg, serverState, checker, waker)

	// Verify wake was NOT called since server was already up
	if waker.wakeCalled {
		t.Error("Expected wake to not be called when server is already online")
	}
}

// TestIntegrationServerNeedsWakeUp tests the flow when server needs to be woken up
func TestIntegrationServerNeedsWakeUp(t *testing.T) {
	t.Skip("Skipping integration test that requires real network connection")

	cfg := config.Config{
		ForwardIp:            "192.168.1.100",
		ForwardPort:          8096,
		WakeUpIp:             "192.168.1.100",
		WakeUpPort:           80,
		WakeUpEndpoints:      []string{"/videos/*"},
		ServerWakeUpTimeout:  10,
		ServerWakeUpTicker:   1,
		PostPingDelaySeconds: 0,
	}

	log := logrus.New()
	log.SetLevel(logrus.FatalLevel)

	serverState := &server_state.ServerState{}
	checker := &integrationServerStateChecker{serverUp: false}
	waker := &integrationWaker{}

	req := httptest.NewRequest("GET", "/videos/test.mp4", nil)
	rr := httptest.NewRecorder()

	Handler(rr, req, log, cfg, serverState, checker, waker)

	// Verify wake WAS called
	if !waker.wakeCalled {
		t.Error("Expected wake to be called when server is offline")
	}

	// The wake/wait logic should have been triggered
	// We can't verify status code because the actual proxy will fail,
	// but the wake logic is what we're testing
}

// TestIntegrationNonWakeEndpoint tests that non-wake endpoints don't trigger wake
func TestIntegrationNonWakeEndpoint(t *testing.T) {
	t.Skip("Skipping integration test that requires real network connection")

	cfg := config.Config{
		ForwardIp:       "192.168.1.100",
		ForwardPort:     8096,
		WakeUpIp:        "192.168.1.100",
		WakeUpPort:      80,
		WakeUpEndpoints: []string{"/videos/*"},
	}

	log := logrus.New()
	log.SetLevel(logrus.FatalLevel)

	serverState := &server_state.ServerState{}
	checker := &integrationServerStateChecker{serverUp: false}
	waker := &integrationWaker{}

	// Request to non-wake endpoint
	req := httptest.NewRequest("GET", "/api/users", nil)
	rr := httptest.NewRecorder()

	Handler(rr, req, log, cfg, serverState, checker, waker)

	// Wake should NOT be called for non-wake endpoints
	if waker.wakeCalled {
		t.Error("Expected wake to not be called for non-wake endpoint")
	}
}

// TestIntegrationConcurrentWakeRequests tests that multiple concurrent wake requests are handled correctly
func TestIntegrationConcurrentWakeRequests(t *testing.T) {
	t.Skip("Skipping integration test that requires real network connection")

	cfg := config.Config{
		ForwardIp:            "192.168.1.100",
		ForwardPort:          8096,
		WakeUpIp:             "192.168.1.100",
		WakeUpPort:           80,
		WakeUpEndpoints:      []string{"/videos/*"},
		ServerWakeUpTimeout:  10,
		ServerWakeUpTicker:   1,
		PostPingDelaySeconds: 0,
	}

	log := logrus.New()
	log.SetLevel(logrus.FatalLevel)

	serverState := &server_state.ServerState{}
	checker := &integrationServerStateChecker{serverUp: false}

	// Track wake calls
	wakeCount := 0
	waker := &mockWakerWithCounter{
		wakeCount: &wakeCount,
	}

	// Simulate 5 concurrent requests
	concurrentRequests := 5
	done := make(chan bool, concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		go func(id int) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/videos/test%d.mp4", id), nil)
			rr := httptest.NewRecorder()
			Handler(rr, req, log, cfg, serverState, checker, waker)
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < concurrentRequests; i++ {
		<-done
	}

	// Only one wake should have been initiated
	if wakeCount != 1 {
		t.Errorf("Expected exactly 1 wake call for concurrent requests, got %d", wakeCount)
	}
}

type mockWakerWithCounter struct {
	wakeCount *int
}

func (m *mockWakerWithCounter) WakeServer(logger *logrus.Logger, macAddress string, broadcastAddress string, config config.Config, serverState *server_state.ServerState) bool {
	if serverState.StartWakingUp() {
		*m.wakeCount++
		return true
	}
	return false
}

// TestIntegrationStateTransition tests state changes during wake process
func TestIntegrationStateTransition(t *testing.T) {
	cfg := config.Config{
		ForwardIp:            "192.168.1.100",
		ForwardPort:          8096,
		WakeUpIp:             "192.168.1.100",
		WakeUpPort:           80,
		WakeUpEndpoints:      []string{"/videos/*"},
		ServerWakeUpTimeout:  10,
		ServerWakeUpTicker:   1,
		PostPingDelaySeconds: 0,
	}

	log := logrus.New()
	log.SetLevel(logrus.FatalLevel)

	serverState := &server_state.ServerState{}

	// Test state transitions
	if serverState.IsWakingUp() {
		t.Error("Expected initial state to not be waking up")
	}

	// Simulate wake process
	serverOnline := false
	checker := &mockCheckerWithState{online: &serverOnline}
	waker := &integrationWaker{}

	req := httptest.NewRequest("GET", "/videos/movie.mp4", nil)
	rr := httptest.NewRecorder()

	// This will trigger wake
	go Handler(rr, req, log, cfg, serverState, checker, waker)

	// Give it a moment to start waking
	time.Sleep(50 * time.Millisecond)

	// Verify wake was called
	if !waker.wakeCalled {
		t.Error("Expected wake to be called")
	}
}

type mockCheckerWithState struct {
	online *bool
}

func (m *mockCheckerWithState) IsServerUp(logger *logrus.Logger, address string) bool {
	return *m.online
}
