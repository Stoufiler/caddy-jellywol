package wol

import (
	"bytes"
	"testing"

	"github.com/Stoufiler/JellyWolProxy/internal/config"
	"github.com/Stoufiler/JellyWolProxy/internal/server_state"
	"github.com/sirupsen/logrus"
)

func TestWakeServer(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("wake server logic", func(t *testing.T) {
		cfg := config.Config{
			MacAddress:       "00:11:22:33:44:55",
			BroadcastAddress: "192.168.1.255:9",
		}
		serverState := &server_state.ServerState{}
		if !WakeServer(logger, cfg.MacAddress, cfg.BroadcastAddress, cfg, serverState) {
			t.Error("expected WakeServer to return true")
		}
	})

	t.Run("already waking up", func(t *testing.T) {
		cfg := config.Config{
			MacAddress:       "00:11:22:33:44:55",
			BroadcastAddress: "192.168.1.255:9",
		}
		serverState := &server_state.ServerState{}

		// Start a wake-up
		serverState.StartWakingUp()

		// Second call should return false (already waking)
		result := WakeServer(logger, cfg.MacAddress, cfg.BroadcastAddress, cfg, serverState)
		if result {
			t.Error("expected WakeServer to return false when already waking up")
		}

		// Clean up
		serverState.DoneWakingUp()
	})

	t.Run("invalid mac address", func(t *testing.T) {
		cfg := config.Config{
			MacAddress:       "invalid-mac",
			BroadcastAddress: "192.168.1.255:9",
		}
		serverState := &server_state.ServerState{}

		// Should still return true (wake was started) but log warning
		result := WakeServer(logger, cfg.MacAddress, cfg.BroadcastAddress, cfg, serverState)
		if !result {
			t.Error("expected WakeServer to return true even with invalid MAC")
		}

		serverState.DoneWakingUp()
	})

	t.Run("empty mac address", func(t *testing.T) {
		cfg := config.Config{
			MacAddress:       "",
			BroadcastAddress: "192.168.1.255:9",
		}
		serverState := &server_state.ServerState{}

		result := WakeServer(logger, cfg.MacAddress, cfg.BroadcastAddress, cfg, serverState)
		if !result {
			t.Error("expected WakeServer to return true")
		}

		serverState.DoneWakingUp()
	})

	t.Run("concurrent wake attempts", func(t *testing.T) {
		cfg := config.Config{
			MacAddress:       "00:11:22:33:44:55",
			BroadcastAddress: "192.168.1.255:9",
		}
		serverState := &server_state.ServerState{}

		// First call should succeed
		result1 := WakeServer(logger, cfg.MacAddress, cfg.BroadcastAddress, cfg, serverState)
		if !result1 {
			t.Error("first wake should succeed")
		}

		// Second concurrent call should fail
		result2 := WakeServer(logger, cfg.MacAddress, cfg.BroadcastAddress, cfg, serverState)
		if result2 {
			t.Error("second concurrent wake should fail")
		}

		// Clean up
		serverState.DoneWakingUp()

		// After cleanup, new wake should succeed
		result3 := WakeServer(logger, cfg.MacAddress, cfg.BroadcastAddress, cfg, serverState)
		if !result3 {
			t.Error("wake after cleanup should succeed")
		}
		serverState.DoneWakingUp()
	})
}
