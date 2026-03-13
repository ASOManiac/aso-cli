package shared

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

type appResolutionRoundTripFunc func(*http.Request) (*http.Response, error)

func (f appResolutionRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newAppResolutionTestClient(t *testing.T, transport appResolutionRoundTripFunc) *asc.Client {
	t.Helper()

	keyPath := filepath.Join(t.TempDir(), "key.p8")
	writeECDSAPEM(t, keyPath)

	httpClient := &http.Client{Transport: transport}
	client, err := asc.NewClientWithHTTPClient("KEY123", "ISS456", keyPath, httpClient)
	if err != nil {
		t.Fatalf("NewClientWithHTTPClient() error: %v", err)
	}

	return client
}

func appResolutionJSONResponse(status int, body string) (*http.Response, error) {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}

func TestResolveAppStoreVersionIDAndState_PrefersAppVersionState(t *testing.T) {
	client := newAppResolutionTestClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/apps/app-1/appStoreVersions" {
			t.Fatalf("expected /v1/apps/app-1/appStoreVersions, got %s", req.URL.Path)
		}

		query := req.URL.Query()
		if query.Get("filter[versionString]") != "1.2.3" {
			t.Fatalf("expected filter[versionString]=1.2.3, got %q", query.Get("filter[versionString]"))
		}
		if query.Get("filter[platform]") != "IOS" {
			t.Fatalf("expected filter[platform]=IOS, got %q", query.Get("filter[platform]"))
		}
		if query.Get("limit") != "10" {
			t.Fatalf("expected limit=10, got %q", query.Get("limit"))
		}

		return appResolutionJSONResponse(http.StatusOK, `{"data":[{"type":"appStoreVersions","id":"ver-123","attributes":{"appVersionState":"PREORDER_READY_FOR_SALE","appStoreState":"READY_FOR_SALE"}}]}`)
	})

	versionID, versionState, err := ResolveAppStoreVersionIDAndState(context.Background(), client, "app-1", "1.2.3", "IOS")
	if err != nil {
		t.Fatalf("ResolveAppStoreVersionIDAndState() error: %v", err)
	}
	if versionID != "ver-123" {
		t.Fatalf("expected version ID ver-123, got %q", versionID)
	}
	if versionState != "PREORDER_READY_FOR_SALE" {
		t.Fatalf("expected state PREORDER_READY_FOR_SALE, got %q", versionState)
	}
}

func TestResolveAppStoreVersionIDAndState_FallsBackToTrimmedAppStoreState(t *testing.T) {
	client := newAppResolutionTestClient(t, func(req *http.Request) (*http.Response, error) {
		return appResolutionJSONResponse(http.StatusOK, `{"data":[{"type":"appStoreVersions","id":"ver-456","attributes":{"appVersionState":"   ","appStoreState":" READY_FOR_SALE "}}]}`)
	})

	_, versionState, err := ResolveAppStoreVersionIDAndState(context.Background(), client, "app-1", "1.2.3", "IOS")
	if err != nil {
		t.Fatalf("ResolveAppStoreVersionIDAndState() error: %v", err)
	}
	if versionState != "READY_FOR_SALE" {
		t.Fatalf("expected fallback state READY_FOR_SALE, got %q", versionState)
	}
}
