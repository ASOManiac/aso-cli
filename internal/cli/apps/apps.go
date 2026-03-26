package apps

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/asc"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
	cliweb "github.com/ASOManiac/aso-cli/internal/cli/web"
)

func appsListFlags(fs *flag.FlagSet) (output shared.OutputFlags, bundleID *string, name *string, sku *string, sort *string, limit *int, next *string, paginate *bool) {
	output = shared.BindOutputFlags(fs)
	bundleID = fs.String("bundle-id", "", "Filter by bundle ID(s), comma-separated")
	name = fs.String("name", "", "Filter by app name(s), comma-separated")
	sku = fs.String("sku", "", "Filter by SKU(s), comma-separated")
	sort = fs.String("sort", "", "Sort by name, -name, bundleId, or -bundleId")
	limit = fs.Int("limit", 0, "Maximum results per page (1-200)")
	next = fs.String("next", "", "Fetch next page using a links.next URL")
	paginate = fs.Bool("paginate", false, "Automatically fetch all pages (aggregate results)")
	return
}

// AppsCommand returns the apps command factory.
func AppsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("apps", flag.ExitOnError)

	output, bundleID, name, sku, sort, limit, next, paginate := appsListFlags(fs)

	return &ffcli.Command{
		Name:       "apps",
		ShortUsage: "aso apps <subcommand> [flags]",
		ShortHelp:  "List and manage apps in App Store Connect.",
		LongHelp: `List and manage apps in App Store Connect.

Examples:
  aso apps
  aso apps list --bundle-id "com.example.app"
  aso web apps create --name "My App" --bundle-id "com.example.app" --sku "MYAPP123"
  aso apps wall
  aso apps wall submit --app "1234567890" --confirm
  aso apps public view --app "1234567890"
  aso apps public search --term "focus" --country us
  aso apps public storefronts list
  aso apps get --id "APP_ID"
  aso apps info view --app "APP_ID"
  aso apps info edit --app "APP_ID" --locale "en-US" --whats-new "Bug fixes"
  aso apps ci-product get --id "APP_ID"
  aso apps update --id "APP_ID" --bundle-id "com.example.app"
  aso apps update --id "APP_ID" --primary-locale "en-US"
  aso apps subscription-grace-period get --app "APP_ID"
  aso apps content-rights edit --app "APP_ID" --uses-third-party-content=false
  aso apps --limit 10
  aso apps --sort name
  aso apps --output table
  aso apps --next "<links.next>"
  aso apps --paginate`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			AppsListCommand(),
			AppsCreateCommand(),
			AppsWallCommand(),
			AppsPublicCommand(),
			AppsGetCommand(),
			AppsInfoCommand(),
			AppsCIProductCommand(),
			AppsUpdateCommand(),
			AppsRemoveBetaTestersCommand(),
			AppsSubscriptionGracePeriodCommand(),
			AppsSearchKeywordsCommand(),
			AppEncryptionDeclarationsCommand(),
			AppsContentRightsCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			if len(args) > 0 {
				fmt.Fprintf(os.Stderr, "Error: unknown subcommand %q\n", strings.TrimSpace(args[0]))
				return flag.ErrHelp
			}
			return appsList(ctx, *output.Output, *output.Pretty, *bundleID, *name, *sku, *sort, *limit, *next, *paginate)
		},
	}
}

// AppsListCommand returns the apps list subcommand.
func AppsListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("apps list", flag.ExitOnError)

	output, bundleID, name, sku, sort, limit, next, paginate := appsListFlags(fs)

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "aso apps list [flags]",
		ShortHelp:  "List apps from App Store Connect.",
		LongHelp: `List apps from App Store Connect.

Examples:
  aso apps list
  aso apps list --bundle-id "com.example.app"
  aso apps list --name "My App"
  aso apps list --limit 10
  aso apps list --sort name
  aso apps list --output table
  aso apps list --next "<links.next>"
  aso apps list --paginate`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return appsList(ctx, *output.Output, *output.Pretty, *bundleID, *name, *sku, *sort, *limit, *next, *paginate)
		},
	}
}

// AppsGetCommand returns the apps get subcommand.
func AppsGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("apps get", flag.ExitOnError)

	id := fs.String("id", "", "App Store Connect app ID")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "aso apps get --id APP_ID",
		ShortHelp:  "Get app details by ID.",
		LongHelp: `Get app details by ID.

Examples:
  aso apps get --id "APP_ID"
  aso apps get --id "APP_ID" --output table`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			idValue := strings.TrimSpace(*id)
			if idValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}

			client, err := shared.GetASCClient()
			if err != nil {
				return fmt.Errorf("apps get: %w", err)
			}

			requestCtx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			app, err := client.GetApp(requestCtx, idValue)
			if err != nil {
				return fmt.Errorf("apps get: failed to fetch: %w", err)
			}

			return shared.PrintOutput(app, *output.Output, *output.Pretty)
		},
	}
}

var runAppsCreateShimFn = cliweb.RunAppsCreate

const (
	appsCreateDeprecationWarning = "Warning: `aso apps create` is deprecated and will be removed after one release cycle."
	appsCreateMigrationGuidance  = "Use `aso web apps create` instead. Legacy ASC_IRIS_SESSION_CACHE entries are imported into the web session cache automatically during the transition."
)

// AppsCreateCommand returns the apps create subcommand.
// TODO(next-release-cycle): remove this shim after the deprecation window closes.
func AppsCreateCommand() *ffcli.Command {
	fs := flag.NewFlagSet("apps create", flag.ExitOnError)

	name := fs.String("name", "", "App name")
	bundleID := fs.String("bundle-id", "", "Bundle ID (e.g., com.example.app)")
	sku := fs.String("sku", "", "Unique SKU for the app")
	primaryLocale := fs.String("primary-locale", "", "Primary locale (e.g., en-US)")
	platform := fs.String("platform", "", "Platform: IOS, MAC_OS, TV_OS, UNIVERSAL")
	version := fs.String("version", "1.0", "Initial version string")
	companyName := fs.String("company-name", "", "Company name (optional)")
	appleID := fs.String("apple-id", "", "Apple ID (email) for authentication")
	password := fs.String("password", "", "Apple ID password (will prompt if not provided)")
	twoFactorCode := fs.String("two-factor-code", "", "2FA verification code (if prompted)")
	twoFactorCodeCommand := fs.String("two-factor-code-command", "", "Shell command that prints the 2FA code to stdout if verification is required")
	autoRename := fs.Bool("auto-rename", true, "Auto-retry with a unique app name when the chosen name is already in use (default: true)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "aso apps create [flags]",
		ShortHelp:  "[deprecated] Create a new app via the unofficial web-session shim.",
		LongHelp: `DEPRECATED: Use ` + "`aso web apps create`" + `.

This compatibility shim forwards to the canonical unofficial web-session app
creation flow and will be removed after one release cycle.

App creation requires Apple web-session authentication (not API key).
If 2FA is enabled on your account, you may need to complete authentication in a browser first.
The canonical web flow also supports --two-factor-code-command or
ASC_WEB_2FA_CODE_COMMAND when a fresh login requires verification.
Legacy ` + "`ASC_IRIS_SESSION_CACHE*`" + ` entries are imported into the web
session cache automatically during the deprecation window.

If flags are not provided, an interactive prompt will guide you through the required fields.
This deprecated shim preserves the old Apple-ID-only contract and assumes the
bundle ID already exists. Use ` + "`aso web apps create`" + ` if you want the
new official-auth bundle-ID preflight and auto-create behavior.

Examples:
  aso web apps create
  aso web apps create --name "My App" --bundle-id "com.example.myapp" --sku "MYAPP123"
  aso apps create --name "My App" --bundle-id "com.example.myapp" --sku "MYAPP123"
  aso apps create --apple-id "user@example.com" --password`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			fmt.Fprintln(os.Stderr, appsCreateDeprecationWarning)
			fmt.Fprintln(os.Stderr, appsCreateMigrationGuidance)
			err := runAppsCreateShimFn(ctx, cliweb.AppsCreateRunOptions{
				Name:                         *name,
				BundleID:                     *bundleID,
				SKU:                          *sku,
				PrimaryLocale:                *primaryLocale,
				Platform:                     *platform,
				Version:                      *version,
				CompanyName:                  *companyName,
				AppleID:                      *appleID,
				Password:                     *password,
				TwoFactorCode:                *twoFactorCode,
				TwoFactorCodeCommand:         *twoFactorCodeCommand,
				AutoRename:                   *autoRename,
				Output:                       *output.Output,
				Pretty:                       *output.Pretty,
				PromptForAppleIDWithPassword: true,
				DisableBundleIDPreflight:     true,
			})
			if err == nil || errors.Is(err, flag.ErrHelp) {
				return err
			}
			return fmt.Errorf("apps create: %w", err)
		},
	}
}

// AppsUpdateCommand returns the apps update subcommand.
func AppsUpdateCommand() *ffcli.Command {
	fs := flag.NewFlagSet("apps update", flag.ExitOnError)

	id := fs.String("id", "", "App Store Connect app ID")
	bundleID := fs.String("bundle-id", "", "Update bundle ID")
	primaryLocale := fs.String("primary-locale", "", "Update primary locale (e.g., en-US)")
	contentRights := fs.String("content-rights", "", "Content rights declaration: DOES_NOT_USE_THIRD_PARTY_CONTENT or USES_THIRD_PARTY_CONTENT")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "update",
		ShortUsage: "aso apps update --id APP_ID [--bundle-id BUNDLE_ID] [--primary-locale LOCALE] [--content-rights DECLARATION]",
		ShortHelp:  "Update an app's bundle ID, primary locale, or content rights declaration.",
		LongHelp: `Update an app's bundle ID, primary locale, or content rights declaration.

Examples:
  aso apps update --id "APP_ID" --bundle-id "com.example.app"
  aso apps update --id "APP_ID" --primary-locale "en-US"
  aso apps update --id "APP_ID" --content-rights "DOES_NOT_USE_THIRD_PARTY_CONTENT"`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			idValue := strings.TrimSpace(*id)
			if idValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}

			attrs := asc.AppUpdateAttributes{}
			if bundleValue := strings.TrimSpace(*bundleID); bundleValue != "" {
				attrs.BundleID = &bundleValue
			}
			if localeValue := strings.TrimSpace(*primaryLocale); localeValue != "" {
				attrs.PrimaryLocale = &localeValue
			}
			if rightsValue := strings.TrimSpace(*contentRights); rightsValue != "" {
				normalizedRights := asc.ContentRightsDeclaration(strings.ToUpper(rightsValue))
				switch normalizedRights {
				case asc.ContentRightsDeclarationDoesNotUseThirdPartyContent,
					asc.ContentRightsDeclarationUsesThirdPartyContent:
					attrs.ContentRightsDeclaration = &normalizedRights
				default:
					fmt.Fprintf(os.Stderr, "Error: --content-rights must be %s or %s\n", asc.ContentRightsDeclarationDoesNotUseThirdPartyContent, asc.ContentRightsDeclarationUsesThirdPartyContent)
					return flag.ErrHelp
				}
			}
			if attrs.BundleID == nil && attrs.PrimaryLocale == nil && attrs.ContentRightsDeclaration == nil {
				fmt.Fprintln(os.Stderr, "Error: --bundle-id, --primary-locale, or --content-rights is required")
				return flag.ErrHelp
			}

			client, err := shared.GetASCClient()
			if err != nil {
				return fmt.Errorf("apps update: %w", err)
			}

			requestCtx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			app, err := client.UpdateApp(requestCtx, idValue, attrs)
			if err != nil {
				return fmt.Errorf("apps update: failed to update: %w", err)
			}

			return shared.PrintOutput(app, *output.Output, *output.Pretty)
		},
	}
}

func appsList(ctx context.Context, output string, pretty bool, bundleID string, name string, sku string, sort string, limit int, next string, paginate bool) error {
	if limit != 0 && (limit < 1 || limit > 200) {
		return fmt.Errorf("apps: --limit must be between 1 and 200")
	}
	if err := shared.ValidateNextURL(next); err != nil {
		return fmt.Errorf("apps: %w", err)
	}
	if err := shared.ValidateSort(sort, "name", "-name", "bundleId", "-bundleId"); err != nil {
		return fmt.Errorf("apps: %w", err)
	}

	client, err := shared.GetASCClient()
	if err != nil {
		return fmt.Errorf("apps: %w", err)
	}

	requestCtx, cancel := shared.ContextWithTimeout(ctx)
	defer cancel()

	opts := []asc.AppsOption{
		asc.WithAppsBundleIDs(shared.SplitCSV(bundleID)),
		asc.WithAppsNames(shared.SplitCSV(name)),
		asc.WithAppsSKUs(shared.SplitCSV(sku)),
		asc.WithAppsLimit(limit),
		asc.WithAppsNextURL(next),
	}
	if strings.TrimSpace(sort) != "" {
		opts = append(opts, asc.WithAppsSort(sort))
	}

	if paginate {
		paginateOpts := append(opts, asc.WithAppsLimit(200))
		apps, err := shared.PaginateWithSpinner(requestCtx,
			func(ctx context.Context) (asc.PaginatedResponse, error) {
				return client.GetApps(ctx, paginateOpts...)
			},
			func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
				return client.GetApps(ctx, asc.WithAppsNextURL(nextURL))
			},
		)
		if err != nil {
			return fmt.Errorf("apps: %w", err)
		}

		return shared.PrintOutput(apps, output, pretty)
	}

	apps, err := client.GetApps(requestCtx, opts...)
	if err != nil {
		return fmt.Errorf("apps: failed to fetch: %w", err)
	}

	return shared.PrintOutput(apps, output, pretty)
}
