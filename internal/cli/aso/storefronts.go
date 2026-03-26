package aso

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/asomaniac"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// StorefrontsCommand returns the "storefronts" subcommand.
func StorefrontsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso aso storefronts", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "storefronts",
		ShortUsage: "aso aso storefronts",
		ShortHelp:  "List all supported App Store storefronts.",
		LongHelp: `Print all supported App Store storefront codes as JSON.

No authentication required — this is a free command.

Example:
  aso aso storefronts`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return runStorefronts(os.Stdout)
		},
	}
}

func runStorefronts(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(asomaniac.Storefronts)
}
