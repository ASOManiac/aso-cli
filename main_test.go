package main

import (
	"strings"
	"testing"

	"github.com/ASOManiac/aso-cli/cmd"
)

func TestVersionInfoString(t *testing.T) {
	prevVersion, prevCommit, prevDate := version, commit, date
	t.Cleanup(func() {
		version = prevVersion
		commit = prevCommit
		date = prevDate
	})

	version = "v1.2.3"
	commit = "abc123"
	date = "2026-02-10T00:00:00Z"

	got := versionInfoString()
	if !strings.Contains(got, "v1.2.3") || !strings.Contains(got, "abc123") || !strings.Contains(got, "2026-02-10T00:00:00Z") {
		t.Fatalf("unexpected version info string: %q", got)
	}
}

func TestRunVersionFlagReturnsSuccess(t *testing.T) {
	code := run([]string{"--version"})
	if code != cmd.ExitSuccess {
		t.Fatalf("expected exit success (%d), got %d", cmd.ExitSuccess, code)
	}
}
