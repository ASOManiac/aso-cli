package aso

import (
	"context"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// AuthCommand returns the top-level `aso auth` command. The slim build only
// hosts the asomaniac.com auth flow under `aso auth maniac`.
func AuthCommand() *ffcli.Command {
	fs := flag.NewFlagSet("auth", flag.ContinueOnError)
	return &ffcli.Command{
		Name:       "auth",
		ShortUsage: "aso auth <subcommand> [flags]",
		ShortHelp:  "Manage authentication for the ASO Maniac API.",
		LongHelp: `Manage authentication for the ASO Maniac API.

See 'aso auth maniac --help' for the available subcommands.`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			ManiacAuthCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return flag.ErrHelp
			}
			return fmt.Errorf("unknown subcommand %q. Run 'aso auth --help' for usage", args[0])
		},
	}
}
