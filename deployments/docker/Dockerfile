# Build stage
FROM caddy:builder AS builder

WORKDIR /build

# Copy source code
COPY . .

# Build arguments
ARG VERSION=dev

# Build Caddy with the local jellywol plugin
RUN xcaddy build \
    --with github.com/Stoufiler/caddy-jellywol=. \
    --output /caddy-jellywol

# Runtime stage
FROM caddy:alpine

# Copy binary from builder
COPY --from=builder /caddy-jellywol /usr/bin/caddy

# Set entrypoint and default command
ENTRYPOINT ["/usr/bin/caddy"]
CMD ["run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]
