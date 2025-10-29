package util

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		pattern  string
		want     bool
	}{
		{"exact match", "/api/users", "/api/users", true},
		{"simple wildcard", "/api/users/123/profile", "/api/users/*/profile", true},
		{"no match", "/api/posts", "/api/users/*", false},
		{"suffix match only", "/other/api/users/123/profile", "/api/users/*/profile", false},
		{"prefix match only", "/api/users/123/settings", "/api/users/*/profile", false},
		{"invalid pattern", "/api/users/123", "/api/*/*", false},
		{"empty endpoint", "", "/api/*", false},
		{"empty pattern", "/api/users", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldWakeServer(tt.endpoint, []string{tt.pattern}); got != tt.want {
				t.Errorf("ShouldWakeServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldWakeServer(t *testing.T) {
	tests := []struct {
		name            string
		endpoint        string
		wakeUpEndpoints []string
		want            bool
	}{
		{
			name:            "exact match in endpoints",
			endpoint:        "/api/play",
			wakeUpEndpoints: []string{"/api/stop", "/api/play"},
			want:            true,
		},
		{
			name:            "wildcard match in endpoints",
			endpoint:        "/videos/12345/stream.m3u8",
			wakeUpEndpoints: []string{"/videos/*/stream.m3u8"},
			want:            true,
		},
		{
			name:            "no match",
			endpoint:        "/api/status",
			wakeUpEndpoints: []string{"/api/play", "/videos/*"},
			want:            false,
		},
		{
			name:            "empty endpoints list",
			endpoint:        "/api/play",
			wakeUpEndpoints: []string{},
			want:            false,
		},
		{
			name:            "empty endpoint string",
			endpoint:        "",
			wakeUpEndpoints: []string{"/api/play"},
			want:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldWakeServer(tt.endpoint, tt.wakeUpEndpoints); got != tt.want {
				t.Errorf("ShouldWakeServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsServerUp(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("server is up", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		if !IsServerUp(logger, server.Listener.Addr().String()) {
			t.Error("expected server to be up")
		}
	})

	t.Run("server is down", func(t *testing.T) {
		if IsServerUp(logger, "localhost:12345") {
			t.Error("expected server to be down")
		}
	})
}
