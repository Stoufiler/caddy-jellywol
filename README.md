# JellyWolProxy

JellyWolProxy is a smart proxy server that seamlessly integrates Jellyfin media server with Wake-on-LAN capabilities. It automatically manages your Jellyfin server's power state by waking it up on-demand when media content is requested and forwarding traffic efficiently.

## Key Features

- **Smart Power Management**: Automatically wakes up your Jellyfin server using Wake-on-LAN when media is requested
- **Transparent Proxying**: Seamlessly forwards requests to your Jellyfin server
- **Energy Efficient**: Allows your media server to sleep when not in use
- **Configurable Wake Triggers**: Customize which endpoints trigger server wake-up
- **Simple Setup**: Easy configuration through JSON file

## Installation

1. Ensure you have Go 1.23.5 or later installed
2. Clone the repository:
   ```bash
   git clone https://github.com/StephanGR/JellyWolProxy.git
   cd JellyWolProxy
   ```
3. Build the project:
   ```bash
   go build -o jellywolproxy
   ```

## Configuration

Create a `config.json` file with the following structure:

```json
{
  "jellyfinUrl": "your.jellyfin.domain",
  "apiKey": "your-jellyfin-api-key",
  "macAddress": "XX:XX:XX:XX:XX:XX",
  "broadcastAddress": "255.255.255.255:9",
  "wakeUpIp": "192.168.0.x",
  "wakeUpPort": 81,
  "forwardIp": "192.168.0.x",
  "forwardPort": 1234,
  "wakeUpEndpoints": [
    "/videos/*/main.m3u8",
    "/Videos/*/stream"
  ]
}
```

### Configuration Parameters

- `jellyfinUrl`: Your Jellyfin server's domain
- `apiKey`: Jellyfin API key for authentication
- `macAddress`: MAC address of the Jellyfin server for Wake-on-LAN
- `broadcastAddress`: Network broadcast address for WoL packets
- `wakeUpIp`: IP address for wake-up requests
- `wakeUpPort`: Port for wake-up requests
- `forwardIp`: Jellyfin server IP for request forwarding
- `forwardPort`: Jellyfin server port for request forwarding
- `wakeUpEndpoints`: List of endpoints that trigger server wake-up

## How It Works

1. When a request matches one of the configured wake-up endpoints, JellyWolProxy sends a Wake-on-LAN packet to your Jellyfin server
2. The proxy then forwards the request to your Jellyfin server
3. All subsequent requests are forwarded normally until the server goes back to sleep

## Monitoring

JellyWolProxy provides comprehensive monitoring capabilities through health check endpoints and Prometheus metrics.

### Health Check Endpoints

- `/health`: Basic health check endpoint
  ```json
  {
    "status": "UP",
    "timestamp": "2025-01-18T22:22:33+01:00",
    "version": "1.0.0"
  }
  ```

- `/health/ready`: Detailed readiness check that verifies Jellyfin connectivity
  ```json
  {
    "status": "UP",
    "timestamp": "2025-01-18T22:22:33+01:00",
    "version": "1.0.0",
    "checks": {
      "jellyfin": {
        "status": "UP",
        "message": "Jellyfin server is reachable"
      }
    }
  }
  ```

### Prometheus Metrics

The `/metrics` endpoint exposes the following metrics in Prometheus format:

| Metric Name | Type | Description | Labels |
|------------|------|-------------|---------|
| `jellywolproxy_requests_total` | Counter | Total number of requests processed | `path`, `method`, `status` |
| `jellywolproxy_wakeup_attempts_total` | Counter | Total number of wake-up attempts | - |
| `jellywolproxy_wakeup_success_total` | Counter | Total number of successful wake-ups | - |
| `jellywolproxy_server_state` | Gauge | Current server state (1 = up, 0 = down) | - |
| `jellywolproxy_request_duration_seconds` | Histogram | Request duration in seconds | `path`, `method` |

### Monitoring Setup

1. Configure Prometheus to scrape the `/metrics` endpoint:
   ```yaml
   scrape_configs:
     - job_name: 'jellywolproxy'
       static_configs:
         - targets: ['localhost:8080']
   ```

2. Use the health check endpoints for uptime monitoring:
   - `/health` for basic uptime checks
   - `/health/ready` for detailed health status including Jellyfin connectivity

3. Set up alerting based on metrics:
   - Server state changes (`jellywolproxy_server_state`)
   - Failed wake-up attempts (difference between `jellywolproxy_wakeup_attempts_total` and `jellywolproxy_wakeup_success_total`)
   - High request latency (`jellywolproxy_request_duration_seconds`)

## Project Structure

```
JellyWolProxy/
├── cmd/           # Application entry points
├── internal/      # Internal packages
│   ├── config/    # Configuration management
│   ├── handlers/  # HTTP request handlers
│   ├── jellyfin/  # Jellyfin-specific logic
│   ├── logger/    # Logging functionality
│   ├── server/    # Server implementation
│   └── wol/       # Wake-on-LAN functionality
└── config.json    # Configuration file
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

GPL License
