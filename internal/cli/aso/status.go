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
	fs := flag.NewFlagSet("aso auth maniac status", flag.ContinueOnError)
	return &ffcli.Command{
		Name:       "status",
		ShortUsage: "aso auth maniac status",
		ShortHelp:  "Check connection and authentication health.",
		LongHelp: `Verify that your ASO Maniac credentials are valid and the API is reachable.

Example:
  aso auth maniac status`,
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

	// Show whether a key exists in the config file.
	fileCfg, fileErr := asomaniac.ReadConfig(configPath)
	if fileErr != nil {
		fmt.Fprintf(w, "Stored:  no config file found\n")
	} else if fileCfg.APIKey == "" {
		fmt.Fprintf(w, "Stored:  config file exists, no key stored\n")
	} else {
		fmt.Fprintf(w, "Stored:  key present in config file\n")
	}

	// Show whether env var is set.
	if os.Getenv(asomaniac.EnvAPIKey) != "" {
		fmt.Fprintf(w, "Env:     %s is set (takes priority)\n", asomaniac.EnvAPIKey)
	} else {
		fmt.Fprintf(w, "Env:     %s is not set\n", asomaniac.EnvAPIKey)
	}

	if !resolved.IsAuthenticated() {
		fmt.Fprintf(w, "Auth:    not logged in\n")
		fmt.Fprintf(w, "\nRun 'aso auth maniac login' to authenticate.\n")
		return fmt.Errorf("not authenticated")
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
		fmt.Fprintf(w, "\nRun 'aso auth maniac login' to re-authenticate.\n")
		return fmt.Errorf("authentication invalid")
	}

	name := profile.Email
	if profile.Name != nil && *profile.Name != "" {
		name = *profile.Name
	}
	fmt.Fprintf(w, "Auth:    valid (%s, %s plan)\n", name, profile.Plan)

	return nil
}
