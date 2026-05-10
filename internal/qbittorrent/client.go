package qbittorrent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

// Client handles qBittorrent Web API v2 interactions
type Client struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

// Torrent represents a qBittorrent torrent
type Torrent struct {
	Hash     string `json:"hash"`
	Name     string `json:"name"`
	SavePath string `json:"save_path"`
	State    string `json:"state"`
}

// New creates a new qBittorrent client
func New(baseURL, username, password string) (*Client, error) {
	// Ensure baseURL doesn't have trailing slash
	baseURL = strings.TrimRight(baseURL, "/")

	// Create cookie jar for session management
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		client: &http.Client{
			Jar: jar,
		},
	}, nil
}

// Login authenticates with qBittorrent
func (c *Client) Login() error {
	endpoint := fmt.Sprintf("%s/api/v2/auth/login", c.baseURL)

	data := url.Values{
		"username": {c.username},
		"password": {c.password},
	}

	resp, err := c.client.PostForm(endpoint, data)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetTorrents retrieves all torrents from qBittorrent
func (c *Client) GetTorrents() ([]Torrent, error) {
	endpoint := fmt.Sprintf("%s/api/v2/torrents/info", c.baseURL)

	resp, err := c.client.Get(endpoint)
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			// Session expired, re-login
			if err := c.Login(); err != nil {
				return nil, fmt.Errorf("re-authentication failed: %w", err)
			}
			// Retry
			return c.GetTorrents()
		}
		return nil, fmt.Errorf("failed to get torrents: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get torrents failed with status %d: %s", resp.StatusCode, string(body))
	}

	var torrents []Torrent
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		return nil, fmt.Errorf("failed to parse torrents: %w", err)
	}

	return torrents, nil
}

// FindTorrentByPath finds a torrent by its save path
// Returns nil if not found (not an error)
func (c *Client) FindTorrentByPath(path string) (*Torrent, error) {
	// Normalize path for comparison
	searchPath := normalizePath(path)

	torrents, err := c.GetTorrents()
	if err != nil {
		return nil, err
	}

	for i := range torrents {
		torrentPath := normalizePath(torrents[i].SavePath)
		if torrentPath == searchPath {
			return &torrents[i], nil
		}
	}

	return nil, nil
}

// SetLocation updates a torrent's save location
func (c *Client) SetLocation(hash, location string) error {
	endpoint := fmt.Sprintf("%s/api/v2/torrents/setLocation", c.baseURL)

	data := url.Values{
		"hashes":   {hash},
		"location": {location},
	}

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("set location request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set location failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetTorrent retrieves a specific torrent by hash
func (c *Client) GetTorrent(hash string) (*Torrent, error) {
	endpoint := fmt.Sprintf("%s/api/v2/torrents/info?hashes=%s", c.baseURL, hash)

	resp, err := c.client.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get torrent: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get torrent failed with status %d", resp.StatusCode)
	}

	var torrents []Torrent
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		return nil, fmt.Errorf("failed to parse torrent: %w", err)
	}

	if len(torrents) == 0 {
		return nil, nil
	}

	return &torrents[0], nil
}

// normalizePath normalizes a file path for comparison
// Removes trailing slashes and returns absolute paths
func normalizePath(path string) string {
	return strings.TrimRight(path, "/")
}
