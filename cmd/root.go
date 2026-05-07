package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/aso"
	"github.com/ASOManiac/aso-cli/internal/cli/completion"
	"github.com/ASOManiac/aso-cli/internal/cli/docs"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
	"github.com/ASOManiac/aso-cli/internal/cli/shared/suggest"
)

var versionRequested bool

// RootCommand returns the root command tree for the slim aso CLI.
func RootCommand(version string) *ffcli.Command {
	versionRequested = false

	subcommands := []*ffcli.Command{
		aso.AuthCommand(),
		aso.KeywordsCommand(),
		aso.StorefrontsCommand(),
		docs.DocsCommand(),
	}

	root := &ffcli.Command{
		Name:        "aso",
		ShortUsage:  "aso <subcommand> [flags]",
		ShortHelp:   "ASO Maniac CLI: keyword intelligence powered by asomaniac.com.",
		FlagSet:     flag.NewFlagSet("aso", flag.ExitOnError),
		UsageFunc:   RootUsageFunc,
		Subcommands: subcommands,
	}

	// Completion needs the root subcommand list to enumerate names.
	root.Subcommands = append(root.Subcommands, completion.CompletionCommand(root.Subcommands))

	for _, subcommand := range root.Subcommands {
		shared.WrapCommandOutputValidation(subcommand)
	}

	root.FlagSet.BoolVar(&versionRequested, "version", false, "Print version and exit")
	shared.BindRootFlags(root.FlagSet)

	var (
		rootSubcommandNames     []string
		rootSubcommandNamesOnce sync.Once
	)

	root.Exec = func(ctx context.Context, args []string) error {
		if versionRequested {
			fmt.Fprintln(os.Stdout, version)
			return nil
		}
		if len(args) > 0 {
			rootSubcommandNamesOnce.Do(func() {
				rootSubcommandNames = make([]string, 0, len(root.Subcommands))
				for _, sub := range root.Subcommands {
					rootSubcommandNames = append(rootSubcommandNames, sub.Name)
				}
			})
			unknown := shared.SanitizeTerminal(args[0])
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", unknown)
			if suggestions := suggest.Commands(args[0], rootSubcommandNames); len(suggestions) > 0 {
				for i, s := range suggestions {
					suggestions[i] = shared.SanitizeTerminal(s)
				}
				fmt.Fprintf(os.Stderr, "Did you mean: %s\n\n", strings.Join(suggestions, ", "))
			}
		}
		return flag.ErrHelp
	}

	return root
}
