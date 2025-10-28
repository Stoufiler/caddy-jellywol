package server

import (
	"time"

	"github.com/StephanGR/JellyWolProxy/internal/config"
	"github.com/StephanGR/JellyWolProxy/internal/jellyfin"
	"github.com/StephanGR/JellyWolProxy/internal/util"
	"github.com/sirupsen/logrus"
)

func WaitServerOnline(logger *logrus.Logger, serverAddress string, config *config.Config) bool {
	timeoutDuration := time.Duration(config.ServerWakeUpTimeout) * time.Second
	if config.ServerWakeUpTimeout == 0 {
		timeoutDuration = 2 * time.Minute
	}
	tickerDuration := time.Duration(config.ServerWakeUpTicker) * time.Second
	if config.ServerWakeUpTicker == 0 {
		tickerDuration = 5 * time.Second
	}

	timeout := time.After(timeoutDuration)
	ticker := time.NewTicker(tickerDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if util.IsServerUp(logger, serverAddress) {
				logger.Info("Server is up !")
				jellyfin.SendJellyfinMessagesToAllSessions(logger, config.JellyfinUrl, config.ApiKey, "Information ", "\nLe serveur est démarré !\nBon film !")
				return true
			}
		case <-timeout:
			logger.Info("Timeout reached, server did not wake up.")
			return false
		}
	}
}
