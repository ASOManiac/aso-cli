package categories

import (
	"context"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/asc"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// CategoriesCommand returns the categories command with subcommands.
func CategoriesCommand() *ffcli.Command {
	fs := flag.NewFlagSet("categories", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "categories",
		ShortUsage: "aso categories <subcommand> [flags]",
		ShortHelp:  "Manage App Store categories.",
		LongHelp: `Manage App Store categories.

Examples:
  aso categories list
  aso categories get --category-id "GAMES"
  aso categories parent --category-id "GAMES"
  aso categories subcategories --category-id "GAMES"
  aso categories set --app APP_ID --primary GAMES`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			CategoriesListCommand(),
			CategoriesGetCommand(),
			CategoriesParentCommand(),
			CategoriesSubcategoriesCommand(),
			CategoriesSetCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// CategoriesListCommand returns the categories list subcommand.
func CategoriesListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("categories list", flag.ExitOnError)

	limit := fs.Int("limit", 200, "Maximum results to fetch (1-200)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "aso categories list [flags]",
		ShortHelp:  "List available App Store categories.",
		LongHelp: `List available App Store categories.

Category IDs can be used when updating app information to set primary
and secondary categories.

Examples:
  aso categories list
  aso categories list --output table`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit < 1 || *limit > 200 {
				return fmt.Errorf("categories list: --limit must be between 1 and 200")
			}

			client, err := shared.GetASCClient()
			if err != nil {
				return fmt.Errorf("categories list: %w", err)
			}

			requestCtx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			categories, err := client.GetAppCategories(requestCtx, asc.WithAppCategoriesLimit(*limit))
			if err != nil {
				return fmt.Errorf("categories list: %w", err)
			}

			return shared.PrintOutput(categories, *output.Output, *output.Pretty)
		},
	}
}

// CategoriesSetCommand returns the categories set subcommand.
func CategoriesSetCommand() *ffcli.Command {
	return shared.NewCategoriesSetCommand(shared.CategoriesSetCommandConfig{
		FlagSetName: "categories set",
		ShortUsage:  "aso categories set --app APP_ID --primary CATEGORY_ID [--secondary CATEGORY_ID] [flags]",
		ShortHelp:   "Set primary and secondary categories for an app.",
		LongHelp: `Set the primary and secondary categories for an app.

Use 'aso categories list' to find valid category IDs.
Use 'aso categories subcategories --category-id GAMES' to find valid subcategory IDs.

Note: The app must have an editable version in PREPARE_FOR_SUBMISSION state.

Examples:
  aso categories set --app 123456789 --primary GAMES
  aso categories set --app 123456789 --primary GAMES --secondary ENTERTAINMENT
  aso categories set --app 123456789 --primary GAMES --primary-subcategory-one GAMES_ACTION --primary-subcategory-two GAMES_SIMULATION
  aso categories set --app 123456789 --primary GAMES --primary-subcategory-one GAMES_ACTION --secondary ENTERTAINMENT
  aso categories set --app 123456789 --primary PHOTO_AND_VIDEO`,
		ErrorPrefix:    "categories set",
		IncludeAppInfo: true,
	})
}
