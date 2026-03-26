package builds

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

func deprecatedBuildsRelationshipsAliasCommand() *ffcli.Command {
	fs := BuildsRelationshipsCommand().FlagSet

	return &ffcli.Command{
		Name:       "relationships",
		ShortUsage: "aso builds links <subcommand> [flags]",
		ShortHelp:  "DEPRECATED: use `aso builds links ...`.",
		LongHelp:   "Deprecated compatibility alias for `aso builds links ...`.",
		FlagSet:    fs,
		UsageFunc:  shared.DeprecatedUsageFunc,
		Subcommands: []*ffcli.Command{
			shared.DeprecatedAliasLeafCommand(
				BuildsRelationshipsGetCommand(),
				"get",
				"aso builds links view --build \"BUILD_ID\" --type \"RELATIONSHIP\" [flags]",
				"aso builds links view",
				"Warning: `aso builds relationships get` is deprecated. Use `aso builds links view`.",
			),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
