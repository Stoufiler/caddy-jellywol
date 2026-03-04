package jellywol

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/mdlayher/wol"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(JellyWol{})
	httpcaddyfile.RegisterHandlerDirective("jellywol", parseCaddyfile)
}

// JellyWol is a Caddy HTTP middleware that wakes up a remote server
// using a Wake-On-LAN (WOL) magic packet when it is unreachable.
type JellyWol struct {
	// Configuration fields parsed from Caddyfile
	Mac          string   `json:"mac,omitempty"`
	Broadcast    string   `json:"broadcast,omitempty"`
	PingIP       string   `json:"ping_ip,omitempty"`
	PingPort     int      `json:"ping_port,omitempty"`
	Timeout      string   `json:"timeout,omitempty"`
	BlockPaths   []string `json:"block_paths,omitempty"`
	TriggerPaths []string `json:"trigger_paths,omitempty"`
	WolCount     int      `json:"wol_count,omitempty"`
	RetryAfter   int      `json:"retry_after,omitempty"`

	// Internal state
	logger  *zap.Logger
	macAddr net.HardwareAddr
	timeout time.Duration

	// wakingUp prevents sending multiple WOL packets simultaneously.
	wakingUp atomic.Bool
}

// ProvisionMock is strictly used for integration testing to inject a mock logger and bypass Caddy context.
func (j *JellyWol) ProvisionMock(logger *zap.Logger, customTimeout time.Duration) {
	j.logger = logger
	j.timeout = customTimeout
	j.macAddr, _ = net.ParseMAC(j.Mac)
	j.Broadcast = "255.255.255.255:9"
	if j.WolCount <= 0 {
		j.WolCount = 1
	}
	if j.RetryAfter <= 0 {
		j.RetryAfter = 10
	}
}

// CaddyModule returns the Caddy module information.
//
//nolint:govet // Caddy requires passing struct by value in RegisterModule
func (JellyWol) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.jellywol",
		New: func() caddy.Module { return new(JellyWol) },
	}
}

// Provision sets up the module and parses its configuration.
func (j *JellyWol) Provision(ctx caddy.Context) error {
	j.logger = ctx.Logger()

	// Default to standard WOL broadcast port if not specified
	if j.Broadcast == "" {
		j.Broadcast = "255.255.255.255:9"
	} else if _, _, err := net.SplitHostPort(j.Broadcast); err != nil {
		j.Broadcast = net.JoinHostPort(j.Broadcast, "9")
	}

	var err error
	j.macAddr, err = net.ParseMAC(j.Mac)
	if err != nil {
		return fmt.Errorf("invalid MAC address %q: %w", j.Mac, err)
	}

	if j.Timeout != "" {
		j.timeout, err = time.ParseDuration(j.Timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout %q: %w", j.Timeout, err)
		}
	} else {
		j.timeout = 2 * time.Second
	}

	if j.WolCount <= 0 {
		j.WolCount = 1
	}

	if j.RetryAfter <= 0 {
		j.RetryAfter = 10
	}

	return nil
}

// Validate ensures the configuration is semantically valid.
func (j *JellyWol) Validate() error {
	if j.Mac == "" {
		return fmt.Errorf("mac address is required")
	}
	if j.PingIP == "" {
		return fmt.Errorf("ping_ip is required")
	}
	if j.PingPort <= 0 || j.PingPort > 65535 {
		return fmt.Errorf("invalid ping_port: must be between 1 and 65535")
	}
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (j *JellyWol) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	// If no paths are configured at all, we act globally for everything (legacy behavior)
	globalMode := len(j.BlockPaths) == 0 && len(j.TriggerPaths) == 0

	isBlockPath := globalMode || j.matchesPath(r, j.BlockPaths)
	isTriggerPath := j.matchesPath(r, j.TriggerPaths)

	// If the request doesn't match any target paths, bypass the middleware
	if !isBlockPath && !isTriggerPath {
		return next.ServeHTTP(w, r)
	}

	// Check if the remote server is currently reachable
	address := net.JoinHostPort(j.PingIP, strconv.Itoa(j.PingPort))
	conn, err := net.DialTimeout("tcp", address, j.timeout)
	if err == nil {
		_ = conn.Close()
		j.wakingUp.Store(false) // Reset state since the server is up
		return next.ServeHTTP(w, r)
	}

	// The server is unreachable, attempt to wake it up
	if j.wakingUp.CompareAndSwap(false, true) {
		go j.sendWOL()
	}

	// If this path is ONLY a trigger path and NOT a block path, let it pass through immediately
	if !isBlockPath && isTriggerPath {
		return next.ServeHTTP(w, r)
	}

	// For BlockPaths, intercept the request and return a 503
	w.Header().Set("Retry-After", strconv.Itoa(j.RetryAfter))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusServiceUnavailable)

	msg := fmt.Sprintf("Server is waking up, please retry in %d seconds.", j.RetryAfter)
	_, _ = w.Write([]byte(msg))

	return nil
}

// matchesPath checks if the request URL matches any of the provided patterns using basic wildcards.
func (j *JellyWol) matchesPath(r *http.Request, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	reqPath := r.URL.Path
	for _, pattern := range patterns {
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(reqPath, prefix) {
				return true
			}
		} else if reqPath == pattern {
			return true
		}
	}
	return false
}

// sendWOL broadcasts the Magic Packet(s) and manages the cooldown.
func (j *JellyWol) sendWOL() {
	j.logger.Info("Target server is down, initiating Wake-On-LAN sequence",
		zap.String("mac", j.Mac),
		zap.Int("count", j.WolCount),
	)

	client, err := wol.NewClient()
	if err != nil {
		j.logger.Error("Failed to initialize WOL client", zap.Error(err))
		// Reset state so we can retry on the next request
		j.wakingUp.Store(false)
		return
	}
	defer func() {
		_ = client.Close()
	}()

	for i := 0; i < j.WolCount; i++ {
		if i > 0 {
			time.Sleep(1 * time.Second)
		}
		if err := client.Wake(j.Broadcast, j.macAddr); err != nil {
			j.logger.Error("Failed to broadcast WOL magic packet",
				zap.Int("attempt", i+1),
				zap.Error(err),
			)
		}
	}
	j.logger.Info("Wake-On-LAN sequence completed")

	// Allow another wake attempt if the server hasn't come online after 60 seconds
	go func() {
		time.Sleep(60 * time.Second)
		if j.wakingUp.CompareAndSwap(true, false) {
			j.logger.Debug("WOL cooldown expired, ready for another attempt if necessary")
		}
	}()
}

// UnmarshalCaddyfile sets up the module from Caddyfile tokens.
func (j *JellyWol) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			if err := j.parseSubdirective(d); err != nil {
				return err
			}
		}
	}
	return nil
}

func (j *JellyWol) parseSubdirective(d *caddyfile.Dispenser) error {
	val := d.Val()

	// Handle list-based arguments
	if val == "block_paths" {
		j.BlockPaths = d.RemainingArgs()
		return nil
	}
	if val == "trigger_paths" {
		j.TriggerPaths = d.RemainingArgs()
		return nil
	}

	// Handle single-value arguments
	if !d.NextArg() {
		return d.ArgErr()
	}

	switch val {
	case "mac":
		j.Mac = d.Val()
	case "broadcast":
		j.Broadcast = d.Val()
	case "ping_ip":
		j.PingIP = d.Val()
	case "timeout":
		j.Timeout = d.Val()
	case "ping_port":
		return j.parseInt(d.Val(), &j.PingPort, "ping_port")
	case "wol_count":
		return j.parseInt(d.Val(), &j.WolCount, "wol_count")
	case "retry_after":
		return j.parseInt(d.Val(), &j.RetryAfter, "retry_after")
	default:
		return d.Errf("unrecognized subdirective: %s", val)
	}

	return nil
}

func (j *JellyWol) parseInt(val string, target *int, field string) error {
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Errorf("invalid %s: %w", field, err)
	}
	*target = parsed
	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var j JellyWol
	err := j.UnmarshalCaddyfile(h.Dispenser)
	return &j, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*JellyWol)(nil)
	_ caddy.Validator             = (*JellyWol)(nil)
	_ caddyhttp.MiddlewareHandler = (*JellyWol)(nil)
	_ caddyfile.Unmarshaler       = (*JellyWol)(nil)
)
