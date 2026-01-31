package websocket

import (
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestIsWebSocketRequest(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected bool
	}{
		{
			name: "valid websocket request",
			headers: map[string]string{
				"Connection": "Upgrade",
				"Upgrade":    "websocket",
			},
			expected: true,
		},
		{
			name: "case insensitive headers",
			headers: map[string]string{
				"Connection": "upgrade",
				"Upgrade":    "WebSocket",
			},
			expected: true,
		},
		{
			name: "missing upgrade header",
			headers: map[string]string{
				"Connection": "Upgrade",
			},
			expected: false,
		},
		{
			name: "missing connection header",
			headers: map[string]string{
				"Upgrade": "websocket",
			},
			expected: false,
		},
		{
			name:     "no headers",
			headers:  map[string]string{},
			expected: false,
		},
		{
			name: "connection keep-alive with upgrade",
			headers: map[string]string{
				"Connection": "keep-alive, Upgrade",
				"Upgrade":    "websocket",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ws", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			result := IsWebSocketRequest(req)
			if result != tt.expected {
				t.Errorf("IsWebSocketRequest() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProxyWebSocket_NonHijackable(t *testing.T) {
	// Regular ResponseRecorder doesn't support Hijack
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")

	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	err := ProxyWebSocket(w, req, "http://localhost:8096", logger)
	if err == nil {
		t.Error("Expected error when ResponseWriter doesn't support Hijack")
	}
}
