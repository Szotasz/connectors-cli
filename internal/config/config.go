package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Token   string
	BaseURL string
	DataDir string
}

func Load() *Config {
	token := os.Getenv("CONNECTORS_HU_TOKEN")
	if token == "" {
		if legacy := os.Getenv("CONN_HU_TOKEN"); legacy != "" {
			fmt.Fprintln(os.Stderr, "Warning: CONN_HU_TOKEN is deprecated, use CONNECTORS_HU_TOKEN")
			token = legacy
		}
	}
	baseURL := os.Getenv("CONNECTORS_HU_URL")
	if baseURL == "" {
		baseURL = os.Getenv("CONN_HU_URL")
	}
	if baseURL == "" {
		baseURL = "https://api.connectors.hu"
	}

	// Don't ever send a Bearer token over plain HTTP. Hostile env vars or
	// a typo could otherwise leak the token to a passive listener. The
	// loopback dev exception is opt-in via CONNECTORS_HU_ALLOW_INSECURE=1.
	if err := validateBaseURL(baseURL); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid CONNECTORS_HU_URL: %v\n", err)
		os.Exit(1)
	}

	home, _ := os.UserHomeDir()
	dataDir := filepath.Join(home, ".config", "connectors-hu")

	return &Config{
		Token:   token,
		BaseURL: baseURL,
		DataDir: dataDir,
	}
}

func validateBaseURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("not a valid URL: %w", err)
	}
	if parsed.Scheme == "https" {
		return nil
	}
	if parsed.Scheme == "http" {
		host := strings.ToLower(parsed.Hostname())
		if host == "localhost" || host == "127.0.0.1" {
			if os.Getenv("CONNECTORS_HU_ALLOW_INSECURE") == "1" {
				return nil
			}
			return fmt.Errorf("http:// loopback requires CONNECTORS_HU_ALLOW_INSECURE=1")
		}
		return fmt.Errorf("http:// not allowed for remote hosts; use https://")
	}
	return fmt.Errorf("unsupported scheme %q; only https (and loopback http) is allowed", parsed.Scheme)
}

func (c *Config) ManifestPath() string {
	return filepath.Join(c.DataDir, "manifest.json")
}

func (c *Config) EnsureDataDir() error {
	return os.MkdirAll(c.DataDir, 0700)
}
