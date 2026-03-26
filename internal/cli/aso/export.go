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

// ExportCommand returns the "export" subcommand.
func ExportCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso export", flag.ExitOnError)
	dataType := fs.String("type", "", "Data type to export: rankings, keywords, or apps (required)")
	format := fs.String("format", "json", "Output format: json or csv")

	return &ffcli.Command{
		Name:       "export",
		ShortUsage: "aso export --type rankings [--format csv]",
		ShortHelp:  "Export keyword data in CSV, JSON, or TSV format.",
		LongHelp: `Export your ASO data for external analysis.

Examples:
  aso export --type rankings --format csv
  aso export --type keywords
  aso export --type apps --format json`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *dataType == "" {
				return fmt.Errorf("--type is required (rankings, keywords, or apps)")
			}
			switch *dataType {
			case "rankings", "keywords", "apps":
			default:
				return fmt.Errorf("invalid --type %q: must be rankings, keywords, or apps", *dataType)
			}
			switch *format {
			case "json", "csv":
			default:
				return fmt.Errorf("invalid --format %q: must be json or csv", *format)
			}
			return runExport(ctx, asomaniac.DefaultConfigPath(), *format, *dataType, os.Stdout)
		},
	}
}

func runExport(ctx context.Context, configPath, format, dataType string, w io.Writer) error {
	cfg, err := asomaniac.ReadConfig(configPath)
	if err != nil {
		return fmt.Errorf("not logged in. Run 'aso login' to authenticate")
	}
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not logged in. Run 'aso login' to authenticate")
	}

	client := asomaniac.NewClientFromConfig(cfg)
	result, err := client.Export(ctx, format, dataType, nil)
	if err != nil {
		return fmt.Errorf("export: %w", err)
	}

	// For CSV exports with inline data, print raw data directly.
	if result.Data != "" {
		_, err := fmt.Fprint(w, result.Data)
		return err
	}

	// Otherwise JSON-encode the result metadata.
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
