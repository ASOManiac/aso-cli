package performance

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// PerformanceCommand returns the performance command group.
func PerformanceCommand() *ffcli.Command {
	fs := flag.NewFlagSet("performance", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "performance",
		ShortUsage: "aso performance <subcommand> [flags]",
		ShortHelp:  "Access performance metrics and diagnostic logs.",
		LongHelp: `Access performance metrics and diagnostic logs.

Examples:
  aso performance metrics list --app "APP_ID"
  aso performance metrics get --build "BUILD_ID"
  aso performance diagnostics list --build "BUILD_ID"
  aso performance diagnostics get --id "SIGNATURE_ID"
  aso performance download --build "BUILD_ID" --output ./metrics.json`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			PerformanceMetricsCommand(),
			PerformanceDiagnosticsCommand(),
			PerformanceDownloadCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
