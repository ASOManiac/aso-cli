package cmdtest

import (
	"context"
	"errors"
	"flag"
	"io"
	"strings"
	"testing"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func TestAppsHelpShowsInfoSubcommand(t *testing.T) {
	root := RootCommand("1.2.3")
	var appsCmd any
	for _, sub := range root.Subcommands {
		if sub != nil && sub.Name == "apps" {
			appsCmd = sub
			break
		}
	}
	if appsCmd == nil {
		t.Fatal("expected apps command in root subcommands")
	}

	usage := appsCmd.(*ffcli.Command).UsageFunc(appsCmd.(*ffcli.Command))
	if !strings.Contains(usage, "info") {
		t.Fatalf("expected apps help to show info subcommand, got %q", usage)
	}
}

func TestRootHelpRemovesAppInfoRoots(t *testing.T) {
	root := RootCommand("1.2.3")

	var runErr error
	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		runErr = root.Run(context.Background())
	})

	if !errors.Is(runErr, flag.ErrHelp) {
		t.Fatalf("expected ErrHelp, got %v", runErr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
	if strings.Contains(stderr, "  app-info:") || strings.Contains(stderr, "  app-infos:") {
		t.Fatalf("expected root help to remove app-info roots, got %q", stderr)
	}
}

func TestAppsInfoHelpShowsNewSurface(t *testing.T) {
	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	var runErr error
	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"apps", "info"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		runErr = root.Run(context.Background())
	})

	if !errors.Is(runErr, flag.ErrHelp) {
		t.Fatalf("expected ErrHelp, got %v", runErr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
	for _, want := range []string{"list", "view", "edit"} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("expected apps info help to contain %q, got %q", want, stderr)
		}
	}
}
