package migrate

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"

	metadatacmd "github.com/ASOManiac/aso-cli/internal/cli/metadata"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// MigrateMetadataCommand provides migration-friendly aliases for metadata workflows.
func MigrateMetadataCommand() *ffcli.Command {
	fs := flag.NewFlagSet("migrate metadata", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "metadata",
		ShortUsage: "aso migrate metadata <pull|push|validate> [flags]",
		ShortHelp:  "Compatibility aliases for aso metadata commands.",
		LongHelp: `Compatibility aliases for aso metadata commands.

These aliases help teams move from fastlane/deliver conventions while
adopting native aso metadata workflows.

Prefer direct commands for new scripts:
  aso metadata pull ...
  aso metadata push ...
  aso metadata validate ...`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			metadatacmd.MetadataPullCommand(),
			metadatacmd.MetadataPushCommand(),
			metadatacmd.MetadataValidateCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			fmt.Fprintln(os.Stderr, "Tip: use `aso metadata ...`; `aso migrate metadata ...` is a compatibility alias.")
			return flag.ErrHelp
		},
	}
}
