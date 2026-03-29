package productpages

import (
	"context"
	"flag"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
	"github.com/peterbourgon/ff/v3/ffcli"
)

const productPagesMaxLimit = 200

// ProductPagesCommand returns the product pages command group.
func ProductPagesCommand() *ffcli.Command {
	fs := flag.NewFlagSet("product-pages", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "product-pages",
		ShortUsage: "aso product-pages <subcommand> [flags]",
		ShortHelp:  "Manage custom product pages and product page experiments.",
		LongHelp: `Manage custom product pages and product page optimization experiments.

Examples:
  aso product-pages custom-pages list --app "APP_ID"
  aso product-pages custom-pages create --app "APP_ID" --name "Summer Campaign"
  aso product-pages experiments list --version-id "VERSION_ID"
  aso product-pages experiments create --version-id "VERSION_ID" --name "Icon Test" --traffic-proportion 25`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			CustomPagesCommand(),
			ExperimentsCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
