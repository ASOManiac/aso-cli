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
	"time"

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
		resp := asomaniac.AnalyzeResponse{
			Data: results,
			Meta: asomaniac.AnalyzeMeta{Pending: []string{}, TimedOut: false, ElapsedMs: 25},
		}
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

	var stdout, stderr bytes.Buffer
	err := runKeywordsAnalyze(context.Background(), configPath, []string{"camera", "photo"}, "US", nil, nil, false, false, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runKeywordsAnalyze: %v", err)
	}

	// stdout should be just the data array, no meta.
	var results []asomaniac.KeywordAnalysis
	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("parse stdout JSON: %v\nraw output:\n%s", err, stdout.String())
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
	if stderr.Len() != 0 {
		t.Errorf("stderr should be empty when no timeout, got: %s", stderr.String())
	}
}

func TestKeywordsAnalyzeRequiresAuth(t *testing.T) {
	// Non-existent config path → not logged in.
	configPath := filepath.Join(t.TempDir(), "nonexistent", "config.json")

	var stdout, stderr bytes.Buffer
	err := runKeywordsAnalyze(context.Background(), configPath, []string{"camera"}, "US", nil, nil, false, false, &stdout, &stderr)
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

	var stdout, stderr bytes.Buffer
	err := runKeywordsAnalyze(context.Background(), configPath, []string{"camera"}, "US", nil, nil, false, false, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for empty API key, got nil")
	}
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("error = %q, want it to contain 'not logged in'", err.Error())
	}
}

func TestKeywordsAnalyzeWarnsOnTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := asomaniac.AnalyzeResponse{
			Data: []asomaniac.KeywordAnalysis{
				{Keyword: "alpha", Storefront: "US", Popularity: 50},
			},
			Meta: asomaniac.AnalyzeMeta{
				Pending:   []string{"beta", "gamma"},
				TimedOut:  true,
				ElapsedMs: 50000,
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	cfg := &asomaniac.Config{APIKey: "test-key", BaseURL: srv.URL + "/api/v1"}
	if err := asomaniac.WriteConfig(configPath, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := runKeywordsAnalyze(context.Background(), configPath, []string{"alpha", "beta", "gamma"}, "US", nil, nil, false, false, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runKeywordsAnalyze: %v", err)
	}

	// stdout: just data array.
	var results []asomaniac.KeywordAnalysis
	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("parse stdout JSON: %v\nraw:\n%s", err, stdout.String())
	}
	if len(results) != 1 {
		t.Errorf("results len = %d, want 1", len(results))
	}

	// stderr: warning with pending list.
	se := stderr.String()
	if !strings.Contains(se, "warning") {
		t.Errorf("stderr missing warning, got: %q", se)
	}
	if !strings.Contains(se, "2 keywords pending") {
		t.Errorf("stderr missing pending count, got: %q", se)
	}
	if !strings.Contains(se, "beta") || !strings.Contains(se, "gamma") {
		t.Errorf("stderr missing pending keywords, got: %q", se)
	}
}

func TestKeywordsAnalyzeRetryPending(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var body asomaniac.AnalyzeKeywordRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		calls++
		switch calls {
		case 1:
			// First call: alpha, beta, gamma → resolve alpha, pending beta and gamma.
			resp := asomaniac.AnalyzeResponse{
				Data: []asomaniac.KeywordAnalysis{
					{Keyword: "alpha", Storefront: "US", Popularity: 50},
				},
				Meta: asomaniac.AnalyzeMeta{
					Pending:   []string{"beta", "gamma"},
					TimedOut:  true,
					ElapsedMs: 50000,
				},
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				t.Fatalf("encode: %v", err)
			}
		case 2:
			// Retry call: beta, gamma → resolve both, no pending.
			if len(body.Keywords) != 2 || body.Keywords[0] != "beta" || body.Keywords[1] != "gamma" {
				t.Errorf("retry keywords = %v, want [beta gamma]", body.Keywords)
			}
			resp := asomaniac.AnalyzeResponse{
				Data: []asomaniac.KeywordAnalysis{
					{Keyword: "beta", Storefront: "US", Popularity: 30},
					{Keyword: "gamma", Storefront: "US", Popularity: 40},
				},
				Meta: asomaniac.AnalyzeMeta{Pending: []string{}, TimedOut: false, ElapsedMs: 1000},
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				t.Fatalf("encode: %v", err)
			}
		default:
			t.Errorf("unexpected call %d", calls)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	cfg := &asomaniac.Config{APIKey: "test-key", BaseURL: srv.URL + "/api/v1"}
	if err := asomaniac.WriteConfig(configPath, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := runKeywordsAnalyze(context.Background(), configPath, []string{"alpha", "beta", "gamma"}, "US", nil, nil, true, false, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runKeywordsAnalyze: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls (initial + retry), got %d", calls)
	}

	var results []asomaniac.KeywordAnalysis
	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("parse stdout JSON: %v\nraw:\n%s", err, stdout.String())
	}
	if len(results) != 3 {
		t.Errorf("merged results len = %d, want 3", len(results))
	}
	// stderr: no warning since retry resolved everything.
	if strings.Contains(stderr.String(), "warning") {
		t.Errorf("stderr should not warn after successful retry, got: %q", stderr.String())
	}
}

func TestKeywordsAnalyzeFullFlag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := asomaniac.AnalyzeResponse{
			Data: []asomaniac.KeywordAnalysis{{Keyword: "alpha", Storefront: "US"}},
			Meta: asomaniac.AnalyzeMeta{Pending: []string{}, TimedOut: false, ElapsedMs: 99},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	cfg := &asomaniac.Config{APIKey: "test-key", BaseURL: srv.URL + "/api/v1"}
	if err := asomaniac.WriteConfig(configPath, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := runKeywordsAnalyze(context.Background(), configPath, []string{"alpha"}, "US", nil, nil, false, true, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runKeywordsAnalyze: %v", err)
	}

	var full asomaniac.AnalyzeResponse
	if err := json.Unmarshal(stdout.Bytes(), &full); err != nil {
		t.Fatalf("parse full response: %v\nraw:\n%s", err, stdout.String())
	}
	if full.Meta.ElapsedMs != 99 {
		t.Errorf("Meta.ElapsedMs = %d, want 99", full.Meta.ElapsedMs)
	}
}

func TestKeywordsRecommendWarnsOnTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := asomaniac.RecommendResponse{
			Data: []asomaniac.KeywordRecommendation{{Keyword: "photo filter", Popularity: 50, Source: "ai"}},
			Meta: asomaniac.RecommendMeta{ElapsedMs: 50000, TimedOut: true},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	cfg := &asomaniac.Config{APIKey: "test-key", BaseURL: srv.URL + "/api/v1"}
	if err := asomaniac.WriteConfig(configPath, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := runKeywordsRecommend(context.Background(), configPath, "photo", "US", 25, nil, false, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runKeywordsRecommend: %v", err)
	}

	// stdout: data array only.
	var recs []asomaniac.KeywordRecommendation
	if err := json.Unmarshal(stdout.Bytes(), &recs); err != nil {
		t.Fatalf("parse stdout: %v\nraw:\n%s", err, stdout.String())
	}
	if len(recs) != 1 {
		t.Errorf("recs len = %d, want 1", len(recs))
	}

	se := stderr.String()
	if !strings.Contains(se, "warning") || !strings.Contains(se, "timed out") || !strings.Contains(se, "50000ms") {
		t.Errorf("stderr missing expected warning, got: %q", se)
	}
}

// TestKeywordsBatchAsyncFlow exercises the async submit + poll + stream flow.
// The mock server returns 202 on POST, then progresses items DONE across two
// polls. The CLI should stream JSONL on stdout for each completed item and
// exit 0 when status flips to COMPLETED.
func TestKeywordsBatchAsyncFlow(t *testing.T) {
	doneCalls := 0
	var serverURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/keywords/batch-analyze"):
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": asomaniac.JobSubmitResponse{
					JobID: "job-1", StatusURL: serverURL + "/api/v1/keywords/jobs/job-1", TotalCount: 2,
				},
			})
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/keywords/jobs/job-1"):
			doneCalls++
			items := []asomaniac.JobItemSummary{
				{Term: "camera", Storefront: "US", Status: "DONE", Result: &asomaniac.KeywordAnalysis{Keyword: "camera", Storefront: "US", Popularity: 41}},
			}
			status := "RUNNING"
			processed := 1
			if doneCalls >= 2 {
				items = append(items, asomaniac.JobItemSummary{
					Term: "photo", Storefront: "US", Status: "DONE",
					Result: &asomaniac.KeywordAnalysis{Keyword: "photo", Storefront: "US", Popularity: 33},
				})
				status = "COMPLETED"
				processed = 2
			} else {
				items = append(items, asomaniac.JobItemSummary{Term: "photo", Storefront: "US", Status: "PENDING"})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": asomaniac.JobPollResponse{
					ID: "job-1", Status: status, Storefronts: []string{"US"},
					TotalCount: 2, ProcessedCount: processed, Items: items,
				},
			})
		default:
			http.Error(w, "unexpected", http.StatusBadRequest)
		}
	}))
	defer srv.Close()
	serverURL = srv.URL

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	cfg := &asomaniac.Config{APIKey: "test-key", BaseURL: srv.URL + "/api/v1"}
	if err := asomaniac.WriteConfig(configPath, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := runKeywordsBatch(context.Background(), batchOptions{
		configPath:   configPath,
		keywords:     []string{"camera", "photo"},
		storefronts:  []string{"US"},
		pollInterval: 10 * time.Millisecond,
		timeout:      5 * time.Second,
		stdout:       &stdout,
		stderr:       &stderr,
	})
	if err != nil {
		t.Fatalf("runKeywordsBatch: %v", err)
	}

	// stdout: JSONL — one line per completed item.
	out := strings.TrimSpace(stdout.String())
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 JSONL lines, got %d:\n%s", len(lines), out)
	}
	for _, line := range lines {
		var ka asomaniac.KeywordAnalysis
		if err := json.Unmarshal([]byte(line), &ka); err != nil {
			t.Fatalf("parse JSONL line %q: %v", line, err)
		}
		if ka.Keyword != "camera" && ka.Keyword != "photo" {
			t.Errorf("unexpected keyword in stream: %q", ka.Keyword)
		}
	}

	se := stderr.String()
	if !strings.Contains(se, "submitted job job-1") {
		t.Errorf("stderr missing submission line, got: %q", se)
	}
	if !strings.Contains(se, "completed (2/2)") {
		t.Errorf("stderr missing completion line, got: %q", se)
	}
}

// TestKeywordsBatchNoWait submits the job and exits immediately when --no-wait
// is set, printing the JobSubmitResponse JSON on stdout.
func TestKeywordsBatchNoWait(t *testing.T) {
	var serverURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": asomaniac.JobSubmitResponse{JobID: "job-2", StatusURL: serverURL + "/api/v1/keywords/jobs/job-2", TotalCount: 1},
		})
	}))
	defer srv.Close()
	serverURL = srv.URL

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	cfg := &asomaniac.Config{APIKey: "test-key", BaseURL: srv.URL + "/api/v1"}
	if err := asomaniac.WriteConfig(configPath, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := runKeywordsBatch(context.Background(), batchOptions{
		configPath:  configPath,
		keywords:    []string{"camera"},
		storefronts: []string{"US"},
		noWait:      true,
		stdout:      &stdout,
		stderr:      &stderr,
	})
	if err != nil {
		t.Fatalf("runKeywordsBatch: %v", err)
	}

	var sub asomaniac.JobSubmitResponse
	if err := json.Unmarshal(stdout.Bytes(), &sub); err != nil {
		t.Fatalf("parse stdout: %v", err)
	}
	if sub.JobID != "job-2" {
		t.Errorf("JobID = %q, want job-2", sub.JobID)
	}
}
