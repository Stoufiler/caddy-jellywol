package config

import (
	"fmt"
	"net"
)

// OIDCConfig holds OIDC configuration for Pocket-ID
type OIDCConfig struct {
	Enabled      bool   `mapstructure:"enabled" json:"enabled"`
	IssuerURL    string `mapstructure:"issuer_url" json:"issuer_url"`
	ClientID     string `mapstructure:"client_id" json:"client_id"`
	ClientSecret string `mapstructure:"client_secret" json:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url" json:"redirect_url"`
	Scopes       string `mapstructure:"scopes" json:"scopes"`
}

type Config struct {
	JellyfinUrl          string     `mapstructure:"jellyfinUrl"`
	ApiKey               string     `mapstructure:"apiKey"`
	MacAddress           string     `mapstructure:"macAddress"`
	BroadcastAddress     string     `mapstructure:"broadcastAddress"`
	WakeUpPort           int        `mapstructure:"wakeUpPort"`
	WakeUpIp             string     `mapstructure:"wakeUpIp"`
	ForwardIp            string     `mapstructure:"forwardIp"`
	ForwardPort          int        `mapstructure:"forwardPort"`
	WakeUpEndpoints      []string   `mapstructure:"wakeUpEndpoints"`
	ServerWakeUpTimeout  int        `mapstructure:"serverWakeUpTimeout"`
	ServerWakeUpTicker   int        `mapstructure:"serverWakeUpTicker"`
	PostPingDelaySeconds int        `mapstructure:"postPingDelaySeconds"`
	LogLevel             string     `mapstructure:"logLevel"`
	LogFile              string     `mapstructure:"logFile"`
	CacheEnabled         bool       `mapstructure:"cacheEnabled"`
	CacheTTLSeconds      int        `mapstructure:"cacheTTLSeconds"`
	DashboardOIDC        OIDCConfig `mapstructure:"dashboardOIDC"`
}

func (c *Config) Validate() error {
	if c.WakeUpIp != "" {
		if net.ParseIP(c.WakeUpIp) == nil {
			return fmt.Errorf("invalid wakeUpIp: %s", c.WakeUpIp)
		}
	}
	if c.ForwardIp != "" {
		if net.ParseIP(c.ForwardIp) == nil {
			return fmt.Errorf("invalid forwardIp: %s", c.ForwardIp)
		}
	}
	if c.MacAddress != "" {
		if _, err := net.ParseMAC(c.MacAddress); err != nil {
			return fmt.Errorf("invalid macAddress: %s", c.MacAddress)
		}
	}
	if c.WakeUpPort < 1 || c.WakeUpPort > 65535 {
		return fmt.Errorf("invalid wakeUpPort: %d", c.WakeUpPort)
	}
	if c.ForwardPort < 1 || c.ForwardPort > 65535 {
		return fmt.Errorf("invalid forwardPort: %d", c.ForwardPort)
	}
	return nil
}
