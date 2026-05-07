package docs

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// DocsListCommand returns the docs list subcommand.
func DocsListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("docs list", flag.ExitOnError)
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "aso docs list [flags]",
		ShortHelp:  "List available embedded documentation guides.",
		LongHelp: `List available embedded documentation guides.

Examples:
  aso docs list
  aso docs list --output table`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			_ = ctx
			if len(args) > 0 {
				fmt.Fprintln(os.Stderr, "Error: docs list does not accept positional arguments")
				return flag.ErrHelp
			}
			headers := []string{"slug", "description"}
			rows := guideRows()
			if err := shared.PrintOutputWithRenderers(
				listGuideSummaries(),
				*output.Output,
				*output.Pretty,
				func() error { return renderTable(os.Stdout, headers, rows) },
				func() error { return renderMarkdown(os.Stdout, headers, rows) },
			); err != nil {
				return fmt.Errorf("docs list: %w", err)
			}
			return nil
		},
	}
}

func renderTable(w *os.File, headers []string, rows [][]string) error {
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	return tw.Flush()
}

func renderMarkdown(w *os.File, headers []string, rows [][]string) error {
	fmt.Fprintln(w, "| "+strings.Join(headers, " | ")+" |")
	sep := make([]string, len(headers))
	for i := range sep {
		sep[i] = "---"
	}
	fmt.Fprintln(w, "| "+strings.Join(sep, " | ")+" |")
	for _, row := range rows {
		fmt.Fprintln(w, "| "+strings.Join(row, " | ")+" |")
	}
	return nil
}
