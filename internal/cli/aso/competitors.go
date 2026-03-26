package aso

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// CompetitorsCommand returns the "competitors" subcommand stub.
func CompetitorsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso aso competitors", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "competitors",
		ShortUsage: "aso aso competitors <subcommand> [flags]",
		ShortHelp:  "Analyze competitor keyword overlap.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
