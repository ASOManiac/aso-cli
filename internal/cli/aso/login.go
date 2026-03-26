package aso

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/asomaniac"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// openBrowserFunc is overridable for tests.
var openBrowserFunc = openBrowser

// LoginCommand returns the "login" subcommand for ASO Maniac authentication.
func LoginCommand() *ffcli.Command {
	fs := flag.NewFlagSet("aso aso login", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "login",
		ShortUsage: "aso aso login",
		ShortHelp:  "Authenticate with your ASO Maniac account via browser.",
		LongHelp: `Opens a browser window to authenticate with asomaniac.com.

After successful authentication, your API key is saved to ~/.asomaniac/config.json.
Free plan includes 100 API calls per month.`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return runLogin(ctx, asomaniac.DefaultConfigPath(), os.Stderr)
		},
	}
}

func runLogin(ctx context.Context, configPath string, w *os.File) error {
	// Generate random state token for CSRF protection.
	state, err := generateState()
	if err != nil {
		return fmt.Errorf("generate state: %w", err)
	}

	// Start localhost HTTP server on a random port.
	resultCh := make(chan callbackResult, 1)
	handler := callbackHandler(state, resultCh)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("start callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	srv := &http.Server{Handler: handler}
	go srv.Serve(listener) //nolint:errcheck
	defer srv.Close()

	// Build the auth URL and open browser.
	authURL := fmt.Sprintf("https://asomaniac.com/cli/auth?state=%s&port=%d",
		url.QueryEscape(state), port)

	fmt.Fprintf(w, "Opening browser to authenticate...\n")
	fmt.Fprintf(w, "If the browser doesn't open, visit:\n  %s\n\n", authURL)

	if err := openBrowserFunc(authURL); err != nil {
		fmt.Fprintf(w, "Could not open browser automatically: %v\n", err)
	}

	fmt.Fprintf(w, "Waiting for authentication...\n")

	// Wait for callback or context cancellation.
	select {
	case result := <-resultCh:
		if result.err != nil {
			return result.err
		}

		// Save the API key.
		cfg := &asomaniac.Config{
			APIKey:  result.key,
			BaseURL: asomaniac.DefaultBaseURL,
		}
		if err := asomaniac.WriteConfig(configPath, cfg); err != nil {
			return fmt.Errorf("save config: %w", err)
		}

		// Test the key by fetching the profile.
		client := asomaniac.NewClientFromConfig(cfg)
		profile, err := client.GetProfile(ctx)
		if err != nil {
			fmt.Fprintf(w, "\nAPI key saved but could not verify: %v\n", err)
			fmt.Fprintf(w, "You may need to run 'aso aso login' again.\n")
			return nil
		}

		name := profile.Email
		if profile.Name != nil && *profile.Name != "" {
			name = *profile.Name
		}

		fmt.Fprintf(w, "\nAuthenticated as %s (%s plan).\n", name, profile.Plan)
		fmt.Fprintf(w, "API key saved to %s\n", configPath)
		return nil

	case <-ctx.Done():
		return ctx.Err()

	case <-time.After(5 * time.Minute):
		return fmt.Errorf("authentication timed out after 5 minutes")
	}
}

// callbackResult holds the result from the OAuth callback.
type callbackResult struct {
	key string
	err error
}

// callbackHandler returns an http.Handler that processes the OAuth callback.
// It validates the state parameter and extracts the API key.
func callbackHandler(expectedState string, resultCh chan<- callbackResult) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotState := r.URL.Query().Get("state")
		key := r.URL.Query().Get("key")

		if gotState != expectedState {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			resultCh <- callbackResult{err: fmt.Errorf("state mismatch: expected %q, got %q", expectedState, gotState)}
			return
		}

		if key == "" {
			http.Error(w, "Missing API key", http.StatusBadRequest)
			resultCh <- callbackResult{err: fmt.Errorf("callback missing API key")}
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<!DOCTYPE html>
<html><head><title>ASO Maniac CLI</title></head>
<body style="font-family:system-ui;text-align:center;padding:60px">
<h1>Authenticated!</h1>
<p>You can close this window and return to the terminal.</p>
</body></html>`)

		resultCh <- callbackResult{key: key}
	})
}

// generateState generates a cryptographically random hex state token.
func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// openBrowser opens the specified URL in the default browser.
func openBrowser(target string) error {
	target = strings.TrimSpace(target)
	if target == "" {
		return fmt.Errorf("empty URL")
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", target)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}
	return cmd.Run()
}
