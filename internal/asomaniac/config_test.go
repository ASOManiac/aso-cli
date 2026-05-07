package asomaniac

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadWriteConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{
		APIKey:            "test-key-123",
		BaseURL:           "https://custom.example.com/api/v1",
		DefaultStorefront: "BG",
		OutputFormat:      "table",
	}

	if err := WriteConfig(path, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	got, err := ReadConfig(path)
	if err != nil {
		t.Fatalf("ReadConfig: %v", err)
	}

	if got.APIKey != cfg.APIKey {
		t.Errorf("APIKey = %q, want %q", got.APIKey, cfg.APIKey)
	}
	if got.BaseURL != cfg.BaseURL {
		t.Errorf("BaseURL = %q, want %q", got.BaseURL, cfg.BaseURL)
	}
	if got.DefaultStorefront != cfg.DefaultStorefront {
		t.Errorf("DefaultStorefront = %q, want %q", got.DefaultStorefront, cfg.DefaultStorefront)
	}
	if got.OutputFormat != cfg.OutputFormat {
		t.Errorf("OutputFormat = %q, want %q", got.OutputFormat, cfg.OutputFormat)
	}
}

func TestWriteConfigCreatesDir(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b", "config.json")

	cfg := &Config{APIKey: "key", BaseURL: DefaultBaseURL}
	if err := WriteConfig(nested, cfg); err != nil {
		t.Fatalf("WriteConfig with nested dir: %v", err)
	}

	if _, err := os.Stat(nested); err != nil {
		t.Errorf("config file not created at nested path: %v", err)
	}
}

func TestWriteConfigPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{APIKey: "secret"}
	if err := WriteConfig(path, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	if path == "" {
		t.Fatal("DefaultConfigPath returned empty string")
	}
	if !strings.HasSuffix(path, filepath.Join(".asomaniac", "config.json")) {
		t.Errorf("DefaultConfigPath = %q, want suffix .asomaniac/config.json", path)
	}
}

func TestReadConfigMissing(t *testing.T) {
	_, err := ReadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Fatal("ReadConfig on missing file should return error")
	}
	if !strings.Contains(err.Error(), "read config") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "read config")
	}
}

func TestReadConfigInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := ReadConfig(path)
	if err == nil {
		t.Fatal("ReadConfig on invalid JSON should return error")
	}
	if !strings.Contains(err.Error(), "parse config") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "parse config")
	}
}

func TestLoadOrDefaultMissing(t *testing.T) {
	cfg := LoadOrDefault("/nonexistent/path/config.json")
	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, DefaultBaseURL)
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("OutputFormat = %q, want %q", cfg.OutputFormat, "json")
	}
	if cfg.APIKey != "" {
		t.Errorf("APIKey = %q, want empty", cfg.APIKey)
	}
}

func TestLoadOrDefaultFillsDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// Write config with empty BaseURL and OutputFormat.
	cfg := &Config{APIKey: "my-key"}
	if err := WriteConfig(path, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	got := LoadOrDefault(path)
	if got.APIKey != "my-key" {
		t.Errorf("APIKey = %q, want %q", got.APIKey, "my-key")
	}
	if got.BaseURL != DefaultBaseURL {
		t.Errorf("BaseURL = %q, want %q", got.BaseURL, DefaultBaseURL)
	}
	if got.OutputFormat != "json" {
		t.Errorf("OutputFormat = %q, want %q", got.OutputFormat, "json")
	}
}

func TestIsAuthenticated(t *testing.T) {
	cfg := &Config{}
	if cfg.IsAuthenticated() {
		t.Error("empty APIKey should not be authenticated")
	}

	cfg.APIKey = "test-key"
	if !cfg.IsAuthenticated() {
		t.Error("non-empty APIKey should be authenticated")
	}
}

func TestDeleteConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{APIKey: "key"}
	if err := WriteConfig(path, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	if err := DeleteConfig(path); err != nil {
		t.Fatalf("DeleteConfig: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("config file still exists after delete")
	}
}
