# caddy-jellywol — Copilot Instructions

## Build, Test, Lint

```bash
make build          # builds to build/caddy-jellywol
make test           # go test -v ./...
make lint           # golangci-lint run ./...
make fmt            # go fmt ./...
make test-coverage  # generates coverage.out and coverage.html
```

Run a single test:
```bash
go test -run TestFunctionName ./internal/handlers/
```

Pre-commit hooks run: `go-fmt`, `go-imports`, `go-cyclo` (max 15), `golangci-lint`, `go-critic`, unit tests, and build. Install with `make setup` or `pre-commit install`.

## Architecture

The proxy is a single binary (`cmd/caddy-jellywol/main.go`) built around `gorilla/mux`. On startup it:
1. Loads and validates `config.json` (via `spf13/viper`; sensitive fields can be overridden by env vars `JELLYFIN_API_KEY`, `SERVER_MAC_ADDRESS`, `JELLYFIN_URL`)
2. Wires together the router with specific route handlers (`/health`, `/health/ready`, `/metrics`, `/status/*`, `/ping`) and a catch-all proxy handler

**Request flow for all other paths:**
- Middleware stack (outermost → innermost): `MetricsMiddleware` → `NetworkStatsMiddleware` → `CacheMiddleware` (optional) → `RequestLoggerMiddleware` → `handlers.Handler`
- `handlers.Handler` checks `util.ShouldWakeServer()` against configured `wakeUpEndpoints` (supports single `*` wildcard patterns)
- If matched and server is down: sends WoL magic packet, returns `503` with `Retry-After: 30` so clients retry automatically
- If server is up (or endpoint doesn't match): proxies via `httputil.ReverseProxy`, rewriting `Location` redirect headers back to the original host

**Key packages:**
| Package | Role |
|---|---|
| `internal/config` | `Config` struct + `Validate()`, `HotReloadableConfig` for SIGHUP reloads |
| `internal/server_state` | Mutex-guarded `ServerState` tracking whether a wake is in progress |
| `internal/services` | `Waker` and `ServerStateChecker` interfaces + concrete implementations (enables mock injection in tests) |
| `internal/wol` | Sends Wake-on-LAN magic packets via `mdlayher/wol` |
| `internal/util` | `ShouldWakeServer` (pattern matching) and `IsServerUp` (TCP dial check) |
| `internal/cache` | In-memory response cache; streaming endpoints (`.m3u8`, `.ts`, `/stream`) are excluded automatically |
| `internal/dashboard` | `/status` page with SSE log streaming and optional OIDC auth |
| `internal/websocket` | Detects and proxies WebSocket upgrades separately from normal HTTP |
| `internal/metrics` | Prometheus metrics (`ServerStateGauge` and others) |

## Key Conventions

- **Interface-based injection**: `Waker` and `ServerStateChecker` in `internal/services` are interfaces. Use them in handler/test code; `ConcreteWaker` and `ConcreteServerStateChecker` are the production implementations.
- **Config is passed by value**: `config.Config` is copied into functions, not passed as a pointer (except `HotReloadableConfig` which wraps it behind a mutex).
- **French comments in proxy code**: Some comments in `handlers/handlers.go` and related files are in French — this is intentional, do not translate them.
- **Wildcard patterns**: `wakeUpEndpoints` supports only a single `*` wildcard per pattern (prefix + suffix matching). This is intentional and documented.
- **503 + Retry-After pattern**: The proxy returns 503 immediately when waking a server rather than holding the connection. Clients (Infuse, Jellyfin apps) handle `Retry-After` natively.
- **Commit messages**: Must follow Conventional Commits (`feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`). Enforced by commitlint pre-commit hook.
- **Cyclomatic complexity limit**: Functions must stay below 15 (enforced by `go-cyclo` in pre-commit).
- **`no-go-testing` hook**: `testing.T` must not appear in non-test files.
