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

// decodeFullResponse reads the HTTP response, handles errors, and unmarshals
// the entire payload (data + meta) into T. Use this for endpoints that return
// a meaningful meta block alongside data.
func decodeFullResponse[T any](resp *http.Response) (*T, error) {
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

	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &out, nil
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
// The API accepts a batch of keywords in a single POST request. The returned
// AnalyzeResponse includes a Meta block describing pending keywords and timeouts.
func (c *Client) AnalyzeKeywords(ctx context.Context, keywords []string, storefront string, fields []string) (*AnalyzeResponse, error) {
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
	return decodeFullResponse[AnalyzeResponse](resp)
}

// GetRecommendations fetches keyword recommendations for a seed keyword.
// The returned RecommendResponse includes a Meta block describing the request
// elapsed time and timeout state.
func (c *Client) GetRecommendations(ctx context.Context, seed, storefront string, limit int) (*RecommendResponse, error) {
	path := fmt.Sprintf("/keywords/recommendations?keyword=%s&storefront=%s&limit=%d",
		url.QueryEscape(seed), url.QueryEscape(storefront), limit)
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return decodeFullResponse[RecommendResponse](resp)
}

// BatchAnalyzeRequest is the request body for batch keyword analysis.
type BatchAnalyzeRequest struct {
	Keywords    []string `json:"keywords"`
	Storefronts []string `json:"storefronts"`
}

// SubmitBatchAnalyze submits an async batch keyword analysis job. The server
// returns 202 Accepted with a JobSubmitResponse; the caller must then poll
// GetJob until the job reaches a terminal state.
func (c *Client) SubmitBatchAnalyze(ctx context.Context, keywords, storefronts []string) (*JobSubmitResponse, error) {
	body := BatchAnalyzeRequest{
		Keywords:    keywords,
		Storefronts: storefronts,
	}
	resp, err := c.do(ctx, http.MethodPost, "/keywords/batch-analyze", body)
	if err != nil {
		return nil, err
	}
	return decodeResponse[JobSubmitResponse](resp)
}

// GetJob fetches the current state of an async keyword analysis job.
func (c *Client) GetJob(ctx context.Context, jobID string) (*JobPollResponse, error) {
	path := "/keywords/jobs/" + url.PathEscape(jobID)
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse[JobPollResponse](resp)
}

// CancelJob marks a job CANCELLED. Idempotent.
func (c *Client) CancelJob(ctx context.Context, jobID string) error {
	path := "/keywords/jobs/" + url.PathEscape(jobID)
	resp, err := c.do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cancel job: http %d: %s", resp.StatusCode, string(data))
	}
	return nil
}
