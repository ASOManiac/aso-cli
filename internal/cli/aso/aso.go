package aso

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// ManiacCommand returns the "maniac" parent command with all ASO Maniac subcommands.
func ManiacCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso maniac", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "maniac",
		ShortUsage: "aso maniac <subcommand> [flags]",
		ShortHelp:  "ASO Maniac — AI-powered keyword intelligence and rank tracking.",
		LongHelp: `Premium ASO commands powered by asomaniac.com.

Free plan: 100 API calls/month. Upgrade at https://asomaniac.com/pricing.

Authentication (pick one):
  aso maniac login                          Browser OAuth (interactive)
  aso maniac login --api-key <KEY>          Paste a key from the dashboard
  export ASO_MANIAC_API_KEY=<KEY>           Env var override (CI/scripts)

Key resolution order: ASO_MANIAC_API_KEY env var > ~/.asomaniac/config.json.

Auth & Account:
  login         Sign in via browser or API key
  logout        Remove stored credentials
  whoami        Show account, plan, and API usage
  status        Check connection, auth health, and key source

Keyword Intelligence:
  keywords      Analyze, suggest, and batch-process keywords
  competitors   Find competitor keyword overlap for any app
  trends        View keyword popularity over time
  rank          Track keyword rankings for your apps

Data:
  dashboard     Portfolio overview — tracked apps, rank changes, alerts
  export        Download keyword data as CSV, JSON, or TSV
  storefronts   List all 60+ supported App Store storefronts`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			LoginCommand(),
			LogoutCommand(),
			WhoamiCommand(),
			StatusCommand(),
			KeywordsCommand(),
			CompetitorsCommand(),
			TrendsCommand(),
			RankCommand(),
			DashboardCommand(),
			ExportCommand(),
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
