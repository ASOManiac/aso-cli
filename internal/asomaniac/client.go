package asomaniac

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client communicates with the ASO Maniac API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new API client with the given base URL and API key.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: http.DefaultClient,
	}
}

// NewClientFromConfig creates a new API client from a Config.
func NewClientFromConfig(cfg *Config) *Client {
	base := cfg.BaseURL
	if base == "" {
		base = DefaultBaseURL
	}
	return NewClient(base, cfg.APIKey)
}

// do executes an HTTP request and returns the response.
func (c *Client) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	u := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "aso-cli")

	return c.httpClient.Do(req)
}

// doAbsolute executes an HTTP request against an absolute URL (not relative to baseURL).
func (c *Client) doAbsolute(ctx context.Context, method, absoluteURL string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, absoluteURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "aso-cli")

	return c.httpClient.Do(req)
}

// decodeResponse reads the HTTP response, handles errors, and unmarshals the data.
func decodeResponse[T any](resp *http.Response) (*T, error) {
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if json.Unmarshal(data, &apiErr) == nil && apiErr.Error.Code != "" {
			return nil, fmt.Errorf("api error %s: %s", apiErr.Error.Code, apiErr.Error.Message)
		}
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(data))
	}

	var wrapped APIResponse[T]
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &wrapped.Data, nil
}

// profileBaseURL derives the auth base URL from the v1 base URL.
// e.g. "https://asomaniac.com/api/v1" -> "https://asomaniac.com"
func (c *Client) profileBaseURL() string {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return c.baseURL
	}
	u.Path = ""
	return u.String()
}

// GetProfile fetches the authenticated user's profile.
// The profile endpoint lives at /api/auth/me (not under /api/v1/).
func (c *Client) GetProfile(ctx context.Context) (*UserProfile, error) {
	profileURL := c.profileBaseURL() + "/api/auth/me"
	resp, err := c.doAbsolute(ctx, http.MethodGet, profileURL, nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse[UserProfile](resp)
}

// GetUsage fetches the current usage stats for the authenticated user.
func (c *Client) GetUsage(ctx context.Context) (*UsageStats, error) {
	resp, err := c.do(ctx, http.MethodGet, "/usage", nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse[UsageStats](resp)
}

// AnalyzeKeywordRequest is the request body for the analyze endpoint.
type AnalyzeKeywordRequest struct {
	Keywords   []string `json:"keywords"`
	Storefront string   `json:"storefront"`
	Fields     []string `json:"fields,omitempty"`
}

// AnalyzeKeywords analyzes one or more keywords in a given storefront.
// The API accepts a batch of keywords in a single POST request.
func (c *Client) AnalyzeKeywords(ctx context.Context, keywords []string, storefront string, fields []string) ([]KeywordAnalysis, error) {
	body := AnalyzeKeywordRequest{
		Keywords:   keywords,
		Storefront: storefront,
	}
	if len(fields) > 0 {
		body.Fields = fields
	}
	resp, err := c.do(ctx, http.MethodPost, "/keywords/analyze", body)
	if err != nil {
		return nil, err
	}
	result, err := decodeResponse[[]KeywordAnalysis](resp)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

// GetRecommendations fetches keyword recommendations for a seed keyword.
func (c *Client) GetRecommendations(ctx context.Context, seed, storefront string, limit int) (*[]KeywordRecommendation, error) {
	path := fmt.Sprintf("/keywords/recommendations?keyword=%s&storefront=%s&limit=%d",
		url.QueryEscape(seed), url.QueryEscape(storefront), limit)
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse[[]KeywordRecommendation](resp)
}

// BatchAnalyzeRequest is the request body for batch keyword analysis.
type BatchAnalyzeRequest struct {
	Keywords    []string `json:"keywords"`
	Storefronts []string `json:"storefronts"`
}

// BatchAnalyze analyzes multiple keywords across multiple storefronts.
func (c *Client) BatchAnalyze(ctx context.Context, keywords, storefronts []string) (*BatchResult, error) {
	body := BatchAnalyzeRequest{
		Keywords:    keywords,
		Storefronts: storefronts,
	}
	resp, err := c.do(ctx, http.MethodPost, "/keywords/batch-analyze", body)
	if err != nil {
		return nil, err
	}
	return decodeResponse[BatchResult](resp)
}

// GetCompetitors finds competitor apps for the given app ID and storefront.
func (c *Client) GetCompetitors(ctx context.Context, appID, storefront string) (*[]CompetitorAnalysis, error) {
	path := fmt.Sprintf("/competitors/%s?storefront=%s",
		url.PathEscape(appID), url.QueryEscape(storefront))
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse[[]CompetitorAnalysis](resp)
}

// TrackAppRequest is the request body for tracking an app.
type TrackAppRequest struct {
	AppID      string   `json:"appId"`
	Storefront string   `json:"storefront"`
	Keywords   []string `json:"keywords,omitempty"`
}

// TrackApp adds an app to the user's tracked portfolio.
func (c *Client) TrackApp(ctx context.Context, appID, storefront string, keywords []string) (*TrackedApp, error) {
	body := TrackAppRequest{
		AppID:      appID,
		Storefront: storefront,
		Keywords:   keywords,
	}
	resp, err := c.do(ctx, http.MethodPost, "/portfolio/track", body)
	if err != nil {
		return nil, err
	}
	return decodeResponse[TrackedApp](resp)
}

// GetDashboard fetches the portfolio dashboard overview.
func (c *Client) GetDashboard(ctx context.Context) (*PortfolioDashboard, error) {
	resp, err := c.do(ctx, http.MethodGet, "/portfolio/dashboard", nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse[PortfolioDashboard](resp)
}

// Export exports data in the specified format.
// dataType is one of: rankings, keywords, apps.
// filters is reserved for future use and may be nil.
func (c *Client) Export(ctx context.Context, format, dataType string, filters map[string]string) (*ExportResult, error) {
	path := fmt.Sprintf("/export?format=%s&type=%s",
		url.QueryEscape(format), url.QueryEscape(dataType))
	for k, v := range filters {
		path += fmt.Sprintf("&%s=%s", url.QueryEscape(k), url.QueryEscape(v))
	}
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse[ExportResult](resp)
}

// GetTrends fetches popularity trends for keywords.
// from and to are optional date strings (YYYY-MM-DD); pass "" to omit.
func (c *Client) GetTrends(ctx context.Context, keywords []string, storefront, from, to string) ([]TrendResult, error) {
	results := make([]TrendResult, 0, len(keywords))
	for _, kw := range keywords {
		path := fmt.Sprintf("/keywords/trends?keyword=%s&storefront=%s",
			url.QueryEscape(kw), url.QueryEscape(storefront))
		if from != "" {
			path += "&from=" + url.QueryEscape(from)
		}
		if to != "" {
			path += "&to=" + url.QueryEscape(to)
		}
		resp, err := c.do(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, fmt.Errorf("trends %q: %w", kw, err)
		}
		r, err := decodeResponse[TrendResult](resp)
		if err != nil {
			return nil, fmt.Errorf("trends %q: %w", kw, err)
		}
		results = append(results, *r)
	}
	return results, nil
}

// GetRankHistory fetches rank history for a tracked app's keyword.
// from and to are optional date strings (YYYY-MM-DD); pass "" to omit.
func (c *Client) GetRankHistory(ctx context.Context, appID, keyword, storefront, from, to string) (*RankHistory, error) {
	path := fmt.Sprintf("/portfolio/rank-history?appId=%s&keyword=%s&storefront=%s",
		url.QueryEscape(appID), url.QueryEscape(keyword), url.QueryEscape(storefront))
	if from != "" {
		path += "&from=" + url.QueryEscape(from)
	}
	if to != "" {
		path += "&to=" + url.QueryEscape(to)
	}
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse[RankHistory](resp)
}
