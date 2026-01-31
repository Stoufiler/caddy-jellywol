package services

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Stoufiler/JellyWolProxy/internal/config"
	"github.com/Stoufiler/JellyWolProxy/internal/server_state"
	"github.com/sirupsen/logrus"
)

func TestConcreteServerStateChecker_IsServerUp(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("server is up", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		checker := &ConcreteServerStateChecker{}
		if !checker.IsServerUp(logger, server.Listener.Addr().String()) {
			t.Error("expected server to be up")
		}
	})

	t.Run("server is down", func(t *testing.T) {
		checker := &ConcreteServerStateChecker{}
		if checker.IsServerUp(logger, "localhost:12345") {
			t.Error("expected server to be down")
		}
	})
}

func TestConcreteWaker_WakeServer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// This test is limited because it can't actually send a WOL packet without network access
	// and potentially elevated permissions. We will just test the logic.
	t.Run("wake server logic", func(t *testing.T) {
		waker := &ConcreteWaker{}
		cfg := config.Config{
			MacAddress:       "00:11:22:33:44:55",
			BroadcastAddress: "192.168.1.255",
		}
		serverState := &server_state.ServerState{}
		// We can't assert that the packet was sent, but we can check that the function returns true
		if !waker.WakeServer(logger, cfg.MacAddress, cfg.BroadcastAddress, cfg, serverState) {
			t.Error("expected WakeServer to return true")
		}
	})
}
