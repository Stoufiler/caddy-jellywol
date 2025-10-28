package config

import (
	"fmt"
	"net"
)

type Config struct {
	JellyfinUrl         string   `mapstructure:"jellyfinUrl"`
	ApiKey              string   `mapstructure:"apiKey"`
	MacAddress          string   `mapstructure:"macAddress"`
	BroadcastAddress    string   `mapstructure:"broadcastAddress"`
	WakeUpPort          int      `mapstructure:"wakeUpPort"`
	WakeUpIp            string   `mapstructure:"wakeUpIp"`
	ForwardIp           string   `mapstructure:"forwardIp"`
	ForwardPort         int      `mapstructure:"forwardPort"`
	WakeUpEndpoints     []string `mapstructure:"wakeUpEndpoints"`
	ServerWakeUpTimeout int      `mapstructure:"serverWakeUpTimeout"`
	ServerWakeUpTicker  int      `mapstructure:"serverWakeUpTicker"`
	LogLevel            string   `mapstructure:"logLevel"`
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