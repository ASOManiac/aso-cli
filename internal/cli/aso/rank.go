package aso

import (
	"context"
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
	fs := flag.NewFlagSet("aso rank", flag.ContinueOnError)
	return &ffcli.Command{
		Name:       "rank",
		ShortUsage: "aso rank <subcommand> [flags]",
		ShortHelp:  "Track keyword rankings and view position history.",
		LongHelp: `Monitor how your apps rank for specific keywords over time.

Subcommands:
  track    Start tracking an app's keyword rankings
  history  View historical rank positions for a keyword`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			rankTrackCommand(),
			rankHistoryCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing subcommand. Run 'aso rank --help' for usage")
			}
			return fmt.Errorf("unknown subcommand %q. Run 'aso rank --help' for usage", args[0])
		},
	}
}

func rankTrackCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso rank track", flag.ContinueOnError)
	storefront := fs.String("storefront", "US", "App Store storefront code")
	keywords := fs.String("keywords", "", "Comma-separated keywords to track")
	excludeTrack := fs.String("exclude", "", "Comma-separated fields to hide from output")

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
			positional := resolveArgs(fs, args, false)
			if len(positional) == 0 {
				return fmt.Errorf("app ID is required")
			}
			var kwList []string
			if *keywords != "" {
				kwList = strings.Split(*keywords, ",")
			}
			return runRankTrack(ctx, asomaniac.DefaultConfigPath(), positional[0], *storefront, kwList, parseExclude(*excludeTrack), os.Stdout)
		},
	}
}

func runRankTrack(ctx context.Context, configPath, appID, storefront string, keywords, exclude []string, w io.Writer) error {
	client, err := requireAuth(configPath)
	if err != nil {
		return err
	}

	result, err := client.TrackApp(ctx, appID, storefront, keywords)
	if err != nil {
		return fmt.Errorf("track app: %w", err)
	}

	return writeJSON(w, result, exclude)
}

func rankHistoryCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso rank history", flag.ContinueOnError)
	keywordID := fs.String("keyword-id", "", "Keyword ID to get rank history for (required)")
	storefront := fs.String("storefront", "US", "App Store storefront code")
	from := fs.String("from", "", "Start date YYYY-MM-DD (required)")
	to := fs.String("to", "", "End date YYYY-MM-DD (required)")
	granularity := fs.String("granularity", "day", "Time granularity: day, week, or month")
	aggregation := fs.String("aggregation", "avg", "Aggregation function: avg, min, or max")
	excludeHist := fs.String("exclude", "", "Comma-separated fields to hide from output")

	return &ffcli.Command{
		Name:       "history",
		ShortUsage: "aso rank history <appId> --keyword-id <id> --from <date> --to <date> [flags]",
		ShortHelp:  "View historical rank data for an app's keyword.",
		LongHelp: `View rank history for a tracked app and keyword.

Examples:
  aso rank history 123456789 --keyword-id kw_abc --from 2026-01-01 --to 2026-03-01
  aso rank history 123456789 --keyword-id kw_abc --from 2026-01-01 --to 2026-03-01 --granularity week`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			positional := resolveArgs(fs, args, false)
			if len(positional) == 0 {
				return fmt.Errorf("app ID is required")
			}
			if *keywordID == "" {
				return fmt.Errorf("--keyword-id is required")
			}
			if *from == "" || *to == "" {
				return fmt.Errorf("--from and --to are required")
			}
			return runRankHistory(ctx, asomaniac.DefaultConfigPath(), positional[0], *keywordID, *storefront, *from, *to, *granularity, *aggregation, parseExclude(*excludeHist), os.Stdout)
		},
	}
}

func runRankHistory(ctx context.Context, configPath, appID, keywordID, storefront, from, to, granularity, aggregation string, exclude []string, w io.Writer) error {
	client, err := requireAuth(configPath)
	if err != nil {
		return err
	}

	result, err := client.GetRankHistory(ctx, appID, keywordID, storefront, from, to, granularity, aggregation)
	if err != nil {
		return fmt.Errorf("get rank history: %w", err)
	}

	return writeJSON(w, result, exclude)
}
