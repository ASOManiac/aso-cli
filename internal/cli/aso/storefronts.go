package aso

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// StorefrontsCommand returns the "storefronts" subcommand stub.
func StorefrontsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso aso storefronts", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "storefronts",
		ShortUsage: "aso aso storefronts [flags]",
		ShortHelp:  "List all supported App Store storefronts.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
