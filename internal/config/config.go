package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the application configuration
type Config struct {
	QBittorrent QBittorrentConfig `json:"qbittorrent"`
	Paths       PathsConfig       `json:"paths"`
}

// QBittorrentConfig holds qBittorrent connection details
type QBittorrentConfig struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// PathsConfig holds path mappings
type PathsConfig struct {
	Television string `json:"television"`
	Movies     string `json:"movies"`
	Zombies    string `json:"zombies"`
}

var globalConfig *Config

// Load loads configuration from a JSON file
func Load(path string) (*Config, error) {
	if path == "" {
		path = "config.json"
	}

	// Expand home directory if needed
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(homeDir, path[1:])
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate configuration
	if cfg.QBittorrent.URL == "" {
		return nil, fmt.Errorf("qbittorrent.url is required")
	}
	if cfg.QBittorrent.Username == "" {
		return nil, fmt.Errorf("qbittorrent.username is required")
	}
	if cfg.QBittorrent.Password == "" {
		return nil, fmt.Errorf("qbittorrent.password is required")
	}
	if cfg.Paths.Television == "" {
		return nil, fmt.Errorf("paths.television is required")
	}
	if cfg.Paths.Movies == "" {
		return nil, fmt.Errorf("paths.movies is required")
	}
	if cfg.Paths.Zombies == "" {
		return nil, fmt.Errorf("paths.zombies is required")
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Get returns the global config instance
func Get() *Config {
	return globalConfig
}

// Set sets the global config instance (useful for testing)
func Set(cfg *Config) {
	globalConfig = cfg
}
