package aso

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ASOManiac/aso-cli/internal/asomaniac"
)

func TestLogoutRemovesConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	// Write a config to delete.
	cfg := &asomaniac.Config{APIKey: "test-key", BaseURL: asomaniac.DefaultBaseURL}
	if err := asomaniac.WriteConfig(configPath, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	var buf bytes.Buffer
	err := runLogout(configPath, &buf)
	if err != nil {
		t.Fatalf("runLogout: %v", err)
	}

	if !strings.Contains(buf.String(), "Logged out") {
		t.Errorf("output = %q, want it to contain 'Logged out'", buf.String())
	}

	// Verify config is actually deleted.
	_, err = asomaniac.ReadConfig(configPath)
	if err == nil {
		t.Error("config file still exists after logout")
	}
}

func TestLogoutNotLoggedIn(t *testing.T) {
	// Point at a non-existent path.
	configPath := filepath.Join(t.TempDir(), "nonexistent", "config.json")

	var buf bytes.Buffer
	err := runLogout(configPath, &buf)
	if err != nil {
		t.Fatalf("runLogout should not error for missing config: %v", err)
	}

	if !strings.Contains(buf.String(), "Not logged in") {
		t.Errorf("output = %q, want it to contain 'Not logged in'", buf.String())
	}
}
