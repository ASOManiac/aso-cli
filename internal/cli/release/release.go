package release

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// ReleaseCommand returns the top-level release command group.
func ReleaseCommand() *ffcli.Command {
	fs := flag.NewFlagSet("release", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "release",
		ShortUsage: "aso release <subcommand> [flags]",
		ShortHelp:  "Run high-level App Store release workflows.",
		LongHelp: `Run high-level App Store release workflows.

release stage prepares a version for review without submitting it. It orchestrates:
  1. Ensure/create version
  2. Apply metadata and localizations
  3. Attach selected build
  4. Run readiness checks

release run is the canonical path for shipping to the App Store. It orchestrates:
  1. Ensure/create version
  2. Apply metadata and localizations
  3. Attach selected build
  4. Run readiness checks
  5. Submit for review

After submission, monitor progress with:
  aso status --app "APP_ID"

For lower-level control, use:
  aso validate --app "APP_ID" --version "VERSION"
  aso submit create --app "APP_ID" --version "VERSION" --build "BUILD_ID" --confirm

Examples:
  aso release stage --app "APP_ID" --version "2.4.0" --build "BUILD_ID" --copy-metadata-from "2.3.2" --dry-run
  aso release run --app "APP_ID" --version "2.4.0" --build "BUILD_ID" --metadata-dir "./metadata/version/2.4.0" --dry-run
  aso release run --app "APP_ID" --version "2.4.0" --build "BUILD_ID" --metadata-dir "./metadata/version/2.4.0" --confirm
  aso status --app "APP_ID"`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			ReleaseRunCommand(),
			ReleaseStageCommand(),
		},
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
	}
}
