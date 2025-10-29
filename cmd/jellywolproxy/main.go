package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/Stoufiler/JellyWolProxy/internal/config"
	"github.com/Stoufiler/JellyWolProxy/internal/handlers"
	"github.com/Stoufiler/JellyWolProxy/internal/health"
	"github.com/Stoufiler/JellyWolProxy/internal/logger"
	"github.com/Stoufiler/JellyWolProxy/internal/middlewares"
	"github.com/Stoufiler/JellyWolProxy/internal/server_state"
	"github.com/Stoufiler/JellyWolProxy/internal/services"
	"github.com/Stoufiler/JellyWolProxy/internal/upgrade"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const version = "0.0.1" // This will be replaced by the build process

func main() {
	log := logger.InitLogger("Info")

	upgradeFlag := flag.Bool("upgrade", false, "Upgrade the application")
	logLevelFlag := flag.String("log-level", "", "Log level (e.g., Debug, Info, Warn, Error)")
	configPath := flag.String("config", "config.json", "path to config file")
	port := flag.Int("port", 3881, "port to run the server on")
	versionFlag := flag.Bool("version", false, "Print the current version")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		return
	}

	if *upgradeFlag {
		upgrade.RunUpgrade(version)
		return
	}

	viper.SetConfigFile(*configPath)
	viper.AutomaticEnv()

	var cfg config.Config
	if err := viper.ReadInConfig(); err != nil {
		log.Warnf("Error reading config file: %v", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
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
		log.Warnf("Invalid log level '%s', falling back to 'info'", finalLogLevelStr)
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	logger.SetLogFile(log, cfg.LogFile)

	log.Info("Configuration successfully loaded")
	log.Infof("Log level set to %s", log.GetLevel().String())

	serverState := &server_state.ServerState{}

	// Create concrete service implementations
	checker := &services.ConcreteServerStateChecker{}
	waker := &services.ConcreteWaker{}
	waiter := &services.ConcreteServerWaiter{}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", health.HealthHandler)
	mux.HandleFunc("/health/ready", health.ReadinessHandler(log, &cfg, checker))
	mux.Handle("/metrics", promhttp.Handler())

	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.Handler(w, r, log, cfg, serverState, checker, waker, waiter)
	})

	mux.Handle("/", middlewares.MetricsMiddleware(
		middlewares.RequestLoggerMiddleware(log, mainHandler),
	))

	mux.HandleFunc("/ping", handlers.PingHandler)

	loggedMux := middlewares.RequestLoggerMiddleware(log, mux)

	serverAddress := fmt.Sprintf(":%d", *port)
	log.Infof("Starting app on port %d..", *port)
	log.Fatal(http.ListenAndServe(serverAddress, loggedMux))
}
