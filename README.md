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

## How It Works

1.  When a request is received that matches one of the configured `wakeUpEndpoints`, JellyWolProxy checks if the Jellyfin server is online.
2.  If the server is offline, the proxy sends a Wake-on-LAN (WoL) magic packet to wake it up.
3.  The proxy then holds the request and waits for the server to become available, periodically checking its status.
4.  Once the server is online, the original request is forwarded to it.
5.  All subsequent requests are forwarded directly to the Jellyfin server.

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
