package aso

import (
	"context"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// AuthCommand returns the top-level `aso auth` command, hosting login,
// logout, status, and whoami for the asomaniac.com API.
func AuthCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso auth", flag.ContinueOnError)
	return &ffcli.Command{
		Name:       "auth",
		ShortUsage: "aso auth <subcommand> [flags]",
		ShortHelp:  "Manage authentication for the ASO Maniac API.",
		LongHelp: `Manage authentication for the ASO Maniac keyword intelligence API.

Free plan: 100 API calls/month. Upgrade at https://asomaniac.com/pricing.

Authentication (pick one):
  aso auth login                          Browser OAuth (interactive)
  aso auth login --api-key <KEY>          Paste a key from the dashboard
  export ASO_MANIAC_API_KEY=<KEY>         Env var override (CI/scripts)

Key resolution order: ASO_MANIAC_API_KEY env var > ~/.asomaniac/config.json.`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			LoginCommand(),
			LogoutCommand(),
			StatusCommand(),
			WhoamiCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return flag.ErrHelp
			}
			return fmt.Errorf("unknown subcommand %q. Run 'aso auth --help' for usage", args[0])
		},
	}
}
