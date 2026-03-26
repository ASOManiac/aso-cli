package offercodes

import (
	"context"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/asc"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// OfferCodePricesCommand returns the prices command group.
func OfferCodePricesCommand() *ffcli.Command {
	fs := flag.NewFlagSet("prices", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "prices",
		ShortUsage: "aso offer-codes prices <subcommand> [flags]",
		ShortHelp:  "Manage offer code prices.",
		LongHelp: `Manage offer code prices.

Examples:
  aso offer-codes prices list --offer-code-id "OFFER_CODE_ID"`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			OfferCodePricesListCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// OfferCodePricesListCommand returns the prices list subcommand.
func OfferCodePricesListCommand() *ffcli.Command {
	return shared.BuildPaginatedListCommand(shared.PaginatedListCommandConfig{
		FlagSetName: "list",
		Name:        "list",
		ShortUsage:  "aso offer-codes prices list [flags]",
		ShortHelp:   "List prices for a subscription offer code.",
		LongHelp: `List prices for a subscription offer code.

Examples:
  aso offer-codes prices list --offer-code-id "OFFER_CODE_ID"
  aso offer-codes prices list --offer-code-id "OFFER_CODE_ID" --limit 50
  aso offer-codes prices list --offer-code-id "OFFER_CODE_ID" --paginate`,
		ParentFlag:  "offer-code-id",
		ParentUsage: "Subscription offer code ID (required)",
		LimitMax:    offerCodesMaxLimit,
		ErrorPrefix: "offer-codes prices list",
		FetchPage: func(ctx context.Context, client *asc.Client, offerCodeID string, limit int, next string) (asc.PaginatedResponse, error) {
			opts := []asc.SubscriptionOfferCodePricesOption{
				asc.WithSubscriptionOfferCodePricesLimit(limit),
				asc.WithSubscriptionOfferCodePricesNextURL(next),
			}
			resp, err := client.GetSubscriptionOfferCodePrices(ctx, offerCodeID, opts...)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch: %w", err)
			}
			return resp, nil
		},
	})
}
