package server

import (
	"net/http"
	"time"

	"github.com/Stoufiler/JellyWolProxy/internal/config"
	"github.com/Stoufiler/JellyWolProxy/internal/util"
	"github.com/sirupsen/logrus"
)

func WaitServerOnline(logger *logrus.Logger, serverAddress string, config *config.Config, w http.ResponseWriter) bool {
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

	flusher, ok := w.(http.Flusher)

	for {
		select {
		case <-ticker.C:
			if ok {
				w.WriteHeader(http.StatusProcessing)
				flusher.Flush()
			}
			if util.IsServerUp(logger, serverAddress) {
				logger.Info("Server is up !")
				if config.PostPingDelaySeconds > 0 {
					logger.Infof("Waiting for %d seconds as configured...", config.PostPingDelaySeconds)
					time.Sleep(time.Duration(config.PostPingDelaySeconds) * time.Second)
				}
				return true
			}
		case <-timeout:
			logger.Info("Timeout reached, server did not wake up.")
			return false
		}
	}
}
