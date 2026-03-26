package apps

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// RemovedAppInfoCommand preserves the legacy app-info surface as a deprecated alias.
func RemovedAppInfoCommand() *ffcli.Command {
	fs := flag.NewFlagSet("app-info", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "app-info",
		ShortUsage: "aso app-info <subcommand> [flags]",
		ShortHelp:  "DEPRECATED: use `aso apps info ...`.",
		LongHelp: `Deprecated compatibility alias for the new app-scoped info commands.

- ` + "`aso apps info list --app \"APP_ID\"`" + ` to inspect app info records
- ` + "`aso apps info view --app \"APP_ID\"`" + ` to read metadata/localizations
- ` + "`aso apps info edit --app \"APP_ID\" --locale \"en-US\" --whats-new \"Bug fixes\"`" + ` to update metadata

This alias still works during the transition, but new usage should move to ` + "`aso apps info ...`" + `.`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			deprecatedAppInfoGetAliasCommand(),
			deprecatedAppInfoSetAliasCommand(),
			deprecatedAppInfoRelationshipsAliasCommand(),
			deprecatedAppInfoTerritoryAgeRatingsAliasCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			fmt.Fprintf(os.Stderr, "Warning: `aso app-info` is deprecated. Use `%s`.\n", removedAppInfoSuggestion(args))
			return flag.ErrHelp
		},
	}
}

// RemovedAppInfosCommand preserves the legacy app-infos surface as a deprecated alias.
func RemovedAppInfosCommand() *ffcli.Command {
	fs := flag.NewFlagSet("app-infos", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "app-infos",
		ShortUsage: "aso app-infos <subcommand> [flags]",
		ShortHelp:  "DEPRECATED: use `aso apps info list`.",
		LongHelp: `Deprecated compatibility alias for ` + "`aso apps info list`" + `.

Use ` + "`aso apps info list --app \"APP_ID\"`" + ` to inspect app info records for an app.`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			deprecatedAppInfosListAliasCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			fmt.Fprintln(os.Stderr, `Warning: `+"`aso app-infos`"+` is deprecated. Use `+"`aso apps info list --app \"APP_ID\"`"+`.`)
			return flag.ErrHelp
		},
	}
}

func deprecatedAppInfoGetAliasCommand() *ffcli.Command {
	return deprecatedAliasLeafCommand(
		AppsInfoViewCommand(),
		"get",
		"aso app-info get [flags]",
		"aso apps info view",
		"Warning: `aso app-info get` is deprecated. Use `aso apps info view`.",
	)
}

func deprecatedAppInfoSetAliasCommand() *ffcli.Command {
	return deprecatedAliasLeafCommand(
		AppsInfoEditCommand(),
		"set",
		"aso app-info set [flags]",
		"aso apps info edit",
		"Warning: `aso app-info set` is deprecated. Use `aso apps info edit`.",
	)
}

func deprecatedAppInfosListAliasCommand() *ffcli.Command {
	return deprecatedAliasLeafCommand(
		AppsInfoListCommand(),
		"list",
		"aso app-infos list [flags]",
		"aso apps info list",
		"Warning: `aso app-infos list` is deprecated. Use `aso apps info list`.",
	)
}

func deprecatedAppInfoRelationshipsAliasCommand() *ffcli.Command {
	cmd := AppsInfoRelationshipsCommand()
	cmd.Name = "relationships"
	cmd.ShortUsage = "aso app-info relationships <subcommand> [flags]"
	cmd.ShortHelp = "DEPRECATED: use `aso apps info relationships ...`."
	cmd.LongHelp = "Deprecated compatibility alias for `aso apps info relationships ...`."
	cmd.UsageFunc = shared.DefaultUsageFunc
	for i, sub := range cmd.Subcommands {
		if sub == nil {
			continue
		}
		cmd.Subcommands[i] = deprecatedAliasLeafCommand(
			sub,
			sub.Name,
			fmt.Sprintf("aso app-info relationships %s [flags]", sub.Name),
			fmt.Sprintf("aso apps info relationships %s", sub.Name),
			fmt.Sprintf("Warning: `aso app-info relationships %s` is deprecated. Use `aso apps info relationships %s`.", sub.Name, sub.Name),
		)
	}
	return cmd
}

func deprecatedAppInfoTerritoryAgeRatingsAliasCommand() *ffcli.Command {
	cmd := AppsInfoTerritoryAgeRatingsCommand()
	cmd.Name = "territory-age-ratings"
	cmd.ShortUsage = "aso app-info territory-age-ratings <subcommand> [flags]"
	cmd.ShortHelp = "DEPRECATED: use `aso apps info territory-age-ratings ...`."
	cmd.LongHelp = "Deprecated compatibility alias for `aso apps info territory-age-ratings ...`."
	cmd.UsageFunc = shared.DefaultUsageFunc
	for i, sub := range cmd.Subcommands {
		if sub == nil {
			continue
		}
		cmd.Subcommands[i] = deprecatedAliasLeafCommand(
			sub,
			sub.Name,
			fmt.Sprintf("aso app-info territory-age-ratings %s [flags]", sub.Name),
			fmt.Sprintf("aso apps info territory-age-ratings %s", sub.Name),
			fmt.Sprintf("Warning: `aso app-info territory-age-ratings %s` is deprecated. Use `aso apps info territory-age-ratings %s`.", sub.Name, sub.Name),
		)
	}
	return cmd
}

func deprecatedAliasLeafCommand(cmd *ffcli.Command, name, shortUsage, newCommand, warning string) *ffcli.Command {
	clone := *cmd
	clone.Name = name
	clone.ShortUsage = shortUsage
	clone.ShortHelp = fmt.Sprintf("DEPRECATED: use `%s`.", newCommand)
	clone.LongHelp = fmt.Sprintf("Deprecated compatibility alias for `%s`.", newCommand)
	clone.UsageFunc = shared.DefaultUsageFunc
	origExec := cmd.Exec
	clone.Exec = func(ctx context.Context, args []string) error {
		fmt.Fprintln(os.Stderr, warning)
		return origExec(ctx, args)
	}
	return &clone
}

func removedAppInfoSuggestion(args []string) string {
	if len(args) == 0 {
		return "aso apps info"
	}

	switch strings.TrimSpace(args[0]) {
	case "list":
		return `aso apps info list --app "APP_ID"`
	case "get", "view":
		return `aso apps info view --app "APP_ID"`
	case "set", "edit":
		return `aso apps info edit --app "APP_ID" --locale "en-US" --whats-new "Bug fixes"`
	case "relationships":
		if len(args) > 1 && strings.TrimSpace(args[1]) != "" {
			return fmt.Sprintf(`aso apps info relationships %s --app "APP_ID"`, strings.TrimSpace(args[1]))
		}
		return `aso apps info relationships primary-category --app "APP_ID"`
	case "territory-age-ratings":
		return `aso apps info territory-age-ratings list --app "APP_ID"`
	default:
		return "aso apps info"
	}
}
