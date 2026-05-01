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
# CGO_ENABLED=0 for static binary that works in alpine
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o rarclean ./cmd/rarclean

# Final stage
FROM alpine:latest

# Install 7z for RAR extraction (7zip package = modern 7-Zip with RAR5 support)
# The 7zip package provides 7zz; symlink to 7z for compatibility
RUN apk add --no-cache 7zip && ln -sf /usr/bin/7zz /usr/bin/7z

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/rarclean .

# Copy example config
COPY config.example.json ./config.example.json

# Create necessary directories
RUN mkdir -p /data /logs

# Set permissions
RUN chmod +x rarclean

# Use non-root user
RUN addgroup -g 1000 rarclean && \
    adduser -D -u 1000 -G rarclean rarclean && \
    chown -R rarclean:rarclean /app

USER rarclean

ENTRYPOINT ["./rarclean"]

