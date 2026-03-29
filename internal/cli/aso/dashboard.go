package aso

import (
	"context"
	"encoding/json"
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
	fs := flag.NewFlagSet("aso maniac dashboard", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "dashboard",
		ShortUsage: "aso maniac dashboard",
		ShortHelp:  "Portfolio overview — tracked apps, rank changes, and alerts.",
		LongHelp: `Display a summary of your tracked apps, keyword rankings, recent rank
changes, and any alerts that need attention.

Example:
  aso maniac dashboard`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return runDashboard(ctx, asomaniac.DefaultConfigPath(), os.Stdout)
		},
	}
}

func runDashboard(ctx context.Context, configPath string, w io.Writer) error {
	client, err := requireAuth(configPath)
	if err != nil {
		return err
	}

	result, err := client.GetDashboard(ctx)
	if err != nil {
		return fmt.Errorf("get dashboard: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
