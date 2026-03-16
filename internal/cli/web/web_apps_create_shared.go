package web

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
	webcore "github.com/rudrankriyam/App-Store-Connect-CLI/internal/web"
)

// AppsCreateRunOptions configures the canonical web-backed app-create flow.
type AppsCreateRunOptions struct {
	Name          string
	BundleID      string
	SKU           string
	PrimaryLocale string
	Platform      string
	Version       string
	CompanyName   string

	AppleID       string
	Password      string
	TwoFactorCode string

	AutoRename bool
	Output     string
	Pretty     bool
}

const (
	appCreateDefaultPrimaryLocale = "en-US"
	appCreateDefaultPlatform      = "IOS"
	appCreateDefaultVersion       = "1.0"
)

var (
	appCreateAskOneFn                 = survey.AskOne
	resolveAppCreateSessionFn         = resolveAppCreateSession
	appCreateCanPromptInteractivelyFn = appCreateCanPromptInteractively
)

func appCreateCanPromptInteractively() bool {
	if tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		_ = tty.Close()
		return true
	}
	return termIsTerminalFn(int(os.Stdin.Fd()))
}

func trimAppsCreateRunOptions(opts AppsCreateRunOptions) AppsCreateRunOptions {
	opts.Name = strings.TrimSpace(opts.Name)
	opts.BundleID = strings.TrimSpace(opts.BundleID)
	opts.SKU = strings.TrimSpace(opts.SKU)
	opts.PrimaryLocale = strings.TrimSpace(opts.PrimaryLocale)
	opts.Platform = strings.ToUpper(strings.TrimSpace(opts.Platform))
	opts.Version = strings.TrimSpace(opts.Version)
	opts.CompanyName = strings.TrimSpace(opts.CompanyName)
	opts.AppleID = strings.TrimSpace(opts.AppleID)
	opts.TwoFactorCode = strings.TrimSpace(opts.TwoFactorCode)
	opts.Output = strings.TrimSpace(opts.Output)
	return opts
}

func normalizeAppsCreateRunOptions(opts AppsCreateRunOptions) AppsCreateRunOptions {
	opts = trimAppsCreateRunOptions(opts)
	if opts.PrimaryLocale == "" {
		opts.PrimaryLocale = appCreateDefaultPrimaryLocale
	}
	if opts.Platform == "" {
		opts.Platform = appCreateDefaultPlatform
	}
	if opts.Version == "" {
		opts.Version = appCreateDefaultVersion
	}
	return opts
}

func promptAppsCreateFields(opts *AppsCreateRunOptions) error {
	if opts == nil {
		return fmt.Errorf("app create options are required")
	}
	fullWizard := strings.TrimSpace(opts.Name) == "" && strings.TrimSpace(opts.BundleID) == "" && strings.TrimSpace(opts.SKU) == ""

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Create a new app in App Store Connect")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Note: App creation uses Apple's unofficial web-session create flow.")
	fmt.Fprintln(os.Stderr)

	nameValue := strings.TrimSpace(opts.Name)
	if nameValue == "" {
		if err := appCreateAskOneFn(&survey.Input{
			Message: "App name:",
			Help:    "The name of your app as it will appear in App Store Connect",
		}, &nameValue, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
	}

	bundleIDValue := strings.TrimSpace(opts.BundleID)
	if bundleIDValue == "" {
		if err := appCreateAskOneFn(&survey.Input{
			Message: "Bundle ID:",
			Help:    "The bundle identifier (for example, com.example.myapp). Must match an App ID in your developer account.",
		}, &bundleIDValue, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
	}

	skuValue := strings.TrimSpace(opts.SKU)
	if skuValue == "" {
		if err := appCreateAskOneFn(&survey.Input{
			Message: "SKU:",
			Help:    "A unique identifier for your app used internally by Apple",
		}, &skuValue, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
	}

	localeValue := strings.TrimSpace(opts.PrimaryLocale)
	platformValue := strings.TrimSpace(opts.Platform)
	if fullWizard {
		if localeValue == "" {
			localeValue = appCreateDefaultPrimaryLocale
		}
		if err := appCreateAskOneFn(&survey.Input{
			Message: "Primary locale:",
			Default: appCreateDefaultPrimaryLocale,
			Help:    "The primary language for your app (for example, en-US, en-GB, de-DE)",
		}, &localeValue); err != nil {
			return err
		}

		if platformValue == "" {
			platformValue = appCreateDefaultPlatform
		}
		if err := appCreateAskOneFn(&survey.Select{
			Message: "Platform:",
			Options: []string{"IOS", "MAC_OS", "TV_OS", "UNIVERSAL"},
			Default: platformValue,
			Help:    "The primary platform for your app",
		}, &platformValue); err != nil {
			return err
		}
	}

	opts.Name = strings.TrimSpace(nameValue)
	opts.BundleID = strings.TrimSpace(bundleIDValue)
	opts.SKU = strings.TrimSpace(skuValue)
	opts.PrimaryLocale = strings.TrimSpace(localeValue)
	opts.Platform = strings.ToUpper(strings.TrimSpace(platformValue))
	return nil
}

func promptAppsCreateAppleID(appleID *string) error {
	if appleID == nil {
		return fmt.Errorf("apple id target is required")
	}
	value := strings.TrimSpace(*appleID)
	if err := appCreateAskOneFn(&survey.Input{
		Message: "Apple ID (email):",
		Help:    "Your Apple ID email address",
	}, &value, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	*appleID = strings.TrimSpace(value)
	return nil
}

func promptAppsCreatePassword(password *string) error {
	if password == nil {
		return fmt.Errorf("password target is required")
	}
	value := *password
	if err := appCreateAskOneFn(&survey.Password{
		Message: "Apple ID password:",
		Help:    "Your Apple ID password",
	}, &value, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	*password = value
	return nil
}

func resolveAppCreateSession(ctx context.Context, appleID, password, twoFactorCode string) (*webcore.AuthSession, string, error) {
	shared.ApplyRootLoggingOverrides()

	appleID = strings.TrimSpace(appleID)
	twoFactorCode = strings.TrimSpace(twoFactorCode)

	cacheExpired := false
	if appleID != "" {
		if resumed, ok, err := tryResumeSessionFn(ctx, appleID); err == nil && ok {
			return resumed, "cache", nil
		} else if errors.Is(err, webcore.ErrCachedSessionExpired) {
			cacheExpired = true
		}
	} else {
		if resumed, ok, err := tryResumeLastFn(ctx); err == nil && ok {
			return resumed, "cache", nil
		} else if errors.Is(err, webcore.ErrCachedSessionExpired) {
			cacheExpired = true
		}
	}
	if cacheExpired {
		printExpiredSessionNotice(sessionExpiredWriter)
	}

	if appleID == "" {
		if !appCreateCanPromptInteractivelyFn() {
			return nil, "", shared.UsageError("--apple-id is required when no cached web session is available")
		}
		if err := promptAppsCreateAppleID(&appleID); err != nil {
			return nil, "", err
		}
	}

	if password == "" {
		password = os.Getenv(webPasswordEnv)
		if password == "" {
			if !appCreateCanPromptInteractivelyFn() {
				return nil, "", shared.UsageError(fmt.Sprintf("password is required: run in a terminal for an interactive prompt or set %s", webPasswordEnvDisplay()))
			}
			if err := promptAppsCreatePassword(&password); err != nil {
				return nil, "", err
			}
		}
	}

	session, err := loginWithOptionalTwoFactor(ctx, appleID, password, twoFactorCode)
	if err != nil {
		return nil, "", fmt.Errorf("web auth login failed: %w", err)
	}
	if err := webcore.PersistSession(session); err != nil {
		return nil, "", fmt.Errorf("web auth login succeeded but failed to cache session: %w", err)
	}
	return session, "fresh", nil
}

// RunAppsCreate executes the canonical web-backed app-create flow.
func RunAppsCreate(ctx context.Context, opts AppsCreateRunOptions) error {
	opts = trimAppsCreateRunOptions(opts)

	missingName := opts.Name == ""
	missingBundleID := opts.BundleID == ""
	missingSKU := opts.SKU == ""
	if missingName || missingBundleID || missingSKU {
		if !appCreateCanPromptInteractivelyFn() {
			missingFlags := make([]string, 0, 3)
			if missingName {
				missingFlags = append(missingFlags, "--name")
			}
			if missingBundleID {
				missingFlags = append(missingFlags, "--bundle-id")
			}
			if missingSKU {
				missingFlags = append(missingFlags, "--sku")
			}
			return shared.UsageError(fmt.Sprintf("missing required flags: %s", strings.Join(missingFlags, ", ")))
		}
		if err := promptAppsCreateFields(&opts); err != nil {
			return err
		}
	}

	opts = normalizeAppsCreateRunOptions(opts)

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "  Name:      %s\n", opts.Name)
	fmt.Fprintf(os.Stderr, "  Bundle ID: %s\n", opts.BundleID)
	fmt.Fprintf(os.Stderr, "  SKU:       %s\n", opts.SKU)
	fmt.Fprintf(os.Stderr, "  Locale:    %s\n", opts.PrimaryLocale)
	if opts.Platform != "" {
		fmt.Fprintf(os.Stderr, "  Platform:  %s\n", opts.Platform)
	}
	fmt.Fprintln(os.Stderr)

	session, source, err := resolveAppCreateSessionFn(ctx, opts.AppleID, opts.Password, opts.TwoFactorCode)
	if err != nil {
		return err
	}

	requestCtx, cancel := shared.ContextWithTimeout(ctx)
	defer cancel()

	if source == "fresh" {
		fmt.Fprintln(os.Stderr, "Authenticated via fresh web login.")
	} else {
		fmt.Fprintln(os.Stderr, "Using cached web session.")
	}

	client := newWebClientFn(session)
	attrs := webcore.AppCreateAttributes{
		Name:          opts.Name,
		BundleID:      opts.BundleID,
		SKU:           opts.SKU,
		PrimaryLocale: opts.PrimaryLocale,
		Platform:      opts.Platform,
		VersionString: opts.Version,
		CompanyName:   opts.CompanyName,
	}

	createdBundleID, err := withWebSpinnerValue("Checking or creating Bundle ID", func() (bool, error) {
		return ensureBundleIDFn(requestCtx, opts.BundleID, opts.Name, opts.Platform)
	})
	if err != nil {
		if errors.Is(err, shared.ErrMissingAuth) {
			fmt.Fprintln(os.Stderr, "Skipping Bundle ID preflight because official ASC API authentication is not configured.")
			createdBundleID = false
		} else {
			return fmt.Errorf("web apps create failed: bundle id preflight failed: %w", err)
		}
	}
	if createdBundleID {
		fmt.Fprintf(os.Stderr, "Bundle ID %q was missing; created automatically.\n", opts.BundleID)
	}

	app, err := withWebSpinnerValue("Creating app via Apple web API", func() (*webcore.AppResponse, error) {
		return createWebAppFn(requestCtx, client, attrs)
	})
	if err != nil && opts.AutoRename && webcore.IsDuplicateAppNameError(err) {
		suffix := bundleIDNameSuffix(opts.BundleID)
		if suffix != "" {
			for i := 0; i < 5; i++ {
				trySuffix := suffix
				if i > 0 {
					trySuffix = fmt.Sprintf("%s-%d", suffix, i+1)
				}
				tryName := formatAppNameWithSuffix(opts.Name, trySuffix)
				if tryName == "" || tryName == attrs.Name {
					continue
				}
				fmt.Fprintf(os.Stderr, "App name in use; retrying with %q...\n", tryName)
				attrs.Name = tryName
				app, err = withWebSpinnerValue("Creating app via Apple web API", func() (*webcore.AppResponse, error) {
					return createWebAppFn(requestCtx, client, attrs)
				})
				if err == nil || !webcore.IsDuplicateAppNameError(err) {
					break
				}
			}
		}
	}
	if err != nil {
		return fmt.Errorf("web apps create failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Created app successfully (id=%s)\n", strings.TrimSpace(app.Data.ID))
	return shared.PrintOutput(app, opts.Output, opts.Pretty)
}
