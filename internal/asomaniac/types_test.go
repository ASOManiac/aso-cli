package asomaniac

import (
	"encoding/json"
	"testing"
)

func TestKeywordAnalysisUnmarshal(t *testing.T) {
	raw := `{
		"keyword": "photo editor",
		"storefront": "US",
		"popularity": 72,
		"difficulty": 85,
		"confidence": "high",
		"totalApps": 1500,
		"topApps": [
			{
				"appId": "123456789",
				"name": "SnapEdit",
				"developer": "SnapDev Inc.",
				"icon": "https://example.com/icon.png",
				"rating": 4.7,
				"reviewCount": 12000,
				"price": "Free",
				"rank": 1
			}
		],
		"relatedSearches": ["photo filter", "image editor"]
	}`

	var ka KeywordAnalysis
	if err := json.Unmarshal([]byte(raw), &ka); err != nil {
		t.Fatalf("unmarshal KeywordAnalysis: %v", err)
	}

	if ka.Keyword != "photo editor" {
		t.Errorf("Keyword = %q, want %q", ka.Keyword, "photo editor")
	}
	if ka.Storefront != "US" {
		t.Errorf("Storefront = %q, want %q", ka.Storefront, "US")
	}
	if ka.Popularity != 72 {
		t.Errorf("Popularity = %d, want %d", ka.Popularity, 72)
	}
	if ka.Difficulty != 85 {
		t.Errorf("Difficulty = %d, want %d", ka.Difficulty, 85)
	}
	if ka.Confidence != "high" {
		t.Errorf("Confidence = %q, want %q", ka.Confidence, "high")
	}
	if ka.TotalApps != 1500 {
		t.Errorf("TotalApps = %d, want %d", ka.TotalApps, 1500)
	}
	if len(ka.TopApps) != 1 {
		t.Fatalf("TopApps len = %d, want 1", len(ka.TopApps))
	}
	app := ka.TopApps[0]
	if app.AppID != "123456789" {
		t.Errorf("TopApps[0].AppID = %q, want %q", app.AppID, "123456789")
	}
	if app.Rating != 4.7 {
		t.Errorf("TopApps[0].Rating = %f, want %f", app.Rating, 4.7)
	}
	if app.Rank != 1 {
		t.Errorf("TopApps[0].Rank = %d, want %d", app.Rank, 1)
	}
	if len(ka.RelatedSearches) != 2 {
		t.Fatalf("RelatedSearches len = %d, want 2", len(ka.RelatedSearches))
	}
	if ka.RelatedSearches[0] != "photo filter" {
		t.Errorf("RelatedSearches[0] = %q, want %q", ka.RelatedSearches[0], "photo filter")
	}
}

func TestKeywordAnalysisUnmarshalMinimal(t *testing.T) {
	raw := `{"keyword":"vpn","storefront":"GB","popularity":50,"difficulty":30,"confidence":"low","totalApps":200}`

	var ka KeywordAnalysis
	if err := json.Unmarshal([]byte(raw), &ka); err != nil {
		t.Fatalf("unmarshal minimal KeywordAnalysis: %v", err)
	}

	if ka.Keyword != "vpn" {
		t.Errorf("Keyword = %q, want %q", ka.Keyword, "vpn")
	}
	if ka.TopApps != nil {
		t.Errorf("TopApps = %v, want nil", ka.TopApps)
	}
	if ka.RelatedSearches != nil {
		t.Errorf("RelatedSearches = %v, want nil", ka.RelatedSearches)
	}
}

func TestAPIErrorUnmarshal(t *testing.T) {
	raw := `{
		"error": {
			"code": "UNAUTHORIZED",
			"message": "Invalid API key"
		}
	}`

	var apiErr APIError
	if err := json.Unmarshal([]byte(raw), &apiErr); err != nil {
		t.Fatalf("unmarshal APIError: %v", err)
	}

	if apiErr.Error.Code != "UNAUTHORIZED" {
		t.Errorf("Error.Code = %q, want %q", apiErr.Error.Code, "UNAUTHORIZED")
	}
	if apiErr.Error.Message != "Invalid API key" {
		t.Errorf("Error.Message = %q, want %q", apiErr.Error.Message, "Invalid API key")
	}
}

func TestAPIErrorUnmarshalRateLimit(t *testing.T) {
	raw := `{
		"error": {
			"code": "RATE_LIMITED",
			"message": "Too many requests. Please try again in 60 seconds."
		}
	}`

	var apiErr APIError
	if err := json.Unmarshal([]byte(raw), &apiErr); err != nil {
		t.Fatalf("unmarshal APIError: %v", err)
	}

	if apiErr.Error.Code != "RATE_LIMITED" {
		t.Errorf("Error.Code = %q, want %q", apiErr.Error.Code, "RATE_LIMITED")
	}
}

func TestAPIResponseUnmarshal(t *testing.T) {
	raw := `{
		"data": {
			"keyword": "game",
			"storefront": "JP",
			"popularity": 90,
			"difficulty": 95,
			"confidence": "high",
			"totalApps": 5000,
			"topApps": [],
			"relatedSearches": []
		}
	}`

	var resp APIResponse[KeywordAnalysis]
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal APIResponse[KeywordAnalysis]: %v", err)
	}

	if resp.Data.Keyword != "game" {
		t.Errorf("Data.Keyword = %q, want %q", resp.Data.Keyword, "game")
	}
	if resp.Data.Popularity != 90 {
		t.Errorf("Data.Popularity = %d, want %d", resp.Data.Popularity, 90)
	}
}

func TestStorefrontsCount(t *testing.T) {
	if len(Storefronts) < 58 {
		t.Errorf("Storefronts has %d entries, want at least 58", len(Storefronts))
	}

	// Check a few known storefronts are present.
	want := map[string]bool{"US": false, "GB": false, "JP": false, "BG": false}
	for _, sf := range Storefronts {
		if _, ok := want[sf]; ok {
			want[sf] = true
		}
	}
	for code, found := range want {
		if !found {
			t.Errorf("expected storefront %q not found in Storefronts", code)
		}
	}
}
