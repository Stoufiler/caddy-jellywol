package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/StephanGR/JellyWolProxy/internal/config"
	"github.com/StephanGR/JellyWolProxy/internal/handlers"
	"github.com/StephanGR/JellyWolProxy/internal/health"
	"github.com/StephanGR/JellyWolProxy/internal/logger"
	"github.com/StephanGR/JellyWolProxy/internal/middlewares"
	"github.com/StephanGR/JellyWolProxy/internal/server_state"
	"github.com/StephanGR/JellyWolProxy/internal/services"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logger := logger.InitLogger("Info")

	logLevelFlag := flag.String("log-level", "", "Log level (e.g., Debug, Info, Warn, Error)")
	configPath := flag.String("config", "config.json", "path to config file")
	port := flag.Int("port", 3881, "port to run the server on")
	flag.Parse()

	viper.SetConfigFile(*configPath)
	viper.AutomaticEnv()

	var cfg config.Config
	if err := viper.ReadInConfig(); err != nil {
		logger.Warnf("Error reading config file: %v", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		logger.Fatalf("Unable to decode into struct: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		logger.Fatalf("Invalid configuration: %v", err)
	}

	// Determine final log level
	finalLogLevelStr := "Info" // Default
	if cfg.LogLevel != "" {
		finalLogLevelStr = cfg.LogLevel
	}
	if *logLevelFlag != "" {
		finalLogLevelStr = *logLevelFlag
	}

	// Set the final log level
	level, err := logrus.ParseLevel(finalLogLevelStr)
	if err != nil {
		logger.Warnf("Invalid log level '%s', falling back to 'info'", finalLogLevelStr)
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	logger.Info("Configuration successfully loaded")
	logger.Infof("Log level set to %s", logger.GetLevel().String())

	serverState := &server_state.ServerState{}

	// Create concrete service implementations
	checker := &services.ConcreteServerStateChecker{}
	waker := &services.ConcreteWaker{}
	waiter := &services.ConcreteServerWaiter{}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", health.HealthHandler)
	mux.HandleFunc("/health/ready", health.ReadinessHandler(logger, &cfg, checker))
	mux.Handle("/metrics", promhttp.Handler())

	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.Handler(w, r, logger, cfg, serverState, checker, waker, waiter)
	})

	mux.Handle("/", middlewares.MetricsMiddleware(
		middlewares.RequestLoggerMiddleware(logger, mainHandler),
	))

	mux.HandleFunc("/ping", handlers.PingHandler)

	loggedMux := middlewares.RequestLoggerMiddleware(logger, mux)

	serverAddress := fmt.Sprintf(":%d", *port)
	logger.Infof("Starting app on port %d..", *port)
	logger.Fatal(http.ListenAndServe(serverAddress, loggedMux))
}
