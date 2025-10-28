package util

import "testing"

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		pattern  string
		want     bool
	}{
		{"exact match", "/api/users", "/api/users", false}, // Note: matchesPattern is for wildcard patterns
		{"simple wildcard", "/api/users/123/profile", "/api/users/*/profile", true},
		{"no match", "/api/posts", "/api/users/*", false},
		{"suffix match only", "/other/api/users/123/profile", "/api/users/*/profile", false},
		{"prefix match only", "/api/users/123/settings", "/api/users/*/profile", false},
		{"invalid pattern", "/api/users/123", "/api/*/*", false},
		{"empty endpoint", "", "/api/*", false},
		{"empty pattern", "/api/users", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesPattern(tt.endpoint, tt.pattern); got != tt.want {
				t.Errorf("matchesPattern() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldWakeServer(t *testing.T) {
	tests := []struct {
		name            string
		endpoint        string
		wakeUpEndpoints []string
		want            bool
	}{
		{
			name:            "exact match in endpoints",
			endpoint:        "/api/play",
			wakeUpEndpoints: []string{"/api/stop", "/api/play"},
			want:            true,
		},
		{
			name:            "wildcard match in endpoints",
			endpoint:        "/videos/12345/stream.m3u8",
			wakeUpEndpoints: []string{"/videos/*/stream.m3u8"},
			want:            true,
		},
		{
			name:            "no match",
			endpoint:        "/api/status",
			wakeUpEndpoints: []string{"/api/play", "/videos/*"},
			want:            false,
		},
		{
			name:            "empty endpoints list",
			endpoint:        "/api/play",
			wakeUpEndpoints: []string{},
			want:            false,
		},
		{
			name:            "empty endpoint string",
			endpoint:        "",
			wakeUpEndpoints: []string{"/api/play"},
			want:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldWakeServer(tt.endpoint, tt.wakeUpEndpoints); got != tt.want {
				t.Errorf("ShouldWakeServer() = %v, want %v", got, tt.want)
			}
		})
	}
}
