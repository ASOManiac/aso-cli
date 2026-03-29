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

// CompetitorsCommand returns the "competitors" subcommand.
func CompetitorsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso maniac competitors", flag.ExitOnError)
	storefront := fs.String("storefront", "US", "App Store storefront code")

	return &ffcli.Command{
		Name:       "competitors",
		ShortUsage: "aso maniac competitors <appId> [flags]",
		ShortHelp:  "Find competitor apps and compare keyword overlap.",
		LongHelp: `Discover competitor apps for a given app ID and compare shared vs unique
keywords, ranking positions, and keyword overlap.

Examples:
  aso maniac competitors 123456789
  aso maniac competitors 123456789 --storefront GB`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("app ID is required")
			}
			return runCompetitors(ctx, asomaniac.DefaultConfigPath(), args[0], *storefront, os.Stdout)
		},
	}
}

func runCompetitors(ctx context.Context, configPath, appID, storefront string, w io.Writer) error {
	client, err := requireAuth(configPath)
	if err != nil {
		return err
	}

	result, err := client.GetCompetitors(ctx, appID, storefront)
	if err != nil {
		return fmt.Errorf("get competitors: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
