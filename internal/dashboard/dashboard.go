package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Stoufiler/JellyWolProxy/internal/jellyfin"
	"github.com/Stoufiler/JellyWolProxy/internal/server_state"
	"github.com/sirupsen/logrus"
)

// Stats holds the dashboard statistics
type Stats struct {
	mu sync.RWMutex

	// Server state
	ServerState   string    `json:"serverState"` // online, offline, waking
	LastWakeUp    time.Time `json:"lastWakeUp,omitempty"`
	LastOnline    time.Time `json:"lastOnline,omitempty"`
	WakeUpCount   int64     `json:"wakeUpCount"`
	AvgWakeUpTime float64   `json:"avgWakeUpTimeSeconds"`

	// Request stats
	TotalRequests int64 `json:"totalRequests"`
	CacheHits     int64 `json:"cacheHits"`
	CacheMisses   int64 `json:"cacheMisses"`

	// Network stats
	BytesIn             int64     `json:"bytesIn"`
	BytesOut            int64     `json:"bytesOut"`
	LastBytesIn         int64     `json:"-"`
	LastBytesOut        int64     `json:"-"`
	BandwidthIn         int64     `json:"bandwidthIn"`  // bytes per second
	BandwidthOut        int64     `json:"bandwidthOut"` // bytes per second
	LastBandwidthUpdate time.Time `json:"-"`

	// Uptime
	StartTime time.Time `json:"startTime"`

	// Internal for calculating average
	totalWakeUpTime time.Duration
}

// Global stats instance
var globalStats = &Stats{
	StartTime: time.Now(),
}

// Global Jellyfin client
var jellyfinClient *jellyfin.Client

// SetJellyfinClient sets the global Jellyfin client
func SetJellyfinClient(client *jellyfin.Client) {
	jellyfinClient = client
}

// GetStats returns the global stats instance
func GetStats() *Stats {
	return globalStats
}

// RecordRequest increments the request counter
func (s *Stats) RecordRequest() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalRequests++
}

// RecordCacheHit increments the cache hit counter
func (s *Stats) RecordCacheHit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CacheHits++
}

// RecordCacheMiss increments the cache miss counter
func (s *Stats) RecordCacheMiss() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CacheMisses++
}

// RecordBytes records network bytes in/out
func (s *Stats) RecordBytes(bytesIn, bytesOut int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BytesIn += bytesIn
	s.BytesOut += bytesOut
}

// UpdateBandwidth calculates current bandwidth based on bytes since last update
func (s *Stats) UpdateBandwidth() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if s.LastBandwidthUpdate.IsZero() {
		s.LastBandwidthUpdate = now
		s.LastBytesIn = s.BytesIn
		s.LastBytesOut = s.BytesOut
		return
	}

	elapsed := now.Sub(s.LastBandwidthUpdate).Seconds()
	if elapsed >= 1.0 {
		s.BandwidthIn = int64(float64(s.BytesIn-s.LastBytesIn) / elapsed)
		s.BandwidthOut = int64(float64(s.BytesOut-s.LastBytesOut) / elapsed)
		s.LastBytesIn = s.BytesIn
		s.LastBytesOut = s.BytesOut
		s.LastBandwidthUpdate = now
	}
}

// RecordWakeUp records a wake-up event
func (s *Stats) RecordWakeUp() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastWakeUp = time.Now()
	s.WakeUpCount++
}

// RecordWakeUpComplete records that a wake-up completed
func (s *Stats) RecordWakeUpComplete(duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastOnline = time.Now()
	s.totalWakeUpTime += duration
	if s.WakeUpCount > 0 {
		s.AvgWakeUpTime = s.totalWakeUpTime.Seconds() / float64(s.WakeUpCount)
	}
}

// SetServerState updates the server state
func (s *Stats) SetServerState(state string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ServerState = state
	if state == "online" {
		s.LastOnline = time.Now()
	}
}

// GetSnapshot returns a copy of the current stats
func (s *Stats) GetSnapshot() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Stats{
		ServerState:   s.ServerState,
		LastWakeUp:    s.LastWakeUp,
		LastOnline:    s.LastOnline,
		WakeUpCount:   s.WakeUpCount,
		AvgWakeUpTime: s.AvgWakeUpTime,
		TotalRequests: s.TotalRequests,
		CacheHits:     s.CacheHits,
		CacheMisses:   s.CacheMisses,
		BytesIn:       s.BytesIn,
		BytesOut:      s.BytesOut,
		BandwidthIn:   s.BandwidthIn,
		BandwidthOut:  s.BandwidthOut,
		StartTime:     s.StartTime,
	}
}

// StatusResponse is the JSON response for the status endpoint
type StatusResponse struct {
	ServerState   string             `json:"serverState"`
	Uptime        string             `json:"uptime"`
	StartTime     time.Time          `json:"startTime"`
	LastWakeUp    string             `json:"lastWakeUp,omitempty"`
	LastOnline    string             `json:"lastOnline,omitempty"`
	WakeUpCount   int64              `json:"wakeUpCount"`
	AvgWakeUpTime float64            `json:"avgWakeUpTimeSeconds"`
	TotalRequests int64              `json:"totalRequests"`
	CacheHits     int64              `json:"cacheHits"`
	CacheMisses   int64              `json:"cacheMisses"`
	CacheHitRate  float64            `json:"cacheHitRate"`
	BytesIn       int64              `json:"bytesIn"`
	BytesOut      int64              `json:"bytesOut"`
	BandwidthIn   int64              `json:"bandwidthIn"`
	BandwidthOut  int64              `json:"bandwidthOut"`
	System        SystemInfo         `json:"system"`
	Process       ProcessInfo        `json:"process"`
	Sessions      []jellyfin.Session `json:"sessions"`
}

// StatusAPIHandler returns JSON status data
func StatusAPIHandler(logger *logrus.Logger, serverState *server_state.ServerState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Update bandwidth calculation
		globalStats.UpdateBandwidth()
		stats := globalStats.GetSnapshot()

		// Determine current server state
		var currentState string
		switch {
		case serverState.IsWakingUp():
			currentState = "waking"
		case stats.ServerState == "online":
			currentState = "online"
		default:
			currentState = "offline"
		}

		// Calculate cache hit rate
		var cacheHitRate float64
		totalCacheReqs := stats.CacheHits + stats.CacheMisses
		if totalCacheReqs > 0 {
			cacheHitRate = float64(stats.CacheHits) / float64(totalCacheReqs) * 100
		}

		response := StatusResponse{
			ServerState:   currentState,
			Uptime:        time.Since(stats.StartTime).Round(time.Second).String(),
			StartTime:     stats.StartTime,
			WakeUpCount:   stats.WakeUpCount,
			AvgWakeUpTime: stats.AvgWakeUpTime,
			TotalRequests: stats.TotalRequests,
			CacheHits:     stats.CacheHits,
			CacheMisses:   stats.CacheMisses,
			CacheHitRate:  cacheHitRate,
			BytesIn:       stats.BytesIn,
			BytesOut:      stats.BytesOut,
			BandwidthIn:   stats.BandwidthIn,
			BandwidthOut:  stats.BandwidthOut,
			System:        GetSystemInfo(),
			Process:       GetProcessInfo(stats.StartTime),
			Sessions:      []jellyfin.Session{},
		}

		// Get Jellyfin sessions if client is configured
		if jellyfinClient != nil {
			if sessions, err := jellyfinClient.GetActiveSessions(); err == nil {
				response.Sessions = sessions
			} else {
				logger.Debugf("Failed to get Jellyfin sessions: %v", err)
			}
		}

		if !stats.LastWakeUp.IsZero() {
			response.LastWakeUp = stats.LastWakeUp.Format(time.RFC3339)
		}
		if !stats.LastOnline.IsZero() {
			response.LastOnline = stats.LastOnline.Format(time.RFC3339)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Errorf("Failed to encode status response: %v", err)
		}
	}
}

// StatusPageHandler returns the HTML dashboard page
func StatusPageHandler(logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, statusPageHTML)
	}
}
