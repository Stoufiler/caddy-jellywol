package services

import (
	"github.com/Stoufiler/JellyWolProxy/internal/config"
	"github.com/Stoufiler/JellyWolProxy/internal/server_state"
	"github.com/Stoufiler/JellyWolProxy/internal/util"
	"github.com/Stoufiler/JellyWolProxy/internal/wol"
	"github.com/sirupsen/logrus"
)

// ConcreteWaker implements the Waker interface.
type ConcreteWaker struct{}

func (w *ConcreteWaker) WakeServer(logger *logrus.Logger, macAddress string, broadcastAddress string, config config.Config, serverState *server_state.ServerState) bool {
	return wol.WakeServer(logger, macAddress, broadcastAddress, config, serverState)
}

// ConcreteServerStateChecker implements the ServerStateChecker interface.
type ConcreteServerStateChecker struct{}

func (c *ConcreteServerStateChecker) IsServerUp(logger *logrus.Logger, address string) bool {
	return util.IsServerUp(logger, address)
}
