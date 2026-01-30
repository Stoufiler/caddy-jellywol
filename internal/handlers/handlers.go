package handlers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/Stoufiler/JellyWolProxy/internal/config"
	"github.com/Stoufiler/JellyWolProxy/internal/server_state"
	"github.com/Stoufiler/JellyWolProxy/internal/services"
	"github.com/Stoufiler/JellyWolProxy/internal/util"
	"github.com/sirupsen/logrus"
)

func handleDomainProxy(w http.ResponseWriter, r *http.Request, config config.Config) {
	targetURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", config.ForwardIp, config.ForwardPort),
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Capturer l'hôte original de la requête pour réécrire les redirections
	originalHost := r.Host
	if originalHost == "" {
		originalHost = r.Header.Get("Host")
	}

	// Modifier les headers de la requête
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetURL.Host
	}

	// Modifier la réponse pour réécrire les redirections Location
	proxy.ModifyResponse = func(resp *http.Response) error {
		if location := resp.Header.Get("Location"); location != "" {
			// Remplacer l'IP locale par l'hôte original
			if parsedLocation, err := url.Parse(location); err == nil {
				// Si la redirection pointe vers l'IP locale, la remplacer
				if parsedLocation.Host == fmt.Sprintf("%s:%d", config.ForwardIp, config.ForwardPort) {
					parsedLocation.Host = originalHost
					parsedLocation.Scheme = "http"
					resp.Header.Set("Location", parsedLocation.String())
				}
			}
		}
		return nil
	}

	proxy.ServeHTTP(w, r)
}

func Handler(w http.ResponseWriter, r *http.Request, logger *logrus.Logger, config config.Config, serverState *server_state.ServerState, checker services.ServerStateChecker, waker services.Waker, waiter services.ServerWaiter) {
	logger.Debug("Request received for path: ", r.URL.Path)
	if util.ShouldWakeServer(r.URL.Path, config.WakeUpEndpoints) {
		serverAddress := fmt.Sprintf("%s:%d", config.WakeUpIp, config.WakeUpPort)
		logger.Debug("Wake-up endpoint matched, checking server status...")
		if !checker.IsServerUp(logger, serverAddress) {
			// Send 102 Processing to tell the client to wait
			if flusher, ok := w.(http.Flusher); ok {
				w.WriteHeader(http.StatusProcessing)
				flusher.Flush()
			}

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
