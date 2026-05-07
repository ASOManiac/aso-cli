package asomaniac

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestReadConfigPermissionDenied verifies that ReadConfig returns a wrapped
// "read config" error (rather than panicking or returning nil) when the
// underlying file cannot be opened due to permissions.
func TestReadConfigPermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission semantics not applicable on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("permission checks are bypassed when running as root")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// Write a config, then strip read permission so ReadConfig fails at open.
	cfg := &Config{APIKey: "secret"}
	if err := WriteConfig(path, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })

	_, err := ReadConfig(path)
	if err == nil {
		t.Fatal("expected error from ReadConfig with no read permission")
	}
	if !strings.Contains(err.Error(), "read config") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "read config")
	}
}
