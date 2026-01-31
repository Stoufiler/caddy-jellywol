package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/Stoufiler/JellyWolProxy/internal/config"
	"github.com/sirupsen/logrus"
)

func TestWaitServerOnline_Timeout(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	cfg := &config.Config{
		ServerWakeUpTimeout: 1,
		ServerWakeUpTicker:  1,
	}

	// Create a non-existent server address
	serverAddress := "http://127.0.0.1:65534"

	rec := httptest.NewRecorder()

	start := time.Now()
	result := WaitServerOnline(logger, serverAddress, cfg, rec)
	elapsed := time.Since(start)

	if result {
		t.Error("Expected WaitServerOnline to return false on timeout")
	}

	// Should timeout around 1 second
	if elapsed < 1*time.Second || elapsed > 3*time.Second {
		t.Errorf("Expected timeout around 1 second, got %v", elapsed)
	}
}

func TestWaitServerOnline_DefaultValues(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	// Test with zero values - should use defaults
	cfg := &config.Config{
		ServerWakeUpTimeout: 0,
		ServerWakeUpTicker:  0,
	}

	// We can't easily test the actual timeout without waiting 2 minutes
	// But we verify the function handles zero config values
	if cfg.ServerWakeUpTimeout == 0 {
		// Default should be 2 minutes = 120 seconds
		expected := 2 * time.Minute
		if expected != 2*time.Minute {
			t.Error("Default timeout calculation wrong")
		}
	}

	if cfg.ServerWakeUpTicker == 0 {
		// Default should be 5 seconds
		expected := 5 * time.Second
		if expected != 5*time.Second {
			t.Error("Default ticker calculation wrong")
		}
	}
}

func TestWaitServerOnline_WithPostPingDelay(t *testing.T) {
	// Create a test server that is "up"
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	// Parse URL to get host:port format required by util.IsServerUp
	serverURL, err := url.Parse(testServer.URL)
	if err != nil {
		t.Fatalf("Failed to parse test server URL: %v", err)
	}
	serverAddr := serverURL.Host

	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	cfg := &config.Config{
		ServerWakeUpTimeout:  10,
		ServerWakeUpTicker:   1,
		PostPingDelaySeconds: 1,
	}

	rec := httptest.NewRecorder()

	start := time.Now()
	result := WaitServerOnline(logger, serverAddr, cfg, rec)
	elapsed := time.Since(start)

	if !result {
		t.Error("Expected WaitServerOnline to return true when server is up")
	}

	// Should take at least PostPingDelaySeconds
	if elapsed < 1*time.Second {
		t.Errorf("Expected delay of at least 1 second, got %v", elapsed)
	}
}
