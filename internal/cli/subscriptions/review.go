package subscriptions

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// SubscriptionsReviewCommand returns the canonical review family.
func SubscriptionsReviewCommand() *ffcli.Command {
	fs := flag.NewFlagSet("review", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "review",
		ShortUsage: "aso subscriptions review <subcommand> [flags]",
		ShortHelp:  "Manage subscription review workflows.",
		LongHelp: `Manage subscription review workflows.

Examples:
  aso subscriptions review screenshots create --subscription-id "SUB_ID" --file "./screenshot.png"
  aso subscriptions review app-store-screenshot get --subscription-id "SUB_ID"
  aso subscriptions review submit --subscription-id "SUB_ID" --confirm
  aso subscriptions review submit-group --group-id "GROUP_ID" --confirm`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			wrapSubscriptionsCommand(
				SubscriptionsReviewScreenshotsCommand(),
				"aso subscriptions review-screenshots",
				"aso subscriptions review screenshots",
				"screenshots",
				"Manage subscription App Store review screenshots.",
			),
			wrapSubscriptionsCommand(
				SubscriptionsAppStoreReviewScreenshotCommand(),
				"aso subscriptions app-store-review-screenshot",
				"aso subscriptions review app-store-screenshot",
				"app-store-screenshot",
				"Inspect the App Store review screenshot for a subscription.",
			),
			wrapSubscriptionsCommand(
				SubscriptionsSubmitCommand(),
				"aso subscriptions submit",
				"aso subscriptions review submit",
				"submit",
				"Submit a subscription for review.",
			),
			wrapSubscriptionsCommand(
				SubscriptionsGroupsSubmitCommand(),
				"aso subscriptions groups submit",
				"aso subscriptions review submit-group",
				"submit-group",
				"Submit a subscription group for review.",
			),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
