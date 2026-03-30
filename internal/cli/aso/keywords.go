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

// KeywordsCommand returns the "keywords" subcommand with analyze, recommend, and batch.
func KeywordsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso keywords", flag.ContinueOnError)
	return &ffcli.Command{
		Name:       "keywords",
		ShortUsage: "aso keywords <subcommand> [flags]",
		ShortHelp:  "Analyze popularity, difficulty, and get AI-powered suggestions.",
		LongHelp: `Keyword intelligence commands powered by ASO Maniac.

Subcommands:
  analyze    Score keyword popularity, difficulty, and top-ranking apps
  recommend  Get AI-powered keyword suggestions from a seed
  batch      Analyze multiple keywords across multiple storefronts at once`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			keywordsAnalyzeCommand(),
			keywordsRecommendCommand(),
			keywordsBatchCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing subcommand. Run 'aso keywords --help' for usage")
			}
			return fmt.Errorf("unknown subcommand %q. Run 'aso keywords --help' for usage", args[0])
		},
	}
}

func keywordsAnalyzeCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso keywords analyze", flag.ContinueOnError)
	storefront := fs.String("storefront", "US", "App Store storefront code")
	fields := fs.String("fields", "", "Comma-separated fields: popularity,difficulty,topApps,relatedSearches")

	return &ffcli.Command{
		Name:       "analyze",
		ShortUsage: "aso keywords analyze <keyword> [<keyword>...] [flags]",
		ShortHelp:  "Score keyword popularity, difficulty, and top-ranking apps.",
		LongHelp: `Analyze one or more keywords for a given storefront. Returns popularity
score (0-100), difficulty score, competition data, and top-ranking apps.

Examples:
  aso keywords analyze "photo editor"
  aso keywords analyze camera photo --storefront GB
  aso keywords analyze vpn --fields popularity,difficulty`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			keywords := resolveArgs(fs, args, true)
			if len(keywords) == 0 {
				return fmt.Errorf("at least one keyword is required")
			}
			var fieldSlice []string
			if *fields != "" {
				fieldSlice = strings.Split(*fields, ",")
			}
			return runKeywordsAnalyze(ctx, asomaniac.DefaultConfigPath(), keywords, *storefront, fieldSlice, os.Stdout)
		},
	}
}

func runKeywordsAnalyze(ctx context.Context, configPath string, keywords []string, storefront string, fields []string, w io.Writer) error {
	client, err := requireAuth(configPath)
	if err != nil {
		return err
	}

	results, err := client.AnalyzeKeywords(ctx, keywords, storefront, fields)
	if err != nil {
		return fmt.Errorf("analyze keywords: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func keywordsRecommendCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso keywords recommend", flag.ContinueOnError)
	storefront := fs.String("storefront", "US", "App Store storefront code")
	limit := fs.Int("limit", 50, "Maximum number of recommendations")

	return &ffcli.Command{
		Name:       "recommend",
		ShortUsage: "aso keywords recommend <seed> [flags]",
		ShortHelp:  "Get AI-powered keyword suggestions from a seed.",
		LongHelp: `Generate keyword suggestions based on a seed keyword. Returns related
keywords ranked by popularity and difficulty.

Examples:
  aso keywords recommend "photo editor"
  aso keywords recommend camera --storefront GB --limit 25`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			positional := resolveArgs(fs, args, false)
			if len(positional) == 0 {
				return fmt.Errorf("seed keyword is required")
			}
			return runKeywordsRecommend(ctx, asomaniac.DefaultConfigPath(), positional[0], *storefront, *limit, os.Stdout)
		},
	}
}

func runKeywordsRecommend(ctx context.Context, configPath string, seed, storefront string, limit int, w io.Writer) error {
	client, err := requireAuth(configPath)
	if err != nil {
		return err
	}

	results, err := client.GetRecommendations(ctx, seed, storefront, limit)
	if err != nil {
		return fmt.Errorf("get recommendations: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func keywordsBatchCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso keywords batch", flag.ContinueOnError)
	storefronts := fs.String("storefronts", "US", "Comma-separated storefront codes")

	return &ffcli.Command{
		Name:       "batch",
		ShortUsage: "aso keywords batch <keyword> [<keyword>...] --storefronts US,GB,DE",
		ShortHelp:  "Analyze multiple keywords across multiple storefronts at once.",
		LongHelp: `Batch-analyze keywords across one or more storefronts in a single request.
More efficient than calling analyze repeatedly.

Examples:
  aso keywords batch camera photo vpn
  aso keywords batch "photo editor" "video editor" --storefronts US,GB,DE`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			keywords := resolveArgs(fs, args, true)
			if len(keywords) == 0 {
				return fmt.Errorf("at least one keyword is required")
			}
			sfList := strings.Split(*storefronts, ",")
			return runKeywordsBatch(ctx, asomaniac.DefaultConfigPath(), keywords, sfList, os.Stdout)
		},
	}
}

func runKeywordsBatch(ctx context.Context, configPath string, keywords, storefronts []string, w io.Writer) error {
	client, err := requireAuth(configPath)
	if err != nil {
		return err
	}

	result, err := client.BatchAnalyze(ctx, keywords, storefronts)
	if err != nil {
		return fmt.Errorf("batch analyze: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
