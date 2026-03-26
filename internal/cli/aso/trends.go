package aso

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// TrendsCommand returns the "trends" subcommand stub.
func TrendsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso aso trends", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "trends",
		ShortUsage: "aso aso trends <subcommand> [flags]",
		ShortHelp:  "View keyword popularity trends over time.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
