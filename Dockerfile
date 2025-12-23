# Multi-stage build
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -o rarclean ./cmd/rarclean

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/rarclean .

# Copy example config
COPY config.example.json ./config.json

# Create necessary directories
RUN mkdir -p /data /logs

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/root/rarclean", "-health"]

ENTRYPOINT ["./rarclean"]
CMD ["-config", "/root/config.json"]
