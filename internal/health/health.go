package health

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Stoufiler/JellyWolProxy/internal/config"
	"github.com/Stoufiler/JellyWolProxy/internal/services"
	"github.com/sirupsen/logrus"
)

type HealthResponse struct {
	Status    string           `json:"status"`
	Timestamp time.Time        `json:"timestamp"`
	Version   string           `json:"version"`
	Checks    map[string]Check `json:"checks"`
}

type Check struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

const (
	StatusUp   = "UP"
	StatusDown = "DOWN"
)

// HealthHandler returns basic health status
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    StatusUp,
		Timestamp: time.Now(),
		Version:   "1.0.0", // TODO: Get from build info
		Checks:    make(map[string]Check),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ReadinessHandler checks if all components are ready
func ReadinessHandler(logger *logrus.Logger, config *config.Config, checker services.ServerStateChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checks := make(map[string]Check)
		status := StatusUp

		// Check Jellyfin server connectivity
		serverAddress := config.ForwardIp + ":" + strconv.Itoa(config.ForwardPort)
		if !checker.IsServerUp(logger, serverAddress) {
			checks["jellyfin"] = Check{
				Status:  StatusDown,
				Message: "Jellyfin server is not reachable",
			}
			status = StatusDown
		} else {
			checks["jellyfin"] = Check{
				Status:  StatusUp,
				Message: "Jellyfin server is reachable",
			}
		}

		response := HealthResponse{
			Status:    status,
			Timestamp: time.Now(),
			Version:   "1.0.0", // TODO: Get from build info
			Checks:    checks,
		}

		w.Header().Set("Content-Type", "application/json")
		if status == StatusDown {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
