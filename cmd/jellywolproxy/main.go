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
	"github.com/Stoufiler/JellyWolProxy/internal/dashboard"
	"github.com/Stoufiler/JellyWolProxy/internal/handlers"
	"github.com/Stoufiler/JellyWolProxy/internal/health"
	"github.com/Stoufiler/JellyWolProxy/internal/jellyfin"
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

func loadConfig(log *logrus.Logger, configPath string) config.Config {
	viper.SetConfigFile(configPath)
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

	return cfg
}

func setupLogLevel(log *logrus.Logger, cfg config.Config, flagLevel string) {
	finalLogLevelStr := "Info"
	if cfg.LogLevel != "" {
		finalLogLevelStr = cfg.LogLevel
	}
	if flagLevel != "" {
		finalLogLevelStr = flagLevel
	}

	level, err := logrus.ParseLevel(finalLogLevelStr)
	if err != nil {
		log.Warnf("Invalid log level '%s', falling back to 'info'", finalLogLevelStr)
		level = logrus.InfoLevel
	}
	log.SetLevel(level)
}

func setupDashboardRoutes(r *mux.Router, cfg config.Config, log *logrus.Logger, serverState *server_state.ServerState) {
	dashboardRouter := r.PathPrefix("/status").Subrouter()

	if cfg.DashboardOIDC.Enabled {
		oidcProvider, err := dashboard.NewOIDCProvider(cfg.DashboardOIDC, log)
		if err != nil {
			log.Fatalf("Failed to initialize OIDC provider: %v", err)
		}
		dashboardRouter.HandleFunc("/callback", oidcProvider.CallbackHandler())
		dashboardRouter.HandleFunc("/logout", oidcProvider.LogoutHandler())
		dashboardRouter.Handle("", oidcProvider.AuthMiddleware(dashboard.StatusPageHandler(log)))
		dashboardRouter.Handle("/", oidcProvider.AuthMiddleware(dashboard.StatusPageHandler(log)))
		dashboardRouter.Handle("/api", oidcProvider.AuthMiddleware(dashboard.StatusAPIHandler(log, serverState)))
		dashboardRouter.Handle("/logs", oidcProvider.AuthMiddleware(dashboard.LogStreamHandler(log)))
		log.Info("Dashboard SSO authentication enabled")
	} else {
		dashboardRouter.HandleFunc("", dashboard.StatusPageHandler(log))
		dashboardRouter.HandleFunc("/", dashboard.StatusPageHandler(log))
		dashboardRouter.HandleFunc("/api", dashboard.StatusAPIHandler(log, serverState))
		dashboardRouter.HandleFunc("/logs", dashboard.LogStreamHandler(log))
		log.Info("Dashboard available without authentication")
	}
}

func setupCache(log *logrus.Logger, cfg config.Config) *cache.ResponseCache {
	if !cfg.CacheEnabled {
		log.Info("Response cache disabled")
		return nil
	}

	cacheTTL := time.Duration(cfg.CacheTTLSeconds) * time.Second
	if cfg.CacheTTLSeconds <= 0 {
		cacheTTL = 5 * time.Minute
	}
	responseCache := cache.NewResponseCache(cacheTTL)
	log.Infof("Response cache enabled with TTL of %v", cacheTTL)
	return responseCache
}

func runSignalLoop(log *logrus.Logger, hotConfig *config.HotReloadableConfig, stop, reload chan os.Signal) {
	for {
		select {
		case <-reload:
			handleConfigReload(log, hotConfig)
		case <-stop:
			log.Info("Shutdown signal received, starting graceful shutdown...")
			return
		}
	}
}

func handleConfigReload(log *logrus.Logger, hotConfig *config.HotReloadableConfig) {
	log.Info("SIGHUP received, reloading configuration...")
	if err := hotConfig.Reload(); err != nil {
		log.Errorf("Failed to reload config: %v", err)
		return
	}
	newCfg := hotConfig.Get()
	if newCfg.LogLevel != "" {
		level, err := logrus.ParseLevel(newCfg.LogLevel)
		if err == nil {
			log.SetLevel(level)
			log.Infof("Log level updated to %s", level.String())
		}
	}
	log.Info("Configuration reloaded. Some changes may require restart.")
}

func main() {
	log := logger.InitLogger("Info")
	log.AddHook(dashboard.NewLogrusHook())

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

	cfg := loadConfig(log, *configPath)
	setupLogLevel(log, cfg, *logLevelFlag)
	logger.SetLogFile(log, cfg.LogFile)

	log.Info("Configuration successfully loaded")
	log.Infof("Log level set to %s", log.GetLevel().String())

	// Initialize Jellyfin client
	jellyfinClient := jellyfin.NewClient(cfg, log)
	dashboard.SetJellyfinClient(jellyfinClient)
	log.Info("Jellyfin client initialized for session monitoring")

	serverState := &server_state.ServerState{}
	checker := &services.ConcreteServerStateChecker{}
	waker := &services.ConcreteWaker{}
	responseCache := setupCache(log, cfg)

	r := mux.NewRouter()
	r.HandleFunc("/health", health.HealthHandler)
	r.HandleFunc("/health/ready", health.ReadinessHandler(log, &cfg, checker))
	r.Handle("/metrics", promhttp.Handler())
	r.HandleFunc("/ping", handlers.PingHandler)

	setupDashboardRoutes(r, cfg, log, serverState)

	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.Handler(w, r, log, cfg, serverState, checker, waker)
	})

	var wrappedHandler http.Handler = mainHandler
	wrappedHandler = middlewares.RequestLoggerMiddleware(log, wrappedHandler)
	if responseCache != nil {
		wrappedHandler = middlewares.CacheMiddleware(log, responseCache, wrappedHandler)
	}
	wrappedHandler = middlewares.NetworkStatsMiddleware(wrappedHandler)
	wrappedHandler = middlewares.MetricsMiddleware(wrappedHandler)
	r.PathPrefix("/").Handler(wrappedHandler)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: r,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	reload := make(chan os.Signal, 1)
	signal.Notify(reload, syscall.SIGHUP)

	hotConfig := config.NewHotReloadableConfig(cfg, log)

	go func() {
		log.Infof("Starting app on port %d..", *port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	log.Info("Server started successfully, waiting for shutdown signal...")
	log.Info("Send SIGHUP to reload configuration")

	runSignalLoop(log, hotConfig, stop, reload)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	} else {
		log.Info("Server stopped gracefully")
	}
}
