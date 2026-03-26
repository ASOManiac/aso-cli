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
	fs := flag.NewFlagSet("aso aso dashboard", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "dashboard",
		ShortUsage: "aso aso dashboard",
		ShortHelp:  "View your ASO portfolio dashboard overview.",
		LongHelp: `Display a summary of your tracked apps, keywords, and recent changes.

Example:
  aso aso dashboard`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return runDashboard(ctx, asomaniac.DefaultConfigPath(), os.Stdout)
		},
	}
}

func runDashboard(ctx context.Context, configPath string, w io.Writer) error {
	cfg, err := asomaniac.ReadConfig(configPath)
	if err != nil {
		return fmt.Errorf("not logged in. Run 'aso aso login' to authenticate")
	}
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not logged in. Run 'aso aso login' to authenticate")
	}

	client := asomaniac.NewClientFromConfig(cfg)
	result, err := client.GetDashboard(ctx)
	if err != nil {
		return fmt.Errorf("get dashboard: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
