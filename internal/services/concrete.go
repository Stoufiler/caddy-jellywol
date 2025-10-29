package services

import (
	"net/http"
	"github.com/StephanGR/JellyWolProxy/internal/config"
	"github.com/StephanGR/JellyWolProxy/internal/server"
	"github.com/StephanGR/JellyWolProxy/internal/server_state"
	"github.com/StephanGR/JellyWolProxy/internal/util"
	"github.com/StephanGR/JellyWolProxy/internal/wol"
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

// ConcreteServerWaiter implements the ServerWaiter interface.
type ConcreteServerWaiter struct{}

func (w *ConcreteServerWaiter) WaitServerOnline(logger *logrus.Logger, serverAddress string, config *config.Config, rw http.ResponseWriter) bool {
	return server.WaitServerOnline(logger, serverAddress, config, rw)
}
