# rarclean

A Go CLI tool that automates RAR extraction, file relocation, and qBittorrent torrent tracking updates to maintain seeding while organizing media libraries.

## Problem Statement

When downloading media via torrents as RAR archives:
1. RAR files must be extracted for Plex/media servers
2. Original RAR files must remain for torrent seeding
3. Moving RAR files breaks qBittorrent's file tracking
4. Manual relocation and API updates are tedious and error-prone

## Solution

rarclean executes a five-step automated workflow:

1. **Locate** all RAR files in target directory
2. **Extract** archives using 7z (handles multi-part archives automatically)
3. **Query** qBittorrent API to find associated torrent by path
4. **Relocate** RAR directory to "zombies" location for continued seeding
5. **Update** qBittorrent torrent location via API to maintain tracking

## Features

- ✅ **Multi-part RAR support**: Handles .rar, .part01.rar, .part001.rar patterns
- ✅ **Automatic torrent matching**: Finds torrents by save path
- ✅ **Session management**: qBittorrent API authentication with cookie persistence
- ✅ **Safe operations**: Validation before mutations, atomic moves
- ✅ **Dry-run mode**: Preview changes without executing
- ✅ **Docker ready**: Multi-stage build, lightweight alpine container
- ✅ **CI/CD integrated**: GitHub Actions with multi-arch builds

## Installation

### Prerequisites

- **Local**: Go 1.23+, 7z installed
- **Docker**: Docker and Docker Compose

### Binary Installation

Download pre-built binary from [Releases](https://github.com/killakam3084/rarclean/releases):

```bash
# macOS (Apple Silicon)
wget https://github.com/killakam3084/rarclean/releases/download/v1.0.0/rarclean-darwin-arm64
chmod +x rarclean-darwin-arm64
./rarclean-darwin-arm64 --help

# Linux (x86_64)
wget https://github.com/killakam3084/rarclean/releases/download/v1.0.0/rarclean-linux-amd64
chmod +x rarclean-linux-amd64
./rarclean-linux-amd64 --help
```

### Building from Source

```bash
# Clone repository
git clone https://github.com/killakam3084/rarclean.git
cd rarclean

# Build binary
go build -o rarclean ./cmd/rarclean

# Run
./rarclean --path /path/to/downloads --config config.json
```

### Docker

```bash
# Build image
docker build -t rarclean:latest .

# Run container
docker run --rm \
  -v /mnt/downloads:/mnt/downloads \
  -v ./config.json:/config/config.json:ro \
  rarclean:latest --path /mnt/downloads --config /config/config.json
```

## Configuration

Create `config.json` from `config.example.json`:

```json
{
  "qbittorrent": {
    "url": "http://gluetun:8080",    // qBittorrent Web UI URL
    "username": "admin",              // API username
    "password": "adminPassword"       // API password
  },
  "paths": {
    "television": "/mnt/media/tv",   // TV show directory
    "movies": "/mnt/media/movies",   // Movie directory
    "zombies": "/mnt/media/zombies"  // Seeding RAR storage
  }
}
```

### Configuration Details

| Setting | Purpose | Example |
|---------|---------|---------|
| `qbittorrent.url` | Base URL of qBittorrent Web API | `http://localhost:8080` or `http://gluetun:8080` |
| `qbittorrent.username` | Authentication username | `admin` |
| `qbittorrent.password` | Authentication password | `securePassword` |
| `paths.television` | Directory containing TV shows | `/media/tv` |
| `paths.movies` | Directory containing movies | `/media/movies` |
| `paths.zombies` | Directory for archived RAR files | `/media/zombies` |

## Usage

### Basic Usage

```bash
rarclean --path /mnt/media/tv --config config.json
```

### Options

```bash
# Show help
rarclean --help

# Dry-run mode (preview without executing)
rarclean --path /path --dry-run

# Custom config location
rarclean --path /path --config /etc/rarclean/config.json
```

### Example Output

```
=== rarclean - RAR Extraction & qBittorrent Manager ===

Step 1: Scanning directory for RAR files: /mnt/media/tv/ShowName
Found 1 RAR archive(s):
  1. ShowName.rar (in ShowName)

Authenticating with qBittorrent...

Processing RAR archive 1 of 1: showname
  Step 2: Extracting with 7z...
  Step 3: Finding torrent in qBittorrent...
    Found torrent: Show.Name.S01E01.720p
    Current location: /mnt/media/tv/ShowName
  Step 4: Relocating RAR directory to zombies...
    From: /mnt/media/tv/ShowName
    To:   /mnt/media/zombies/ShowName
  Step 5: Updating qBittorrent torrent location...
    Updated torrent location: /mnt/media/zombies/ShowName
  ✓ Archive processed successfully

=== Processing complete ===
```

## Deployment

### TrueNAS SCALE (Primary Use Case)

Add to docker-compose.yml with gluetun networking:

```yaml
services:
  rarclean:
    image: ghcr.io/killakam3084/rarclean:latest
    network_mode: "container:gluetun"
    volumes:
      - /mnt/cell_block_d:/mnt/cell_block_d
      - ./config.json:/config/config.json:ro
    restart: "no"
```

### Kubernetes CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: rarclean
spec:
  schedule: "0 */6 * * *"  # Every 6 hours
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: rarclean
            image: ghcr.io/killakam3084/rarclean:latest
            args: ["--path", "/media/downloads", "--config", "/etc/rarclean/config.json"]
            volumeMounts:
            - name: media
              mountPath: /media
            - name: config
              mountPath: /etc/rarclean
          volumes:
          - name: media
            hostPath:
              path: /mnt/media
          - name: config
            configMap:
              name: rarclean-config
          restartPolicy: OnFailure
```

### Cron Job (Linux)

```bash
# /etc/cron.d/rarclean
0 */6 * * * user /usr/local/bin/rarclean --path /media/downloads --config /etc/rarclean/config.json >> /var/log/rarclean.log 2>&1
```

## Architecture

### Code Organization

```
rarclean/
├── cmd/rarclean/
│   └── main.go              # Application bootstrap
├── internal/
│   ├── cmd/
│   │   └── root.go          # Cobra root command, workflow orchestration
│   ├── config/
│   │   └── config.go        # Configuration loading with validation
│   ├── extractor/
│   │   └── extractor.go     # RAR discovery and 7z extraction
│   ├── qbittorrent/
│   │   └── client.go        # qBittorrent Web API v2 client
│   └── mover/
│       └── mover.go         # Safe directory relocation
├── .github/workflows/
│   ├── build.yml            # CI: build & test on push
│   └── release.yml          # CD: Docker build & release on tag
├── Dockerfile               # Multi-stage container build
├── docker-compose.yml       # Local testing setup
├── Makefile                 # Build helpers
├── config.example.json      # Configuration template
├── go.mod / go.sum          # Go dependencies
└── README.md                # This file
```

### Workflow Steps

#### Step 1: Find RAR Files
- Walk directory tree recursively
- Collect all `.rar` files (case-insensitive)
- Identify first part of multi-part archives
- Group by archive base name

#### Step 2: Extract
- Execute: `7z x -y -o<dir> <first_part.rar>`
- 7z auto-detects and processes all parts
- Extracted files remain in original directory

#### Step 3: Find Torrent
- GET `/api/v2/torrents/info` (all torrents)
- Filter by exact path match (normalized)
- Return matching torrent or nil if not found

#### Step 4: Relocate
- Validate source directory exists
- Create destination parent if needed
- Move entire directory using `os.Rename` (atomic)
- Verify move completed successfully

#### Step 5: Update Torrent
- POST to `/api/v2/torrents/setLocation`
- Body: `hashes=<hash>&location=<new_path>`
- Verify with GET `/api/v2/torrents/info?hashes=<hash>`

### API Client Implementation

The qBittorrent client handles:

- **Authentication**: POST to `/api/v2/auth/login`, session management via cookies
- **Session persistence**: Reuses cookies across requests
- **Auto-login**: Re-authenticates on 403 Unauthorized
- **Error handling**: Wrapped errors with context for debugging

```go
// Core operations
client.GetTorrents()              // List all torrents
client.FindTorrentByPath(path)    // Find by save path
client.SetLocation(hash, newLoc)  // Update location
client.GetTorrent(hash)           // Get specific torrent
```

## Performance

### Resource Usage

| Metric | Value |
|--------|-------|
| Memory | 5-10 MB |
| CPU | Minimal (I/O bound) |
| Disk | Depends on 7z extraction |
| Network | <1 KB per API call |

### Bottlenecks

- **7z extraction**: Disk I/O is primary bottleneck
- **Network**: Negligible (API calls are lightweight)
- **qBittorrent API**: Fast (<100ms per call)

## Error Handling

| Scenario | Behavior |
|----------|----------|
| No RAR files found | Log message, exit gracefully (exit 0) |
| Extraction fails | Log error, skip relocation, continue to next archive |
| Torrent not found | Log warning, skip relocation, continue |
| Move fails | Don't update qBittorrent, preserve consistency |
| Auth fails | Clear session, return error, user retries |
| Validation fails | Return error with details, no mutations |

## Security

### Credentials

- Store `config.json` outside git (in `.gitignore`)
- Mount as read-only in Docker
- Consider environment variables for CI/CD

### File Operations

- Validate paths before operations
- Check file permissions before moves
- Use atomic operations (`os.Rename`)
- Prevent directory traversal attacks

### Network

- API calls stay within Docker network
- No external exposure required
- TLS not needed (internal-only)

## Testing

### Unit Tests

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -v ./internal/extractor -run TestFindRARFiles
```

### Integration Testing

```bash
# Dry-run mode for safe execution
rarclean --path /test/data --dry-run

# Test data in testdata/ directory (future)
docker-compose up  # Local qBittorrent for testing
```

## Development

### Building

```bash
# Build binary
make build

# Run with config
make run

# Clean build artifacts
make clean
```

### Docker

```bash
# Build image
make docker-build

# Run container
make docker-run

# View logs
make docker-logs

# Stop container
make docker-stop
```

### Code Style

- Go conventions: `golangci-lint`
- Format: `go fmt ./...`
- Imports: `goimports`

```bash
# Format code
make fmt

# Run linter
make lint
```

## Troubleshooting

### qBittorrent Connection Fails

```bash
# Check connectivity
curl -X POST http://qbittorrent-url:8080/api/v2/auth/login \
  -d "username=admin&password=password"

# Verify credentials in config.json
# Check firewall rules if using VPN container
```

### 7z Not Found

```bash
# Install 7z
ubuntu/debian: sudo apt-get install p7zip-full
alpine: apk add p7zip
macOS: brew install p7zip
```

### RAR Files Not Detected

```bash
# Verify file permissions
ls -la /path/to/rars

# Check file extensions
file /path/to/archive.rar

# Manually test 7z extraction
7z x /path/to/archive.rar
```

### Torrent Not Found After Relocation

```bash
# Check if path matches exactly
# Use qBittorrent UI to verify save path
# Manually update torrent location in UI if needed
```

## Performance Optimization

### Future: Parallel Extraction

```go
// Extract multiple archives concurrently
semaphore := make(chan struct{}, 4) // 4 concurrent
for _, rar := range rarFiles {
    go extractConcurrent(rar, semaphore)
}
```

### Future: Daemon Mode

```bash
# Watch directories for new RAR files
rarclean daemon --watch /media/downloads
```

### Future: Webhook Integration

```bash
# HTTP server for qBittorrent webhooks
rarclean webhook --port 8000
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Roadmap

- [ ] Daemon mode with fsnotify watching
- [ ] Discord webhook notifications
- [ ] Plex library refresh integration
- [ ] Configuration hot-reload
- [ ] Prometheus metrics export
- [ ] Web UI for monitoring

## License

MIT License - see LICENSE file

## Support

- **Issues**: [GitHub Issues](https://github.com/killakam3084/rarclean/issues)
- **Discussions**: [GitHub Discussions](https://github.com/killakam3084/rarclean/discussions)
- **Wiki**: [Project Wiki](https://github.com/killakam3084/rarclean/wiki)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history.

## Migration from Manual Workflow

### Before
1. Download RAR files via qBittorrent (5-30 min)
2. SSH into server
3. Manually run 7z extraction (5-10 min)
4. Create zombies directory
5. Move RAR files manually (1-5 min)
6. Open qBittorrent UI
7. Right-click torrent → Set Location (1 min)
8. Verify seeding resumed (1 min)

**Total Time**: ~15-50 minutes per download

### After
```bash
rarclean --path /mnt/downloads
```

**Total Time**: ~1-2 minutes (automatic)

**Time Saved**: 93-98% reduction in manual work

