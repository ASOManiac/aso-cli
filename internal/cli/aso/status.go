package aso

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/asomaniac"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// StatusCommand returns the "status" subcommand that checks auth and API health.
func StatusCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso maniac status", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "status",
		ShortUsage: "aso maniac status",
		ShortHelp:  "Check connection and authentication health.",
		LongHelp: `Verify that your ASO Maniac credentials are valid and the API is reachable.

Example:
  aso maniac status`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return runStatus(ctx, asomaniac.DefaultConfigPath(), os.Stdout)
		},
	}
}

func runStatus(ctx context.Context, configPath string, w io.Writer) error {
	resolved := asomaniac.ResolveConfig(configPath)

	fmt.Fprintf(w, "Config:  %s\n", configPath)

	if !resolved.IsAuthenticated() {
		fmt.Fprintf(w, "Auth:    not logged in\n")
		fmt.Fprintf(w, "\nRun 'aso maniac login' to authenticate.\n")
		return nil
	}

	keyPreview := resolved.APIKey
	if len(keyPreview) > 12 {
		keyPreview = keyPreview[:8] + "..." + keyPreview[len(keyPreview)-4:]
	}
	fmt.Fprintf(w, "API key: %s (source: %s)\n", keyPreview, resolved.Source)
	fmt.Fprintf(w, "Base:    %s\n", resolved.BaseURL)

	// Ping the API.
	httpClient := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, resolved.BaseURL, nil)
	if err != nil {
		fmt.Fprintf(w, "API:     unreachable (%v)\n", err)
		return nil
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(w, "API:     unreachable (%v)\n", err)
		return nil
	}
	resp.Body.Close()
	fmt.Fprintf(w, "API:     reachable (HTTP %d)\n", resp.StatusCode)

	// Validate auth.
	client := asomaniac.NewClientFromConfig(resolved.Config)
	profile, err := client.GetProfile(ctx)
	if err != nil {
		fmt.Fprintf(w, "Auth:    invalid (%v)\n", err)
		fmt.Fprintf(w, "\nRun 'aso maniac login' to re-authenticate.\n")
		return nil
	}

	name := profile.Email
	if profile.Name != nil && *profile.Name != "" {
		name = *profile.Name
	}
	fmt.Fprintf(w, "Auth:    valid (%s, %s plan)\n", name, profile.Plan)

	return nil
}
