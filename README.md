# rarclean

A QoL service for TrueNAS Scale that automatically extracts RAR archives from qBittorrent downloads and organizes the extracted files.

## Features

- 🔍 **RAR Detection**: Automatically detects RAR archives in qBittorrent downloads
- 📦 **Extraction**: Extracts RAR files automatically
- 📁 **Organization**: Organizes extracted files according to configurable patterns
- 🧹 **Cleanup**: Removes empty directories after extraction
- ⏸️ **Torrent Control**: Can pause/resume torrents during extraction
- 🐳 **Docker Ready**: Containerized for easy TrueNAS Scale deployment

## Installation

### Docker (Recommended for TrueNAS Scale)

1. Clone the repository
2. Copy `config.example.json` to `config.json`
3. Configure your settings in `config.json`
4. Run with Docker Compose:

```bash
docker-compose up -d
```

### Binary Build

```bash
make build
make run
```

## Configuration

Edit `config.json` to configure:

- **qBittorrent**: Connection details and credentials
- **Paths**: Download, extraction, and organization directories
- **Extraction**: Enable/disable extraction, cleanup options
- **Organization**: File organization patterns
- **Service**: Polling interval and logging

## Building

```bash
# Build binary
make build

# Build Docker image
make docker-build

# Run tests
make test

# Format code
make fmt
```

## Development

The project is organized as follows:

- `cmd/rarclean/` - Main application entry point
- `internal/extractor/` - RAR extraction logic
- `internal/qbittorrent/` - qBittorrent API client
- `internal/mover/` - File movement and organization

## License

MIT
