package alternativedistribution

import (
	"context"
	"flag"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
	"github.com/peterbourgon/ff/v3/ffcli"
)

// AlternativeDistributionCommand returns the alternative distribution command group.
func AlternativeDistributionCommand() *ffcli.Command {
	fs := flag.NewFlagSet("alternative-distribution", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "alternative-distribution",
		ShortUsage: "aso alternative-distribution <subcommand> [flags]",
		ShortHelp:  "Manage alternative distribution resources.",
		LongHelp: `Manage alternative distribution resources.

Examples:
  aso alternative-distribution domains list
  aso alternative-distribution domains create --domain "example.com" --reference-name "Example"
  aso alternative-distribution keys list
  aso alternative-distribution keys create --app "APP_ID" --public-key-path "./key.pem"
  aso alternative-distribution packages get --package-id "PACKAGE_ID"`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			AlternativeDistributionDomainsCommand(),
			AlternativeDistributionKeysCommand(),
			AlternativeDistributionPackagesCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
