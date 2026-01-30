package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Stoufiler/JellyWolProxy/internal/cache"
	"github.com/Stoufiler/JellyWolProxy/internal/config"
	"github.com/Stoufiler/JellyWolProxy/internal/handlers"
	"github.com/Stoufiler/JellyWolProxy/internal/health"
	"github.com/Stoufiler/JellyWolProxy/internal/logger"
	"github.com/Stoufiler/JellyWolProxy/internal/middlewares"
	"github.com/Stoufiler/JellyWolProxy/internal/server_state"
	"github.com/Stoufiler/JellyWolProxy/internal/services"
	"github.com/Stoufiler/JellyWolProxy/internal/upgrade"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var version = "0.0.1" // This will be replaced by the build process

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

	// Set environment variable bindings for sensitive data
	_ = viper.BindEnv("apiKey", "JELLYFIN_API_KEY")
	_ = viper.BindEnv("macAddress", "SERVER_MAC_ADDRESS")
	_ = viper.BindEnv("jellyfinUrl", "JELLYFIN_URL")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var cfg config.Config
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

	// Initialize cache if enabled
	var responseCache *cache.ResponseCache
	if cfg.CacheEnabled {
		cacheTTL := time.Duration(cfg.CacheTTLSeconds) * time.Second
		if cfg.CacheTTLSeconds <= 0 {
			cacheTTL = 5 * time.Minute // Default 5 minutes
		}
		responseCache = cache.NewResponseCache(cacheTTL)
		log.Infof("Response cache enabled with TTL of %v", cacheTTL)
	} else {
		log.Info("Response cache disabled")
	}

	r := mux.NewRouter()

	r.HandleFunc("/health", health.HealthHandler)
	r.HandleFunc("/health/ready", health.ReadinessHandler(log, &cfg, checker))
	r.Handle("/metrics", promhttp.Handler())
	r.HandleFunc("/ping", handlers.PingHandler)

	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.Handler(w, r, log, cfg, serverState, checker, waker, waiter)
	})

	// Apply middlewares conditionally
	var wrappedHandler http.Handler = mainHandler
	wrappedHandler = middlewares.RequestLoggerMiddleware(log, wrappedHandler)
	if responseCache != nil {
		wrappedHandler = middlewares.CacheMiddleware(log, responseCache, wrappedHandler)
	}
	wrappedHandler = middlewares.MetricsMiddleware(wrappedHandler)

	r.PathPrefix("/").Handler(wrappedHandler)

	serverAddress := fmt.Sprintf(":%d", *port)
	log.Infof("Starting app on port %d..", *port)

	srv := &http.Server{
		Addr:    serverAddress,
		Handler: r,
	}

	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Run server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	log.Info("Server started successfully, waiting for shutdown signal...")

	// Wait for interrupt signal
	<-stop
	log.Info("Shutdown signal received, starting graceful shutdown...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	} else {
		log.Info("Server stopped gracefully")
	}
}
