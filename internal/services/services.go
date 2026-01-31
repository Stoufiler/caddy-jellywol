package services

import (
	"github.com/Stoufiler/JellyWolProxy/internal/config"
	"github.com/Stoufiler/JellyWolProxy/internal/server_state"
	"github.com/sirupsen/logrus"
)

type Waker interface {
	WakeServer(logger *logrus.Logger, macAddress string, broadcastAddress string, config config.Config, serverState *server_state.ServerState) bool
}

type ServerStateChecker interface {
	IsServerUp(logger *logrus.Logger, address string) bool
}
