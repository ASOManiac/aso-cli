package testflight

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
)

// TestFlightCommand returns the testflight command with subcommands.
func TestFlightCommand() *ffcli.Command {
	fs := flag.NewFlagSet("testflight", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "testflight",
		ShortUsage: "aso testflight <subcommand> [flags]",
		ShortHelp:  "Manage TestFlight workflows.",
		LongHelp: `Manage TestFlight workflows.

Examples:
  aso testflight groups list --app "APP_ID"
  aso testflight testers list --app "APP_ID"
  aso testflight feedback list --app "APP_ID"
  aso testflight crashes view --submission-id "SUBMISSION_ID"
  aso testflight crashes log --submission-id "SUBMISSION_ID"
  aso testflight review view --app "APP_ID"
  aso testflight distribution view --build "BUILD_ID"
  aso testflight metrics group-testers --group "GROUP_ID"
  aso testflight metrics app-testers --app "APP_ID"
  aso testflight agreements view --app "APP_ID"
  aso testflight notifications send --build "BUILD_ID"
  aso testflight config export --app "APP_ID" --output "./testflight.yaml"
  aso testflight app-localizations list --app "APP_ID"
  aso testflight pre-release list --app "APP_ID"`,
		FlagSet:   fs,
		UsageFunc: testflightVisibleUsageFunc,
		Subcommands: []*ffcli.Command{
			RemovedTestFlightAppsCommand(),
			TestFlightGroupsCommand(),
			TestFlightTestersCommand(),
			TestFlightFeedbackCommand(),
			TestFlightCrashesCommand(),
			TestFlightAgreementsCommand(),
			TestFlightNotificationsCommand(),
			TestFlightReviewSurfaceCommand(),
			TestFlightDistributionCommand(),
			TestFlightRecruitmentCommand(),
			TestFlightMetricsSurfaceCommand(),
			TestFlightConfigCommand(),
			TestFlightAppLocalizationsCommand(),
			TestFlightPreReleaseCommand(),
			DeprecatedBetaGroupsAliasCommand(),
			DeprecatedBetaTestersAliasCommand(),
			DeprecatedBetaFeedbackAliasCommand(),
			DeprecatedBetaCrashLogsAliasCommand(),
			DeprecatedBetaDetailsAliasCommand(),
			DeprecatedBetaLicenseAgreementsAliasCommand(),
			DeprecatedBetaNotificationsAliasCommand(),
			DeprecatedTestFlightSyncAliasCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
