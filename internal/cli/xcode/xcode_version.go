package xcode

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
	localxcode "github.com/rudrankriyam/App-Store-Connect-CLI/internal/xcode"
)

// XcodeVersionCommand returns the xcode version command group.
func XcodeVersionCommand() *ffcli.Command {
	fs := flag.NewFlagSet("version", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "version",
		ShortUsage: "asc xcode version <subcommand> [flags]",
		ShortHelp:  "Read and modify Xcode project version numbers.",
		LongHelp: `Read and modify Xcode project version and build numbers using agvtool.

Requires Apple Generic Versioning to be enabled in the Xcode project.
macOS only.

Examples:
  asc xcode version get
  asc xcode version get --project-dir ./MyApp
  asc xcode version set --version "1.3.0"
  asc xcode version set --build-number "42"
  asc xcode version set --version "1.3.0" --build-number "42"
  asc xcode version bump --type patch
  asc xcode version bump --type minor
  asc xcode version bump --type major
  asc xcode version bump --type build`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			xcodeVersionGetCommand(),
			xcodeVersionSetCommand(),
			xcodeVersionBumpCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

func xcodeVersionGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("get", flag.ExitOnError)

	projectDir := fs.String("project-dir", ".", "Path to directory containing .xcodeproj")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc xcode version get [--project-dir DIR]",
		ShortHelp:  "Read current version and build number.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) > 0 {
				return shared.UsageErrorf("unexpected argument(s): %s", strings.Join(args, " "))
			}

			dir := strings.TrimSpace(*projectDir)
			if dir == "" {
				dir = "."
			}

			result, err := localxcode.GetVersion(ctx, dir)
			if err != nil {
				return fmt.Errorf("xcode version get: %w", err)
			}

			return shared.PrintOutputWithRenderers(
				result,
				*output.Output,
				*output.Pretty,
				func() error {
					fmt.Printf("Version: %s\n", result.Version)
					fmt.Printf("Build:   %s\n", result.BuildNumber)
					return nil
				},
				func() error {
					fmt.Printf("**Version:** %s\n\n**Build:** %s\n", result.Version, result.BuildNumber)
					return nil
				},
			)
		},
	}
}

func xcodeVersionSetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("set", flag.ExitOnError)

	projectDir := fs.String("project-dir", ".", "Path to directory containing .xcodeproj")
	version := fs.String("version", "", "Marketing version (CFBundleShortVersionString)")
	buildNumber := fs.String("build-number", "", "Build number (CFBundleVersion)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "set",
		ShortUsage: "asc xcode version set [--version VER] [--build-number NUM] [--project-dir DIR]",
		ShortHelp:  "Set version and/or build number.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) > 0 {
				return shared.UsageErrorf("unexpected argument(s): %s", strings.Join(args, " "))
			}

			v := strings.TrimSpace(*version)
			b := strings.TrimSpace(*buildNumber)
			if v == "" && b == "" {
				return shared.UsageError("--version or --build-number is required")
			}

			dir := strings.TrimSpace(*projectDir)
			if dir == "" {
				dir = "."
			}

			result, err := localxcode.SetVersion(ctx, localxcode.SetVersionOptions{
				ProjectDir:  dir,
				Version:     v,
				BuildNumber: b,
			})
			if err != nil {
				return fmt.Errorf("xcode version set: %w", err)
			}

			return shared.PrintOutput(result, *output.Output, *output.Pretty)
		},
	}
}

func xcodeVersionBumpCommand() *ffcli.Command {
	fs := flag.NewFlagSet("bump", flag.ExitOnError)

	projectDir := fs.String("project-dir", ".", "Path to directory containing .xcodeproj")
	bumpType := fs.String("type", "", "Bump type: major, minor, patch, or build (required)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "bump",
		ShortUsage: "asc xcode version bump --type TYPE [--project-dir DIR]",
		ShortHelp:  "Increment version or build number.",
		LongHelp: `Increment the version or build number in an Xcode project.

Bump types:
  major   1.2.3 → 2.0.0
  minor   1.2.3 → 1.3.0
  patch   1.2.3 → 1.2.4
  build   Increment CFBundleVersion (build number)

Examples:
  asc xcode version bump --type patch
  asc xcode version bump --type minor --project-dir ./MyApp
  asc xcode version bump --type build`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) > 0 {
				return shared.UsageErrorf("unexpected argument(s): %s", strings.Join(args, " "))
			}

			parsed, err := localxcode.ParseBumpType(*bumpType)
			if err != nil {
				return shared.UsageError(err.Error())
			}

			dir := strings.TrimSpace(*projectDir)
			if dir == "" {
				dir = "."
			}

			result, err := localxcode.BumpVersion(ctx, localxcode.BumpVersionOptions{
				ProjectDir: dir,
				BumpType:   parsed,
			})
			if err != nil {
				return fmt.Errorf("xcode version bump: %w", err)
			}

			return shared.PrintOutput(result, *output.Output, *output.Pretty)
		},
	}
}
