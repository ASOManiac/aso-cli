package aso

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// DashboardCommand returns the "dashboard" subcommand stub.
func DashboardCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso aso dashboard", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "dashboard",
		ShortUsage: "aso aso dashboard [flags]",
		ShortHelp:  "View your ASO portfolio dashboard overview.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
