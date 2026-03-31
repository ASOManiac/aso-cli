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

func TestCompetitorsJSON(t *testing.T) {
	competitors := []asomaniac.CompetitorAnalysis{
		{
			App: asomaniac.AppMetadata{
				AppID:     "111",
				Name:      "MyApp",
				Developer: "Dev1",
				Rating:    4.5,
			},
			Competitor: asomaniac.AppMetadata{
				AppID:     "222",
				Name:      "RivalApp",
				Developer: "Dev2",
				Rating:    4.2,
			},
			SharedKeywords:     15,
			UniqueToApp:        8,
			UniqueToCompetitor: 12,
			KeywordOverlap: []asomaniac.KeywordOverlap{
				{Keyword: "camera", Storefront: "US", AppRank: 3, CompetitorRank: 7, Popularity: 72},
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/api/v1/competitors/111" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if sf := r.URL.Query().Get("storefront"); sf != "US" {
			t.Errorf("storefront = %q, want %q", sf, "US")
		}
		if err := json.NewEncoder(w).Encode(asomaniac.APIResponse[[]asomaniac.CompetitorAnalysis]{Data: competitors}); err != nil {
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
	err := runCompetitors(context.Background(), configPath, "111", "US", nil, &buf)
	if err != nil {
		t.Fatalf("runCompetitors: %v", err)
	}

	var results []asomaniac.CompetitorAnalysis
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("parse output JSON: %v\nraw output:\n%s", err, buf.String())
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].SharedKeywords != 15 {
		t.Errorf("SharedKeywords = %d, want 15", results[0].SharedKeywords)
	}
	if results[0].Competitor.Name != "RivalApp" {
		t.Errorf("Competitor.Name = %q, want %q", results[0].Competitor.Name, "RivalApp")
	}
}

func TestCompetitorsRequiresAuth(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "nonexistent", "config.json")

	var buf bytes.Buffer
	err := runCompetitors(context.Background(), configPath, "111", "US", nil, &buf)
	if err == nil {
		t.Fatal("expected error for unauthenticated user, got nil")
	}
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("error = %q, want it to contain 'not logged in'", err.Error())
	}
}
