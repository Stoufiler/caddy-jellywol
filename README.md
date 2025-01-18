# JellyWolProxy

JellyWolProxy is a smart proxy server for Jellyfin media server that integrates Wake-on-LAN (WoL) capabilities. It automatically wakes up your Jellyfin server when media content is requested and handles request forwarding efficiently.

## Features

- Automatic Wake-on-LAN when accessing specific media endpoints
- Intelligent request forwarding to Jellyfin server
- Configurable endpoints that trigger wake-up
- Simple JSON-based configuration

## Installation

1. Ensure you have Go 1.21.3 or later installed
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

[Add your license information here]
