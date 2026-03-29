package asomaniac

// APIResponse wraps all successful API responses.
type APIResponse[T any] struct {
	Data T `json:"data"`
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

// CompetitorAnalysis holds the result of comparing two apps.
type CompetitorAnalysis struct {
	App                AppMetadata      `json:"app"`
	Competitor         AppMetadata      `json:"competitor"`
	SharedKeywords     int              `json:"sharedKeywords"`
	UniqueToApp        int              `json:"uniqueToApp"`
	UniqueToCompetitor int              `json:"uniqueToCompetitor"`
	KeywordOverlap     []KeywordOverlap `json:"keywordOverlap"`
}

// AppMetadata describes an app in the store.
type AppMetadata struct {
	AppID       string   `json:"appId"`
	BundleID    string   `json:"bundleId"`
	Name        string   `json:"name"`
	Developer   string   `json:"developer"`
	Category    string   `json:"category"`
	Price       string   `json:"price"`
	Rating      float64  `json:"rating"`
	ReviewCount int      `json:"reviewCount"`
	Icon        string   `json:"icon"`
	Screenshots []string `json:"screenshots"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	LastUpdated string   `json:"lastUpdated"`
}

// KeywordOverlap describes a keyword shared between two apps.
type KeywordOverlap struct {
	Keyword        string `json:"keyword"`
	Storefront     string `json:"storefront"`
	AppRank        int    `json:"appRank"`
	CompetitorRank int    `json:"competitorRank"`
	Popularity     int    `json:"popularity"`
}

// PortfolioDashboard is the overview for all tracked apps.
type PortfolioDashboard struct {
	TotalApps        int                  `json:"totalApps"`
	TotalKeywords    int                  `json:"totalKeywords"`
	AverageRank      *float64             `json:"averageRank"`
	RankImprovements int                  `json:"rankImprovements"`
	RankDeclines     int                  `json:"rankDeclines"`
	Alerts           []DashboardAlert     `json:"alerts"`
	TopPerformers    []DashboardPerformer `json:"topPerformers"`
	RecentChanges    []DashboardChange    `json:"recentChanges"`
}

// DashboardPerformer is a top-performing app on the dashboard.
type DashboardPerformer struct {
	AppName  string  `json:"appName"`
	AppID    string  `json:"appId"`
	AvgRank  float64 `json:"avgRank"`
	Keywords int     `json:"keywords"`
}

// DashboardChange is a recent rank change on the dashboard.
type DashboardChange struct {
	AppName   string `json:"appName"`
	Keyword   string `json:"keyword"`
	Change    int    `json:"change"`
	Direction string `json:"direction"`
}

// DashboardAlert is a notification on the dashboard.
type DashboardAlert struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	AppID     string `json:"appId"`
	Timestamp string `json:"timestamp"`
	Severity  string `json:"severity"`
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

// ExportResult holds the result of an export request.
type ExportResult struct {
	ExportID    string `json:"exportId"`
	Format      string `json:"format"`
	Status      string `json:"status"`
	Data        string `json:"data,omitempty"`
	RecordCount int    `json:"recordCount"`
	GeneratedAt string `json:"generatedAt"`
}

// TrendResult holds popularity trend data for a keyword.
type TrendResult struct {
	Keyword    string           `json:"keyword"`
	Storefront string           `json:"storefront"`
	DataPoints []TrendDataPoint `json:"dataPoints"`
}

// TrendDataPoint is a single data point in a trend series.
type TrendDataPoint struct {
	Date       string `json:"date"`
	Popularity int    `json:"popularity"`
}

// TrackedApp is an app being tracked in the user's portfolio.
type TrackedApp struct {
	AppID           string   `json:"appId"`
	Name            string   `json:"name"`
	Storefront      string   `json:"storefront"`
	TrackedKeywords []string `json:"trackedKeywords"`
	AddedAt         string   `json:"addedAt"`
}

// RankHistory holds historical rank data for a keyword.
type RankHistory struct {
	Keyword    string          `json:"keyword"`
	Storefront string          `json:"storefront"`
	DataPoints []RankDataPoint `json:"dataPoints"`
}

// RankDataPoint is a single data point in a rank history series.
type RankDataPoint struct {
	Date string `json:"date"`
	Rank int    `json:"rank"`
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
