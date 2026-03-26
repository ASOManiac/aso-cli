package asomaniac

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DefaultBaseURL is the base URL for the ASO Maniac API v1.
const DefaultBaseURL = "https://asomaniac.com/api/v1"

// Config holds the CLI configuration for connecting to the ASO Maniac API.
type Config struct {
	APIKey            string `json:"api_key"`
	BaseURL           string `json:"base_url"`
	DefaultStorefront string `json:"default_storefront,omitempty"`
	OutputFormat      string `json:"output_format,omitempty"`
}

// DefaultConfigPath returns the default path for the config file.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".asomaniac", "config.json")
	}
	return filepath.Join(home, ".asomaniac", "config.json")
}

// ReadConfig reads and parses a config file from the given path.
func ReadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// WriteConfig writes the config to the given path, creating directories as needed.
func WriteConfig(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

// LoadOrDefault reads the config from path, returning sensible defaults on failure.
func LoadOrDefault(path string) *Config {
	cfg, err := ReadConfig(path)
	if err != nil {
		return &Config{BaseURL: DefaultBaseURL, OutputFormat: "json"}
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.OutputFormat == "" {
		cfg.OutputFormat = "json"
	}
	return cfg
}

// DeleteConfig removes the config file at the given path.
func DeleteConfig(path string) error {
	return os.Remove(path)
}

// IsAuthenticated returns true if the config has an API key set.
func (c *Config) IsAuthenticated() bool {
	return c.APIKey != ""
}
