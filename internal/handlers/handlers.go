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
	"github.com/Stoufiler/JellyWolProxy/internal/websocket"
	"github.com/sirupsen/logrus"
)

func handleDomainProxy(w http.ResponseWriter, r *http.Request, cfg config.Config, logger *logrus.Logger) {
	targetHost := fmt.Sprintf("%s:%d", cfg.ForwardIp, cfg.ForwardPort)

	// Handle WebSocket connections specially
	if websocket.IsWebSocketRequest(r) {
		logger.Debug("WebSocket upgrade request detected, proxying WebSocket connection")
		if err := websocket.ProxyWebSocket(w, r, targetHost, logger); err != nil {
			logger.Errorf("WebSocket proxy error: %v", err)
			http.Error(w, "WebSocket proxy error", http.StatusBadGateway)
		}
		return
	}

	targetURL := &url.URL{
		Scheme: "http",
		Host:   targetHost,
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
				if parsedLocation.Host == fmt.Sprintf("%s:%d", cfg.ForwardIp, cfg.ForwardPort) {
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

func Handler(w http.ResponseWriter, r *http.Request, logger *logrus.Logger, cfg config.Config, serverState *server_state.ServerState, checker services.ServerStateChecker, waker services.Waker) {
	logger.Debug("Request received for path: ", r.URL.Path)
	if util.ShouldWakeServer(r.URL.Path, cfg.WakeUpEndpoints) {
		serverAddress := fmt.Sprintf("%s:%d", cfg.WakeUpIp, cfg.WakeUpPort)
		logger.Debug("Wake-up endpoint matched, checking server status...")
		if !checker.IsServerUp(logger, serverAddress) {
			// Server is down - trigger WOL asynchronously and return 503 for immediate retry
			if waker.WakeServer(logger, cfg.MacAddress, cfg.BroadcastAddress, cfg, serverState) {
				// WOL packet sent, start background wake process
				go func() {
					defer serverState.DoneWakingUp()
					logger.Info("Server is offline, Wake On LAN packet sent")
				}()
			} else {
				logger.Info("Server is already waking up...")
			}

			// Return 503 Service Unavailable with Retry-After header
			// This allows clients (Infuse, Jellyfin) to automatically retry
			w.Header().Set("Retry-After", "30")
			http.Error(w, "Server is waking up, please retry in 30 seconds", http.StatusServiceUnavailable)
			return
		}
		logger.Debug("Server is already online, handling domain proxy...")
		handleDomainProxy(w, r, cfg, logger)
	} else {
		logger.Debug("No wake-up endpoint matched, handling domain proxy...")
		handleDomainProxy(w, r, cfg, logger)
	}
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
