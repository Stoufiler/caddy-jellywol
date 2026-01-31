package websocket

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// IsWebSocketRequest checks if the request is a WebSocket upgrade request
func IsWebSocketRequest(r *http.Request) bool {
	connection := strings.ToLower(r.Header.Get("Connection"))
	upgrade := strings.ToLower(r.Header.Get("Upgrade"))
	return strings.Contains(connection, "upgrade") && upgrade == "websocket"
}

// ProxyWebSocket handles WebSocket connection proxying
func ProxyWebSocket(w http.ResponseWriter, r *http.Request, targetHost string, logger *logrus.Logger) error {
	logger.Debugf("Proxying WebSocket connection to %s", targetHost)

	// Hijack the client connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return fmt.Errorf("webserver doesn't support hijacking")
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		return fmt.Errorf("failed to hijack connection: %v", err)
	}
	defer func() { _ = clientConn.Close() }()

	// Connect to the target server
	targetConn, err := net.Dial("tcp", targetHost)
	if err != nil {
		return fmt.Errorf("failed to connect to target: %v", err)
	}
	defer func() { _ = targetConn.Close() }()

	// Forward the original request to the target
	if err := r.Write(targetConn); err != nil {
		return fmt.Errorf("failed to write request to target: %v", err)
	}

	// Bidirectional copy
	errChan := make(chan error, 2)

	// Copy from target to client
	go func() {
		_, err := io.Copy(clientConn, targetConn)
		errChan <- err
	}()

	// Copy from client to target
	go func() {
		_, err := io.Copy(targetConn, clientConn)
		errChan <- err
	}()

	// Wait for either direction to complete or error
	err = <-errChan
	if err != nil && err != io.EOF {
		logger.Debugf("WebSocket connection closed: %v", err)
	}

	return nil
}
