package caddy_jellywol_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	caddy_jellywol "github.com/Stoufiler/caddy-jellywol"
	"go.uber.org/zap"
)

// mockHandler is a dummy Caddy handler to simulate the next middleware (the actual reverse_proxy)
type mockHandler struct {
	called bool
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	m.called = true
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Proxy Success"))
	return nil
}

// setupMockServer starts a local TCP server to simulate Jellyfin being "UP"
func setupMockServer(t *testing.T) (string, int, func()) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}

	port := l.Addr().(*net.TCPAddr).Port
	ip := "127.0.0.1"

	var wg sync.WaitGroup
	wg.Add(1)

	// Accept connections and close them immediately to simulate a healthy ping response
	go func() {
		defer wg.Done()
		for {
			conn, err := l.Accept()
			if err != nil {
				return // Server closed
			}
			_ = conn.Close()
		}
	}()

	cleanup := func() {
		_ = l.Close()
		wg.Wait()
	}

	return ip, port, cleanup
}

func TestIntegration_ServerIsUp(t *testing.T) {
	ip, port, cleanup := setupMockServer(t)
	defer cleanup()

	// 1. Setup our Plugin
	plugin := &caddy_jellywol.JellyWol{
		Mac:        "00:11:22:33:44:55",
		PingIP:     ip,
		PingPort:   port,
		BlockPaths: []string{"/Videos/*"},
	}

	// Mock the Caddy Provision phase
	plugin.ProvisionMock(zap.NewNop(), 2*time.Second)

	// 2. Setup the request and recorder
	req := httptest.NewRequest(http.MethodGet, "/Videos/123/stream", nil)
	rr := httptest.NewRecorder()
	next := &mockHandler{}

	// 3. Execute
	err := plugin.ServeHTTP(rr, req, next)

	// 4. Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !next.called {
		t.Error("Expected the next handler (proxy) to be called because server is UP")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %v", rr.Code)
	}
}

func TestIntegration_ServerIsDown_BlockPath(t *testing.T) {
	// Pick a random unassigned port to simulate a DOWN server
	ip := "127.0.0.1"
	port := 45678

	plugin := &caddy_jellywol.JellyWol{
		Mac:        "00:11:22:33:44:55",
		PingIP:     ip,
		PingPort:   port,
		BlockPaths: []string{"/Videos/*"},
		RetryAfter: 15,
	}
	plugin.ProvisionMock(zap.NewNop(), 100*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/Videos/123/stream", nil)
	rr := httptest.NewRecorder()
	next := &mockHandler{}

	err := plugin.ServeHTTP(rr, req, next)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if next.called {
		t.Error("Expected the next handler to NOT be called because server is DOWN on a BlockPath")
	}

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503 Service Unavailable, got %v", rr.Code)
	}

	if retry := rr.Header().Get("Retry-After"); retry != "15" {
		t.Errorf("Expected Retry-After: 15, got %v", retry)
	}
}

func TestIntegration_ServerIsDown_TriggerPath(t *testing.T) {
	ip := "127.0.0.1"
	port := 45678

	plugin := &caddy_jellywol.JellyWol{
		Mac:          "00:11:22:33:44:55",
		PingIP:       ip,
		PingPort:     port,
		TriggerPaths: []string{"/Library/*"},
	}
	plugin.ProvisionMock(zap.NewNop(), 100*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/Library/movies", nil)
	rr := httptest.NewRecorder()
	next := &mockHandler{}

	err := plugin.ServeHTTP(rr, req, next)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Crucial difference: even if down, a trigger path MUST let the request pass through!
	if !next.called {
		t.Error("Expected the next handler to BE called because this is only a TriggerPath")
	}
}

func TestIntegration_UnmatchedPath(t *testing.T) {
	ip := "127.0.0.1"
	port := 45678

	plugin := &caddy_jellywol.JellyWol{
		Mac:          "00:11:22:33:44:55",
		PingIP:       ip,
		PingPort:     port,
		BlockPaths:   []string{"/Videos/*"},
		TriggerPaths: []string{"/Library/*"},
	}
	plugin.ProvisionMock(zap.NewNop(), 100*time.Millisecond)

	// Path matches neither Block nor Trigger
	req := httptest.NewRequest(http.MethodGet, "/web/index.html", nil)
	rr := httptest.NewRecorder()
	next := &mockHandler{}

	err := plugin.ServeHTTP(rr, req, next)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !next.called {
		t.Error("Expected the next handler to BE called because path is ignored by middleware")
	}
}
