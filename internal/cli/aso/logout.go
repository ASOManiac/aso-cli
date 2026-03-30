package aso

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/asomaniac"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// LogoutCommand returns the "logout" subcommand that removes stored credentials.
func LogoutCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso auth maniac logout", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "logout",
		ShortUsage: "aso auth maniac logout",
		ShortHelp:  "Remove stored ASO Maniac credentials.",
		LongHelp:   `Deletes the API key stored at ~/.asomaniac/config.json.`,
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return runLogout(asomaniac.DefaultConfigPath(), os.Stdout)
		},
	}
}

func runLogout(configPath string, w io.Writer) error {
	err := asomaniac.DeleteConfig(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(w, "Not logged in (no credentials found).")
			return nil
		}
		return fmt.Errorf("remove credentials: %w", err)
	}

	fmt.Fprintln(w, "Logged out. Credentials removed.")
	return nil
}
