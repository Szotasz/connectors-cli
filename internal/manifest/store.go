package manifest

import (
	"encoding/json"
	"os"

	"github.com/Szotasz/conn-cli/internal/api"
	"github.com/Szotasz/conn-cli/internal/config"
)

func Save(cfg *config.Config, m *api.Manifest) error {
	if err := cfg.EnsureDataDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfg.ManifestPath(), data, 0600)
}

func Load(cfg *config.Config) (*api.Manifest, error) {
	data, err := os.ReadFile(cfg.ManifestPath())
	if err != nil {
		return nil, err
	}
	var m api.Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
