package aso

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// ASOCommand returns the top-level "aso" command group for ASO Maniac subcommands.
func ASOCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "aso",
		ShortUsage: "aso aso <subcommand> [flags]",
		ShortHelp:  "ASO Maniac — AI-powered App Store Optimization commands.",
		LongHelp: `Premium ASO commands powered by asomaniac.com.

Run 'aso aso login' to authenticate. Free plan: 100 API calls/month.`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			LoginCommand(),
			LogoutCommand(),
			WhoamiCommand(),
			KeywordsCommand(),
			CompetitorsCommand(),
			RankCommand(),
			TrendsCommand(),
			ExportCommand(),
			DashboardCommand(),
			StorefrontsCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return flag.ErrHelp
			}
			fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n\n", args[0])
			return flag.ErrHelp
		},
	}
}
