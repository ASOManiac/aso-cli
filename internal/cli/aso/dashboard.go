package aso

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/asomaniac"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// DashboardCommand returns the "dashboard" subcommand.
func DashboardCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso dashboard", flag.ContinueOnError)
	exclude := fs.String("exclude", "", "Comma-separated fields to hide from output")
	return &ffcli.Command{
		Name:       "dashboard",
		ShortUsage: "aso dashboard",
		ShortHelp:  "Portfolio overview — tracked apps, rank changes, and alerts.",
		LongHelp: `Display a summary of your tracked apps, keyword rankings, recent rank
changes, and any alerts that need attention.

Example:
  aso dashboard`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return runDashboard(ctx, asomaniac.DefaultConfigPath(), parseExclude(*exclude), os.Stdout)
		},
	}
}

func runDashboard(ctx context.Context, configPath string, exclude []string, w io.Writer) error {
	client, err := requireAuth(configPath)
	if err != nil {
		return err
	}

	result, err := client.GetDashboard(ctx)
	if err != nil {
		return fmt.Errorf("get dashboard: %w", err)
	}

	return writeJSON(w, result, exclude)
}
