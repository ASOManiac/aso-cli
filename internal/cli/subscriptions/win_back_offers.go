package subscriptions

import (
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
	"github.com/ASOManiac/aso-cli/internal/cli/winbackoffers"
)

// SubscriptionsWinBackOffersCommand returns the canonical nested win-back offers tree.
func SubscriptionsWinBackOffersCommand() *ffcli.Command {
	return shared.RewriteCommandTreePath(
		winbackoffers.WinBackOffersCommand(),
		"aso win-back-offers",
		"aso subscriptions win-back-offers",
	)
}
