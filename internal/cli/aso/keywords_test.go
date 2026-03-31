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

func TestKeywordsAnalyzeJSON(t *testing.T) {
	kw1 := asomaniac.KeywordAnalysis{
		Keyword:    "camera",
		Storefront: "US",
		Popularity: 72,
		Difficulty: 65,
		Confidence: "high",
		TotalApps:  1200,
		TopApps: []asomaniac.TopApp{
			{AppID: "111", Name: "CamApp", Developer: "CamDev", Rating: 4.5, ReviewCount: 5000, Price: "Free", Rank: 1},
		},
		RelatedSearches: []string{"photo", "picture"},
	}
	kw2 := asomaniac.KeywordAnalysis{
		Keyword:    "photo",
		Storefront: "US",
		Popularity: 85,
		Difficulty: 90,
		Confidence: "high",
		TotalApps:  3000,
		TopApps: []asomaniac.TopApp{
			{AppID: "222", Name: "PhotoApp", Developer: "PhotoDev", Rating: 4.8, ReviewCount: 20000, Price: "Free", Rank: 1},
		},
		RelatedSearches: []string{"camera", "image"},
	}

	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/api/v1/keywords/analyze" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body asomaniac.AnalyzeKeywordRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		// Return analysis for each requested keyword.
		results := make([]asomaniac.KeywordAnalysis, 0, len(body.Keywords))
		for _, kw := range body.Keywords {
			switch kw {
			case "camera":
				results = append(results, kw1)
			case "photo":
				results = append(results, kw2)
			default:
				t.Errorf("unexpected keyword: %s", kw)
				http.Error(w, "bad keyword", http.StatusBadRequest)
				return
			}
		}
		callCount++
		resp := asomaniac.APIResponse[[]asomaniac.KeywordAnalysis]{Data: results}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer srv.Close()

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
	err := runKeywordsAnalyze(context.Background(), configPath, []string{"camera", "photo"}, "US", nil, nil, &buf)
	if err != nil {
		t.Fatalf("runKeywordsAnalyze: %v", err)
	}

	// Parse the JSON output.
	var results []asomaniac.KeywordAnalysis
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("parse output JSON: %v\nraw output:\n%s", err, buf.String())
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Keyword != "camera" {
		t.Errorf("results[0].Keyword = %q, want %q", results[0].Keyword, "camera")
	}
	if results[0].Popularity != 72 {
		t.Errorf("results[0].Popularity = %d, want 72", results[0].Popularity)
	}
	if results[1].Keyword != "photo" {
		t.Errorf("results[1].Keyword = %q, want %q", results[1].Keyword, "photo")
	}
	if results[1].Popularity != 85 {
		t.Errorf("results[1].Popularity = %d, want 85", results[1].Popularity)
	}
	if callCount != 1 {
		t.Errorf("expected 1 API call (batch POST), got %d", callCount)
	}
}

func TestKeywordsAnalyzeRequiresAuth(t *testing.T) {
	// Non-existent config path → not logged in.
	configPath := filepath.Join(t.TempDir(), "nonexistent", "config.json")

	var buf bytes.Buffer
	err := runKeywordsAnalyze(context.Background(), configPath, []string{"camera"}, "US", nil, nil, &buf)
	if err == nil {
		t.Fatal("expected error for unauthenticated user, got nil")
	}
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("error = %q, want it to contain 'not logged in'", err.Error())
	}
}

func TestKeywordsAnalyzeEmptyAPIKey(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	cfg := &asomaniac.Config{
		APIKey:  "",
		BaseURL: asomaniac.DefaultBaseURL,
	}
	if err := asomaniac.WriteConfig(configPath, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	var buf bytes.Buffer
	err := runKeywordsAnalyze(context.Background(), configPath, []string{"camera"}, "US", nil, nil, &buf)
	if err == nil {
		t.Fatal("expected error for empty API key, got nil")
	}
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("error = %q, want it to contain 'not logged in'", err.Error())
	}
}
