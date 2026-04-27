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
	"sync"
	"time"
)

// Client communicates with the ASO Maniac API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// DefaultHTTPTimeout is the default timeout for API requests.
const DefaultHTTPTimeout = 30 * time.Second

// Version is the CLI version string, set at initialization time.
// Used in the User-Agent header. Defaults to "dev" if not set.
var Version = "dev"

// NewClient creates a new API client with the given base URL and API key.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: DefaultHTTPTimeout,
		},
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

// retryMaxAttempts is the maximum number of attempts (1 initial + 2 retries).
const retryMaxAttempts = 3

// retryBaseDelay is the initial backoff delay before the first retry.
const retryBaseDelay = 1 * time.Second

// isRetryableStatus returns true for HTTP status codes that warrant a retry.
func isRetryableStatus(code int) bool {
	return code == http.StatusTooManyRequests || code == http.StatusBadGateway || code == http.StatusServiceUnavailable
}

// isRetryableError returns true for transient network/connection errors.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	// Context cancellation/deadline should not be retried.
	if ctx_err := context.Canceled; err == ctx_err {
		return false
	}
	if ctx_err := context.DeadlineExceeded; err == ctx_err {
		return false
	}
	// All other errors (connection refused, DNS, TLS, etc.) are retryable.
	return true
}

// do executes an HTTP request with retry and returns the response.
func (c *Client) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	u := c.baseURL + path
	return c.executeWithRetry(ctx, method, u, body)
}

// doAbsolute executes an HTTP request against an absolute URL with retry.
func (c *Client) doAbsolute(ctx context.Context, method, absoluteURL string, body any) (*http.Response, error) {
	return c.executeWithRetry(ctx, method, absoluteURL, body)
}

// executeWithRetry sends an HTTP request, retrying on transient failures.
// Retries up to 2 times (3 total attempts) with exponential backoff (1s, 2s)
// on 429, 502, 503 status codes or connection errors.
func (c *Client) executeWithRetry(ctx context.Context, method, url string, body any) (*http.Response, error) {
	var bodyData []byte
	if body != nil {
		var err error
		bodyData, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
	}

	var lastErr error
	for attempt := range retryMaxAttempts {
		if attempt > 0 {
			delay := retryBaseDelay * time.Duration(1<<(attempt-1)) // 1s, 2s
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		var bodyReader io.Reader
		if bodyData != nil {
			bodyReader = bytes.NewReader(bodyData)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
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
		req.Header.Set("User-Agent", "aso-cli/"+Version)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if !isRetryableError(err) {
				return nil, err
			}
			continue
		}

		if isRetryableStatus(resp.StatusCode) {
			// On the last attempt, return the response so the caller can
			// decode the API error body (e.g. RATE_LIMITED).
			if attempt == retryMaxAttempts-1 {
				return resp, nil
			}
			resp.Body.Close()
			lastErr = fmt.Errorf("http %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", retryMaxAttempts, lastErr)
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
	resp, err := c.do(ctx, http.MethodPost, "/apps/track", body)
	if err != nil {
		return nil, err
	}
	return decodeResponse[TrackedApp](resp)
}

// GetDashboard fetches the portfolio dashboard overview.
func (c *Client) GetDashboard(ctx context.Context) (*PortfolioDashboard, error) {
	resp, err := c.do(ctx, http.MethodGet, "/dashboard", nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse[PortfolioDashboard](resp)
}

// ExportRequestBody is the JSON body for the export endpoint.
type ExportRequestBody struct {
	Format  string            `json:"format"`
	Type    string            `json:"type"`
	Filters map[string]string `json:"filters,omitempty"`
}

// Export exports data in the specified format.
// dataType is one of: rankings, keywords, apps.
// filters is reserved for future use and may be nil.
func (c *Client) Export(ctx context.Context, format, dataType string, filters map[string]string) (*ExportResult, error) {
	body := ExportRequestBody{
		Format:  format,
		Type:    dataType,
		Filters: filters,
	}
	resp, err := c.do(ctx, http.MethodPost, "/export", body)
	if err != nil {
		return nil, err
	}
	return decodeResponse[ExportResult](resp)
}

// trendsConcurrency is the maximum number of parallel trend requests.
const trendsConcurrency = 5

// GetTrends fetches popularity trends for keywords in parallel.
// appID is the App Store ID; from and to are optional date strings (YYYY-MM-DD); pass "" to omit.
func (c *Client) GetTrends(ctx context.Context, keywords []string, storefront, appID, from, to string) ([]TrendResult, error) {
	results := make([]TrendResult, len(keywords))
	errs := make([]error, len(keywords))

	sem := make(chan struct{}, trendsConcurrency)
	var wg sync.WaitGroup

	for i, kw := range keywords {
		wg.Add(1)
		go func(idx int, keyword string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			path := fmt.Sprintf("/trends?keyword=%s&storefront=%s&appId=%s",
				url.QueryEscape(keyword), url.QueryEscape(storefront), url.QueryEscape(appID))
			if from != "" {
				path += "&from=" + url.QueryEscape(from)
			}
			if to != "" {
				path += "&to=" + url.QueryEscape(to)
			}
			resp, err := c.do(ctx, http.MethodGet, path, nil)
			if err != nil {
				errs[idx] = fmt.Errorf("trends %q: %w", keyword, err)
				return
			}
			r, err := decodeResponse[TrendResult](resp)
			if err != nil {
				errs[idx] = fmt.Errorf("trends %q: %w", keyword, err)
				return
			}
			results[idx] = *r
		}(i, kw)
	}
	wg.Wait()

	// Return the first error encountered (preserves keyword order).
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

// GetRankHistory fetches rank history for a tracked app's keyword.
// from and to are date strings (YYYY-MM-DD). granularity is "day", "week", or "month".
// aggregation is "avg", "min", or "max".
func (c *Client) GetRankHistory(ctx context.Context, appID, keywordID, storefront, from, to, granularity, aggregation string) (*RankHistory, error) {
	path := fmt.Sprintf("/timeseries/rankings?appId=%s&keywordId=%s&storefront=%s&startDate=%s&endDate=%s&granularity=%s&aggregation=%s",
		url.QueryEscape(appID), url.QueryEscape(keywordID), url.QueryEscape(storefront),
		url.QueryEscape(from), url.QueryEscape(to),
		url.QueryEscape(granularity), url.QueryEscape(aggregation))
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse[RankHistory](resp)
}
