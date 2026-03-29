package analytics

import (
	"context"
	"flag"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
	"github.com/peterbourgon/ff/v3/ffcli"
)

// AnalyticsCommand returns the analytics command with subcommands.
func AnalyticsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("analytics", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "analytics",
		ShortUsage: "aso analytics <subcommand> [flags]",
		ShortHelp:  "Request and download analytics and sales reports.",
		LongHelp: `Request and download analytics and sales reports.

Examples:
  aso analytics sales --vendor "12345678" --type SALES --subtype SUMMARY --frequency DAILY --date "2024-01-20"
  aso analytics request --app "APP_ID" --access-type ONGOING
  aso analytics requests --app "APP_ID"
  aso analytics get --request-id "REQUEST_ID"
  aso analytics reports get --report-id "REPORT_ID"
  aso analytics instances links --instance-id "INSTANCE_ID"
  aso analytics download --request-id "REQUEST_ID" --instance-id "INSTANCE_ID"`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			AnalyticsSalesCommand(),
			AnalyticsRequestCommand(),
			AnalyticsRequestsCommand(),
			AnalyticsGetCommand(),
			AnalyticsReportsCommand(),
			AnalyticsInstancesCommand(),
			AnalyticsSegmentsCommand(),
			AnalyticsDownloadCommand(),
			AnalyticsCompareCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
