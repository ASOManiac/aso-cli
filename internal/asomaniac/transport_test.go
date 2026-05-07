package asomaniac

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// TestRateLimitNoSilentRetry documents the current contract: a 429 response
// surfaces immediately as an error rather than being silently retried by the
// client. Callers (or skills) are expected to back off and retry themselves.
func TestRateLimitNoSilentRetry(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(APIError{
			Error: struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				Code:    "RATE_LIMITED",
				Message: "Too many requests",
			},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	_, err := client.AnalyzeKeywords(context.Background(), []string{"vpn"}, "US", nil)
	if err == nil {
		t.Fatal("expected error for 429 response")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("server received %d requests, want exactly 1 (no silent retry)", got)
	}
	if !strings.Contains(err.Error(), "RATE_LIMITED") {
		t.Errorf("error = %q, want it to mention RATE_LIMITED", err.Error())
	}
}

// TestAuthHeaderInjection verifies the API key is injected as a Bearer token
// on every request, including non-keyword endpoints like /usage.
func TestAuthHeaderInjection(t *testing.T) {
	const key = "asm_k_live_abc123"

	var seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(APIResponse[UsageStats]{
			Data: UsageStats{},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, key)
	if _, err := client.GetUsage(context.Background()); err != nil {
		t.Fatalf("GetUsage: %v", err)
	}

	want := "Bearer " + key
	if seen != want {
		t.Errorf("Authorization header = %q, want %q", seen, want)
	}
}

// TestRequireAuthMissingKey verifies that ResolveConfig surfaces an
// unauthenticated state when no API key is configured (no env var, no file).
func TestRequireAuthMissingKey(t *testing.T) {
	t.Setenv(EnvAPIKey, "")

	resolved := ResolveConfig("/nonexistent/path/config.json")
	if resolved.IsAuthenticated() {
		t.Fatal("expected unauthenticated state when no key is configured")
	}
	if resolved.Source != KeySourceNone {
		t.Errorf("Source = %q, want %q", resolved.Source, KeySourceNone)
	}
}
