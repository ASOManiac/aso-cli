package aso

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// ExportCommand returns the "export" subcommand stub.
func ExportCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso aso export", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "export",
		ShortUsage: "aso aso export [flags]",
		ShortHelp:  "Export keyword data in CSV, JSON, or TSV format.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
