package web

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
	webcore "github.com/rudrankriyam/App-Store-Connect-CLI/internal/web"
)

func TestWebAppsCreatePassesPasswordCompatibilityFlagToSessionResolver(t *testing.T) {
	origResolveAppCreateSession := resolveAppCreateSessionFn
	origNewWebClient := newWebClientFn
	origEnsureBundleID := ensureBundleIDFn
	origCreateWebApp := createWebAppFn
	t.Cleanup(func() {
		resolveAppCreateSessionFn = origResolveAppCreateSession
		newWebClientFn = origNewWebClient
		ensureBundleIDFn = origEnsureBundleID
		createWebAppFn = origCreateWebApp
	})

	var (
		receivedID   string
		receivedPass string
	)
	resolveAppCreateSessionFn = func(ctx context.Context, appleID, password, twoFactorCode string) (*webcore.AuthSession, string, error) {
		receivedID = appleID
		receivedPass = password
		return &webcore.AuthSession{}, "cache", nil
	}
	newWebClientFn = func(session *webcore.AuthSession) *webcore.Client {
		return &webcore.Client{}
	}
	ensureBundleIDFn = func(ctx context.Context, bundleID, appName, platform string) (bool, error) {
		return false, nil
	}
	createWebAppFn = func(ctx context.Context, client *webcore.Client, attrs webcore.AppCreateAttributes) (*webcore.AppResponse, error) {
		resp := &webcore.AppResponse{}
		resp.Data.ID = "app-123"
		return resp, nil
	}
	passwordFlag := "--" + "password"

	cmd := WebAppsCreateCommand()
	if err := cmd.FlagSet.Parse([]string{
		"--name", "My App",
		"--bundle-id", "com.example.app",
		"--sku", "SKU123",
		"--apple-id", "user@example.com",
		passwordFlag, "  fixture-password  ",
	}); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if err := cmd.Exec(context.Background(), nil); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if receivedID != "user@example.com" {
		t.Fatalf("expected apple ID %q, got %q", "user@example.com", receivedID)
	}
	if receivedPass != "  fixture-password  " {
		t.Fatalf("expected password %q, got %q", "  fixture-password  ", receivedPass)
	}
}

func TestResolveAppCreateSessionUsesCacheEvenWhenPasswordProvided(t *testing.T) {
	origTryResume := tryResumeSessionFn
	origTryResumeLast := tryResumeLastFn
	origWebLogin := webLoginFn
	t.Cleanup(func() {
		tryResumeSessionFn = origTryResume
		tryResumeLastFn = origTryResumeLast
		webLoginFn = origWebLogin
	})

	expected := &webcore.AuthSession{UserEmail: "cached@example.com"}
	tryResumeSessionFn = func(ctx context.Context, username string) (*webcore.AuthSession, bool, error) {
		if username != "user@example.com" {
			t.Fatalf("expected cache lookup for %q, got %q", "user@example.com", username)
		}
		return expected, true, nil
	}
	tryResumeLastFn = func(ctx context.Context) (*webcore.AuthSession, bool, error) {
		t.Fatal("did not expect last-session cache lookup when apple-id is provided")
		return nil, false, nil
	}
	webLoginFn = func(ctx context.Context, creds webcore.LoginCredentials) (*webcore.AuthSession, error) {
		t.Fatal("did not expect fresh login when cache hit succeeds")
		return nil, nil
	}

	session, source, err := resolveAppCreateSession(context.Background(), "user@example.com", "fixture-password", "")
	if err != nil {
		t.Fatalf("resolveAppCreateSession returned error: %v", err)
	}
	if source != "cache" {
		t.Fatalf("expected source %q, got %q", "cache", source)
	}
	if session != expected {
		t.Fatal("expected cached session pointer to be returned")
	}
}

func TestResolveAppCreateSessionUsesPasswordEnvWithoutTrimming(t *testing.T) {
	origTryResume := tryResumeSessionFn
	origTryResumeLast := tryResumeLastFn
	origWebLogin := webLoginFn
	t.Cleanup(func() {
		tryResumeSessionFn = origTryResume
		tryResumeLastFn = origTryResumeLast
		webLoginFn = origWebLogin
	})

	tryResumeSessionFn = func(ctx context.Context, username string) (*webcore.AuthSession, bool, error) {
		return nil, false, nil
	}
	tryResumeLastFn = func(ctx context.Context) (*webcore.AuthSession, bool, error) {
		return nil, false, nil
	}

	var received webcore.LoginCredentials
	webLoginFn = func(ctx context.Context, creds webcore.LoginCredentials) (*webcore.AuthSession, error) {
		received = creds
		return &webcore.AuthSession{UserEmail: creds.Username}, nil
	}

	t.Setenv(webPasswordEnv, "  env-password  ")

	session, source, err := resolveAppCreateSession(context.Background(), "user@example.com", "", "")
	if err != nil {
		t.Fatalf("resolveAppCreateSession returned error: %v", err)
	}
	if source != "fresh" {
		t.Fatalf("expected fresh source, got %q", source)
	}
	if session == nil {
		t.Fatal("expected session")
	}
	if received.Password != "  env-password  " {
		t.Fatalf("expected env password %q, got %q", "  env-password  ", received.Password)
	}
}

func TestPromptAppsCreatePasswordPreservesWhitespace(t *testing.T) {
	origAskOne := appCreateAskOneFn
	t.Cleanup(func() {
		appCreateAskOneFn = origAskOne
	})

	appCreateAskOneFn = func(p survey.Prompt, response interface{}, _ ...survey.AskOpt) error {
		prompt, ok := p.(*survey.Password)
		if !ok {
			t.Fatalf("expected password prompt, got %T", p)
		}
		if prompt.Message != "Apple ID password:" {
			t.Fatalf("unexpected prompt message %q", prompt.Message)
		}
		target, ok := response.(*string)
		if !ok {
			t.Fatalf("expected *string response, got %T", response)
		}
		*target = "  prompted-password  "
		return nil
	}

	password := ""
	if err := promptAppsCreatePassword(&password); err != nil {
		t.Fatalf("promptAppsCreatePassword returned error: %v", err)
	}
	if password != "  prompted-password  " {
		t.Fatalf("expected prompted password %q, got %q", "  prompted-password  ", password)
	}
}

func TestWebAppsCreateResolvesSessionBeforeTimeoutContext(t *testing.T) {
	origResolveAppCreateSession := resolveAppCreateSessionFn
	t.Cleanup(func() {
		resolveAppCreateSessionFn = origResolveAppCreateSession
	})

	resolveErr := errors.New("stop before network call")
	hadDeadline := false
	resolveAppCreateSessionFn = func(ctx context.Context, appleID, password, twoFactorCode string) (*webcore.AuthSession, string, error) {
		_, hadDeadline = ctx.Deadline()
		return nil, "", resolveErr
	}

	cmd := WebAppsCreateCommand()
	if err := cmd.FlagSet.Parse([]string{
		"--name", "My App",
		"--bundle-id", "com.example.app",
		"--sku", "SKU123",
		"--apple-id", "user@example.com",
	}); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	err := cmd.Exec(context.Background(), nil)
	if !errors.Is(err, resolveErr) {
		t.Fatalf("expected resolveSession error %v, got %v", resolveErr, err)
	}
	if hadDeadline {
		t.Fatal("expected resolveSession to run before timeout context creation")
	}
}

func TestWebAppsCreateInteractiveWizardPromptsForMissingFields(t *testing.T) {
	origAskOne := appCreateAskOneFn
	origResolveAppCreateSession := resolveAppCreateSessionFn
	origNewWebClient := newWebClientFn
	origEnsureBundleID := ensureBundleIDFn
	origCreateWebApp := createWebAppFn
	origCanPrompt := appCreateCanPromptInteractivelyFn
	t.Cleanup(func() {
		appCreateAskOneFn = origAskOne
		resolveAppCreateSessionFn = origResolveAppCreateSession
		newWebClientFn = origNewWebClient
		ensureBundleIDFn = origEnsureBundleID
		createWebAppFn = origCreateWebApp
		appCreateCanPromptInteractivelyFn = origCanPrompt
	})

	promptOrder := []string{}
	appCreateCanPromptInteractivelyFn = func() bool { return true }
	appCreateAskOneFn = func(p survey.Prompt, response interface{}, _ ...survey.AskOpt) error {
		switch prompt := p.(type) {
		case *survey.Input:
			promptOrder = append(promptOrder, prompt.Message)
			target, ok := response.(*string)
			if !ok {
				t.Fatalf("expected *string response for input prompt %q", prompt.Message)
			}
			switch prompt.Message {
			case "App name:":
				*target = "My App"
			case "Bundle ID:":
				*target = "com.example.app"
			case "SKU:":
				*target = "SKU123"
			case "Primary locale:":
				*target = "en-US"
			default:
				t.Fatalf("unexpected input prompt %q", prompt.Message)
			}
		case *survey.Select:
			promptOrder = append(promptOrder, prompt.Message)
			target, ok := response.(*string)
			if !ok {
				t.Fatalf("expected *string response for select prompt %q", prompt.Message)
			}
			if prompt.Message != "Platform:" {
				t.Fatalf("unexpected select prompt %q", prompt.Message)
			}
			*target = "IOS"
		default:
			t.Fatalf("unexpected prompt type %T", p)
		}
		return nil
	}
	resolveAppCreateSessionFn = func(ctx context.Context, appleID, password, twoFactorCode string) (*webcore.AuthSession, string, error) {
		return &webcore.AuthSession{}, "cache", nil
	}
	newWebClientFn = func(session *webcore.AuthSession) *webcore.Client {
		return &webcore.Client{}
	}
	ensureBundleIDFn = func(ctx context.Context, bundleID, appName, platform string) (bool, error) {
		if bundleID != "com.example.app" {
			t.Fatalf("expected prompted bundle id, got %q", bundleID)
		}
		if appName != "My App" {
			t.Fatalf("expected prompted app name, got %q", appName)
		}
		if platform != "IOS" {
			t.Fatalf("expected prompted platform, got %q", platform)
		}
		return false, nil
	}
	createWebAppFn = func(ctx context.Context, client *webcore.Client, attrs webcore.AppCreateAttributes) (*webcore.AppResponse, error) {
		if attrs.Name != "My App" {
			t.Fatalf("expected prompted name, got %q", attrs.Name)
		}
		if attrs.BundleID != "com.example.app" {
			t.Fatalf("expected prompted bundle id, got %q", attrs.BundleID)
		}
		if attrs.SKU != "SKU123" {
			t.Fatalf("expected prompted sku, got %q", attrs.SKU)
		}
		if attrs.PrimaryLocale != "en-US" {
			t.Fatalf("expected prompted locale, got %q", attrs.PrimaryLocale)
		}
		if attrs.Platform != "IOS" {
			t.Fatalf("expected prompted platform, got %q", attrs.Platform)
		}
		resp := &webcore.AppResponse{}
		resp.Data.ID = "app-123"
		return resp, nil
	}

	cmd := WebAppsCreateCommand()
	if err := cmd.FlagSet.Parse([]string{"--output", "json"}); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if err := cmd.Exec(context.Background(), nil); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	wantOrder := []string{"App name:", "Bundle ID:", "SKU:", "Primary locale:", "Platform:"}
	if len(promptOrder) != len(wantOrder) {
		t.Fatalf("expected prompt order %v, got %v", wantOrder, promptOrder)
	}
	for i := range wantOrder {
		if promptOrder[i] != wantOrder[i] {
			t.Fatalf("expected prompt order %v, got %v", wantOrder, promptOrder)
		}
	}
}

func TestWebAppsCreateSkipsBundleIDPreflightWhenOfficialAuthMissing(t *testing.T) {
	origResolveAppCreateSession := resolveAppCreateSessionFn
	origNewWebClient := newWebClientFn
	origEnsureBundleID := ensureBundleIDFn
	origCreateWebApp := createWebAppFn
	t.Cleanup(func() {
		resolveAppCreateSessionFn = origResolveAppCreateSession
		newWebClientFn = origNewWebClient
		ensureBundleIDFn = origEnsureBundleID
		createWebAppFn = origCreateWebApp
	})

	resolveAppCreateSessionFn = func(ctx context.Context, appleID, password, twoFactorCode string) (*webcore.AuthSession, string, error) {
		return &webcore.AuthSession{}, "cache", nil
	}
	newWebClientFn = func(session *webcore.AuthSession) *webcore.Client {
		return &webcore.Client{}
	}

	createCalled := false
	ensureBundleIDFn = func(ctx context.Context, bundleID, appName, platform string) (bool, error) {
		return false, shared.ErrMissingAuth
	}
	createWebAppFn = func(ctx context.Context, client *webcore.Client, attrs webcore.AppCreateAttributes) (*webcore.AppResponse, error) {
		createCalled = true
		resp := &webcore.AppResponse{}
		resp.Data.ID = "app-123"
		return resp, nil
	}

	cmd := WebAppsCreateCommand()
	if err := cmd.FlagSet.Parse([]string{
		"--name", "My App",
		"--bundle-id", "com.example.app",
		"--sku", "SKU123",
		"--apple-id", "user@example.com",
	}); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if err := cmd.Exec(context.Background(), nil); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !createCalled {
		t.Fatal("expected app creation to continue when official bundle-id auth is unavailable")
	}
}

func TestWebAppsCreateEnsuresBundleIDBeforeCreateApp(t *testing.T) {
	origResolveAppCreateSession := resolveAppCreateSessionFn
	origNewWebClient := newWebClientFn
	origEnsureBundleID := ensureBundleIDFn
	origCreateWebApp := createWebAppFn
	t.Cleanup(func() {
		resolveAppCreateSessionFn = origResolveAppCreateSession
		newWebClientFn = origNewWebClient
		ensureBundleIDFn = origEnsureBundleID
		createWebAppFn = origCreateWebApp
	})

	resolveAppCreateSessionFn = func(ctx context.Context, appleID, password, twoFactorCode string) (*webcore.AuthSession, string, error) {
		return &webcore.AuthSession{}, "cache", nil
	}
	newWebClientFn = func(session *webcore.AuthSession) *webcore.Client {
		return &webcore.Client{}
	}

	callOrder := make([]string, 0, 2)
	ensureBundleIDFn = func(ctx context.Context, bundleID, appName, platform string) (bool, error) {
		callOrder = append(callOrder, "ensure")
		if bundleID != "com.example.app" {
			t.Fatalf("expected bundle id %q, got %q", "com.example.app", bundleID)
		}
		if appName != "My App" {
			t.Fatalf("expected app name %q, got %q", "My App", appName)
		}
		if platform != "IOS" {
			t.Fatalf("expected platform %q, got %q", "IOS", platform)
		}
		return true, nil
	}
	createWebAppFn = func(ctx context.Context, client *webcore.Client, attrs webcore.AppCreateAttributes) (*webcore.AppResponse, error) {
		callOrder = append(callOrder, "create")
		resp := &webcore.AppResponse{}
		resp.Data.ID = "app-123"
		return resp, nil
	}

	cmd := WebAppsCreateCommand()
	if err := cmd.FlagSet.Parse([]string{
		"--name", "My App",
		"--bundle-id", "com.example.app",
		"--sku", "SKU123",
		"--apple-id", "user@example.com",
	}); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if err := cmd.Exec(context.Background(), nil); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(callOrder) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(callOrder))
	}
	if callOrder[0] != "ensure" || callOrder[1] != "create" {
		t.Fatalf("expected ensure before create, got %v", callOrder)
	}
}

func TestWebAppsCreateFailsWhenBundleIDPreflightFails(t *testing.T) {
	origResolveAppCreateSession := resolveAppCreateSessionFn
	origNewWebClient := newWebClientFn
	origEnsureBundleID := ensureBundleIDFn
	origCreateWebApp := createWebAppFn
	t.Cleanup(func() {
		resolveAppCreateSessionFn = origResolveAppCreateSession
		newWebClientFn = origNewWebClient
		ensureBundleIDFn = origEnsureBundleID
		createWebAppFn = origCreateWebApp
	})

	resolveAppCreateSessionFn = func(ctx context.Context, appleID, password, twoFactorCode string) (*webcore.AuthSession, string, error) {
		return &webcore.AuthSession{}, "cache", nil
	}
	newWebClientFn = func(session *webcore.AuthSession) *webcore.Client {
		return &webcore.Client{}
	}

	preflightErr := errors.New("preflight failed")
	ensureBundleIDFn = func(ctx context.Context, bundleID, appName, platform string) (bool, error) {
		return false, preflightErr
	}
	createCalled := false
	createWebAppFn = func(ctx context.Context, client *webcore.Client, attrs webcore.AppCreateAttributes) (*webcore.AppResponse, error) {
		createCalled = true
		return nil, nil
	}

	cmd := WebAppsCreateCommand()
	if err := cmd.FlagSet.Parse([]string{
		"--name", "My App",
		"--bundle-id", "com.example.app",
		"--sku", "SKU123",
		"--apple-id", "user@example.com",
	}); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	err := cmd.Exec(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "bundle id preflight failed") {
		t.Fatalf("expected bundle preflight message, got %v", err)
	}
	if !errors.Is(err, preflightErr) {
		t.Fatalf("expected wrapped preflight error, got %v", err)
	}
	if createCalled {
		t.Fatal("expected create app to be skipped on preflight failure")
	}
}

func TestBundleIDPlatformForWebApp(t *testing.T) {
	t.Run("maps UNIVERSAL to IOS for bundle id create", func(t *testing.T) {
		got, err := bundleIDPlatformForWebApp("UNIVERSAL")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != asc.PlatformIOS {
			t.Fatalf("expected %q, got %q", asc.PlatformIOS, got)
		}
	})

	t.Run("keeps explicit mac platform", func(t *testing.T) {
		got, err := bundleIDPlatformForWebApp("MAC_OS")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != asc.PlatformMacOS {
			t.Fatalf("expected %q, got %q", asc.PlatformMacOS, got)
		}
	})

	t.Run("rejects invalid platform with web command contract", func(t *testing.T) {
		_, err := bundleIDPlatformForWebApp("VISION_OS")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "IOS, MAC_OS, TV_OS, UNIVERSAL") {
			t.Fatalf("expected web platform list in error, got %v", err)
		}
	})
}
