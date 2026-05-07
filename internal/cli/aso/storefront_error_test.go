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

// TestKeywordsAnalyzeInvalidStorefront verifies the error path when the API
// rejects an unknown storefront code: the server's error code/message is
// surfaced verbatim to the caller rather than being swallowed.
func TestKeywordsAnalyzeInvalidStorefront(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(asomaniac.APIError{
			Error: struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				Code:    "INVALID_STOREFRONT",
				Message: "Unknown storefront code: ZZ",
			},
		})
	}))
	defer srv.Close()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	cfg := &asomaniac.Config{APIKey: "test-key", BaseURL: srv.URL + "/api/v1"}
	if err := asomaniac.WriteConfig(configPath, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := runKeywordsAnalyze(context.Background(), configPath, []string{"vpn"}, "ZZ", nil, nil, false, false, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid storefront, got nil")
	}
	if !strings.Contains(err.Error(), "INVALID_STOREFRONT") {
		t.Errorf("error = %q, want it to contain INVALID_STOREFRONT", err.Error())
	}
}
