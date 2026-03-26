package subscriptions

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// SubscriptionsPricingCommand returns the canonical pricing family.
func SubscriptionsPricingCommand() *ffcli.Command {
	fs := flag.NewFlagSet("pricing", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "pricing",
		ShortUsage: "aso subscriptions pricing <subcommand> [flags]",
		ShortHelp:  "Manage subscription pricing.",
		LongHelp: `Manage subscription pricing.

Examples:
  aso subscriptions pricing summary --app "APP_ID"
  aso subscriptions pricing prices list --subscription-id "SUB_ID"
  aso subscriptions pricing prices set --subscription-id "SUB_ID" --price-point "PRICE_POINT_ID"
  aso subscriptions pricing price-points list --subscription-id "SUB_ID" --territory "USA"
  aso subscriptions pricing availability view --subscription-id "SUB_ID"
  aso subscriptions pricing equalize --subscription-id "SUB_ID" --base-price "3.49"`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			SubscriptionsPricingSummaryCommand(),
			SubscriptionsPricingPricesCommand(),
			SubscriptionsPricingPricePointsCommand(),
			SubscriptionsPricingAvailabilityCommand(),
			SubscriptionsPricingEqualizeCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// SubscriptionsPricingPricesCommand returns the canonical prices subgroup.
func SubscriptionsPricingPricesCommand() *ffcli.Command {
	fs := flag.NewFlagSet("pricing prices", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "prices",
		ShortUsage: "aso subscriptions pricing prices <subcommand> [flags]",
		ShortHelp:  "Manage subscription price records.",
		LongHelp: `Manage subscription price records.

Examples:
  aso subscriptions pricing prices list --subscription-id "SUB_ID"
  aso subscriptions pricing prices set --subscription-id "SUB_ID" --price-point "PRICE_POINT_ID"
  aso subscriptions pricing prices import --subscription-id "SUB_ID" --input "./prices.csv"
  aso subscriptions pricing prices delete --price-id "PRICE_ID" --confirm`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			wrapSubscriptionsCommand(
				SubscriptionsPricesListCommand(),
				"aso subscriptions prices list",
				"aso subscriptions pricing prices list",
				"",
				"",
			),
			wrapSubscriptionsCommand(
				SubscriptionsPricesAddCommand(),
				"aso subscriptions prices add",
				"aso subscriptions pricing prices set",
				"set",
				"Set a subscription price.",
			),
			wrapSubscriptionsCommand(
				SubscriptionsPricesImportCommand(),
				"aso subscriptions prices import",
				"aso subscriptions pricing prices import",
				"",
				"",
			),
			wrapSubscriptionsCommand(
				SubscriptionsPricesDeleteCommand(),
				"aso subscriptions prices delete",
				"aso subscriptions pricing prices delete",
				"",
				"",
			),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// SubscriptionsPricingPricePointsCommand returns the canonical price points subgroup.
func SubscriptionsPricingPricePointsCommand() *ffcli.Command {
	return wrapSubscriptionsCommand(
		SubscriptionsPricePointsCommand(),
		"aso subscriptions price-points",
		"aso subscriptions pricing price-points",
		"price-points",
		"Manage subscription price points.",
	)
}

// SubscriptionsPricingAvailabilityCommand returns the canonical availability subgroup.
func SubscriptionsPricingAvailabilityCommand() *ffcli.Command {
	cmd := wrapSubscriptionsCommand(
		SubscriptionsAvailabilityCommand(),
		"aso subscriptions availability",
		"aso subscriptions pricing availability",
		"availability",
		"Manage subscription availability.",
	)
	if viewCmd := findSubscriptionsSubcommand(cmd, "view"); viewCmd != nil {
		cmd.Subcommands = append(cmd.Subcommands, shared.DeprecatedAliasLeafCommand(
			viewCmd,
			"get",
			"aso subscriptions pricing availability get [flags]",
			"aso subscriptions pricing availability view",
			"Warning: `aso subscriptions pricing availability get` is deprecated. Use `aso subscriptions pricing availability view`.",
		))
	}
	if editCmd := findSubscriptionsSubcommand(cmd, "edit"); editCmd != nil {
		cmd.Subcommands = append(cmd.Subcommands, shared.DeprecatedAliasLeafCommand(
			editCmd,
			"set",
			"aso subscriptions pricing availability set [flags]",
			"aso subscriptions pricing availability edit",
			"Warning: `aso subscriptions pricing availability set` is deprecated. Use `aso subscriptions pricing availability edit`.",
		))
	}
	return cmd
}
