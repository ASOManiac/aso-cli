package docs

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// DocsCommand returns the docs command group.
func DocsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("docs", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "docs",
		ShortUsage: "aso docs <subcommand> [flags]",
		ShortHelp:  "Access embedded documentation guides and reference helpers.",
		LongHelp: `Access embedded documentation guides and reference helpers.

Examples:
  aso docs list
  aso docs show api-notes
  aso docs init
  aso docs init --path ./ASC.md
  aso docs init --force --link=false`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			DocsListCommand(),
			DocsShowCommand(),
			DocsInitCommand(),
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
