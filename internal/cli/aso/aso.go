package aso

import (
	"context"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// ManiacAuthCommand returns the "maniac" subcommand for aso auth, grouping
// ASO Maniac authentication commands (login, logout, whoami, status).
func ManiacAuthCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso auth maniac", flag.ContinueOnError)
	return &ffcli.Command{
		Name:       "maniac",
		ShortUsage: "aso auth maniac <subcommand> [flags]",
		ShortHelp:  "Manage authentication for the ASO Maniac API.",
		LongHelp: `Manage authentication for the ASO Maniac keyword intelligence API.

Free plan: 100 API calls/month. Upgrade at https://asomaniac.com/pricing.

Authentication (pick one):
  aso auth maniac login                          Browser OAuth (interactive)
  aso auth maniac login --api-key <KEY>          Paste a key from the dashboard
  export ASO_MANIAC_API_KEY=<KEY>                Env var override (CI/scripts)

Key resolution order: ASO_MANIAC_API_KEY env var > ~/.asomaniac/config.json.`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			LoginCommand(),
			LogoutCommand(),
			WhoamiCommand(),
			StatusCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing subcommand. Run 'aso auth maniac --help' for usage")
			}
			return fmt.Errorf("unknown subcommand %q. Run 'aso auth maniac --help' for usage", args[0])
		},
	}
}
