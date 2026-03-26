package aso

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// KeywordsCommand returns the "keywords" subcommand stub.
func KeywordsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso aso keywords", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "keywords",
		ShortUsage: "aso aso keywords <subcommand> [flags]",
		ShortHelp:  "Analyze, suggest, and batch-process keywords.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
