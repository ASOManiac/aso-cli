package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/registry"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
	"github.com/ASOManiac/aso-cli/internal/cli/shared/suggest"
)

var versionRequested bool

// RootCommand returns the root command
func RootCommand(version string) *ffcli.Command {
	versionRequested = false
	subcommands := registry.Subcommands(version)
	root := &ffcli.Command{
		Name:        "aso",
		ShortUsage:  "aso <subcommand> [flags]",
		ShortHelp:   "ASO CLI — App Store Optimization + App Store Connect.",
		LongHelp:    "",
		FlagSet:     flag.NewFlagSet("aso", flag.ExitOnError),
		UsageFunc:   RootUsageFunc,
		Subcommands: subcommands,
	}

	for _, subcommand := range subcommands {
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
					if shouldHideRootCommand(sub) {
						continue
					}
					rootSubcommandNames = append(rootSubcommandNames, sub.Name)
				}
			})
			unknown := shared.SanitizeTerminal(args[0])
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", unknown)
			if suggestions := suggest.Commands(args[0], rootSubcommandNames); len(suggestions) > 0 {
				for i, suggestion := range suggestions {
					suggestions[i] = shared.SanitizeTerminal(suggestion)
				}
				fmt.Fprintf(os.Stderr, "Did you mean: %s\n\n", strings.Join(suggestions, ", "))
			}
		}
		return flag.ErrHelp
	}

	return root
}
