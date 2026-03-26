package subscriptions

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// SubscriptionsOffersCommand returns the canonical offers family.
func SubscriptionsOffersCommand() *ffcli.Command {
	fs := flag.NewFlagSet("offers", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "offers",
		ShortUsage: "aso subscriptions offers <subcommand> [flags]",
		ShortHelp:  "Manage subscription offers.",
		LongHelp: `Manage subscription offers.

Examples:
  aso subscriptions offers introductory list --subscription-id "SUB_ID"
  aso subscriptions offers promotional create --subscription-id "SUB_ID" --offer-code "SPRING" --name "Spring" --offer-duration ONE_MONTH --offer-mode FREE_TRIAL --number-of-periods 1 --prices "PRICE_ID"
  aso subscriptions offers offer-codes generate --offer-code-id "OFFER_CODE_ID" --quantity 10 --expiration-date "2026-02-01"
  aso subscriptions offers win-back list --subscription-id "SUB_ID"`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			wrapSubscriptionsCommand(
				SubscriptionsIntroductoryOffersCommand(),
				"aso subscriptions introductory-offers",
				"aso subscriptions offers introductory",
				"introductory",
				"Manage subscription introductory offers.",
			),
			wrapSubscriptionsCommand(
				SubscriptionsPromotionalOffersCommand(),
				"aso subscriptions promotional-offers",
				"aso subscriptions offers promotional",
				"promotional",
				"Manage subscription promotional offers.",
			),
			wrapSubscriptionsCommand(
				SubscriptionsOfferCodesCommand(),
				"aso subscriptions offer-codes",
				"aso subscriptions offers offer-codes",
				"offer-codes",
				"Manage subscription offer codes.",
			),
			wrapSubscriptionsCommand(
				SubscriptionsWinBackOffersCommand(),
				"aso subscriptions win-back-offers",
				"aso subscriptions offers win-back",
				"win-back",
				"Manage subscription win-back offers.",
			),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
