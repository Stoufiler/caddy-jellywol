package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LogEntry represents a log entry for streaming
type LogEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

// LogBroadcaster manages log streaming to SSE clients
type LogBroadcaster struct {
	mu       sync.RWMutex
	clients  map[chan LogEntry]struct{}
	buffer   []LogEntry
	bufferMu sync.RWMutex
}

// NewLogBroadcaster creates a new log broadcaster
func NewLogBroadcaster() *LogBroadcaster {
	return &LogBroadcaster{
		clients: make(map[chan LogEntry]struct{}),
		buffer:  make([]LogEntry, 0, 50),
	}
}

// Subscribe adds a new client to receive log entries
func (lb *LogBroadcaster) Subscribe() chan LogEntry {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	ch := make(chan LogEntry, 100)
	lb.clients[ch] = struct{}{}

	// Send buffered entries
	lb.bufferMu.RLock()
	for _, entry := range lb.buffer {
		select {
		case ch <- entry:
		default:
		}
	}
	lb.bufferMu.RUnlock()

	return ch
}

// Unsubscribe removes a client
func (lb *LogBroadcaster) Unsubscribe(ch chan LogEntry) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	delete(lb.clients, ch)
	close(ch)
}

// Broadcast sends a log entry to all clients
func (lb *LogBroadcaster) Broadcast(entry LogEntry) {
	// Add to buffer
	lb.bufferMu.Lock()
	lb.buffer = append(lb.buffer, entry)
	if len(lb.buffer) > 50 {
		lb.buffer = lb.buffer[1:]
	}
	lb.bufferMu.Unlock()

	// Send to all clients
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	for ch := range lb.clients {
		select {
		case ch <- entry:
		default:
			// Client is slow, skip this entry
		}
	}
}

// Global broadcaster instance
var globalBroadcaster = NewLogBroadcaster()

// GetLogBroadcaster returns the global log broadcaster
func GetLogBroadcaster() *LogBroadcaster {
	return globalBroadcaster
}

// LogrusHook is a logrus hook that broadcasts log entries
type LogrusHook struct {
	broadcaster *LogBroadcaster
}

// NewLogrusHook creates a new logrus hook for broadcasting logs
func NewLogrusHook() *LogrusHook {
	return &LogrusHook{
		broadcaster: globalBroadcaster,
	}
}

// Levels returns the log levels this hook handles
func (h *LogrusHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is called when a log entry is made
func (h *LogrusHook) Fire(entry *logrus.Entry) error {
	message := entry.Message

	// If there are fields, append them to the message
	if len(entry.Data) > 0 {
		for k, v := range entry.Data {
			message += fmt.Sprintf(" %s=%v", k, v)
		}
	}

	logEntry := LogEntry{
		Time:    entry.Time.Format("15:04:05"),
		Level:   entry.Level.String(),
		Message: message,
	}
	h.broadcaster.Broadcast(logEntry)
	return nil
}

// LogStreamHandler returns an SSE handler for streaming logs
func LogStreamHandler(logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE not supported", http.StatusInternalServerError)
			return
		}

		// Subscribe to log entries
		ch := globalBroadcaster.Subscribe()
		defer globalBroadcaster.Unsubscribe(ch)

		// Send initial connection message
		_, _ = fmt.Fprintf(w, "data: %s\n\n", `{"time":"now","level":"info","message":"Connected to log stream"}`)
		flusher.Flush()

		// Stream logs
		for {
			select {
			case <-r.Context().Done():
				return
			case entry, ok := <-ch:
				if !ok {
					return
				}
				data, err := json.Marshal(entry)
				if err != nil {
					continue
				}
				_, _ = fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			case <-time.After(30 * time.Second):
				// Send keepalive
				_, _ = fmt.Fprintf(w, ": keepalive\n\n")
				flusher.Flush()
			}
		}
	}
}
