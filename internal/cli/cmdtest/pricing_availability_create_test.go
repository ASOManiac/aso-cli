package cmdtest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestPricingAvailabilityCreateSuccess(t *testing.T) {
	setupAuth(t)

	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	var seenPayload struct {
		Data struct {
			Attributes struct {
				AvailableInNewTerritories bool `json:"availableInNewTerritories"`
			} `json:"attributes"`
			Relationships struct {
				App struct {
					Data struct {
						ID string `json:"id"`
					} `json:"data"`
				} `json:"app"`
				TerritoryAvailabilities struct {
					Data []struct {
						ID string `json:"id"`
					} `json:"data"`
				} `json:"territoryAvailabilities"`
			} `json:"relationships"`
		} `json:"data"`
		Included []struct {
			Attributes struct {
				Available bool `json:"available"`
			} `json:"attributes"`
			Relationships struct {
				Territory struct {
					Data struct {
						ID string `json:"id"`
					} `json:"data"`
				} `json:"territory"`
			} `json:"relationships"`
		} `json:"included"`
	}

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.Path != "/v2/appAvailabilities" {
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.String())
		}
		if err := json.NewDecoder(req.Body).Decode(&seenPayload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}

		body := `{"data":{"type":"appAvailabilities","id":"availability-1","attributes":{"availableInNewTerritories":true}},"links":{"self":"https://api.appstoreconnect.apple.com/v2/appAvailabilities/availability-1"}}`
		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})

	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{
			"pricing", "availability", "create",
			"--app", "app-1",
			"--territory", "usa,gbr",
			"--available", "true",
			"--available-in-new-territories", "true",
		}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := root.Run(context.Background()); err != nil {
			t.Fatalf("run error: %v", err)
		}
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, `"id":"availability-1"`) {
		t.Fatalf("expected availability output, got %q", stdout)
	}
	if seenPayload.Data.Relationships.App.Data.ID != "app-1" {
		t.Fatalf("expected app id app-1, got %q", seenPayload.Data.Relationships.App.Data.ID)
	}
	if !seenPayload.Data.Attributes.AvailableInNewTerritories {
		t.Fatal("expected availableInNewTerritories=true")
	}
	if len(seenPayload.Data.Relationships.TerritoryAvailabilities.Data) != 2 {
		t.Fatalf("expected 2 territory availability relationships, got %d", len(seenPayload.Data.Relationships.TerritoryAvailabilities.Data))
	}
	if len(seenPayload.Included) != 2 {
		t.Fatalf("expected 2 included territory availability resources, got %d", len(seenPayload.Included))
	}
	if seenPayload.Included[0].Relationships.Territory.Data.ID != "USA" {
		t.Fatalf("expected first territory ID USA, got %q", seenPayload.Included[0].Relationships.Territory.Data.ID)
	}
	if !seenPayload.Included[0].Attributes.Available || !seenPayload.Included[1].Attributes.Available {
		t.Fatal("expected included territory availability resources to set available=true")
	}
}
