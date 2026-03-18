package encryption

import (
	"os"
	"path/filepath"
	"testing"

	"howett.net/plist"
)

func TestUpdatePlistExemption_ExistingTrueSetsFalse(t *testing.T) {
	plistPath := writeTestInfoPlist(t, plist.XMLFormat, map[string]any{
		"CFBundleIdentifier":            "com.example.app",
		"ITSAppUsesNonExemptEncryption": true,
	})

	if err := updatePlistExemption(plistPath); err != nil {
		t.Fatalf("updatePlistExemption() error: %v", err)
	}

	format, payload := readTestInfoPlist(t, plistPath)
	if format != plist.XMLFormat {
		t.Fatalf("expected XML plist format, got %d", format)
	}

	value, ok := payload["ITSAppUsesNonExemptEncryption"].(bool)
	if !ok {
		t.Fatalf("expected boolean ITSAppUsesNonExemptEncryption, got %#v", payload["ITSAppUsesNonExemptEncryption"])
	}
	if value {
		t.Fatal("expected ITSAppUsesNonExemptEncryption to be set to false")
	}
}

func TestUpdatePlistExemption_UpdatesBinaryPlist(t *testing.T) {
	plistPath := writeTestInfoPlist(t, plist.BinaryFormat, map[string]any{
		"CFBundleIdentifier": "com.example.binary",
	})

	if err := updatePlistExemption(plistPath); err != nil {
		t.Fatalf("updatePlistExemption() error: %v", err)
	}

	format, payload := readTestInfoPlist(t, plistPath)
	if format != plist.BinaryFormat {
		t.Fatalf("expected binary plist format, got %d", format)
	}

	value, ok := payload["ITSAppUsesNonExemptEncryption"].(bool)
	if !ok {
		t.Fatalf("expected boolean ITSAppUsesNonExemptEncryption, got %#v", payload["ITSAppUsesNonExemptEncryption"])
	}
	if value {
		t.Fatal("expected ITSAppUsesNonExemptEncryption to be set to false")
	}
}

func writeTestInfoPlist(t *testing.T, format int, payload map[string]any) string {
	t.Helper()

	data, err := plist.Marshal(payload, format)
	if err != nil {
		t.Fatalf("plist.Marshal() error: %v", err)
	}

	plistPath := filepath.Join(t.TempDir(), "Info.plist")
	if err := os.WriteFile(plistPath, data, 0o644); err != nil {
		t.Fatalf("os.WriteFile() error: %v", err)
	}

	return plistPath
}

func readTestInfoPlist(t *testing.T, plistPath string) (int, map[string]any) {
	t.Helper()

	data, err := os.ReadFile(plistPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error: %v", err)
	}

	var payload map[string]any
	format, err := plist.Unmarshal(data, &payload)
	if err != nil {
		t.Fatalf("plist.Unmarshal() error: %v", err)
	}

	return format, payload
}
