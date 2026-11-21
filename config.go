package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the connection settings read from the config file.
type Config struct {
	// Endpoint is the Proxmox VE URL, e.g. "https://192.168.1.12:8006".
	Endpoint string `json:"endpoint"`
	// Token is the API token in the form "user@realm!tokenid=secret".
	Token string `json:"token"`
	// Insecure skips TLS certificate verification (for self-signed PVE certificates).
	Insecure bool `json:"insecure"`
}

// defaultConfigPath returns ~/.config/pvebar/config.json.
func defaultConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "config.json"
	}

	return filepath.Join(dir, "pvebar", "config.json")
}

// LoadConfig reads and validates the config file at path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	if cfg.Endpoint == "" {
		return nil, errors.New("config: endpoint is required")
	}

	if cfg.Token == "" {
		return nil, errors.New("config: token is required")
	}

	return &cfg, nil
}
