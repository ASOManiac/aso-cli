package aso

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// RankCommand returns the "rank" subcommand stub.
func RankCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso aso rank", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "rank",
		ShortUsage: "aso aso rank <subcommand> [flags]",
		ShortHelp:  "Track and view keyword ranking history.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
