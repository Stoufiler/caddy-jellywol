package jellyfin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestSendJellyfinMessage(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("sends correct request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/Sessions/test-session-id/message" {
				t.Errorf("expected path /Sessions/test-session-id/message, got %s", r.URL.Path)
			}
			if r.Method != http.MethodPost {
				t.Errorf("expected method POST, got %s", r.Method)
			}
			if r.Header.Get("X-MediaBrowser-Token") != "test-api-key" {
				t.Errorf("missing or incorrect api key header")
			}

			var msg JellyfinMessage
			if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if msg.Header != "Test Header" || msg.Text != "Test Text" {
				t.Errorf("incorrect message content: %+v", msg)
			}

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		SendJellyfinMessage(logger, server.URL, "test-api-key", "test-session-id", "Test Header", "Test Text")
	})
}

func TestSendJellyfinMessagesToAllSessions(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("sends messages to all sessions", func(t *testing.T) {
		var messageCount int32

		mux := http.NewServeMux()
		mux.HandleFunc("/Sessions", func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-MediaBrowser-Token") != "test-api-key" {
				t.Errorf("missing or incorrect api key header for /Sessions")
			}
			w.Header().Set("Content-Type", "application/json")
			response := []map[string]interface{}{
				{"Id": "session1"},
				{"Id": "session2"},
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Fatalf("failed to encode sessions response: %v", err)
			}
		})

		mux.HandleFunc("/Sessions/session1/message", func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&messageCount, 1)
			w.WriteHeader(http.StatusNoContent)
		})
		mux.HandleFunc("/Sessions/session2/message", func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&messageCount, 1)
			w.WriteHeader(http.StatusNoContent)
		})

		server := httptest.NewServer(mux)
		defer server.Close()

		SendJellyfinMessagesToAllSessions(logger, server.URL, "test-api-key", "Header", "Text")

		if atomic.LoadInt32(&messageCount) != 2 {
			t.Errorf("expected 2 messages to be sent, but got %d", messageCount)
		}
	})
}
