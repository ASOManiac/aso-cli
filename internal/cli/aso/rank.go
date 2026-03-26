package aso

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/asomaniac"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// RankCommand returns the "rank" subcommand with track and history.
func RankCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso rank", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "rank",
		ShortUsage: "aso rank <subcommand> [flags]",
		ShortHelp:  "Track and view keyword ranking history.",
		LongHelp: `Rank tracking commands for your app portfolio.

Subcommands:
  track    Add an app to rank tracking with specific keywords
  history  View historical rank data for an app's keyword`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			rankTrackCommand(),
			rankHistoryCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return flag.ErrHelp
			}
			fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n\n", args[0])
			return flag.ErrHelp
		},
	}
}

func rankTrackCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso rank track", flag.ExitOnError)
	storefront := fs.String("storefront", "US", "App Store storefront code")
	keywords := fs.String("keywords", "", "Comma-separated keywords to track")

	return &ffcli.Command{
		Name:       "track",
		ShortUsage: "aso rank track <appId> --keywords camera,photo [flags]",
		ShortHelp:  "Add an app to rank tracking with specific keywords.",
		LongHelp: `Start tracking an app's keyword rankings.

Examples:
  aso rank track 123456789 --keywords camera,photo
  aso rank track 123456789 --keywords vpn --storefront GB`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("app ID is required")
			}
			var kwList []string
			if *keywords != "" {
				kwList = strings.Split(*keywords, ",")
			}
			return runRankTrack(ctx, asomaniac.DefaultConfigPath(), args[0], *storefront, kwList, os.Stdout)
		},
	}
}

func runRankTrack(ctx context.Context, configPath, appID, storefront string, keywords []string, w io.Writer) error {
	cfg, err := asomaniac.ReadConfig(configPath)
	if err != nil {
		return fmt.Errorf("not logged in. Run 'aso login' to authenticate")
	}
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not logged in. Run 'aso login' to authenticate")
	}

	client := asomaniac.NewClientFromConfig(cfg)
	result, err := client.TrackApp(ctx, appID, storefront, keywords)
	if err != nil {
		return fmt.Errorf("track app: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func rankHistoryCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso rank history", flag.ExitOnError)
	keyword := fs.String("keyword", "", "Keyword to get rank history for (required)")
	storefront := fs.String("storefront", "US", "App Store storefront code")
	from := fs.String("from", "", "Start date (YYYY-MM-DD)")
	to := fs.String("to", "", "End date (YYYY-MM-DD)")

	return &ffcli.Command{
		Name:       "history",
		ShortUsage: "aso rank history <appId> --keyword camera [flags]",
		ShortHelp:  "View historical rank data for an app's keyword.",
		LongHelp: `View rank history for a tracked app and keyword.

Examples:
  aso rank history 123456789 --keyword camera
  aso rank history 123456789 --keyword vpn --from 2026-01-01 --to 2026-03-01`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("app ID is required")
			}
			if *keyword == "" {
				return fmt.Errorf("--keyword is required")
			}
			return runRankHistory(ctx, asomaniac.DefaultConfigPath(), args[0], *keyword, *storefront, *from, *to, os.Stdout)
		},
	}
}

func runRankHistory(ctx context.Context, configPath, appID, keyword, storefront, from, to string, w io.Writer) error {
	cfg, err := asomaniac.ReadConfig(configPath)
	if err != nil {
		return fmt.Errorf("not logged in. Run 'aso login' to authenticate")
	}
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not logged in. Run 'aso login' to authenticate")
	}

	client := asomaniac.NewClientFromConfig(cfg)
	result, err := client.GetRankHistory(ctx, appID, keyword, storefront, from, to)
	if err != nil {
		return fmt.Errorf("get rank history: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
