# Multi-stage build
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install git for go mod download
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 for a static binary
RUN CGO_ENABLED=0 go build \
    -ldflags="-w -s" \
    -o rarclean ./cmd/rarclean

# Final stage - Debian slim for unrar (non-free) support
FROM debian:bookworm-slim

# Install unrar from Debian non-free (supports RAR3, RAR4, RAR5)
RUN echo "deb http://deb.debian.org/debian bookworm main non-free non-free-firmware" > /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends unrar curl bash && \
    rm -rf /var/lib/apt/lists/*

# Infisical CLI
RUN curl -1sLf 'https://artifacts-cli.infisical.com/setup.deb.sh' | bash \
  && apt-get install -y --no-install-recommends infisical \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/rarclean .
COPY config.example.json ./config.example.json
COPY infisical-run.sh .

# Create necessary directories
RUN mkdir -p /data /logs

# Set permissions
RUN chmod +x rarclean infisical-run.sh

# Use non-root user
RUN groupadd -g 1000 rarclean && \
    useradd -u 1000 -g rarclean -M -s /sbin/nologin rarclean && \
    chown -R rarclean:rarclean /app

USER rarclean

ENTRYPOINT ["./rarclean"]

