package handlers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/StephanGR/JellyWolProxy/internal/config"
	"github.com/StephanGR/JellyWolProxy/internal/server_state"
	"github.com/StephanGR/JellyWolProxy/internal/services"
	"github.com/StephanGR/JellyWolProxy/internal/util"
	"github.com/sirupsen/logrus"
)

func handleDomainProxy(w http.ResponseWriter, r *http.Request, config config.Config) {
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", config.ForwardIp, config.ForwardPort),
	})

	r.URL.Host = fmt.Sprintf("%s:%d", config.ForwardIp, config.ForwardPort)
	r.URL.Scheme = "http"
	r.Host = fmt.Sprintf("%s:%d", config.ForwardIp, config.ForwardPort)
	proxy.ServeHTTP(w, r)
}

func Handler(w http.ResponseWriter, r *http.Request, logger *logrus.Logger, config config.Config, serverState *server_state.ServerState, checker services.ServerStateChecker, waker services.Waker, waiter services.ServerWaiter) {
	logger.Debug("Request received for path: ", r.URL.Path)
	if util.ShouldWakeServer(r.URL.Path, config.WakeUpEndpoints) {
		serverAddress := fmt.Sprintf("%s:%d", config.WakeUpIp, config.WakeUpPort)
		logger.Debug("Wake-up endpoint matched, checking server status...")
		if !checker.IsServerUp(logger, serverAddress) {
			if waker.WakeServer(logger, config.MacAddress, config.BroadcastAddress, config, serverState) {
				defer serverState.DoneWakingUp()
				logger.Info("Server is offline, trying to wake up using Wake On Lan")
			} else {
				logger.Info("Server is waking up, waiting for it to be online...")
			}

			if waiter.WaitServerOnline(logger, serverAddress, &config, w) {
				logger.Info("Server is now online, proxying request")
				handleDomainProxy(w, r, config)
			} else {
				logger.Error("Timeout reached, server did not wake up. Aborting request.")
				http.Error(w, "Server did not come online in time", http.StatusGatewayTimeout)
			}
		} else {
			logger.Debug("Server is already online, handling domain proxy...")
			handleDomainProxy(w, r, config)
		}
	} else {
		logger.Debug("No wake-up endpoint matched, handling domain proxy...")
		handleDomainProxy(w, r, config)
	}
}
func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
