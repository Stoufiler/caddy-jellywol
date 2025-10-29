package services

import (
	"net/http"
	"github.com/StephanGR/JellyWolProxy/internal/config"
	"github.com/StephanGR/JellyWolProxy/internal/server_state"
	"github.com/sirupsen/logrus"
)

type Waker interface {
	WakeServer(logger *logrus.Logger, macAddress string, broadcastAddress string, config config.Config, serverState *server_state.ServerState) bool
}

type ServerStateChecker interface {
	IsServerUp(logger *logrus.Logger, address string) bool
}

type ServerWaiter interface {
	WaitServerOnline(logger *logrus.Logger, serverAddress string, config *config.Config, w http.ResponseWriter) bool
}
