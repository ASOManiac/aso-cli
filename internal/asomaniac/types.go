package asomaniac

// APIResponse wraps all successful API responses.
type APIResponse[T any] struct {
	Data T `json:"data"`
}

// AnalyzeMeta is the meta block returned by POST /keywords/analyze.
type AnalyzeMeta struct {
	Pending   []string `json:"pending"`
	TimedOut  bool     `json:"timedOut"`
	ElapsedMs int64    `json:"elapsedMs"`
}

// AnalyzeResponse is the full response from POST /keywords/analyze.
type AnalyzeResponse struct {
	Data []KeywordAnalysis `json:"data"`
	Meta AnalyzeMeta       `json:"meta"`
}

// BatchPerStorefront is the per-storefront meta entry in a batch-analyze response.
type BatchPerStorefront struct {
	ResolvedCount int   `json:"resolvedCount"`
	PendingCount  int   `json:"pendingCount"`
	TimedOut      bool  `json:"timedOut"`
	ElapsedMs     int64 `json:"elapsedMs"`
}

// BatchMeta is the meta block returned by POST /keywords/batch-analyze.
type BatchMeta struct {
	Pending       map[string][]string           `json:"pending"`
	TimedOut      bool                          `json:"timedOut"`
	ElapsedMs     int64                         `json:"elapsedMs"`
	PerStorefront map[string]BatchPerStorefront `json:"perStorefront"`
}

// BatchAnalyzeResponse is the full response from POST /keywords/batch-analyze.
type BatchAnalyzeResponse struct {
	Data BatchResult `json:"data"`
	Meta BatchMeta   `json:"meta"`
}

// JobSubmitResponse is returned by POST /keywords/batch-analyze (202 Accepted).
type JobSubmitResponse struct {
	JobID      string `json:"jobId"`
	StatusURL  string `json:"statusUrl"`
	TotalCount int    `json:"totalCount"`
}

// JobItemSummary is one entry in a job poll response.
type JobItemSummary struct {
	Term       string           `json:"term"`
	Storefront string           `json:"storefront"`
	Status     string           `json:"status"` // PENDING|RUNNING|DONE|FAILED|CANCELLED
	Error      *string          `json:"error,omitempty"`
	Result     *KeywordAnalysis `json:"result,omitempty"`
}

// JobPollResponse is the body returned by GET /keywords/jobs/:id.
type JobPollResponse struct {
	ID             string           `json:"id"`
	Status         string           `json:"status"` // QUEUED|RUNNING|COMPLETED|FAILED|CANCELLED
	Storefronts    []string         `json:"storefronts"`
	TotalCount     int              `json:"totalCount"`
	ProcessedCount int              `json:"processedCount"`
	CreatedAt      string           `json:"createdAt"`
	StartedAt      *string          `json:"startedAt,omitempty"`
	CompletedAt    *string          `json:"completedAt,omitempty"`
	Error          *string          `json:"error,omitempty"`
	Items          []JobItemSummary `json:"items"`
}

// RecommendMeta is the meta block returned by GET /keywords/recommendations.
type RecommendMeta struct {
	ElapsedMs int64 `json:"elapsedMs"`
	TimedOut  bool  `json:"timedOut"`
}

// RecommendResponse is the full response from GET /keywords/recommendations.
type RecommendResponse struct {
	Data []KeywordRecommendation `json:"data"`
	Meta RecommendMeta           `json:"meta"`
}

// APIError is the error shape from the ASO Maniac API.
type APIError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// KeywordAnalysis holds the result of analyzing a single keyword.
type KeywordAnalysis struct {
	Keyword         string   `json:"keyword"`
	Storefront      string   `json:"storefront"`
	Popularity      int      `json:"popularity"`
	Difficulty      int      `json:"difficulty"`
	Confidence      string   `json:"confidence"`
	TotalApps       int      `json:"totalApps"`
	TopApps         []TopApp `json:"topApps"`
	RelatedSearches []string `json:"relatedSearches"`
}

// TopApp represents a top-ranking app for a keyword.
type TopApp struct {
	AppID       string  `json:"appId"`
	Name        string  `json:"name"`
	Developer   string  `json:"developer"`
	Icon        string  `json:"icon"`
	Rating      float64 `json:"rating"`
	ReviewCount int     `json:"reviewCount"`
	Price       string  `json:"price"`
	Rank        int     `json:"rank"`
}

// KeywordRecommendation is a suggested keyword from the API.
type KeywordRecommendation struct {
	Keyword    string `json:"keyword"`
	Popularity int    `json:"popularity"`
	Difficulty int    `json:"difficulty,omitempty"`
	Source     string `json:"source"`
}

// BatchResult holds results for a batch keyword analysis request.
type BatchResult struct {
	Results          []BatchKeywordResult `json:"results"`
	TotalKeywords    int                  `json:"totalKeywords"`
	TotalStorefronts int                  `json:"totalStorefronts"`
}

// BatchKeywordResult holds analysis results for a single keyword across storefronts.
type BatchKeywordResult struct {
	Keyword     string                     `json:"keyword"`
	Storefronts map[string]KeywordAnalysis `json:"storefronts"`
}

// UsageStats describes the current API usage for the authenticated user.
type UsageStats struct {
	Plan     string `json:"plan"`
	APICalls struct {
		Today     int `json:"today"`
		ThisMonth int `json:"thisMonth"`
		Limit     int `json:"limit"`
	} `json:"apiCalls"`
	TrackedApps struct {
		Current int `json:"current"`
		Limit   int `json:"limit"`
	} `json:"trackedApps"`
	TrackedKeywords struct {
		Current int `json:"current"`
		Limit   int `json:"limit"`
	} `json:"trackedKeywords"`
}

// UserProfile is the authenticated user's profile.
type UserProfile struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      *string `json:"name"`
	Avatar    *string `json:"avatar"`
	Plan      string  `json:"plan"`
	CreatedAt string  `json:"createdAt"`
}

// Storefronts lists all 58 supported App Store storefront codes.
var Storefronts = []string{
	"US", "GB", "CA", "AU", "NZ",
	"DE", "FR", "IT", "ES", "PT", "NL", "BE", "AT", "CH", "SE", "NO", "DK", "FI", "IE",
	"PL", "CZ", "HU", "RO", "BG", "GR", "TR", "RU", "UA",
	"JP", "KR", "CN", "TW", "HK",
	"SG", "MY", "TH", "VN", "ID", "PH",
	"IN", "PK",
	"SA", "AE", "IL", "EG", "ZA", "KE",
	"BR", "MX", "AR", "CL", "CO", "PE",
	"LK", "NP", "MO", "AM",
	"HR", "SK", "LT", "LV", "EE",
}
