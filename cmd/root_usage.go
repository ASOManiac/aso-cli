package cmd

import (
	"flag"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// RootUsageFunc renders the root help screen.
func RootUsageFunc(c *ffcli.Command) string {
	var b strings.Builder

	shortHelp := strings.TrimSpace(c.ShortHelp)
	longHelp := strings.TrimSpace(c.LongHelp)
	if shortHelp == "" && longHelp != "" {
		shortHelp = longHelp
		longHelp = ""
	}

	if shortHelp != "" {
		b.WriteString(shared.Bold("DESCRIPTION"))
		b.WriteString("\n  ")
		b.WriteString(shortHelp)
		b.WriteString("\n\n")
	}

	usage := strings.TrimSpace(c.ShortUsage)
	if usage == "" {
		usage = strings.TrimSpace(c.Name)
	}
	if usage != "" {
		b.WriteString(shared.Bold("USAGE"))
		b.WriteString("\n  ")
		b.WriteString(usage)
		b.WriteString("\n\n")
	}

	if longHelp != "" {
		if shortHelp != "" && strings.HasPrefix(longHelp, shortHelp) {
			longHelp = strings.TrimSpace(strings.TrimPrefix(longHelp, shortHelp))
		}
		if longHelp != "" {
			b.WriteString(longHelp)
			b.WriteString("\n\n")
		}
	}

	if len(c.Subcommands) > 0 {
		b.WriteString(shared.Bold("SUBCOMMANDS"))
		b.WriteString("\n")
		tw := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
		for _, sub := range c.Subcommands {
			fmt.Fprintf(tw, "  %s\t%s\n", sub.Name, sub.ShortHelp)
		}
		tw.Flush()
		b.WriteString("\n")
	}

	if c.FlagSet != nil {
		var flags []*flag.Flag
		c.FlagSet.VisitAll(func(f *flag.Flag) {
			flags = append(flags, f)
		})
		if len(flags) > 0 {
			b.WriteString(shared.Bold("FLAGS"))
			b.WriteString("\n")
			tw := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
			for _, f := range flags {
				if f.DefValue != "" {
					fmt.Fprintf(tw, "  --%s\t%s (default: %s)\n", f.Name, f.Usage, f.DefValue)
				} else {
					fmt.Fprintf(tw, "  --%s\t%s\n", f.Name, f.Usage)
				}
			}
			tw.Flush()
			b.WriteString("\n")
		}
	}

	return b.String()
}
