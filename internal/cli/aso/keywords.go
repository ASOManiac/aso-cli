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
	fs := flag.NewFlagSet("aso keywords", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "keywords",
		ShortUsage: "aso keywords <subcommand> [flags]",
		ShortHelp:  "Analyze, suggest, and batch-process keywords.",
		LongHelp: `Keyword intelligence commands powered by ASO Maniac.

Subcommands:
  analyze    Analyze keyword popularity, difficulty, and competition
  recommend  Get keyword recommendations from a seed keyword
  batch      Analyze multiple keywords across multiple storefronts`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			keywordsAnalyzeCommand(),
			keywordsRecommendCommand(),
			keywordsBatchCommand(),
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

func keywordsAnalyzeCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso keywords analyze", flag.ExitOnError)
	storefront := fs.String("storefront", "US", "App Store storefront code")
	fields := fs.String("fields", "", "Comma-separated fields: popularity,difficulty,topApps,relatedSearches")

	return &ffcli.Command{
		Name:       "analyze",
		ShortUsage: "aso keywords analyze <keyword> [<keyword>...] [flags]",
		ShortHelp:  "Analyze keyword popularity, difficulty, and competition.",
		LongHelp: `Analyze one or more keywords for a given storefront.

Examples:
  aso keywords analyze "photo editor"
  aso keywords analyze camera photo --storefront GB
  aso keywords analyze vpn --fields popularity,difficulty`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("at least one keyword is required")
			}
			var fieldSlice []string
			if *fields != "" {
				fieldSlice = strings.Split(*fields, ",")
			}
			return runKeywordsAnalyze(ctx, asomaniac.DefaultConfigPath(), args, *storefront, fieldSlice, os.Stdout)
		},
	}
}

func runKeywordsAnalyze(ctx context.Context, configPath string, keywords []string, storefront string, fields []string, w io.Writer) error {
	cfg, err := asomaniac.ReadConfig(configPath)
	if err != nil {
		return fmt.Errorf("not logged in. Run 'aso login' to authenticate")
	}
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not logged in. Run 'aso login' to authenticate")
	}

	client := asomaniac.NewClientFromConfig(cfg)
	results, err := client.AnalyzeKeywords(ctx, keywords, storefront, fields)
	if err != nil {
		return fmt.Errorf("analyze keywords: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func keywordsRecommendCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso keywords recommend", flag.ExitOnError)
	storefront := fs.String("storefront", "US", "App Store storefront code")
	limit := fs.Int("limit", 50, "Maximum number of recommendations")

	return &ffcli.Command{
		Name:       "recommend",
		ShortUsage: "aso keywords recommend <seed> [flags]",
		ShortHelp:  "Get keyword recommendations from a seed keyword.",
		LongHelp: `Get AI-powered keyword recommendations based on a seed keyword.

Examples:
  aso keywords recommend "photo editor"
  aso keywords recommend camera --storefront GB --limit 25`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("seed keyword is required")
			}
			return runKeywordsRecommend(ctx, asomaniac.DefaultConfigPath(), args[0], *storefront, *limit, os.Stdout)
		},
	}
}

func runKeywordsRecommend(ctx context.Context, configPath string, seed, storefront string, limit int, w io.Writer) error {
	cfg, err := asomaniac.ReadConfig(configPath)
	if err != nil {
		return fmt.Errorf("not logged in. Run 'aso login' to authenticate")
	}
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not logged in. Run 'aso login' to authenticate")
	}

	client := asomaniac.NewClientFromConfig(cfg)
	results, err := client.GetRecommendations(ctx, seed, storefront, limit)
	if err != nil {
		return fmt.Errorf("get recommendations: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func keywordsBatchCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso keywords batch", flag.ExitOnError)
	storefronts := fs.String("storefronts", "US", "Comma-separated storefront codes")

	return &ffcli.Command{
		Name:       "batch",
		ShortUsage: "aso keywords batch <keyword> [<keyword>...] --storefronts US,GB,DE",
		ShortHelp:  "Analyze multiple keywords across multiple storefronts.",
		LongHelp: `Batch-analyze keywords across one or more storefronts in a single request.

Examples:
  aso keywords batch camera photo vpn
  aso keywords batch "photo editor" "video editor" --storefronts US,GB,DE`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("at least one keyword is required")
			}
			sfList := strings.Split(*storefronts, ",")
			return runKeywordsBatch(ctx, asomaniac.DefaultConfigPath(), args, sfList, os.Stdout)
		},
	}
}

func runKeywordsBatch(ctx context.Context, configPath string, keywords, storefronts []string, w io.Writer) error {
	cfg, err := asomaniac.ReadConfig(configPath)
	if err != nil {
		return fmt.Errorf("not logged in. Run 'aso login' to authenticate")
	}
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not logged in. Run 'aso login' to authenticate")
	}

	client := asomaniac.NewClientFromConfig(cfg)
	result, err := client.BatchAnalyze(ctx, keywords, storefronts)
	if err != nil {
		return fmt.Errorf("batch analyze: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
