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

// TrendsCommand returns the "trends" subcommand.
func TrendsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso trends", flag.ExitOnError)
	appID := fs.String("app-id", "", "App Store app ID (required)")
	storefront := fs.String("storefront", "US", "App Store storefront code")
	from := fs.String("from", "", "Start date (YYYY-MM-DD)")
	to := fs.String("to", "", "End date (YYYY-MM-DD)")

	return &ffcli.Command{
		Name:       "trends",
		ShortUsage: "aso trends --app-id <id> <keyword> [<keyword>...] [flags]",
		ShortHelp:  "View how keyword popularity changes over time.",
		LongHelp: `View historical popularity trends for one or more keywords.
Useful for spotting seasonal patterns and rising opportunities.

Examples:
  aso trends --app-id 123456789 camera
  aso trends --app-id 123456789 camera photo --storefront GB
  aso trends --app-id 123456789 vpn --from 2026-01-01 --to 2026-03-01`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *appID == "" {
				return fmt.Errorf("--app-id is required")
			}
			if len(args) == 0 {
				return fmt.Errorf("at least one keyword is required")
			}
			return runTrends(ctx, asomaniac.DefaultConfigPath(), args, *storefront, *appID, *from, *to, os.Stdout)
		},
	}
}

func runTrends(ctx context.Context, configPath string, keywords []string, storefront, appID, from, to string, w io.Writer) error {
	client, err := requireAuth(configPath)
	if err != nil {
		return err
	}

	results, err := client.GetTrends(ctx, keywords, storefront, appID, from, to)
	if err != nil {
		return fmt.Errorf("get trends: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}
