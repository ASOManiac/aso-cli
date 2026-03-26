package aso

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/asomaniac"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// WhoamiCommand returns the "whoami" subcommand that displays the current user.
func WhoamiCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso aso whoami", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "whoami",
		ShortUsage: "aso aso whoami",
		ShortHelp:  "Display the current authenticated ASO Maniac user.",
		LongHelp:   `Shows your email, plan, and API usage statistics.`,
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return runWhoami(ctx, asomaniac.DefaultConfigPath(), os.Stdout)
		},
	}
}

func runWhoami(ctx context.Context, configPath string, w io.Writer) error {
	cfg, err := asomaniac.ReadConfig(configPath)
	if err != nil {
		return fmt.Errorf("not logged in. Run 'aso aso login' to authenticate")
	}

	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not logged in. Run 'aso aso login' to authenticate")
	}

	client := asomaniac.NewClientFromConfig(cfg)

	profile, err := client.GetProfile(ctx)
	if err != nil {
		return fmt.Errorf("fetch profile: %w", err)
	}

	usage, err := client.GetUsage(ctx)
	if err != nil {
		return fmt.Errorf("fetch usage: %w", err)
	}

	name := profile.Email
	if profile.Name != nil && *profile.Name != "" {
		name = *profile.Name
	}

	fmt.Fprintf(w, "Email:   %s\n", profile.Email)
	if profile.Name != nil && *profile.Name != "" {
		fmt.Fprintf(w, "Name:    %s\n", name)
	}
	fmt.Fprintf(w, "Plan:    %s\n", profile.Plan)
	fmt.Fprintf(w, "Member:  %s\n", profile.CreatedAt)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "API calls today:    %d\n", usage.APICalls.Today)
	fmt.Fprintf(w, "API calls (month):  %d / %d\n", usage.APICalls.ThisMonth, usage.APICalls.Limit)
	fmt.Fprintf(w, "Tracked apps:       %d / %d\n", usage.TrackedApps.Current, usage.TrackedApps.Limit)
	fmt.Fprintf(w, "Tracked keywords:   %d / %d\n", usage.TrackedKeywords.Current, usage.TrackedKeywords.Limit)

	return nil
}
