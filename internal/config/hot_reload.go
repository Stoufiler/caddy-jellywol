package config

import (
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// HotReloadableConfig wraps Config with hot-reload capabilities
type HotReloadableConfig struct {
	mu     sync.RWMutex
	config Config
	logger *logrus.Logger
}

// NewHotReloadableConfig creates a new hot-reloadable config wrapper
func NewHotReloadableConfig(cfg Config, logger *logrus.Logger) *HotReloadableConfig {
	return &HotReloadableConfig{
		config: cfg,
		logger: logger,
	}
}

// Get returns the current configuration (thread-safe)
func (h *HotReloadableConfig) Get() Config {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.config
}

// GetPtr returns a pointer to the current configuration (thread-safe)
func (h *HotReloadableConfig) GetPtr() *Config {
	h.mu.RLock()
	defer h.mu.RUnlock()
	cfg := h.config
	return &cfg
}

// Reload reloads the configuration from file
func (h *HotReloadableConfig) Reload() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Re-read the config file
	if err := viper.ReadInConfig(); err != nil {
		h.logger.Errorf("Failed to reload config: %v", err)
		return err
	}

	var newConfig Config
	if err := viper.Unmarshal(&newConfig); err != nil {
		h.logger.Errorf("Failed to unmarshal config: %v", err)
		return err
	}

	if err := newConfig.Validate(); err != nil {
		h.logger.Errorf("Invalid configuration: %v", err)
		return err
	}

	// Log what changed
	h.logConfigChanges(newConfig)

	h.config = newConfig
	h.logger.Info("Configuration reloaded successfully")
	return nil
}

// logConfigChanges logs the differences between old and new config
func (h *HotReloadableConfig) logConfigChanges(newConfig Config) {
	old := h.config

	if old.LogLevel != newConfig.LogLevel {
		h.logger.Infof("LogLevel changed: %s -> %s", old.LogLevel, newConfig.LogLevel)
	}
	if old.CacheEnabled != newConfig.CacheEnabled {
		h.logger.Infof("CacheEnabled changed: %v -> %v", old.CacheEnabled, newConfig.CacheEnabled)
	}
	if old.CacheTTLSeconds != newConfig.CacheTTLSeconds {
		h.logger.Infof("CacheTTLSeconds changed: %d -> %d", old.CacheTTLSeconds, newConfig.CacheTTLSeconds)
	}
	if old.ServerWakeUpTimeout != newConfig.ServerWakeUpTimeout {
		h.logger.Infof("ServerWakeUpTimeout changed: %d -> %d", old.ServerWakeUpTimeout, newConfig.ServerWakeUpTimeout)
	}
	if old.ServerWakeUpTicker != newConfig.ServerWakeUpTicker {
		h.logger.Infof("ServerWakeUpTicker changed: %d -> %d", old.ServerWakeUpTicker, newConfig.ServerWakeUpTicker)
	}
	if old.PostPingDelaySeconds != newConfig.PostPingDelaySeconds {
		h.logger.Infof("PostPingDelaySeconds changed: %d -> %d", old.PostPingDelaySeconds, newConfig.PostPingDelaySeconds)
	}
	// Note: Some fields like MacAddress, IPs cannot be changed at runtime safely
	if old.MacAddress != newConfig.MacAddress {
		h.logger.Warn("MacAddress changed - requires restart to take effect")
	}
	if old.ForwardIp != newConfig.ForwardIp || old.ForwardPort != newConfig.ForwardPort {
		h.logger.Warn("Forward IP/Port changed - requires restart to take effect")
	}
	if old.WakeUpIp != newConfig.WakeUpIp || old.WakeUpPort != newConfig.WakeUpPort {
		h.logger.Warn("WakeUp IP/Port changed - requires restart to take effect")
	}
}

// OnConfigChange registers a callback for config changes (for viper.WatchConfig)
func (h *HotReloadableConfig) OnConfigChange(callback func(Config)) {
	viper.OnConfigChange(func(e fsnotify.Event) {
		h.logger.Info("Config file changed, reloading...")
		if err := h.Reload(); err != nil {
			h.logger.Errorf("Failed to reload config on change: %v", err)
			return
		}
		if callback != nil {
			callback(h.Get())
		}
	})
}
