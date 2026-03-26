package aso

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ASOManiac/aso-cli/internal/asomaniac"
)

func TestWhoamiAuthenticated(t *testing.T) {
	name := "Hristo"
	profile := asomaniac.UserProfile{
		ID:        "user-123",
		Email:     "hristo@example.com",
		Name:      &name,
		Plan:      "pro",
		CreatedAt: "2026-01-15T10:00:00Z",
	}
	usage := asomaniac.UsageStats{
		Plan: "pro",
	}
	usage.APICalls.Today = 5
	usage.APICalls.ThisMonth = 42
	usage.APICalls.Limit = 10000
	usage.TrackedApps.Current = 3
	usage.TrackedApps.Limit = 50
	usage.TrackedKeywords.Current = 25
	usage.TrackedKeywords.Limit = 500

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.URL.Path == "/api/auth/me":
			json.NewEncoder(w).Encode(asomaniac.APIResponse[asomaniac.UserProfile]{Data: profile})
		case r.URL.Path == "/api/v1/usage":
			json.NewEncoder(w).Encode(asomaniac.APIResponse[asomaniac.UsageStats]{Data: usage})
		default:
			t.Errorf("unexpected request path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// Write a config pointing at the test server.
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	cfg := &asomaniac.Config{
		APIKey:  "test-key",
		BaseURL: srv.URL + "/api/v1",
	}
	if err := asomaniac.WriteConfig(configPath, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	var buf bytes.Buffer
	err := runWhoami(context.Background(), configPath, &buf)
	if err != nil {
		t.Fatalf("runWhoami: %v", err)
	}

	output := buf.String()

	// Check key fields are present in output.
	checks := []string{
		"hristo@example.com",
		"Hristo",
		"pro",
		"5",
		"42 / 10000",
		"3 / 50",
		"25 / 500",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("output missing %q\nfull output:\n%s", check, output)
		}
	}
}

func TestWhoamiUnauthenticated(t *testing.T) {
	// Use a non-existent config path.
	configPath := filepath.Join(t.TempDir(), "nonexistent", "config.json")

	var buf bytes.Buffer
	err := runWhoami(context.Background(), configPath, &buf)
	if err == nil {
		t.Fatal("expected error for unauthenticated user, got nil")
	}
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("error = %q, want it to contain 'not logged in'", err.Error())
	}
}

func TestWhoamiEmptyAPIKey(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	// Write config with empty API key.
	cfg := &asomaniac.Config{
		APIKey:  "",
		BaseURL: asomaniac.DefaultBaseURL,
	}
	if err := asomaniac.WriteConfig(configPath, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	var buf bytes.Buffer
	err := runWhoami(context.Background(), configPath, &buf)
	if err == nil {
		t.Fatal("expected error for empty API key, got nil")
	}
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("error = %q, want it to contain 'not logged in'", err.Error())
	}
}
