package jellywol

import (
	"testing"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func TestJellyWol_Validate(t *testing.T) {
	tests := []struct {
		name    string
		module  *JellyWol
		wantErr bool
	}{
		{
			name: "valid configuration",
			module: &JellyWol{
				Mac:      "aa:bb:cc:dd:ee:ff",
				PingIP:   "192.168.1.10",
				PingPort: 8096,
			},
			wantErr: false,
		},
		{
			name: "missing mac",
			module: &JellyWol{
				PingIP:   "192.168.1.10",
				PingPort: 8096,
			},
			wantErr: true,
		},
		{
			name: "missing ping ip",
			module: &JellyWol{
				Mac:      "aa:bb:cc:dd:ee:ff",
				PingPort: 8096,
			},
			wantErr: true,
		},
		{
			name: "invalid ping port",
			module: &JellyWol{
				Mac:      "aa:bb:cc:dd:ee:ff",
				PingIP:   "192.168.1.10",
				PingPort: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.module.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("JellyWol.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJellyWol_UnmarshalCaddyfile(t *testing.T) {
	input := `jellywol {
		mac aa:bb:cc:dd:ee:ff
		broadcast 192.168.1.255:9
		ping_ip 192.168.1.100
		ping_port 8096
		timeout 2s
	}`

	d := caddyfile.NewTestDispenser(input)
	j := &JellyWol{}

	if err := j.UnmarshalCaddyfile(d); err != nil {
		t.Fatalf("Failed to unmarshal Caddyfile: %v", err)
	}

	if j.Mac != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("Expected Mac to be aa:bb:cc:dd:ee:ff, got %v", j.Mac)
	}
	if j.Broadcast != "192.168.1.255:9" {
		t.Errorf("Expected Broadcast to be 192.168.1.255:9, got %v", j.Broadcast)
	}
	if j.PingIP != "192.168.1.100" {
		t.Errorf("Expected PingIP to be 192.168.1.100, got %v", j.PingIP)
	}
	if j.PingPort != 8096 {
		t.Errorf("Expected PingPort to be 8096, got %v", j.PingPort)
	}
	if j.Timeout != "2s" {
		t.Errorf("Expected Timeout to be 2s, got %v", j.Timeout)
	}
}

func TestJellyWol_State(t *testing.T) {
	j := &JellyWol{}

	if !j.wakingUp.CompareAndSwap(false, true) {
		t.Error("Expected to successfully swap wakingUp from false to true")
	}
	if j.wakingUp.CompareAndSwap(false, true) {
		t.Error("Expected CompareAndSwap to fail since it is already true")
	}

	j.wakingUp.Store(false)

	if !j.wakingUp.CompareAndSwap(false, true) {
		t.Error("Expected to successfully swap again after reset")
	}
}
