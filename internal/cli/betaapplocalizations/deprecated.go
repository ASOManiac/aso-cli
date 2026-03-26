package betaapplocalizations

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

const betaAppLocalizationsCanonicalRoot = "aso testflight app-localizations"

// DeprecatedBetaAppLocalizationsCommand returns the hidden compatibility root.
func DeprecatedBetaAppLocalizationsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("beta-app-localizations", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "beta-app-localizations",
		ShortUsage: betaAppLocalizationsCanonicalRoot + " <subcommand> [flags]",
		ShortHelp:  "DEPRECATED: use `aso testflight app-localizations ...`.",
		LongHelp: `Deprecated compatibility alias for TestFlight app localizations.

Canonical workflows now live under ` + "`aso testflight app-localizations ...`" + `.

Examples:
  aso testflight app-localizations list --app "APP_ID"
  aso testflight app-localizations create --app "APP_ID" --locale "en-US" --description "Welcome testers"`,
		FlagSet:   fs,
		UsageFunc: shared.DeprecatedUsageFunc,
		Subcommands: []*ffcli.Command{
			deprecatedBetaAppLocalizationsLeafCommand(
				BetaAppLocalizationsListCommand(),
				betaAppLocalizationsCanonicalRoot+" list [flags]",
				betaAppLocalizationsCanonicalRoot+" list",
				"Warning: `aso beta-app-localizations list` is deprecated. Use `aso testflight app-localizations list`.",
			),
			deprecatedBetaAppLocalizationsLeafCommand(
				BetaAppLocalizationsGetCommand(),
				betaAppLocalizationsCanonicalRoot+" view --id \"LOCALIZATION_ID\"",
				betaAppLocalizationsCanonicalRoot+" view",
				"Warning: `aso beta-app-localizations get` is deprecated. Use `aso testflight app-localizations view`.",
			),
			deprecatedBetaAppLocalizationsAppCommand(),
			deprecatedBetaAppLocalizationsLeafCommand(
				BetaAppLocalizationsCreateCommand(),
				betaAppLocalizationsCanonicalRoot+" create [flags]",
				betaAppLocalizationsCanonicalRoot+" create",
				"Warning: `aso beta-app-localizations create` is deprecated. Use `aso testflight app-localizations create`.",
			),
			deprecatedBetaAppLocalizationsLeafCommand(
				BetaAppLocalizationsUpdateCommand(),
				betaAppLocalizationsCanonicalRoot+" update [flags]",
				betaAppLocalizationsCanonicalRoot+" update",
				"Warning: `aso beta-app-localizations update` is deprecated. Use `aso testflight app-localizations update`.",
			),
			deprecatedBetaAppLocalizationsLeafCommand(
				BetaAppLocalizationsDeleteCommand(),
				betaAppLocalizationsCanonicalRoot+" delete --id \"LOCALIZATION_ID\" --confirm",
				betaAppLocalizationsCanonicalRoot+" delete",
				"Warning: `aso beta-app-localizations delete` is deprecated. Use `aso testflight app-localizations delete`.",
			),
		},
		Exec: func(ctx context.Context, args []string) error {
			fmt.Fprintln(os.Stderr, "Warning: `aso beta-app-localizations` is deprecated. Use `aso testflight app-localizations ...`.")
			return flag.ErrHelp
		},
	}
}

func deprecatedBetaAppLocalizationsAppCommand() *ffcli.Command {
	fs := flag.NewFlagSet("app", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "app",
		ShortUsage: betaAppLocalizationsCanonicalRoot + " app <subcommand> [flags]",
		ShortHelp:  "Compatibility alias: use `aso testflight app-localizations app view`.",
		LongHelp:   "Compatibility alias: use `aso testflight app-localizations app view --id LOCALIZATION_ID`.",
		FlagSet:    fs,
		UsageFunc:  shared.DeprecatedUsageFunc,
		Subcommands: []*ffcli.Command{
			deprecatedBetaAppLocalizationsLeafCommand(
				BetaAppLocalizationsAppGetCommand(),
				betaAppLocalizationsCanonicalRoot+" app view --id \"LOCALIZATION_ID\"",
				betaAppLocalizationsCanonicalRoot+" app view",
				"Warning: `aso beta-app-localizations app get` is deprecated. Use `aso testflight app-localizations app view`.",
			),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

func deprecatedBetaAppLocalizationsLeafCommand(cmd *ffcli.Command, shortUsage, newCommand, warning string) *ffcli.Command {
	if cmd == nil {
		return nil
	}

	clone := *cmd
	clone.ShortUsage = shortUsage
	clone.ShortHelp = fmt.Sprintf("Compatibility alias: use `%s`.", newCommand)
	clone.LongHelp = fmt.Sprintf("Compatibility alias: use `%s`.", shortUsage)
	clone.UsageFunc = shared.DeprecatedUsageFunc

	origExec := cmd.Exec
	clone.Exec = func(ctx context.Context, args []string) error {
		fmt.Fprintln(os.Stderr, warning)
		return origExec(ctx, args)
	}

	return &clone
}
