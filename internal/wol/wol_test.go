package wol

import (
	"testing"

	"github.com/Stoufiler/JellyWolProxy/internal/config"
	"github.com/Stoufiler/JellyWolProxy/internal/server_state"
	"github.com/sirupsen/logrus"
)

func TestWakeServer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("wake server logic", func(t *testing.T) {
		cfg := config.Config{
			MacAddress:       "00:11:22:33:44:55",
			BroadcastAddress: "192.168.1.255",
		}
		serverState := &server_state.ServerState{}
		if !WakeServer(logger, cfg.MacAddress, cfg.BroadcastAddress, cfg, serverState) {
			t.Error("expected WakeServer to return true")
		}
	})
}
