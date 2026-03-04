# caddy-jellywol (Caddy Plugin)

[![Go](https://github.com/Stoufiler/caddy-jellywol/actions/workflows/go.yml/badge.svg)](https://github.com/Stoufiler/caddy-jellywol/actions/workflows/go.yml)
[![Release](https://github.com/Stoufiler/caddy-jellywol/actions/workflows/release.yml/badge.svg)](https://github.com/Stoufiler/caddy-jellywol/actions/workflows/release.yml)
[![License](https://img.shields.io/github/license/Stoufiler/caddy-jellywol)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/stoufiler/caddy-jellywol)](https://github.com/Stoufiler/caddy-jellywol/pkgs/container/caddy-jellywol)

caddy-jellywol is a smart [Caddy](https://caddyserver.com/) plugin that seamlessly integrates your media server (like Jellyfin) or remote storage (like a NAS) with Wake-on-LAN capabilities. It acts as an HTTP middleware for Caddy, automatically managing your server's power state by waking it up on-demand when someone tries to access specific paths.

Instead of reinventing the wheel with a custom proxy, caddy-jellywol leverages the robust, production-ready, and high-performance reverse proxy native to Caddy.

## Key Features

- **Smart Power Management**: Intercepts requests. If your server is offline, it sends a Wake-on-LAN (WoL) magic packet to wake it up.
- **Path Filtering (Trigger vs Block)**: Define which URLs simply trigger the wake-up in the background, and which URLs should be blocked until the server is ready.
- **App Friendly (Infuse/Jellyfin)**: Returns standard `503 Service Unavailable` with `Retry-After` headers, ensuring media players automatically retry without breaking.
- **Native Caddy Power**: Once the server is online, the plugin steps out of the way, letting Caddy handle the proxying.
- **Energy Efficient**: Allows your heavy storage or media server to sleep when not in use.

## Installation

### Using Docker (Recommended)

We provide a pre-built Docker image based on Caddy with the `jellywol` plugin included.

**docker-compose.yml:**

```yaml
version: '3.8'
services:
  jellywolproxy:
    image: ghcr.io/stoufiler/jellywolproxy:latest
    container_name: jellywolproxy
    network_mode: host # Required for Wake-on-LAN to broadcast properly
    volumes:
      - ./deployments/docker/Caddyfile.example:/etc/caddy/Caddyfile:ro
      - ./caddy_data:/data
      - ./caddy_config:/config
    restart: unless-stopped
    environment:
      TZ: Europe/Paris
```

### Build from Source (xcaddy)

If you prefer running the binary directly, you can compile Caddy with this plugin using [xcaddy](https://github.com/caddyserver/xcaddy):

```bash
xcaddy build --with github.com/Stoufiler/caddy-jellywol
```

## Deployment Scenario: Chained Proxy (Sidecar)

If you already have a main Caddy (or another proxy like Nginx/Traefik) handling your HTTPS and domains, you can run caddy-jellywol as a "Sidecar" on a specific port. This is perfect if your Jellyfin server is always on, but your media files are on a sleeping NAS.

**1. Your Main Proxy (e.g., Caddy on Port 443):**
```caddyfile
jellyfin.yourdomain.com {
    # Forward everything to the caddy-jellywol container
    reverse_proxy localhost:3881
}
```

**2. caddy-jellywol (Running on Port 3881):**
```caddyfile
:3881 {
    jellywol {
        mac aa:bb:cc:dd:ee:ff
        ping_ip 192.168.1.50   # IP of your NAS
        ping_port 2049         # NFS Port (to ensure storage is ready)

        # 1. Non-Blocking Trigger:
        # Send WOL in background when browsing libraries, but DON'T block the user.
        trigger_paths /Items* /Library*

        # 2. Blocking Paths:
        # Return 503 + Retry-After for these paths (Streaming/Download)
        # Infuse and Jellyfin apps will automatically retry after X seconds.
        block_paths /Videos/* /Items/*/Download /Items/*/stream*

        retry_after 10         # Ask client to retry in 10 seconds
        wol_count 3            # Send 3 packets for reliability
    }

    # Final proxy to your Jellyfin server (always ON)
    reverse_proxy 192.168.1.10:8096
}
```

This setup keeps your main proxy clean, keeps the Jellyfin UI fast, and delegates the Wake-On-LAN logic specifically for heavy media streaming.

> **Note:** `network_mode: host` is highly recommended for the caddy-jellywol container to ensure WOL packets reach your local network.

#### Configuration Parameters

- `mac`: **Required**. The MAC address of the server to wake up.
- `ping_ip`: **Required**. The IP address of the server (used to check if it's online).
- `ping_port`: **Required**. The TCP port to ping (e.g., 8096 for Jellyfin, 445 for SMB, 2049 for NFS).
- `broadcast`: (Optional) The broadcast address for the WOL packet. Defaults to `255.255.255.255:9`.
- `timeout`: (Optional) The timeout for the TCP ping check. Defaults to `2s`.
- `block_paths`: (Optional) Space-separated list of paths that should be blocked with a 503 error if the server is down. Sends WOL.
- `trigger_paths`: (Optional) Space-separated list of paths that send WOL in the background but let the request pass through to the proxy immediately.
- `wol_count`: (Optional) Number of WOL packets to send in sequence for reliability. Defaults to `1`.
- `retry_after`: (Optional) The number of seconds sent in the `Retry-After` HTTP header. Defaults to `10`.

*Note: If neither `block_paths` nor `trigger_paths` are defined, the plugin acts as a catch-all and blocks all paths if the server is down.*

## How It Works

1. A user visits your Caddy server (e.g., `https://jellyfin.yourdomain.com/Videos/123/stream`).
2. Caddy passes the request to the `jellywol` middleware.
3. The plugin matches the path and attempts a fast TCP ping to `ping_ip:ping_port`.
   - **If successful:** The server is online! The plugin immediately passes the request to Caddy's `reverse_proxy`.
   - **If it fails:** The server is offline. The plugin broadcasts the WOL magic packet(s), and immediately returns a `503 Service Unavailable` status with a `Retry-After: 10` header.
4. The media player (Infuse/Jellyfin App) waits 10 seconds and automatically retries the request while the server boots up.

## License

This project is licensed under the GPL License.
