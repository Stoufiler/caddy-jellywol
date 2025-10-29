package config

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				WakeUpIp:    "192.168.1.1",
				ForwardIp:   "10.0.0.1",
				MacAddress:  "00:11:22:33:44:55",
				WakeUpPort:  80,
				ForwardPort: 8096,
			},
			wantErr: false,
		},
		{
			name: "invalid wake up ip",
			config: Config{
				WakeUpIp: "not-an-ip",
			},
			wantErr: true,
		},
		{
			name: "invalid forward ip",
			config: Config{
				WakeUpIp:  "192.168.1.1",
				ForwardIp: "not-an-ip",
			},
			wantErr: true,
		},
		{
			name: "invalid mac address",
			config: Config{
				WakeUpIp:   "192.168.1.1",
				ForwardIp:  "10.0.0.1",
				MacAddress: "not-a-mac",
			},
			wantErr: true,
		},
		{
			name: "invalid wake up port",
			config: Config{
				WakeUpIp:   "192.168.1.1",
				ForwardIp:  "10.0.0.1",
				MacAddress: "00:11:22:33:44:55",
				WakeUpPort: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid forward port",
			config: Config{
				WakeUpIp:    "192.168.1.1",
				ForwardIp:   "10.0.0.1",
				MacAddress:  "00:11:22:33:44:55",
				WakeUpPort:  80,
				ForwardPort: 70000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
