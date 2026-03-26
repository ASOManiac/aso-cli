package testflight

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/betaapplocalizations"
	"github.com/ASOManiac/aso-cli/internal/cli/prerelease"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

type textReplacement struct {
	old string
	new string
}

type rewrittenCommandError struct {
	message string
	err     error
}

func (e rewrittenCommandError) Error() string {
	return e.message
}

func (e rewrittenCommandError) Unwrap() error {
	return e.err
}

var hiddenTestFlightCommands sync.Map

func testflightVisibleUsageFunc(c *ffcli.Command) string {
	clone := *c
	if len(c.Subcommands) > 0 {
		visible := make([]*ffcli.Command, 0, len(c.Subcommands))
		for _, sub := range c.Subcommands {
			if !isHiddenTestFlightCommand(sub) {
				visible = append(visible, sub)
			}
		}
		clone.Subcommands = visible
	}
	return shared.DefaultUsageFunc(&clone)
}

func hideTestFlightCommand(cmd *ffcli.Command) *ffcli.Command {
	if cmd != nil {
		hiddenTestFlightCommands.Store(cmd, struct{}{})
	}
	return cmd
}

func isHiddenTestFlightCommand(cmd *ffcli.Command) bool {
	if cmd == nil {
		return false
	}
	_, ok := hiddenTestFlightCommands.Load(cmd)
	return ok
}

func rewriteCommandTree(cmd *ffcli.Command, oldRootPath, newRootPath string, nameRenames map[string]string, textReplacements []textReplacement) *ffcli.Command {
	if cmd == nil {
		return nil
	}

	pathReplacements := collectCommandPathReplacements(cmd, oldRootPath, newRootPath, nameRenames)
	replacements := append(pathReplacements, textReplacements...)
	sortTextReplacements(replacements)

	renameCommandNames(cmd, nameRenames)
	rewriteCommandStrings(cmd, replacements)
	rewriteCommandErrors(cmd, replacements)
	return cmd
}

func rewriteCommandPresentation(cmd *ffcli.Command, oldRootPath, newRootPath string, nameRenames map[string]string) *ffcli.Command {
	if cmd == nil {
		return nil
	}

	pathReplacements := collectCommandPathReplacements(cmd, oldRootPath, newRootPath, nameRenames)
	rewriteCommandStrings(cmd, pathReplacements)
	rewriteCommandErrors(cmd, pathReplacements)
	return cmd
}

func rewriteCommandStringsOnly(cmd *ffcli.Command, oldRootPath, newRootPath string, nameRenames map[string]string) *ffcli.Command {
	if cmd == nil {
		return nil
	}

	pathReplacements := collectCommandPathReplacements(cmd, oldRootPath, newRootPath, nameRenames)
	rewriteCommandStrings(cmd, pathReplacements)
	return cmd
}

func collectCommandPathReplacements(cmd *ffcli.Command, oldRootPath, newRootPath string, nameRenames map[string]string) []textReplacement {
	replacements := []textReplacement{}

	var walk func(current *ffcli.Command, oldPath, newPath string)
	walk = func(current *ffcli.Command, oldPath, newPath string) {
		replacements = append(replacements, textReplacement{old: oldPath, new: newPath})
		if strings.HasPrefix(oldPath, "aso testflight ") && strings.HasPrefix(newPath, "aso testflight ") {
			replacements = append(replacements, textReplacement{
				old: strings.TrimPrefix(oldPath, "aso testflight "),
				new: strings.TrimPrefix(newPath, "aso testflight "),
			})
		}

		for _, sub := range current.Subcommands {
			oldChildName := sub.Name
			newChildName := oldChildName
			if renamed, ok := nameRenames[oldChildName]; ok {
				newChildName = renamed
			}
			walk(sub, oldPath+" "+oldChildName, newPath+" "+newChildName)
		}
	}

	walk(cmd, oldRootPath, newRootPath)
	sortTextReplacements(replacements)
	return replacements
}

func sortTextReplacements(replacements []textReplacement) {
	sort.SliceStable(replacements, func(i, j int) bool {
		return len(replacements[i].old) > len(replacements[j].old)
	})
}

func renameCommandNames(cmd *ffcli.Command, nameRenames map[string]string) {
	if cmd == nil {
		return
	}

	if renamed, ok := nameRenames[cmd.Name]; ok {
		cmd.Name = renamed
		if cmd.FlagSet != nil {
			cmd.FlagSet.Init(renamed, cmd.FlagSet.ErrorHandling())
		}
	}
	for _, sub := range cmd.Subcommands {
		renameCommandNames(sub, nameRenames)
	}
}

func rewriteCommandStrings(cmd *ffcli.Command, replacements []textReplacement) {
	if cmd == nil {
		return
	}

	if cmd.ShortUsage != "" {
		cmd.ShortUsage = applyTextReplacements(cmd.ShortUsage, replacements)
	}
	if cmd.ShortHelp != "" {
		cmd.ShortHelp = applyTextReplacements(cmd.ShortHelp, replacements)
	}
	if cmd.LongHelp != "" {
		cmd.LongHelp = applyTextReplacements(cmd.LongHelp, replacements)
	}
	if cmd.FlagSet != nil {
		cmd.FlagSet.VisitAll(func(f *flag.Flag) {
			f.Usage = applyTextReplacements(f.Usage, replacements)
		})
	}
	for _, sub := range cmd.Subcommands {
		rewriteCommandStrings(sub, replacements)
	}
}

func rewriteCommandErrors(cmd *ffcli.Command, replacements []textReplacement) {
	if cmd == nil {
		return
	}

	if cmd.Exec != nil {
		originalExec := cmd.Exec
		cmd.Exec = func(ctx context.Context, args []string) error {
			err := originalExec(ctx, args)
			if err == nil || errors.Is(err, flag.ErrHelp) {
				return err
			}

			rewritten := applyTextReplacements(err.Error(), replacements)
			if rewritten == err.Error() {
				return err
			}
			return rewrittenCommandError{
				message: rewritten,
				err:     err,
			}
		}
	}

	for _, sub := range cmd.Subcommands {
		rewriteCommandErrors(sub, replacements)
	}
}

func applyTextReplacements(input string, replacements []textReplacement) string {
	output := input
	for _, replacement := range replacements {
		output = strings.ReplaceAll(output, replacement.old, replacement.new)
	}
	return output
}

func setUsageFuncRecursively(cmd *ffcli.Command, usageFunc func(*ffcli.Command) string) {
	if cmd == nil {
		return
	}
	cmd.UsageFunc = usageFunc
	for _, sub := range cmd.Subcommands {
		setUsageFuncRecursively(sub, usageFunc)
	}
}

func findSubcommand(cmd *ffcli.Command, name string) *ffcli.Command {
	if cmd == nil {
		return nil
	}
	for _, sub := range cmd.Subcommands {
		if sub.Name == name {
			return sub
		}
	}
	return nil
}

func appendDeprecatedLeafAlias(parent *ffcli.Command, oldName, newName, shortUsage, newCommand, warning string) {
	if parent == nil {
		return
	}
	if alias := shared.DeprecatedAliasLeafCommand(
		findSubcommand(parent, oldName),
		newName,
		shortUsage,
		newCommand,
		warning,
	); alias != nil {
		parent.Subcommands = append(parent.Subcommands, alias)
	}
}

func deprecatedAliasCommand(cmd *ffcli.Command, shortUsage, shortHelp, longHelp string) *ffcli.Command {
	if cmd == nil {
		return nil
	}
	cmd.ShortUsage = shortUsage
	cmd.ShortHelp = shortHelp
	cmd.LongHelp = longHelp
	cmd.UsageFunc = shared.DeprecatedUsageFunc
	return hideTestFlightCommand(cmd)
}

func markCommandTreeDeprecated(cmd *ffcli.Command) {
	if cmd == nil {
		return
	}

	usage := strings.TrimSpace(cmd.ShortUsage)
	if usage == "" {
		usage = strings.TrimSpace(cmd.Name)
	}
	if usage != "" {
		cmd.ShortHelp = fmt.Sprintf("Compatibility alias: use `%s`.", usage)
		cmd.LongHelp = fmt.Sprintf("Compatibility alias: use `%s`.", usage)
	}

	for _, sub := range cmd.Subcommands {
		markCommandTreeDeprecated(sub)
	}
}

func markDeprecatedSubcommands(cmd *ffcli.Command) {
	if cmd == nil {
		return
	}
	for _, sub := range cmd.Subcommands {
		markCommandTreeDeprecated(sub)
	}
}

func RemovedTestFlightAppsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("apps", flag.ExitOnError)

	return hideTestFlightCommand(&ffcli.Command{
		Name:       "apps",
		ShortUsage: "aso apps <subcommand> [flags]",
		ShortHelp:  "REMOVED: use `aso apps`.",
		LongHelp:   "Use `aso apps list` for collection lookup and `aso apps view --id APP_ID` for a single app.",
		FlagSet:    fs,
		UsageFunc:  shared.DeprecatedUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			suggestion := "aso apps"
			if len(args) > 0 {
				switch strings.TrimSpace(args[0]) {
				case "list":
					suggestion = "aso apps list"
				case "get", "view":
					suggestion = "aso apps view --id APP_ID"
				}
			}

			fmt.Fprintf(os.Stderr, "Error: `aso testflight apps` was removed. Use `%s` instead.\n", suggestion)
			return flag.ErrHelp
		},
	})
}

func TestFlightGroupsCommand() *ffcli.Command {
	cmd := rewriteCommandTree(
		BetaGroupsCommand(),
		"aso testflight beta-groups",
		"aso testflight groups",
		map[string]string{
			"beta-groups":               "groups",
			"beta-recruitment-criteria": "recruitment",
			"beta-recruitment-criterion-compatible-build-check": "compatibility",
			"relationships": "links",
			"get":           "view",
			"update":        "edit",
		},
		[]textReplacement{
			{old: "Manage TestFlight beta groups.", new: "Manage TestFlight groups."},
			{old: "Manage TestFlight beta groups", new: "Manage TestFlight groups"},
			{old: "List TestFlight beta groups", new: "List TestFlight groups"},
			{old: "beta testers", new: "testers"},
			{old: "Beta testers", new: "Testers"},
			{old: "beta tester", new: "tester"},
			{old: "Beta tester", new: "Tester"},
			{old: "compatible build status", new: "recruitment compatibility"},
			{old: "Compatible build status", new: "Recruitment compatibility"},
			{old: "beta recruitment criteria", new: "recruitment criteria"},
			{old: "Beta recruitment criteria", new: "Recruitment criteria"},
			{old: "beta recruitment criterion compatible build status", new: "compatible build status"},
			{old: "Beta recruitment criterion compatible build status", new: "Compatible build status"},
			{old: "beta groups", new: "groups"},
			{old: "Beta groups", new: "Groups"},
			{old: "beta group", new: "group"},
			{old: "Beta group", new: "Group"},
			{old: "Get ", new: "View "},
			{old: "get ", new: "view "},
			{old: "Update ", new: "Edit "},
			{old: "update ", new: "edit "},
		},
	)
	setUsageFuncRecursively(cmd, testflightVisibleUsageFunc)
	cmd.Subcommands = append(cmd.Subcommands, deprecatedTestFlightGroupsRelationshipsAliasCommand())
	if compatibilityCmd := findSubcommand(cmd, "compatibility"); compatibilityCmd != nil {
		compatibilityCmd.ShortHelp = "Check recruitment compatibility for a group."
		compatibilityCmd.LongHelp = `Check recruitment compatibility for a group.

Examples:
  aso testflight groups compatibility view --group-id "GROUP_ID"`
		if viewCmd := findSubcommand(compatibilityCmd, "view"); viewCmd != nil {
			viewCmd.ShortHelp = "View recruitment compatibility for a group."
			viewCmd.LongHelp = `View recruitment compatibility for a group.

Examples:
  aso testflight groups compatibility view --group-id "GROUP_ID"`
		}
	}
	return cmd
}

func TestFlightTestersCommand() *ffcli.Command {
	cmd := rewriteCommandTree(
		BetaTestersCommand(),
		"aso testflight beta-testers",
		"aso testflight testers",
		map[string]string{
			"beta-testers":  "testers",
			"beta-groups":   "groups",
			"relationships": "links",
			"get":           "view",
		},
		[]textReplacement{
			{old: "Manage TestFlight beta testers.", new: "Manage TestFlight testers."},
			{old: "Manage TestFlight beta testers", new: "Manage TestFlight testers"},
			{old: "List TestFlight beta testers", new: "List TestFlight testers"},
			{old: "beta groups", new: "groups"},
			{old: "Beta groups", new: "Groups"},
			{old: "beta group", new: "group"},
			{old: "Beta group", new: "Group"},
			{old: "beta testers", new: "testers"},
			{old: "Beta testers", new: "Testers"},
			{old: "beta tester", new: "tester"},
			{old: "Beta tester", new: "Tester"},
			{old: "Get ", new: "View "},
			{old: "get ", new: "view "},
		},
	)
	setUsageFuncRecursively(cmd, testflightVisibleUsageFunc)
	cmd.Subcommands = append(cmd.Subcommands, deprecatedTestFlightTestersRelationshipsAliasCommand())
	return cmd
}

func deprecatedTestFlightGroupsRelationshipsAliasCommand() *ffcli.Command {
	cmd := BetaGroupsRelationshipsCommand()
	if cmd == nil {
		return nil
	}

	cmd.ShortUsage = "aso testflight groups links <subcommand> [flags]"
	cmd.ShortHelp = "DEPRECATED: use `aso testflight groups links ...`."
	cmd.LongHelp = "Deprecated compatibility alias for `aso testflight groups links ...`."
	cmd.UsageFunc = shared.DeprecatedUsageFunc

	if viewCmd := shared.DeprecatedAliasLeafCommand(
		rewriteCommandTree(
			BetaGroupsRelationshipsGetCommand(),
			"aso testflight beta-groups relationships get",
			"aso testflight groups links view",
			map[string]string{"get": "view"},
			nil,
		),
		"view",
		"aso testflight groups links view --group-id \"GROUP_ID\" --type \"RELATIONSHIP\" [flags]",
		"aso testflight groups links view",
		"Warning: `aso testflight groups relationships view` is deprecated. Use `aso testflight groups links view`.",
	); viewCmd != nil {
		cmd.Subcommands = []*ffcli.Command{viewCmd}
	}

	return hideTestFlightCommand(cmd)
}

func deprecatedTestFlightTestersRelationshipsAliasCommand() *ffcli.Command {
	cmd := BetaTestersRelationshipsCommand()
	if cmd == nil {
		return nil
	}

	cmd.ShortUsage = "aso testflight testers links <subcommand> [flags]"
	cmd.ShortHelp = "DEPRECATED: use `aso testflight testers links ...`."
	cmd.LongHelp = "Deprecated compatibility alias for `aso testflight testers links ...`."
	cmd.UsageFunc = shared.DeprecatedUsageFunc

	if viewCmd := shared.DeprecatedAliasLeafCommand(
		rewriteCommandTree(
			BetaTestersRelationshipsGetCommand(),
			"aso testflight beta-testers relationships get",
			"aso testflight testers links view",
			map[string]string{"get": "view"},
			nil,
		),
		"view",
		"aso testflight testers links view --tester-id \"TESTER_ID\" --type \"RELATIONSHIP\" [flags]",
		"aso testflight testers links view",
		"Warning: `aso testflight testers relationships view` is deprecated. Use `aso testflight testers links view`.",
	); viewCmd != nil {
		cmd.Subcommands = []*ffcli.Command{viewCmd}
	}

	return hideTestFlightCommand(cmd)
}

func TestFlightAgreementsCommand() *ffcli.Command {
	cmd := rewriteCommandTree(
		BetaLicenseAgreementsCommand(),
		"aso testflight beta-license-agreements",
		"aso testflight agreements",
		map[string]string{
			"beta-license-agreements": "agreements",
			"get":                     "view",
			"update":                  "edit",
		},
		[]textReplacement{
			{old: "Manage TestFlight beta license agreements.", new: "Manage TestFlight agreements."},
			{old: "Manage TestFlight beta license agreements", new: "Manage TestFlight agreements"},
			{old: "Fields to include (betaLicenseAgreements), comma-separated", new: "Fields to include for agreements, comma-separated"},
			{old: "beta license agreements", new: "agreements"},
			{old: "Beta license agreements", new: "Agreements"},
			{old: "beta license agreement", new: "agreement"},
			{old: "Beta license agreement", new: "Agreement"},
			{old: "Get ", new: "View "},
			{old: "get ", new: "view "},
			{old: "Update ", new: "Edit "},
			{old: "update ", new: "edit "},
		},
	)
	cmd.UsageFunc = testflightVisibleUsageFunc
	if listCmd := findSubcommand(cmd, "list"); listCmd != nil {
		listCmd.ShortHelp = "List agreements."
	}
	if viewCmd := findSubcommand(cmd, "view"); viewCmd != nil {
		viewCmd.ShortHelp = "View an agreement by ID or app."
		viewCmd.LongHelp = `View an agreement by ID or app.

Examples:
  aso testflight agreements view --id "AGREEMENT_ID"
  aso testflight agreements view --app "APP_ID"`
	}
	if editCmd := findSubcommand(cmd, "edit"); editCmd != nil {
		editCmd.ShortHelp = "Edit an agreement."
		editCmd.LongHelp = `Edit an agreement.

Examples:
  aso testflight agreements edit --id "AGREEMENT_ID" --agreement-text "Updated terms"`
	}
	return cmd
}

func TestFlightNotificationsCommand() *ffcli.Command {
	cmd := rewriteCommandTree(
		BetaNotificationsCommand(),
		"aso testflight beta-notifications",
		"aso testflight notifications",
		map[string]string{
			"beta-notifications": "notifications",
			"create":             "send",
		},
		[]textReplacement{
			{old: "Send TestFlight beta build notifications.", new: "Send TestFlight notifications."},
			{old: "Send TestFlight beta build notifications", new: "Send TestFlight notifications"},
			{old: "beta notification", new: "notification"},
			{old: "Beta notification", new: "Notification"},
			{old: "Create ", new: "Send "},
			{old: "create ", new: "send "},
		},
	)
	cmd.UsageFunc = testflightVisibleUsageFunc
	return cmd
}

func TestFlightConfigCommand() *ffcli.Command {
	cmd := rewriteCommandTree(
		TestFlightSyncCommand(),
		"aso testflight sync",
		"aso testflight config",
		map[string]string{
			"sync": "config",
			"pull": "export",
		},
		[]textReplacement{
			{old: "Sync TestFlight configuration.", new: "Export TestFlight configuration."},
			{old: "Sync TestFlight configuration", new: "Export TestFlight configuration"},
			{old: "beta groups", new: "TestFlight groups"},
			{old: "beta group", new: "TestFlight group"},
			{old: "sync pull", new: "config export"},
			{old: "testflight sync", new: "testflight config"},
		},
	)
	cmd.UsageFunc = testflightVisibleUsageFunc
	return cmd
}

func TestFlightReviewSurfaceCommand() *ffcli.Command {
	cmd := rewriteCommandTree(
		TestFlightReviewCommand(),
		"aso testflight review",
		"aso testflight review",
		map[string]string{
			"get":    "view",
			"update": "edit",
		},
		[]textReplacement{
			{old: "Manage TestFlight beta app review details.", new: "Manage TestFlight review details."},
			{old: "Manage TestFlight beta app review details", new: "Manage TestFlight review details"},
			{old: "beta app review details", new: "review details"},
			{old: "Beta app review details", new: "Review details"},
			{old: "beta app review detail", new: "review detail"},
			{old: "Beta app review detail", new: "Review detail"},
			{old: "beta app review submissions", new: "review submissions"},
			{old: "Beta app review submissions", new: "Review submissions"},
			{old: "beta app review submission", new: "review submission"},
			{old: "Beta app review submission", new: "Review submission"},
			{old: "Get ", new: "View "},
			{old: "get ", new: "view "},
			{old: "Fetch ", new: "View "},
			{old: "fetch ", new: "view "},
			{old: "Update ", new: "Edit "},
			{old: "update ", new: "edit "},
			{old: "Submit a build for beta app review.", new: "Submit a build for TestFlight review."},
			{old: "Submit a build for beta app review", new: "Submit a build for TestFlight review"},
		},
	)
	cmd.ShortHelp = "Manage TestFlight review details."
	cmd.LongHelp = `Manage TestFlight review details and submissions.

Examples:
  aso testflight review view --app "APP_ID"
  aso testflight review edit --id "DETAIL_ID" --contact-email "dev@example.com"
  aso testflight review submit --build "BUILD_ID" --confirm
  aso testflight review app view --id "DETAIL_ID"
  aso testflight review submissions list --build "BUILD_ID"
  aso testflight review submissions view --id "SUBMISSION_ID"`
	setUsageFuncRecursively(cmd, testflightVisibleUsageFunc)

	cmd.Subcommands = append(cmd.Subcommands,
		deprecatedAliasCommand(
			TestFlightReviewGetCommand(),
			"aso testflight review view [flags]",
			"Compatibility alias: use `aso testflight review view`.",
			"Compatibility alias: use `aso testflight review view --app APP_ID`.",
		),
		deprecatedAliasCommand(
			TestFlightReviewUpdateCommand(),
			"aso testflight review edit [flags]",
			"Compatibility alias: use `aso testflight review edit`.",
			"Compatibility alias: use `aso testflight review edit --id DETAIL_ID ...`.",
		),
	)

	if appCmd := findSubcommand(cmd, "app"); appCmd != nil {
		appCmd.Subcommands = append(appCmd.Subcommands,
			deprecatedAliasCommand(
				TestFlightReviewAppGetCommand(),
				"aso testflight review app view --id \"DETAIL_ID\"",
				"Compatibility alias: use `aso testflight review app view`.",
				"Compatibility alias: use `aso testflight review app view --id DETAIL_ID`.",
			),
		)
	}

	if submissionsCmd := findSubcommand(cmd, "submissions"); submissionsCmd != nil {
		submissionsCmd.Subcommands = append(submissionsCmd.Subcommands,
			deprecatedAliasCommand(
				TestFlightReviewSubmissionsGetCommand(),
				"aso testflight review submissions view --id \"SUBMISSION_ID\"",
				"Compatibility alias: use `aso testflight review submissions view`.",
				"Compatibility alias: use `aso testflight review submissions view --id SUBMISSION_ID`.",
			),
		)
	}

	return cmd
}

func TestFlightDistributionCommand() *ffcli.Command {
	cmd := rewriteCommandTree(
		TestFlightBetaDetailsCommand(),
		"aso testflight beta-details",
		"aso testflight distribution",
		map[string]string{
			"beta-details": "distribution",
			"get":          "view",
			"update":       "edit",
		},
		[]textReplacement{
			{old: "Manage TestFlight build beta details.", new: "Manage TestFlight distribution settings."},
			{old: "Manage TestFlight build beta details", new: "Manage TestFlight distribution settings"},
			{old: "build beta details", new: "distribution settings"},
			{old: "Build beta details", new: "Distribution settings"},
			{old: "build beta detail", new: "distribution setting"},
			{old: "Build beta detail", new: "Distribution setting"},
			{old: "Get ", new: "View "},
			{old: "get ", new: "view "},
			{old: "Fetch ", new: "View "},
			{old: "fetch ", new: "view "},
			{old: "Update ", new: "Edit "},
			{old: "update ", new: "edit "},
		},
	)
	cmd.ShortHelp = "Manage TestFlight distribution settings."
	cmd.LongHelp = `Manage TestFlight distribution settings.

Examples:
  aso testflight distribution view --build "BUILD_ID"
  aso testflight distribution edit --id "DETAIL_ID" --auto-notify
  aso testflight distribution build view --id "DETAIL_ID"`
	setUsageFuncRecursively(cmd, testflightVisibleUsageFunc)
	return cmd
}

func DeprecatedBetaDetailsAliasCommand() *ffcli.Command {
	cmd := deprecatedAliasCommand(
		rewriteCommandStringsOnly(
			TestFlightBetaDetailsCommand(),
			"aso testflight beta-details",
			"aso testflight distribution",
			map[string]string{
				"beta-details": "distribution",
				"get":          "view",
				"update":       "edit",
			},
		),
		"aso testflight distribution <subcommand> [flags]",
		"Compatibility alias: use `aso testflight distribution`.",
		"Compatibility alias: use `aso testflight distribution ...`.",
	)
	setUsageFuncRecursively(cmd, shared.DeprecatedUsageFunc)
	markDeprecatedSubcommands(cmd)
	if viewCmd := shared.DeprecatedAliasLeafCommand(
		findSubcommand(cmd, "get"),
		"view",
		"aso testflight distribution view [flags]",
		"aso testflight distribution view",
		"Warning: `aso testflight beta-details view` is deprecated. Use `aso testflight distribution view`.",
	); viewCmd != nil {
		cmd.Subcommands = append(cmd.Subcommands, viewCmd)
	}
	if editCmd := shared.DeprecatedAliasLeafCommand(
		findSubcommand(cmd, "update"),
		"edit",
		"aso testflight distribution edit [flags]",
		"aso testflight distribution edit",
		"Warning: `aso testflight beta-details edit` is deprecated. Use `aso testflight distribution edit`.",
	); editCmd != nil {
		cmd.Subcommands = append(cmd.Subcommands, editCmd)
	}
	if buildCmd := findSubcommand(cmd, "build"); buildCmd != nil {
		if viewCmd := shared.DeprecatedAliasLeafCommand(
			findSubcommand(buildCmd, "get"),
			"view",
			"aso testflight distribution build view --id \"DETAIL_ID\"",
			"aso testflight distribution build view",
			"Warning: `aso testflight beta-details build view` is deprecated. Use `aso testflight distribution build view`.",
		); viewCmd != nil {
			buildCmd.Subcommands = append(buildCmd.Subcommands, viewCmd)
		}
	}
	return cmd
}

func TestFlightMetricsSurfaceCommand() *ffcli.Command {
	cmd := TestFlightMetricsCommand()
	cmd.LongHelp = `Fetch TestFlight metrics.

Examples:
  aso testflight metrics public-link --group "GROUP_ID"
  aso testflight metrics group-testers --group "GROUP_ID"
  aso testflight metrics app-testers --app "APP_ID"`
	cmd.UsageFunc = testflightVisibleUsageFunc
	cmd.Subcommands = []*ffcli.Command{
		TestFlightMetricsPublicLinkCommand(),
		TestFlightMetricsGroupTestersCommand(),
		TestFlightMetricsAppTestersCommand(),
		DeprecatedMetricsTestersAliasCommand(),
		DeprecatedMetricsBetaTesterUsagesAliasCommand(),
	}
	return cmd
}

func DeprecatedBetaGroupsAliasCommand() *ffcli.Command {
	cmd := deprecatedAliasCommand(
		rewriteCommandPresentation(
			BetaGroupsCommand(),
			"aso testflight beta-groups",
			"aso testflight groups",
			map[string]string{
				"beta-groups":               "groups",
				"beta-recruitment-criteria": "recruitment",
				"beta-recruitment-criterion-compatible-build-check": "compatibility",
				"relationships": "links",
				"get":           "view",
				"update":        "edit",
			},
		),
		"aso testflight groups <subcommand> [flags]",
		"Compatibility alias: use `aso testflight groups`.",
		"Compatibility alias: use `aso testflight groups ...`.",
	)
	if linksCmd := rewriteCommandTree(
		BetaGroupsRelationshipsCommand(),
		"aso testflight beta-groups relationships",
		"aso testflight groups links",
		map[string]string{
			"relationships": "links",
			"get":           "view",
		},
		nil,
	); linksCmd != nil {
		cmd.Subcommands = append(cmd.Subcommands, linksCmd)
	}
	setUsageFuncRecursively(cmd, shared.DeprecatedUsageFunc)
	markDeprecatedSubcommands(cmd)
	appendDeprecatedLeafAlias(
		cmd,
		"get",
		"view",
		"aso testflight groups view [flags]",
		"aso testflight groups view",
		"Warning: `aso testflight beta-groups view` is deprecated. Use `aso testflight groups view`.",
	)
	appendDeprecatedLeafAlias(
		cmd,
		"update",
		"edit",
		"aso testflight groups edit [flags]",
		"aso testflight groups edit",
		"Warning: `aso testflight beta-groups edit` is deprecated. Use `aso testflight groups edit`.",
	)
	if appCmd := findSubcommand(cmd, "app"); appCmd != nil {
		appendDeprecatedLeafAlias(
			appCmd,
			"get",
			"view",
			"aso testflight groups app view --group-id \"GROUP_ID\"",
			"aso testflight groups app view",
			"Warning: `aso testflight beta-groups app view` is deprecated. Use `aso testflight groups app view`.",
		)
	}
	if recruitmentCmd := findSubcommand(cmd, "beta-recruitment-criteria"); recruitmentCmd != nil {
		appendDeprecatedLeafAlias(
			recruitmentCmd,
			"get",
			"view",
			"aso testflight groups recruitment view --group-id \"GROUP_ID\"",
			"aso testflight groups recruitment view",
			"Warning: `aso testflight beta-groups beta-recruitment-criteria view` is deprecated. Use `aso testflight groups recruitment view`.",
		)
	}
	if compatibilityCmd := findSubcommand(cmd, "beta-recruitment-criterion-compatible-build-check"); compatibilityCmd != nil {
		appendDeprecatedLeafAlias(
			compatibilityCmd,
			"get",
			"view",
			"aso testflight groups compatibility view --group-id \"GROUP_ID\"",
			"aso testflight groups compatibility view",
			"Warning: `aso testflight beta-groups beta-recruitment-criterion-compatible-build-check view` is deprecated. Use `aso testflight groups compatibility view`.",
		)
	}
	return cmd
}

func DeprecatedBetaTestersAliasCommand() *ffcli.Command {
	cmd := deprecatedAliasCommand(
		rewriteCommandPresentation(
			BetaTestersCommand(),
			"aso testflight beta-testers",
			"aso testflight testers",
			map[string]string{
				"beta-testers":  "testers",
				"beta-groups":   "groups",
				"relationships": "links",
				"get":           "view",
			},
		),
		"aso testflight testers <subcommand> [flags]",
		"Compatibility alias: use `aso testflight testers`.",
		"Compatibility alias: use `aso testflight testers ...`.",
	)
	if linksCmd := rewriteCommandTree(
		BetaTestersRelationshipsCommand(),
		"aso testflight beta-testers relationships",
		"aso testflight testers links",
		map[string]string{
			"relationships": "links",
			"get":           "view",
		},
		nil,
	); linksCmd != nil {
		cmd.Subcommands = append(cmd.Subcommands, linksCmd)
	}
	setUsageFuncRecursively(cmd, shared.DeprecatedUsageFunc)
	markDeprecatedSubcommands(cmd)
	appendDeprecatedLeafAlias(
		cmd,
		"get",
		"view",
		"aso testflight testers view [flags]",
		"aso testflight testers view",
		"Warning: `aso testflight beta-testers view` is deprecated. Use `aso testflight testers view`.",
	)
	return cmd
}

func DeprecatedBetaLicenseAgreementsAliasCommand() *ffcli.Command {
	cmd := deprecatedAliasCommand(
		rewriteCommandStringsOnly(
			BetaLicenseAgreementsCommand(),
			"aso testflight beta-license-agreements",
			"aso testflight agreements",
			map[string]string{
				"beta-license-agreements": "agreements",
				"get":                     "view",
				"update":                  "edit",
			},
		),
		"aso testflight agreements <subcommand> [flags]",
		"Compatibility alias: use `aso testflight agreements`.",
		"Compatibility alias: use `aso testflight agreements ...`.",
	)
	setUsageFuncRecursively(cmd, shared.DeprecatedUsageFunc)
	markDeprecatedSubcommands(cmd)
	if viewCmd := shared.DeprecatedAliasLeafCommand(
		findSubcommand(cmd, "get"),
		"view",
		"aso testflight agreements view [flags]",
		"aso testflight agreements view",
		"Warning: `aso testflight beta-license-agreements view` is deprecated. Use `aso testflight agreements view`.",
	); viewCmd != nil {
		cmd.Subcommands = append(cmd.Subcommands, viewCmd)
	}
	if editCmd := shared.DeprecatedAliasLeafCommand(
		findSubcommand(cmd, "update"),
		"edit",
		"aso testflight agreements edit [flags]",
		"aso testflight agreements edit",
		"Warning: `aso testflight beta-license-agreements edit` is deprecated. Use `aso testflight agreements edit`.",
	); editCmd != nil {
		cmd.Subcommands = append(cmd.Subcommands, editCmd)
	}
	return cmd
}

func DeprecatedBetaNotificationsAliasCommand() *ffcli.Command {
	cmd := deprecatedAliasCommand(
		rewriteCommandStringsOnly(
			BetaNotificationsCommand(),
			"aso testflight beta-notifications",
			"aso testflight notifications",
			map[string]string{
				"beta-notifications": "notifications",
				"create":             "send",
			},
		),
		"aso testflight notifications send --build \"BUILD_ID\"",
		"Compatibility alias: use `aso testflight notifications send`.",
		"Compatibility alias: use `aso testflight notifications send --build BUILD_ID`.",
	)
	setUsageFuncRecursively(cmd, shared.DeprecatedUsageFunc)
	markDeprecatedSubcommands(cmd)
	if sendCmd := shared.DeprecatedAliasLeafCommand(
		findSubcommand(cmd, "create"),
		"send",
		"aso testflight notifications send --build \"BUILD_ID\"",
		"aso testflight notifications send",
		"Warning: `aso testflight beta-notifications send` is deprecated. Use `aso testflight notifications send`.",
	); sendCmd != nil {
		cmd.Subcommands = append(cmd.Subcommands, sendCmd)
	}
	return cmd
}

const preReleaseLinksCanonicalRoot = "aso testflight pre-release links"

func testFlightPreReleaseLinksCommand() *ffcli.Command {
	return rewriteCommandTree(
		prerelease.PreReleaseVersionsRelationshipsCommand(),
		"aso pre-release-versions relationships",
		preReleaseLinksCanonicalRoot,
		map[string]string{
			"relationships": "links",
			"get":           "view",
		},
		[]textReplacement{
			{old: "pre-release-versions relationships get", new: "testflight pre-release links view"},
			{old: "Get ", new: "View "},
			{old: "get: ", new: "view: "},
			{old: "get ", new: "view "},
		},
	)
}

func deprecatedPreReleaseRelationshipsAliasCommand() *ffcli.Command {
	cmd := testFlightPreReleaseLinksCommand()
	if cmd == nil {
		return nil
	}

	cmd.Name = "relationships"
	cmd.ShortUsage = preReleaseLinksCanonicalRoot + " <subcommand> [flags]"
	cmd.ShortHelp = "DEPRECATED: use `aso testflight pre-release links ...`."
	cmd.LongHelp = "Deprecated compatibility alias for `aso testflight pre-release links ...`."
	cmd.UsageFunc = shared.DeprecatedUsageFunc

	if viewCmd := findSubcommand(cmd, "view"); viewCmd != nil {
		viewCmd.ShortUsage = preReleaseLinksCanonicalRoot + " view --id \"PR_ID\" --type \"RELATIONSHIP\" [flags]"
		viewCmd.ShortHelp = "Compatibility alias: use `aso testflight pre-release links view`."
		viewCmd.LongHelp = "Compatibility alias: use `aso testflight pre-release links view --id \"PR_ID\" --type \"RELATIONSHIP\" [flags]`."
		viewCmd.UsageFunc = shared.DeprecatedUsageFunc
		origExec := viewCmd.Exec
		viewCmd.Exec = func(ctx context.Context, args []string) error {
			fmt.Fprintln(os.Stderr, "Warning: `aso testflight pre-release relationships view` is deprecated. Use `aso testflight pre-release links view`.")
			return origExec(ctx, args)
		}
	}

	return hideTestFlightCommand(cmd)
}

func TestFlightAppLocalizationsCommand() *ffcli.Command {
	cmd := rewriteCommandTree(
		betaapplocalizations.BetaAppLocalizationsCommand(),
		"aso beta-app-localizations",
		"aso testflight app-localizations",
		map[string]string{
			"beta-app-localizations": "app-localizations",
		},
		[]textReplacement{
			{old: "beta-app-localizations ", new: "testflight app-localizations "},
			{old: "Manage TestFlight beta app localizations.", new: "Manage TestFlight app localizations."},
			{old: "Manage TestFlight beta app localizations", new: "Manage TestFlight app localizations"},
			{old: "List beta app localizations", new: "List app localizations"},
			{old: "Get a beta app localization", new: "Get an app localization"},
			{old: "Create a beta app localization", new: "Create an app localization"},
			{old: "Update a beta app localization", new: "Update an app localization"},
			{old: "Delete a beta app localization", new: "Delete an app localization"},
			{old: "View the app for a beta app localization", new: "View the app for an app localization"},
			{old: "beta app localizations", new: "app localizations"},
			{old: "Beta app localizations", new: "App localizations"},
			{old: "beta app localization", new: "app localization"},
			{old: "Beta app localization", new: "App localization"},
		},
	)
	cmd.ShortHelp = "Manage TestFlight app localizations."
	cmd.LongHelp = `Manage TestFlight app localizations.

Examples:
  aso testflight app-localizations list --app "APP_ID"
  aso testflight app-localizations get --id "LOCALIZATION_ID"
  aso testflight app-localizations app get --id "LOCALIZATION_ID"
  aso testflight app-localizations create --app "APP_ID" --locale "en-US" --description "Welcome testers"`
	setUsageFuncRecursively(cmd, testflightVisibleUsageFunc)
	return cmd
}

func TestFlightPreReleaseCommand() *ffcli.Command {
	cmd := rewriteCommandTree(
		prerelease.PreReleaseVersionsCommand(),
		"aso pre-release-versions",
		"aso testflight pre-release",
		map[string]string{
			"pre-release-versions": "pre-release",
			"relationships":        "links",
			"get":                  "view",
		},
		[]textReplacement{
			{old: "pre-release-versions ", new: "testflight pre-release "},
			{old: "pre-release-versions relationships get", new: "testflight pre-release links view"},
			{old: "Manage TestFlight pre-release versions.", new: "Manage pre-release versions."},
			{old: "List TestFlight pre-release versions", new: "List pre-release versions"},
			{old: "Get a TestFlight pre-release version", new: "View a pre-release version"},
			{old: "Get the app for a pre-release version", new: "View the app for a pre-release version"},
			{old: "Get relationship linkages for a pre-release version", new: "View relationship linkages for a pre-release version"},
			{old: "Get ", new: "View "},
			{old: "get: ", new: "view: "},
			{old: "get ", new: "view "},
		},
	)
	cmd.ShortHelp = "Manage pre-release versions."
	cmd.LongHelp = `Manage pre-release versions.

Examples:
  aso testflight pre-release list --app "APP_ID"
  aso testflight pre-release view --id "PR_ID"
  aso testflight pre-release app view --id "PR_ID"
  aso testflight pre-release builds list --id "PR_ID"
  aso testflight pre-release links view --id "PR_ID" --type "app"`
	setUsageFuncRecursively(cmd, testflightVisibleUsageFunc)
	cmd.Subcommands = append(cmd.Subcommands, deprecatedPreReleaseRelationshipsAliasCommand())
	return cmd
}

func DeprecatedTestFlightSyncAliasCommand() *ffcli.Command {
	cmd := deprecatedAliasCommand(
		rewriteCommandStringsOnly(
			TestFlightSyncCommand(),
			"aso testflight sync",
			"aso testflight config",
			map[string]string{
				"sync": "config",
				"pull": "export",
			},
		),
		"aso testflight config export [flags]",
		"Compatibility alias: use `aso testflight config export`.",
		"Compatibility alias: use `aso testflight config export --app APP_ID --output ./testflight.yaml`.",
	)
	setUsageFuncRecursively(cmd, shared.DeprecatedUsageFunc)
	markDeprecatedSubcommands(cmd)
	if exportCmd := shared.DeprecatedAliasLeafCommand(
		findSubcommand(cmd, "pull"),
		"export",
		"aso testflight config export [flags]",
		"aso testflight config export",
		"Warning: `aso testflight sync export` is deprecated. Use `aso testflight config export`.",
	); exportCmd != nil {
		cmd.Subcommands = append(cmd.Subcommands, exportCmd)
	}
	return cmd
}
