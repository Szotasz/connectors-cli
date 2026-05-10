package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Token   string
	BaseURL string
	DataDir string
}

func Load() *Config {
	token := os.Getenv("CONN_HU_TOKEN")
	baseURL := os.Getenv("CONN_HU_URL")
	if baseURL == "" {
		baseURL = "https://api.connectors.hu"
	}

	home, _ := os.UserHomeDir()
	dataDir := filepath.Join(home, ".config", "conn-hu")

	return &Config{
		Token:   token,
		BaseURL: baseURL,
		DataDir: dataDir,
	}
}

func (c *Config) ManifestPath() string {
	return filepath.Join(c.DataDir, "manifest.json")
}

func (c *Config) EnsureDataDir() error {
	return os.MkdirAll(c.DataDir, 0700)
}
