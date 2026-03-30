package asomaniac

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAnalyzeKeywords(t *testing.T) {
	want := KeywordAnalysis{
		Keyword:    "photo editor",
		Storefront: "US",
		Popularity: 72,
		Difficulty: 85,
		Confidence: "high",
		TotalApps:  1500,
		TopApps: []TopApp{
			{
				AppID:       "123456789",
				Name:        "SnapEdit",
				Developer:   "SnapDev Inc.",
				Icon:        "https://example.com/icon.png",
				Rating:      4.7,
				ReviewCount: 12000,
				Price:       "Free",
				Rank:        1,
			},
		},
		RelatedSearches: []string{"photo filter", "image editor"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/keywords/analyze" {
			t.Errorf("path = %q, want /keywords/analyze", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-api-key" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-api-key")
		}

		var body AnalyzeKeywordRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if len(body.Keywords) != 1 || body.Keywords[0] != "photo editor" {
			t.Errorf("keywords = %v, want [photo editor]", body.Keywords)
		}
		if body.Storefront != "US" {
			t.Errorf("storefront = %q, want US", body.Storefront)
		}

		resp := APIResponse[[]KeywordAnalysis]{Data: []KeywordAnalysis{want}}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-api-key")
	results, err := client.AnalyzeKeywords(context.Background(), []string{"photo editor"}, "US", nil)
	if err != nil {
		t.Fatalf("AnalyzeKeywords: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("results len = %d, want 1", len(results))
	}
	got := results[0]
	if got.Keyword != want.Keyword {
		t.Errorf("Keyword = %q, want %q", got.Keyword, want.Keyword)
	}
	if got.Popularity != want.Popularity {
		t.Errorf("Popularity = %d, want %d", got.Popularity, want.Popularity)
	}
	if got.Difficulty != want.Difficulty {
		t.Errorf("Difficulty = %d, want %d", got.Difficulty, want.Difficulty)
	}
	if len(got.TopApps) != 1 {
		t.Fatalf("TopApps len = %d, want 1", len(got.TopApps))
	}
	if got.TopApps[0].Name != "SnapEdit" {
		t.Errorf("TopApps[0].Name = %q, want %q", got.TopApps[0].Name, "SnapEdit")
	}
}

func TestGetProfile(t *testing.T) {
	name := "Hristo"
	want := UserProfile{
		ID:        "user-123",
		Email:     "hristo@example.com",
		Name:      &name,
		Plan:      "pro",
		CreatedAt: "2026-01-15T10:00:00Z",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GetProfile should hit /api/auth/me, not /api/v1/...
		if r.URL.Path != "/api/auth/me" {
			t.Errorf("path = %q, want /api/auth/me", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer my-key" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer my-key")
		}

		resp := APIResponse[UserProfile]{Data: want}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	// baseURL is srv.URL + "/api/v1" to simulate production config.
	client := NewClient(srv.URL+"/api/v1", "my-key")
	got, err := client.GetProfile(context.Background())
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}

	if got.ID != want.ID {
		t.Errorf("ID = %q, want %q", got.ID, want.ID)
	}
	if got.Email != want.Email {
		t.Errorf("Email = %q, want %q", got.Email, want.Email)
	}
	if got.Name == nil || *got.Name != name {
		t.Errorf("Name = %v, want %q", got.Name, name)
	}
	if got.Plan != want.Plan {
		t.Errorf("Plan = %q, want %q", got.Plan, want.Plan)
	}
}

func TestUnauthorizedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(APIError{
			Error: struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				Code:    "UNAUTHORIZED",
				Message: "Invalid API key",
			},
		}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "bad-key")
	_, err := client.AnalyzeKeywords(context.Background(), []string{"test"}, "US", nil)
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !strings.Contains(err.Error(), "UNAUTHORIZED") {
		t.Errorf("error = %q, want it to contain UNAUTHORIZED", err.Error())
	}
	if !strings.Contains(err.Error(), "Invalid API key") {
		t.Errorf("error = %q, want it to contain 'Invalid API key'", err.Error())
	}
}

func TestRateLimitedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		if err := json.NewEncoder(w).Encode(APIError{
			Error: struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				Code:    "RATE_LIMITED",
				Message: "Too many requests",
			},
		}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "key")
	_, err := client.GetUsage(context.Background())
	if err == nil {
		t.Fatal("expected error for 429 response")
	}
	if !strings.Contains(err.Error(), "RATE_LIMITED") {
		t.Errorf("error = %q, want it to contain RATE_LIMITED", err.Error())
	}
}

func TestNoAuthHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "" {
			t.Errorf("Authorization header should be empty, got %q", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(APIError{
			Error: struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				Code:    "UNAUTHORIZED",
				Message: "Missing API key",
			},
		}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	// Empty API key should not send Authorization header.
	client := NewClient(srv.URL, "")
	_, err := client.AnalyzeKeywords(context.Background(), []string{"test"}, "US", nil)
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
}

func TestNewClientFromConfig(t *testing.T) {
	cfg := &Config{
		APIKey:  "config-key",
		BaseURL: "https://custom.example.com/api/v1",
	}
	client := NewClientFromConfig(cfg)
	if client.apiKey != "config-key" {
		t.Errorf("apiKey = %q, want %q", client.apiKey, "config-key")
	}
	if client.baseURL != "https://custom.example.com/api/v1" {
		t.Errorf("baseURL = %q, want %q", client.baseURL, "https://custom.example.com/api/v1")
	}
}

func TestNewClientFromConfigDefaults(t *testing.T) {
	cfg := &Config{APIKey: "key"}
	client := NewClientFromConfig(cfg)
	if client.baseURL != DefaultBaseURL {
		t.Errorf("baseURL = %q, want default %q", client.baseURL, DefaultBaseURL)
	}
}

func TestBatchAnalyze(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/keywords/batch-analyze" {
			t.Errorf("path = %q, want /keywords/batch-analyze", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}

		var body BatchAnalyzeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if len(body.Keywords) != 2 {
			t.Errorf("keywords len = %d, want 2", len(body.Keywords))
		}

		resp := APIResponse[BatchResult]{
			Data: BatchResult{
				TotalKeywords:    2,
				TotalStorefronts: 1,
				Results: []BatchKeywordResult{
					{Keyword: "vpn", Storefronts: map[string]KeywordAnalysis{}},
					{Keyword: "proxy", Storefronts: map[string]KeywordAnalysis{}},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "key")
	got, err := client.BatchAnalyze(context.Background(), []string{"vpn", "proxy"}, []string{"US"})
	if err != nil {
		t.Fatalf("BatchAnalyze: %v", err)
	}
	if got.TotalKeywords != 2 {
		t.Errorf("TotalKeywords = %d, want 2", got.TotalKeywords)
	}
	if len(got.Results) != 2 {
		t.Errorf("Results len = %d, want 2", len(got.Results))
	}
}

func TestNon400ErrorWithoutAPIErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "key")
	_, err := client.GetDashboard(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "http 500") {
		t.Errorf("error = %q, want it to contain 'http 500'", err.Error())
	}
}
