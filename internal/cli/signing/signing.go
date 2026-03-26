package signing

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// SigningCommand returns the signing command with subcommands.
func SigningCommand() *ffcli.Command {
	fs := flag.NewFlagSet("signing", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "signing",
		ShortUsage: "aso signing <subcommand> [flags]",
		ShortHelp:  "Manage signing certificates and profiles.",
		LongHelp: `Manage signing assets in App Store Connect.

Examples:
  aso signing fetch --bundle-id com.example.app --profile-type IOS_APP_STORE --output ./signing
  aso signing sync push --bundle-id com.example.app --profile-type IOS_APP_STORE --repo git@github.com:team/certs.git
  aso signing sync pull --repo git@github.com:team/certs.git --output-dir ./signing`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			SigningFetchCommand(),
			SigningSyncCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
