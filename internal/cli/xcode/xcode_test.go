package xcode

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
	localxcode "github.com/rudrankriyam/App-Store-Connect-CLI/internal/xcode"
)

func TestXcodeExportWaitRequiresDirectUpload(t *testing.T) {
	restore := overrideXcodeCommandTestHooks(t)
	defer restore()

	isDirectUploadExportOptionsFn = func(string) bool { return false }

	cmd := XcodeExportCommand()
	cmd.FlagSet.SetOutput(io.Discard)
	if err := cmd.FlagSet.Parse([]string{
		"--archive-path", "Demo.xcarchive",
		"--export-options", "ExportOptions.plist",
		"--ipa-path", "Demo.ipa",
		"--wait",
	}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	var runErr error
	_, stderr := captureCommandOutput(t, func() error {
		runErr = cmd.Exec(context.Background(), nil)
		return runErr
	})
	if !errors.Is(runErr, flag.ErrHelp) {
		t.Fatal("expected flag.ErrHelp when --wait is used without direct upload mode")
	}
	if !strings.Contains(stderr, "Error: --wait requires ExportOptions.plist with destination=upload") {
		t.Fatalf("expected direct upload usage error, got %q", stderr)
	}
}

func TestXcodeExportWaitRequiresPositivePollInterval(t *testing.T) {
	restore := overrideXcodeCommandTestHooks(t)
	defer restore()

	isDirectUploadExportOptionsFn = func(string) bool { return true }

	cmd := XcodeExportCommand()
	cmd.FlagSet.SetOutput(io.Discard)
	if err := cmd.FlagSet.Parse([]string{
		"--archive-path", "Demo.xcarchive",
		"--export-options", "ExportOptions.plist",
		"--ipa-path", "Demo.ipa",
		"--wait",
		"--poll-interval", "0s",
	}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	var runErr error
	_, stderr := captureCommandOutput(t, func() error {
		runErr = cmd.Exec(context.Background(), nil)
		return runErr
	})
	if !errors.Is(runErr, flag.ErrHelp) {
		t.Fatal("expected flag.ErrHelp for non-positive poll interval")
	}
	if !strings.Contains(stderr, "Error: --poll-interval must be greater than 0") {
		t.Fatalf("expected poll interval usage error, got %q", stderr)
	}
}

func TestXcodeExportAllowsPollIntervalWithoutWait(t *testing.T) {
	restore := overrideXcodeCommandTestHooks(t)
	defer restore()

	runExport = func(context.Context, localxcode.ExportOptions) (*localxcode.ExportResult, error) {
		return &localxcode.ExportResult{
			ArchivePath: "/tmp/Demo.xcarchive",
			IPAPath:     "/tmp/Demo.ipa",
			BundleID:    "com.example.demo",
			Version:     "1.2.3",
			BuildNumber: "42",
		}, nil
	}

	cmd := XcodeExportCommand()
	cmd.FlagSet.SetOutput(io.Discard)
	if err := cmd.FlagSet.Parse([]string{
		"--archive-path", "Demo.xcarchive",
		"--export-options", "ExportOptions.plist",
		"--ipa-path", "Demo.ipa",
		"--poll-interval", "0s",
		"--output", "json",
	}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	var runErr error
	stdout, stderr := captureCommandOutput(t, func() error {
		runErr = cmd.Exec(context.Background(), nil)
		return runErr
	})
	if runErr != nil {
		t.Fatalf("Exec() error: %v", runErr)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Fatal("expected JSON output")
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected no stderr output without --wait, got %q", stderr)
	}
}

func TestXcodeExportWaitPollsForUploadedBuild(t *testing.T) {
	restore := overrideXcodeCommandTestHooks(t)
	defer restore()

	isDirectUploadExportOptionsFn = func(string) bool { return true }
	runExport = func(context.Context, localxcode.ExportOptions) (*localxcode.ExportResult, error) {
		return &localxcode.ExportResult{
			ArchivePath: "/tmp/Demo.xcarchive",
			IPAPath:     "",
			BundleID:    "com.example.demo",
			Version:     "1.2.3",
			BuildNumber: "42",
		}, nil
	}
	inferArchivePlatformFn = func(string) (string, error) { return "IOS", nil }
	getASCClientFn = func() (*asc.Client, error) { return &asc.Client{}, nil }
	resolveAppIDWithExactLookupFn = func(_ context.Context, _ *asc.Client, app string) (string, error) {
		if app != "com.example.demo" {
			t.Fatalf("expected bundle ID app lookup, got %q", app)
		}
		return "123456789", nil
	}
	resolveBuildUploadIDFn = func(_ context.Context, _ *asc.Client, appID, version, buildNumber, platform string, exportStartedAt time.Time, pollInterval time.Duration) (string, error) {
		if appID != "123456789" {
			t.Fatalf("expected resolved app ID for upload lookup, got %q", appID)
		}
		if version != "1.2.3" || buildNumber != "42" || platform != "IOS" {
			t.Fatalf("unexpected upload lookup params: version=%q build=%q platform=%q", version, buildNumber, platform)
		}
		if pollInterval != 5*time.Second {
			t.Fatalf("expected 5s poll interval, got %s", pollInterval)
		}
		if exportStartedAt.IsZero() {
			t.Fatal("expected export start time for upload lookup")
		}
		return "upload-123", nil
	}
	waitForBuildByNumberOrUploadFailureFn = func(_ context.Context, _ *asc.Client, appID, uploadID, version, buildNumber, platform string, pollInterval time.Duration) (*asc.BuildResponse, error) {
		if appID != "123456789" {
			t.Fatalf("expected resolved app ID, got %q", appID)
		}
		if uploadID != "upload-123" {
			t.Fatalf("expected upload-123 upload ID for xcode export wait, got %q", uploadID)
		}
		if version != "1.2.3" || buildNumber != "42" || platform != "IOS" {
			t.Fatalf("unexpected wait lookup params: version=%q build=%q platform=%q", version, buildNumber, platform)
		}
		if pollInterval != 5*time.Second {
			t.Fatalf("expected 5s poll interval, got %s", pollInterval)
		}
		return &asc.BuildResponse{
			Data: asc.Resource[asc.BuildAttributes]{
				ID: "build-123",
				Attributes: asc.BuildAttributes{
					Version:         "42",
					ProcessingState: asc.BuildProcessingStateValid,
				},
			},
		}, nil
	}
	waitForBuildProcessingFn = func(_ context.Context, _ *asc.Client, buildID string, pollInterval time.Duration) (*asc.BuildResponse, error) {
		if buildID != "build-123" {
			t.Fatalf("expected build-123, got %q", buildID)
		}
		if pollInterval != 5*time.Second {
			t.Fatalf("expected 5s poll interval, got %s", pollInterval)
		}
		return &asc.BuildResponse{
			Data: asc.Resource[asc.BuildAttributes]{
				ID: "build-123",
				Attributes: asc.BuildAttributes{
					Version:         "42",
					ProcessingState: asc.BuildProcessingStateValid,
				},
			},
		}, nil
	}

	cmd := XcodeExportCommand()
	cmd.FlagSet.SetOutput(io.Discard)
	if err := cmd.FlagSet.Parse([]string{
		"--archive-path", "Demo.xcarchive",
		"--export-options", "ExportOptions.plist",
		"--ipa-path", "Demo.ipa",
		"--wait",
		"--poll-interval", "5s",
		"--output", "json",
	}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	var runErr error
	stdout, stderr := captureCommandOutput(t, func() error {
		runErr = cmd.Exec(context.Background(), nil)
		return runErr
	})
	if runErr != nil {
		t.Fatalf("Exec() error: %v", runErr)
	}

	if strings.TrimSpace(stdout) == "" {
		t.Fatal("expected JSON output")
	}
	var payload struct {
		ArchivePath     string `json:"archive_path"`
		IPAPath         string `json:"ipa_path"`
		BuildID         string `json:"build_id"`
		ProcessingState string `json:"processing_state"`
		BundleID        string `json:"bundle_id"`
		Version         string `json:"version"`
		BuildNumber     string `json:"build_number"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error: %v\nstdout=%s", err, stdout)
	}
	if payload.BuildID != "build-123" {
		t.Fatalf("expected build_id build-123, got %q", payload.BuildID)
	}
	if payload.ProcessingState != asc.BuildProcessingStateValid {
		t.Fatalf("expected processing state VALID, got %q", payload.ProcessingState)
	}
	if !strings.Contains(stderr, "Waiting for build 42 (1.2.3) to appear in App Store Connect...") {
		t.Fatalf("expected discovery wait message, got %q", stderr)
	}
	if !strings.Contains(stderr, "Build build-123 discovered; waiting for processing...") {
		t.Fatalf("expected processing wait message, got %q", stderr)
	}
}

func TestXcodeExportWaitRejectsNilProcessedBuildResponse(t *testing.T) {
	restore := overrideXcodeCommandTestHooks(t)
	defer restore()

	isDirectUploadExportOptionsFn = func(string) bool { return true }
	runExport = func(context.Context, localxcode.ExportOptions) (*localxcode.ExportResult, error) {
		return &localxcode.ExportResult{
			ArchivePath: "/tmp/Demo.xcarchive",
			BundleID:    "com.example.demo",
			Version:     "1.2.3",
			BuildNumber: "42",
		}, nil
	}
	inferArchivePlatformFn = func(string) (string, error) { return "IOS", nil }
	getASCClientFn = func() (*asc.Client, error) { return &asc.Client{}, nil }
	resolveAppIDWithExactLookupFn = func(context.Context, *asc.Client, string) (string, error) {
		return "123456789", nil
	}
	resolveBuildUploadIDFn = func(context.Context, *asc.Client, string, string, string, string, time.Time, time.Duration) (string, error) {
		return "upload-123", nil
	}
	waitForBuildByNumberOrUploadFailureFn = func(context.Context, *asc.Client, string, string, string, string, string, time.Duration) (*asc.BuildResponse, error) {
		return &asc.BuildResponse{
			Data: asc.Resource[asc.BuildAttributes]{
				ID: "build-123",
			},
		}, nil
	}
	waitForBuildProcessingFn = func(context.Context, *asc.Client, string, time.Duration) (*asc.BuildResponse, error) {
		return nil, nil
	}

	cmd := XcodeExportCommand()
	cmd.FlagSet.SetOutput(io.Discard)
	if err := cmd.FlagSet.Parse([]string{
		"--archive-path", "Demo.xcarchive",
		"--export-options", "ExportOptions.plist",
		"--ipa-path", "Demo.ipa",
		"--wait",
	}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	var runErr error
	_, _ = captureCommandOutput(t, func() error {
		runErr = cmd.Exec(context.Background(), nil)
		return runErr
	})
	if runErr == nil {
		t.Fatal("expected error for nil processed build response")
	}
	if !strings.Contains(runErr.Error(), "failed to resolve processed build state for build \"build-123\"") {
		t.Fatalf("expected nil processed build error, got %v", runErr)
	}
}

func TestXcodeExportWaitRejectsMissingBuildUploadID(t *testing.T) {
	restore := overrideXcodeCommandTestHooks(t)
	defer restore()

	isDirectUploadExportOptionsFn = func(string) bool { return true }
	runExport = func(context.Context, localxcode.ExportOptions) (*localxcode.ExportResult, error) {
		return &localxcode.ExportResult{
			ArchivePath: "/tmp/Demo.xcarchive",
			BundleID:    "com.example.demo",
			Version:     "1.2.3",
			BuildNumber: "42",
		}, nil
	}
	inferArchivePlatformFn = func(string) (string, error) { return "IOS", nil }
	getASCClientFn = func() (*asc.Client, error) { return &asc.Client{}, nil }
	resolveAppIDWithExactLookupFn = func(context.Context, *asc.Client, string) (string, error) {
		return "123456789", nil
	}
	resolveBuildUploadIDFn = func(context.Context, *asc.Client, string, string, string, string, time.Time, time.Duration) (string, error) {
		return "", nil
	}

	cmd := XcodeExportCommand()
	cmd.FlagSet.SetOutput(io.Discard)
	if err := cmd.FlagSet.Parse([]string{
		"--archive-path", "Demo.xcarchive",
		"--export-options", "ExportOptions.plist",
		"--ipa-path", "Demo.ipa",
		"--wait",
	}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	var runErr error
	_, _ = captureCommandOutput(t, func() error {
		runErr = cmd.Exec(context.Background(), nil)
		return runErr
	})
	if runErr == nil {
		t.Fatal("expected error for missing build upload ID")
	}
	if !strings.Contains(runErr.Error(), "failed to resolve build upload for version \"1.2.3\" build \"42\"") {
		t.Fatalf("expected missing build upload error, got %v", runErr)
	}
}

func TestIsMoreRecentBuildUploadCandidateKeepsCurrentOnOpaqueTies(t *testing.T) {
	base := time.Date(2026, time.March, 16, 12, 0, 0, 0, time.UTC)

	if !isMoreRecentBuildUploadCandidate(base.Add(time.Second), true, base, true) {
		t.Fatal("expected newer observed timestamp to win")
	}
	if !isMoreRecentBuildUploadCandidate(base, true, time.Time{}, false) {
		t.Fatal("expected timestamped candidate to beat an undated current upload")
	}
	if isMoreRecentBuildUploadCandidate(base, true, base, true) {
		t.Fatal("expected identical timestamps to keep the current upload")
	}
	if isMoreRecentBuildUploadCandidate(time.Time{}, false, time.Time{}, false) {
		t.Fatal("expected undated uploads to preserve API order instead of comparing opaque IDs")
	}
}

func TestFindRecentBuildUploadIDIgnoresUndatedUploadsAfterExportStarts(t *testing.T) {
	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	http.DefaultTransport = xcodeCommandRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			return nil, fmt.Errorf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/apps/app-123/buildUploads" {
			return nil, fmt.Errorf("unexpected path: %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("filter[cfBundleShortVersionString]") != "1.2.3" {
			return nil, fmt.Errorf("unexpected short version filter: %q", values.Get("filter[cfBundleShortVersionString]"))
		}
		if values.Get("filter[cfBundleVersion]") != "42" {
			return nil, fmt.Errorf("unexpected build version filter: %q", values.Get("filter[cfBundleVersion]"))
		}
		if values.Get("filter[platform]") != "IOS" {
			return nil, fmt.Errorf("unexpected platform filter: %q", values.Get("filter[platform]"))
		}
		if values.Get("sort") != "-uploadedDate" {
			return nil, fmt.Errorf("unexpected sort: %q", values.Get("sort"))
		}
		if values.Get("limit") != "10" {
			return nil, fmt.Errorf("unexpected limit: %q", values.Get("limit"))
		}
		return xcodeCommandJSONResponse(http.StatusOK, `{
			"data": [
				{
					"type": "buildUploads",
					"id": "stale-undated",
					"attributes": {
						"cfBundleShortVersionString": "1.2.3",
						"cfBundleVersion": "42",
						"platform": "IOS"
					}
				}
			],
			"links": {}
		}`)
	})

	client := newXcodeCommandTestClient(t)
	exportStartedAt := time.Date(2026, time.March, 16, 12, 0, 0, 0, time.UTC)
	uploadID, found, err := findRecentBuildUploadID(context.Background(), client, "app-123", "1.2.3", "42", "IOS", exportStartedAt)
	if err != nil {
		t.Fatalf("findRecentBuildUploadID() error: %v", err)
	}
	if found {
		t.Fatalf("expected undated uploads to be ignored after export start, got upload ID %q", uploadID)
	}
	if uploadID != "" {
		t.Fatalf("expected empty upload ID when only undated uploads exist, got %q", uploadID)
	}
}

func TestFindRecentBuildUploadIDPrefersEarliestUploadInCurrentExportWindow(t *testing.T) {
	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	http.DefaultTransport = xcodeCommandRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			return nil, fmt.Errorf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/apps/app-123/buildUploads" {
			return nil, fmt.Errorf("unexpected path: %s", req.URL.Path)
		}
		return xcodeCommandJSONResponse(http.StatusOK, `{
			"data": [
				{
					"type": "buildUploads",
					"id": "later-retry",
					"attributes": {
						"cfBundleShortVersionString": "1.2.3",
						"cfBundleVersion": "42",
						"platform": "IOS",
						"uploadedDate": "2026-03-16T12:00:30Z"
					}
				},
				{
					"type": "buildUploads",
					"id": "current-export",
					"attributes": {
						"cfBundleShortVersionString": "1.2.3",
						"cfBundleVersion": "42",
						"platform": "IOS",
						"uploadedDate": "2026-03-16T12:00:05Z"
					}
				},
				{
					"type": "buildUploads",
					"id": "older-upload",
					"attributes": {
						"cfBundleShortVersionString": "1.2.3",
						"cfBundleVersion": "42",
						"platform": "IOS",
						"uploadedDate": "2026-03-16T11:59:40Z"
					}
				}
			],
			"links": {}
		}`)
	})

	client := newXcodeCommandTestClient(t)
	exportStartedAt := time.Date(2026, time.March, 16, 12, 0, 0, 0, time.UTC)
	uploadID, found, err := findRecentBuildUploadID(context.Background(), client, "app-123", "1.2.3", "42", "IOS", exportStartedAt)
	if err != nil {
		t.Fatalf("findRecentBuildUploadID() error: %v", err)
	}
	if !found {
		t.Fatal("expected to find a matching upload in the current export window")
	}
	if uploadID != "current-export" {
		t.Fatalf("expected earliest upload in current export window, got %q", uploadID)
	}
}

func overrideXcodeCommandTestHooks(t *testing.T) func() {
	t.Helper()

	originalRunArchive := runArchive
	originalRunExport := runExport
	originalIsDirectUpload := isDirectUploadExportOptionsFn
	originalInferArchivePlatform := inferArchivePlatformFn
	originalGetASCClient := getASCClientFn
	originalResolveAppID := resolveAppIDWithExactLookupFn
	originalResolveBuildUploadID := resolveBuildUploadIDFn
	originalWaitForDiscovery := waitForBuildByNumberOrUploadFailureFn
	originalWaitForProcessing := waitForBuildProcessingFn
	originalWaitTimeout := resolveXcodeExportWaitTimeoutFn

	return func() {
		runArchive = originalRunArchive
		runExport = originalRunExport
		isDirectUploadExportOptionsFn = originalIsDirectUpload
		inferArchivePlatformFn = originalInferArchivePlatform
		getASCClientFn = originalGetASCClient
		resolveAppIDWithExactLookupFn = originalResolveAppID
		resolveBuildUploadIDFn = originalResolveBuildUploadID
		waitForBuildByNumberOrUploadFailureFn = originalWaitForDiscovery
		waitForBuildProcessingFn = originalWaitForProcessing
		resolveXcodeExportWaitTimeoutFn = originalWaitTimeout
	}
}

func captureCommandOutput(t *testing.T, fn func() error) (string, string) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stderr pipe: %v", err)
	}

	os.Stdout = wOut
	os.Stderr = wErr

	outC := make(chan string)
	errC := make(chan string)

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rOut)
		_ = rOut.Close()
		outC <- buf.String()
	}()

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rErr)
		_ = rErr.Close()
		errC <- buf.String()
	}()

	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		_ = wOut.Close()
		_ = wErr.Close()
	}()

	_ = fn()

	_ = wOut.Close()
	_ = wErr.Close()

	stdout := <-outC
	stderr := <-errC

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return stdout, stderr
}

type xcodeCommandRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn xcodeCommandRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func xcodeCommandJSONResponse(status int, body string) (*http.Response, error) {
	return &http.Response{
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}

func newXcodeCommandTestClient(t *testing.T) *asc.Client {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	if pemBytes == nil {
		t.Fatal("encode pem: nil")
	}

	client, err := asc.NewClientFromPEM("KEY_ID", "ISSUER_ID", string(pemBytes))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	return client
}
