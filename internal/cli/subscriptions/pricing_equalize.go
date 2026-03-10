package subscriptions

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
)

const defaultEqualizeWorkers = 8

// SubscriptionsPricingEqualizeCommand returns the equalize subcommand.
func SubscriptionsPricingEqualizeCommand() *ffcli.Command {
	fs := flag.NewFlagSet("equalize", flag.ExitOnError)

	subscriptionID := fs.String("subscription-id", "", "Subscription ID (required)")
	baseTerritory := fs.String("base-territory", "USA", "Territory to use as the pricing base")
	basePrice := fs.String("base-price", "", "Customer price in the base territory (required)")
	dryRun := fs.Bool("dry-run", false, "Show equalized prices without applying them")
	workers := fs.Int("workers", defaultEqualizeWorkers, "Number of concurrent API requests")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "equalize",
		ShortUsage: "asc subscriptions pricing equalize [flags]",
		ShortHelp:  "Set equalized prices for all territories from a base price.",
		LongHelp: `Set equalized prices for all territories from a base price.

Finds the price point matching the given base territory and price, fetches
Apple's equalized prices for all other territories, and sets them in one
operation. This replaces the manual process of exporting equalizations and
importing a CSV.

Examples:
  asc subscriptions pricing equalize --subscription-id "SUB_ID" --base-price "3.49"
  asc subscriptions pricing equalize --subscription-id "SUB_ID" --base-price "38.49" --base-territory "USA"
  asc subscriptions pricing equalize --subscription-id "SUB_ID" --base-price "3.49" --dry-run
  asc subscriptions pricing equalize --subscription-id "SUB_ID" --base-price "3.49" --workers 16`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			subID := strings.TrimSpace(*subscriptionID)
			if subID == "" {
				fmt.Fprintln(os.Stderr, "Error: --subscription-id is required")
				return flag.ErrHelp
			}
			price := strings.TrimSpace(*basePrice)
			if price == "" {
				fmt.Fprintln(os.Stderr, "Error: --base-price is required")
				return flag.ErrHelp
			}
			territory := strings.ToUpper(strings.TrimSpace(*baseTerritory))
			if territory == "" {
				territory = "USA"
			}
			numWorkers := *workers
			if numWorkers < 1 {
				numWorkers = 1
			}
			if numWorkers > 32 {
				numWorkers = 32
			}

			client, err := shared.GetASCClient()
			if err != nil {
				return fmt.Errorf("equalize: %w", err)
			}

			// Step 1: Find the base price point
			fmt.Fprintf(os.Stderr, "Finding %s price point for %s...\n", territory, price)
			pricePointID, err := findPricePoint(ctx, client, subID, territory, price)
			if err != nil {
				return fmt.Errorf("equalize: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Found price point: %s\n", pricePointID)

			// Step 2: Get equalizations for all territories
			fmt.Fprintf(os.Stderr, "Fetching equalized prices for all territories...\n")
			equalizations, err := fetchEqualizations(ctx, client, pricePointID, territory)
			if err != nil {
				return fmt.Errorf("equalize: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Got %d territory equalizations\n", len(equalizations))

			if *dryRun {
				return printEqualizeResult(&equalizeResult{
					SubscriptionID: subID,
					BaseTerritory:  territory,
					BasePrice:      price,
					DryRun:         true,
					Territories:    equalizations,
					Total:          len(equalizations),
				}, *output.Output, *output.Pretty)
			}

			// Step 3: Set prices for all territories concurrently
			fmt.Fprintf(os.Stderr, "Setting prices for %d territories (%d workers)...\n", len(equalizations), numWorkers)
			var succeeded atomic.Int32
			var failed atomic.Int32
			failures := make([]equalizeFailure, 0)
			var mu sync.Mutex

			sem := make(chan struct{}, numWorkers)
			var wg sync.WaitGroup

			for _, eq := range equalizations {
				wg.Add(1)
				go func(e equalization) {
					defer wg.Done()
					sem <- struct{}{}
					defer func() { <-sem }()

					setCtx, setCancel := shared.ContextWithTimeout(ctx)
					defer setCancel()

					_, err := client.CreateSubscriptionPrice(setCtx, subID, e.PricePointID, e.Territory, asc.SubscriptionPriceCreateAttributes{})
					if err != nil {
						failed.Add(1)
						mu.Lock()
						failures = append(failures, equalizeFailure{
							Territory: e.Territory,
							Price:     e.Price,
							Error:     err.Error(),
						})
						mu.Unlock()
						return
					}
					succeeded.Add(1)
				}(eq)
			}

			wg.Wait()

			result := &equalizeResult{
				SubscriptionID: subID,
				BaseTerritory:  territory,
				BasePrice:      price,
				DryRun:         false,
				Total:          len(equalizations),
				Succeeded:      int(succeeded.Load()),
				Failed:         int(failed.Load()),
				Failures:       failures,
			}

			fmt.Fprintf(os.Stderr, "Done: %d succeeded, %d failed\n", result.Succeeded, result.Failed)

			if err := printEqualizeResult(result, *output.Output, *output.Pretty); err != nil {
				return err
			}

			if result.Failed > 0 {
				return fmt.Errorf("equalize: %d of %d territory price updates failed", result.Failed, result.Total)
			}

			return nil
		},
	}
}

type equalization struct {
	Territory    string `json:"territory"`
	Price        string `json:"price"`
	PricePointID string `json:"pricePointId"`
}

type equalizeFailure struct {
	Territory string `json:"territory"`
	Price     string `json:"price"`
	Error     string `json:"error"`
}

type equalizeResult struct {
	SubscriptionID string            `json:"subscriptionId"`
	BaseTerritory  string            `json:"baseTerritory"`
	BasePrice      string            `json:"basePrice"`
	DryRun         bool              `json:"dryRun"`
	Total          int               `json:"total"`
	Succeeded      int               `json:"succeeded,omitempty"`
	Failed         int               `json:"failed,omitempty"`
	Territories    []equalization    `json:"territories,omitempty"`
	Failures       []equalizeFailure `json:"failures,omitempty"`
}

// pricePointIDPayload is the JSON structure encoded in base64 price point IDs.
type pricePointIDPayload struct {
	S string `json:"s"` // subscription ID
	T string `json:"t"` // territory
	P string `json:"p"` // price tier
}

func findPricePoint(ctx context.Context, client *asc.Client, subID, territory, targetPrice string) (string, error) {
	// Parse target price as a float for numeric comparison (e.g., "3.5" matches "3.50").
	targetFloat, targetErr := strconv.ParseFloat(targetPrice, 64)

	// List price points filtered by the base territory, paginating to find the matching price
	var pricePointID string

	firstCtx, firstCancel := shared.ContextWithTimeout(ctx)
	firstPage, err := client.GetSubscriptionPricePoints(firstCtx, subID,
		asc.WithSubscriptionPricePointsTerritory(territory),
		asc.WithSubscriptionPricePointsLimit(200),
	)
	firstCancel()
	if err != nil {
		return "", fmt.Errorf("failed to fetch price points: %w", err)
	}

	// Check first page
	for _, pp := range firstPage.Data {
		if pricesMatch(pp.Attributes.CustomerPrice, targetPrice, targetFloat, targetErr) {
			pricePointID = pp.ID
			return pricePointID, nil
		}
	}

	// Paginate through remaining pages
	err = asc.PaginateEach(ctx, firstPage,
		func(_ context.Context, nextURL string) (asc.PaginatedResponse, error) {
			pageCtx, pageCancel := shared.ContextWithTimeout(ctx)
			defer pageCancel()
			return client.GetSubscriptionPricePoints(pageCtx, subID,
				asc.WithSubscriptionPricePointsNextURL(nextURL),
			)
		},
		func(page asc.PaginatedResponse) error {
			typed, ok := page.(*asc.SubscriptionPricePointsResponse)
			if !ok {
				return nil
			}
			for _, pp := range typed.Data {
				if pricesMatch(pp.Attributes.CustomerPrice, targetPrice, targetFloat, targetErr) {
					pricePointID = pp.ID
					return fmt.Errorf("found") // break pagination
				}
			}
			return nil
		},
	)

	if pricePointID != "" {
		return pricePointID, nil
	}

	if err != nil && err.Error() != "found" {
		return "", err
	}

	return "", fmt.Errorf("no price point found for %s %s", territory, targetPrice)
}

func fetchEqualizations(ctx context.Context, client *asc.Client, pricePointID, baseTerritory string) ([]equalization, error) {
	firstCtx, firstCancel := shared.ContextWithTimeout(ctx)
	resp, err := client.GetSubscriptionPricePointEqualizations(firstCtx, pricePointID,
		asc.WithSubscriptionPricePointsLimit(200),
	)
	firstCancel()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch equalizations: %w", err)
	}

	allPages, err := asc.PaginateAll(ctx, resp, func(_ context.Context, nextURL string) (asc.PaginatedResponse, error) {
		pageCtx, pageCancel := shared.ContextWithTimeout(ctx)
		defer pageCancel()
		return client.GetSubscriptionPricePointEqualizations(pageCtx, pricePointID,
			asc.WithSubscriptionPricePointsNextURL(nextURL),
		)
	})
	if err != nil {
		return nil, fmt.Errorf("paginate equalizations: %w", err)
	}

	typed, ok := allPages.(*asc.SubscriptionPricePointsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type %T", allPages)
	}

	var result []equalization
	for _, pp := range typed.Data {
		territory := decodePricePointTerritory(pp.ID)
		if territory == "" || strings.EqualFold(territory, baseTerritory) {
			continue
		}
		result = append(result, equalization{
			Territory:    territory,
			Price:        strings.TrimSpace(pp.Attributes.CustomerPrice),
			PricePointID: pp.ID,
		})
	}

	return result, nil
}

func decodePricePointTerritory(id string) string {
	// Price point IDs are base64-encoded JSON: {"s":"...","t":"USA","p":"..."}
	// Try with and without padding
	for _, candidate := range []string{id, id + "=", id + "=="} {
		decoded, err := base64.StdEncoding.DecodeString(candidate)
		if err != nil {
			continue
		}
		var payload pricePointIDPayload
		if err := json.Unmarshal(decoded, &payload); err != nil {
			continue
		}
		if payload.T != "" {
			return payload.T
		}
	}
	return ""
}

func printEqualizeResult(result *equalizeResult, format string, pretty bool) error {
	return shared.PrintOutputWithRenderers(
		result,
		format,
		pretty,
		func() error { return printEqualizeTable(result) },
		func() error { return printEqualizeMarkdown(result) },
	)
}

func printEqualizeTable(result *equalizeResult) error {
	if result.DryRun {
		headers := []string{"Territory", "Price", "Price Point ID"}
		rows := make([][]string, 0, len(result.Territories))
		for _, t := range result.Territories {
			rows = append(rows, []string{t.Territory, t.Price, t.PricePointID})
		}
		asc.RenderTable(headers, rows)
		return nil
	}

	fmt.Printf("Subscription: %s\n", result.SubscriptionID)
	fmt.Printf("Base: %s @ %s\n", result.BaseTerritory, result.BasePrice)
	fmt.Printf("Total: %d, Succeeded: %d, Failed: %d\n", result.Total, result.Succeeded, result.Failed)

	if len(result.Failures) > 0 {
		fmt.Println("\nFailures:")
		headers := []string{"Territory", "Price", "Error"}
		rows := make([][]string, 0, len(result.Failures))
		for _, f := range result.Failures {
			rows = append(rows, []string{f.Territory, f.Price, f.Error})
		}
		asc.RenderTable(headers, rows)
	}

	return nil
}

// pricesMatch compares a candidate price string against the target using numeric
// comparison when possible, falling back to exact string match. This handles cases
// like "3.5" matching "3.50".
func pricesMatch(candidate, targetStr string, targetFloat float64, targetErr error) bool {
	candidate = strings.TrimSpace(candidate)
	if candidate == targetStr {
		return true
	}
	if targetErr != nil {
		return false
	}
	candidateFloat, err := strconv.ParseFloat(candidate, 64)
	if err != nil {
		return false
	}
	return candidateFloat == targetFloat
}

func printEqualizeMarkdown(result *equalizeResult) error {
	if result.DryRun {
		headers := []string{"Territory", "Price", "Price Point ID"}
		rows := make([][]string, 0, len(result.Territories))
		for _, t := range result.Territories {
			rows = append(rows, []string{t.Territory, t.Price, t.PricePointID})
		}
		asc.RenderMarkdown(headers, rows)
		return nil
	}

	fmt.Printf("## Equalize Results\n\n")
	fmt.Printf("- **Subscription:** %s\n", result.SubscriptionID)
	fmt.Printf("- **Base:** %s @ %s\n", result.BaseTerritory, result.BasePrice)
	fmt.Printf("- **Total:** %d, **Succeeded:** %d, **Failed:** %d\n\n", result.Total, result.Succeeded, result.Failed)

	if len(result.Failures) > 0 {
		headers := []string{"Territory", "Price", "Error"}
		rows := make([][]string, 0, len(result.Failures))
		for _, f := range result.Failures {
			rows = append(rows, []string{f.Territory, f.Price, f.Error})
		}
		asc.RenderMarkdown(headers, rows)
	}

	return nil
}
