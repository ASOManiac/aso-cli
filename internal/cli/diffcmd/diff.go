package diffcmd

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// DiffCommand returns the root diff command group.
func DiffCommand() *ffcli.Command {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "diff",
		ShortUsage: "aso diff <subcommand> [flags]",
		ShortHelp:  "Generate deterministic non-mutating diff plans.",
		LongHelp: `Generate deterministic non-mutating diff plans.

Examples:
  aso diff localizations --app "APP_ID" --path "./metadata/localizations" --version "VERSION_ID"
  aso diff localizations --app "APP_ID" --from-version "VERSION_ID_A" --to-version "VERSION_ID_B"`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			DiffLocalizationsCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
