package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

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

	// Uptime
	StartTime time.Time `json:"startTime"`

	// Internal for calculating average
	totalWakeUpTime time.Duration
}

// Global stats instance
var globalStats = &Stats{
	StartTime: time.Now(),
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
		StartTime:     s.StartTime,
	}
}

// StatusResponse is the JSON response for the status endpoint
type StatusResponse struct {
	ServerState   string    `json:"serverState"`
	Uptime        string    `json:"uptime"`
	StartTime     time.Time `json:"startTime"`
	LastWakeUp    string    `json:"lastWakeUp,omitempty"`
	LastOnline    string    `json:"lastOnline,omitempty"`
	WakeUpCount   int64     `json:"wakeUpCount"`
	AvgWakeUpTime float64   `json:"avgWakeUpTimeSeconds"`
	TotalRequests int64     `json:"totalRequests"`
	CacheHits     int64     `json:"cacheHits"`
	CacheMisses   int64     `json:"cacheMisses"`
	CacheHitRate  float64   `json:"cacheHitRate"`
}

// StatusAPIHandler returns JSON status data
func StatusAPIHandler(logger *logrus.Logger, serverState *server_state.ServerState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

const statusPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>JellyWolProxy Status</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            color: #e0e0e0;
            min-height: 100vh;
            padding: 20px;
        }
        .container { max-width: 1200px; margin: 0 auto; }
        h1 {
            text-align: center;
            color: #00d4aa;
            margin-bottom: 30px;
            font-size: 2.5em;
        }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 20px; }
        .card {
            background: rgba(255,255,255,0.05);
            border-radius: 12px;
            padding: 24px;
            border: 1px solid rgba(255,255,255,0.1);
            backdrop-filter: blur(10px);
        }
        .card h2 {
            color: #00d4aa;
            font-size: 0.9em;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-bottom: 12px;
        }
        .card .value {
            font-size: 2.2em;
            font-weight: 700;
            color: #fff;
        }
        .card .subtitle { color: #888; font-size: 0.85em; margin-top: 8px; }
        .status-online { color: #00d4aa !important; }
        .status-offline { color: #ff4757 !important; }
        .status-waking { color: #ffa502 !important; }
        .logs {
            background: #0a0a0f;
            border-radius: 8px;
            padding: 16px;
            height: 300px;
            overflow-y: auto;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.85em;
            margin-top: 20px;
        }
        .logs .log-entry { padding: 4px 0; border-bottom: 1px solid rgba(255,255,255,0.05); }
        .logs .log-time { color: #666; }
        .logs .log-info { color: #3498db; }
        .logs .log-warn { color: #f39c12; }
        .logs .log-error { color: #e74c3c; }
        .refresh-indicator {
            text-align: center;
            color: #666;
            font-size: 0.85em;
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>JellyWolProxy Status</h1>
        <div class="grid">
            <div class="card">
                <h2>Server State</h2>
                <div class="value" id="serverState">Loading...</div>
                <div class="subtitle" id="lastOnline"></div>
            </div>
            <div class="card">
                <h2>Uptime</h2>
                <div class="value" id="uptime">-</div>
                <div class="subtitle" id="startTime"></div>
            </div>
            <div class="card">
                <h2>Total Requests</h2>
                <div class="value" id="totalRequests">0</div>
            </div>
            <div class="card">
                <h2>Cache Hit Rate</h2>
                <div class="value" id="cacheHitRate">0%</div>
                <div class="subtitle" id="cacheStats"></div>
            </div>
            <div class="card">
                <h2>Wake-up Count</h2>
                <div class="value" id="wakeUpCount">0</div>
                <div class="subtitle" id="lastWakeUp"></div>
            </div>
            <div class="card">
                <h2>Avg Wake-up Time</h2>
                <div class="value" id="avgWakeUpTime">-</div>
            </div>
        </div>
        <div class="card" style="margin-top: 20px;">
            <h2>Live Logs</h2>
            <div class="logs" id="logs">
                <div class="log-entry"><span class="log-time">Connecting...</span></div>
            </div>
        </div>
        <div class="refresh-indicator">Auto-refresh every 5 seconds</div>
    </div>
    <script>
        function updateStatus() {
            fetch('/status/api')
                .then(r => r.json())
                .then(data => {
                    const stateEl = document.getElementById('serverState');
                    stateEl.textContent = data.serverState.toUpperCase();
                    stateEl.className = 'value status-' + data.serverState;

                    document.getElementById('uptime').textContent = data.uptime;
                    document.getElementById('startTime').textContent = 'Started: ' + new Date(data.startTime).toLocaleString();
                    document.getElementById('totalRequests').textContent = data.totalRequests.toLocaleString();
                    document.getElementById('cacheHitRate').textContent = data.cacheHitRate.toFixed(1) + '%';
                    document.getElementById('cacheStats').textContent = data.cacheHits + ' hits / ' + data.cacheMisses + ' misses';
                    document.getElementById('wakeUpCount').textContent = data.wakeUpCount;
                    document.getElementById('avgWakeUpTime').textContent = data.avgWakeUpTimeSeconds > 0
                        ? data.avgWakeUpTimeSeconds.toFixed(1) + 's' : '-';

                    if (data.lastWakeUp) {
                        document.getElementById('lastWakeUp').textContent = 'Last: ' + new Date(data.lastWakeUp).toLocaleString();
                    }
                    if (data.lastOnline) {
                        document.getElementById('lastOnline').textContent = 'Last seen: ' + new Date(data.lastOnline).toLocaleString();
                    }
                })
                .catch(err => console.error('Failed to fetch status:', err));
        }

        updateStatus();
        setInterval(updateStatus, 5000);

        // SSE for live logs
        const logsEl = document.getElementById('logs');
        const evtSource = new EventSource('/status/logs');
        evtSource.onmessage = function(e) {
            const entry = document.createElement('div');
            entry.className = 'log-entry';
            const data = JSON.parse(e.data);
            let levelClass = 'log-info';
            if (data.level === 'warning') levelClass = 'log-warn';
            if (data.level === 'error') levelClass = 'log-error';
            entry.innerHTML = '<span class="log-time">' + data.time + '</span> ' +
                '<span class="' + levelClass + '">[' + data.level.toUpperCase() + ']</span> ' +
                data.message;
            logsEl.appendChild(entry);
            logsEl.scrollTop = logsEl.scrollHeight;
            // Keep only last 100 entries
            while (logsEl.children.length > 100) {
                logsEl.removeChild(logsEl.firstChild);
            }
        };
        evtSource.onerror = function() {
            logsEl.innerHTML = '<div class="log-entry"><span class="log-warn">Log stream disconnected. Reconnecting...</span></div>';
        };
    </script>
</body>
</html>`
