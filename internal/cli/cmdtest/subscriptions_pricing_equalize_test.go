package cmdtest

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSubscriptionsPricingEqualizeValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing subscription id",
			args:    []string{"subscriptions", "pricing", "equalize", "--base-price", "3.49"},
			wantErr: "Error: --subscription-id is required",
		},
		{
			name:    "missing base price",
			args:    []string{"subscriptions", "pricing", "equalize", "--subscription-id", "sub-1"},
			wantErr: "Error: --base-price is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := RootCommand("1.2.3")
			root.FlagSet.SetOutput(io.Discard)

			stdout, stderr := captureOutput(t, func() {
				if err := root.Parse(test.args); err != nil {
					t.Fatalf("parse error: %v", err)
				}
				err := root.Run(context.Background())
				if !errors.Is(err, flag.ErrHelp) {
					t.Fatalf("expected ErrHelp, got %v", err)
				}
			})

			if stdout != "" {
				t.Fatalf("expected empty stdout, got %q", stdout)
			}
			if !strings.Contains(stderr, test.wantErr) {
				t.Fatalf("expected error %q, got %q", test.wantErr, stderr)
			}
		})
	}
}

func TestSubscriptionsPricingEqualizeDryRunIncludesBaseTerritory(t *testing.T) {
	setupAuth(t)

	basePricePointID := encodedPricePointID(t, "sub-1", "USA", "10036")
	canadaPricePointID := encodedPricePointID(t, "sub-1", "CAN", "10049")

	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Path == "/v1/subscriptions/sub-1/pricePoints" && req.Method == http.MethodGet:
			query := req.URL.Query()
			if query.Get("filter[territory]") != "USA" {
				t.Fatalf("expected filter[territory]=USA, got %q", query.Get("filter[territory]"))
			}
			if query.Get("limit") != "200" {
				t.Fatalf("expected limit=200, got %q", query.Get("limit"))
			}
			body := `{"data":[{"type":"subscriptionPricePoints","id":"` + basePricePointID + `","attributes":{"customerPrice":"3.50"}}],"links":{}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil

		case req.URL.Path == "/v1/subscriptionPricePoints/"+basePricePointID+"/equalizations" && req.Method == http.MethodGet:
			query := req.URL.Query()
			if query.Get("limit") != "200" {
				t.Fatalf("expected limit=200, got %q", query.Get("limit"))
			}
			body := `{"data":[{"type":"subscriptionPricePoints","id":"` + canadaPricePointID + `","attributes":{"customerPrice":"4.99"}}],"links":{}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil

		default:
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
			return nil, nil
		}
	})

	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"subscriptions", "pricing", "equalize", "--subscription-id", "sub-1", "--base-price", "3.5", "--dry-run"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := root.Run(context.Background()); err != nil {
			t.Fatalf("run error: %v", err)
		}
	})

	if !strings.Contains(stderr, "Got 2 territory equalizations") {
		t.Fatalf("expected dry-run progress in stderr, got %q", stderr)
	}

	var result struct {
		Total       int `json:"total"`
		Territories []struct {
			Territory string `json:"territory"`
			Price     string `json:"price"`
		} `json:"territories"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("expected total 2, got %+v", result)
	}
	if len(result.Territories) != 2 {
		t.Fatalf("expected 2 territories, got %+v", result)
	}
	if result.Territories[0].Territory != "USA" || result.Territories[0].Price != "3.5" {
		t.Fatalf("expected USA base territory first, got %+v", result.Territories)
	}
	if result.Territories[1].Territory != "CAN" {
		t.Fatalf("expected CAN equalization second, got %+v", result.Territories)
	}
}

func encodedPricePointID(t *testing.T, subscriptionID, territory, priceTier string) string {
	t.Helper()

	payload := `{"s":"` + subscriptionID + `","t":"` + territory + `","p":"` + priceTier + `"}`
	return base64.RawStdEncoding.EncodeToString([]byte(payload))
}
