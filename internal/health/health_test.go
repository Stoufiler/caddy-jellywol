package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StephanGR/JellyWolProxy/internal/config"
	"github.com/sirupsen/logrus"
)

// Mock service for testing
type mockServerStateChecker struct {
	isUp bool
}

func (m *mockServerStateChecker) IsServerUp(logger *logrus.Logger, address string) bool {
	return m.isUp
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	HealthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if response.Status != StatusUp {
		t.Errorf("expected status %s, got %s", StatusUp, response.Status)
	}
}

func TestReadinessHandler(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	baseConfig := &config.Config{}

	t.Run("jellyfin is up", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health/ready", nil)
		rr := httptest.NewRecorder()

		checker := &mockServerStateChecker{isUp: true}
		handler := ReadinessHandler(logger, baseConfig, checker)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		var response HealthResponse
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("could not decode response: %v", err)
		}

		if response.Status != StatusUp {
			t.Errorf("expected overall status %s, got %s", StatusUp, response.Status)
		}
		if response.Checks["jellyfin"].Status != StatusUp {
			t.Errorf("expected jellyfin status %s, got %s", StatusUp, response.Checks["jellyfin"].Status)
		}
	})

	t.Run("jellyfin is down", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health/ready", nil)
		rr := httptest.NewRecorder()

		checker := &mockServerStateChecker{isUp: false}
		handler := ReadinessHandler(logger, baseConfig, checker)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, rr.Code)
		}

		var response HealthResponse
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("could not decode response: %v", err)
		}

		if response.Status != StatusDown {
			t.Errorf("expected overall status %s, got %s", StatusDown, response.Status)
		}
		if response.Checks["jellyfin"].Status != StatusDown {
			t.Errorf("expected jellyfin status %s, got %s", StatusDown, response.Checks["jellyfin"].Status)
		}
	})
}
