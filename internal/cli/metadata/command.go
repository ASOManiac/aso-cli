package metadata

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// MetadataCommand returns the metadata command group.
func MetadataCommand() *ffcli.Command {
	fs := flag.NewFlagSet("metadata", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "metadata",
		ShortUsage: "aso metadata <subcommand> [flags]",
		ShortHelp:  "Manage app metadata with deterministic workflows and keyword tooling.",
		LongHelp: `Manage app metadata with deterministic workflows and keyword tooling.

Phase 1 scope:
  - app-info localizations: name, subtitle, privacyPolicyUrl, privacyChoicesUrl, privacyPolicyText
  - version localizations: description, keywords, marketingUrl, promotionalText, supportUrl, whatsNew

Keyword workflow:
  - ` + "`aso metadata keywords ...`" + ` manages the canonical version-localization ` + "`keywords`" + ` field
  - raw App Store Connect ` + "`searchKeywords`" + ` relationship APIs remain under
    ` + "`aso apps search-keywords ...`" + ` and ` + "`aso localizations search-keywords ...`" + `

Not yet included in this group:
  - categories, review information, age ratings, screenshots

Note: copyright is managed via "aso versions create --copyright" or "aso versions update --copyright".

Examples:
  aso metadata pull --app "APP_ID" --version "1.2.3" --dir "./metadata"
  aso metadata pull --app "APP_ID" --version "1.2.3" --platform IOS --dir "./metadata"
  aso metadata keywords import --dir "./metadata" --version "1.2.3" --locale "en-US" --input "./keywords.csv"`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			MetadataPullCommand(),
			MetadataKeywordsCommand(),
			MetadataPushCommand(),
			MetadataValidateCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
