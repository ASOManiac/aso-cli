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
	fs := flag.NewFlagSet("aso maniac trends", flag.ExitOnError)
	storefront := fs.String("storefront", "US", "App Store storefront code")
	from := fs.String("from", "", "Start date (YYYY-MM-DD)")
	to := fs.String("to", "", "End date (YYYY-MM-DD)")

	return &ffcli.Command{
		Name:       "trends",
		ShortUsage: "aso maniac trends <keyword> [<keyword>...] [flags]",
		ShortHelp:  "View how keyword popularity changes over time.",
		LongHelp: `View historical popularity trends for one or more keywords.
Useful for spotting seasonal patterns and rising opportunities.

Examples:
  aso maniac trends camera
  aso maniac trends camera photo --storefront GB
  aso maniac trends vpn --from 2026-01-01 --to 2026-03-01`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("at least one keyword is required")
			}
			return runTrends(ctx, asomaniac.DefaultConfigPath(), args, *storefront, *from, *to, os.Stdout)
		},
	}
}

func runTrends(ctx context.Context, configPath string, keywords []string, storefront, from, to string, w io.Writer) error {
	client, err := requireAuth(configPath)
	if err != nil {
		return err
	}

	results, err := client.GetTrends(ctx, keywords, storefront, from, to)
	if err != nil {
		return fmt.Errorf("get trends: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}
