# JellyWolProxy

JellyWolProxy is a smart proxy server that seamlessly integrates Jellyfin media server with Wake-on-LAN capabilities. It automatically manages your Jellyfin server's power state by waking it up on-demand when media content is requested and forwarding traffic efficiently.

## Key Features

- **Smart Power Management**: Automatically wakes up your Jellyfin server using Wake-on-LAN when media is requested.
- **Transparent Proxying**: Holds and forwards requests to your Jellyfin server after it has woken up.
- **Energy Efficient**: Allows your media server to sleep when not in use.
- **Configurable**: Customize wake-up triggers, timeouts, and logging through a configuration file and command-line flags.
- **Robust**: Validates your configuration on startup to prevent errors.
- **Always Up-to-date**: Automatically receives dependency updates via Renovate.

## Installation

1. Ensure you have Go 1.23.5 or later installed.
2. Clone the repository:
   ```bash
   git clone https://github.com/Stoufiler/JellyWolProxy.git
   cd JellyWolProxy
   ```
3. Build the project:
   ```bash
   go build -o jellywolproxy ./cmd/jellywolproxy
   ```

## Usage

Run the proxy using the following command:

```bash
./jellywolproxy [flags]
```

#### Command-line Flags

- `--config`: Path to the `config.json` file. (Default: `config.json`)
- `--port`: Port for the proxy to run on. (Default: `3881`)
- `--log-level`: Set the logging level (`Debug`, `Info`, `Warn`, `Error`). This overrides the `logLevel` in the config file.

## Configuration

Create a `config.json` file by copying the `config.json.example` and filling in the values. The application will validate this configuration on startup and exit if any values are invalid.

```json
{
  "jellyfinUrl": "your.jellyfin.domain",
  "apiKey": "your-jellyfin-api-key",
  "macAddress": "XX:XX:XX:XX:XX:XX",
  "broadcastAddress": "255.255.255.255:9",
  "wakeUpIp": "192.168.0.x",
  "wakeUpPort": 80,
  "forwardIp": "192.168.0.x",
  "forwardPort": 8096,
  "wakeUpEndpoints": [
    "/videos/*/main.m3u8",
    "/Videos/*/stream"
  ],
  "serverWakeUpTimeout": 120,
  "serverWakeUpTicker": 5,
  "postPingDelaySeconds": 0,
  "logLevel": "Info"
}
```

#### Configuration Parameters

- `jellyfinUrl`: Your Jellyfin server's domain.
- `apiKey`: Jellyfin API key for authentication.
- `macAddress`: MAC address of the Jellyfin server for Wake-on-LAN.
- `broadcastAddress`: Network broadcast address for WoL packets.
- `wakeUpIp`: IP address of the server to be woken up.
- `wakeUpPort`: Port used to check if the server is online.
- `forwardIp`: IP address of the Jellyfin server to forward requests to.
- `forwardPort`: Port of the Jellyfin server.
- `wakeUpEndpoints`: List of URL paths that will trigger a server wake-up.
- `serverWakeUpTimeout`: (Optional) The maximum time in seconds to wait for the server to come online. Defaults to `120`.
- `serverWakeUpTicker`: (Optional) The interval in seconds at which to check if the server is online during wake-up. Defaults to `5`.
- `postPingDelaySeconds`: (Optional) The delay in seconds to wait after the server is confirmed to be online before proxying requests. Defaults to `0`.
- `logLevel`: (Optional) The logging level. Can be `Debug`, `Info`, `Warn`, or `Error`. Defaults to `Info`. Can be overridden by the `--log-level` command-line flag.
- `cacheEnabled`: (Optional) Enable response caching for improved performance. Defaults to `false`.
- `cacheTTLSeconds`: (Optional) Cache time-to-live in seconds. Defaults to `300` (5 minutes).

### Environment Variables

For enhanced security, sensitive configuration values can be provided via environment variables instead of storing them in the config file:

- `JELLYFIN_API_KEY`: Overrides the `apiKey` configuration
- `SERVER_MAC_ADDRESS`: Overrides the `macAddress` configuration
- `JELLYFIN_URL`: Overrides the `jellyfinUrl` configuration

**Example:**

```bash
export JELLYFIN_API_KEY="your-secret-api-key"
export SERVER_MAC_ADDRESS="50:91:e3:c9:37:18"
./jellywolproxy --config config.json
```

Environment variables take precedence over values in the config file.

## How It Works

1.  When a request is received that matches one of the configured `wakeUpEndpoints`, JellyWolProxy checks if the Jellyfin server is online.
2.  If the server is offline, the proxy sends a Wake-on-LAN (WoL) magic packet to wake it up.
3.  The proxy then holds the request and waits for the server to become available, periodically checking its status.
4.  Once the server is online, the original request is forwarded to it.
5.  All subsequent requests are forwarded directly to the Jellyfin server.

## Advanced Features

### Response Caching

JellyWolProxy includes an optional response caching layer to improve performance and reduce load on your Jellyfin server. When enabled:

- GET requests are cached for the configured TTL (Time-To-Live)
- Streaming endpoints (`.m3u8`, `.ts`, `/stream`) are automatically excluded from caching
- Cache status is indicated via the `X-Cache` header (`HIT` or `MISS`)

To enable caching, set `cacheEnabled: true` in your configuration and optionally adjust `cacheTTLSeconds` (default: 300 seconds).

### Dashboard & Status Page

JellyWolProxy includes a built-in web dashboard accessible at `/status` that provides:

- **Real-time server state** (online/offline/waking)
- **Statistics**: uptime, total requests, cache hit rate, wake-up count
- **Live log streaming** via Server-Sent Events (SSE)

#### Dashboard Endpoints

- `/status` - Main dashboard page (HTML)
- `/status/api` - JSON API for statistics
- `/status/logs` - Live log stream (SSE)

#### SSO Authentication (Optional)

The dashboard can be protected with OIDC/OAuth2 authentication, compatible with providers like [Pocket-ID](https://github.com/pocket-id/pocket-id), Keycloak, Authentik, etc.

To enable SSO, configure the `dashboardOIDC` section in your config:

```json
{
  "dashboardOIDC": {
    "enabled": true,
    "issuer_url": "https://pocket-id.example.com",
    "client_id": "jellywolproxy",
    "client_secret": "your-client-secret",
    "redirect_url": "http://your-proxy:3881/status/callback",
    "scopes": "openid email profile"
  }
}
```

| Parameter | Description |
|-----------|-------------|
| `enabled` | Enable/disable SSO authentication |
| `issuer_url` | OIDC provider URL (must support `.well-known/openid-configuration`) |
| `client_id` | OAuth2 client ID |
| `client_secret` | OAuth2 client secret |
| `redirect_url` | Callback URL (must point to `/status/callback`) |
| `scopes` | OAuth2 scopes to request (default: `openid email profile`) |

When SSO is disabled, the dashboard is publicly accessible.

### Configuration Hot-Reload

JellyWolProxy supports reloading configuration without restarting the server by sending a `SIGHUP` signal:

```bash
kill -SIGHUP $(pgrep jellywolproxy)
```

The following settings can be hot-reloaded:
- Log level
- Cache TTL
- Wake-up timeouts and intervals

Note: Some settings (port, IP addresses, MAC address) require a full restart.

### Graceful Shutdown

The proxy handles shutdown signals (`SIGINT`, `SIGTERM`) gracefully:

- Stops accepting new connections
- Waits up to 30 seconds for active requests to complete
- Ensures no requests are dropped during deployment or restart

Simply send a termination signal (e.g., `Ctrl+C` or `kill`) and the proxy will shutdown cleanly.

## Monitoring

JellyWolProxy provides comprehensive monitoring capabilities through health check endpoints and Prometheus metrics.

### Health Check Endpoints

- `/health`: Basic health check endpoint.
- `/health/ready`: Detailed readiness check that verifies Jellyfin connectivity.

### Prometheus Metrics

The `/metrics` endpoint exposes key metrics about the proxy's performance and state. You can find more details about the available metrics in the source code.

## Dependency Management

This project uses [Renovate](https://github.com/renovatebot/renovate) to automatically keep dependencies, including Go modules and GitHub Actions, up-to-date. Renovate will create pull requests for any available updates.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Coded with Gemini

This entire project was coded using Google's Gemini. I have only provided the prompts and Gemini has done the rest.

## License

This project is licensed under the GPL License.
