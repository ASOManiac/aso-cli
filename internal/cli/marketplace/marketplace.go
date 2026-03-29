package marketplace

import (
	"context"
	"flag"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
	"github.com/peterbourgon/ff/v3/ffcli"
)

// MarketplaceCommand returns the marketplace command with subcommands.
func MarketplaceCommand() *ffcli.Command {
	fs := flag.NewFlagSet("marketplace", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "marketplace",
		ShortUsage: "aso marketplace <subcommand> [flags]",
		ShortHelp:  "Manage marketplace resources.",
		LongHelp: `Manage marketplace resources.

Examples:
  aso marketplace search-details get --app "APP_ID"
  aso marketplace webhooks list`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			MarketplaceSearchDetailsCommand(),
			MarketplaceWebhooksCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
