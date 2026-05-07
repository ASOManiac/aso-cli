package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
	"github.com/ASOManiac/aso-cli/internal/cli/shared/errfmt"
)

// Run executes the CLI using the provided args and version string.
// It returns the intended process exit code.
func Run(args []string, versionInfo string) int {
	defer shared.CleanupTempPrivateKeys()

	if isVersionOnlyInvocation(args) {
		fmt.Fprintln(os.Stdout, versionInfo)
		return ExitSuccess
	}

	root := RootCommand(versionInfo)
	runCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stopSignals()

	if err := root.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitSuccess
		}
		fmt.Fprint(os.Stderr, errfmt.FormatStderr(err))
		return ExitCodeFromError(err)
	}

	if err := shared.ValidateReportFlags(); err != nil {
		fmt.Fprint(os.Stderr, errfmt.FormatStderr(err))
		return ExitUsage
	}

	if versionRequested {
		if err := root.Run(runCtx); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return ExitUsage
			}
			fmt.Fprint(os.Stderr, errfmt.FormatStderr(err))
			return ExitCodeFromError(err)
		}
		return ExitSuccess
	}

	if !hasPositionalArgs(root.FlagSet, args) {
		fmt.Fprint(os.Stdout, root.UsageFunc(root))
		return ExitSuccess
	}

	commandName := getCommandName(root, args)

	start := time.Now()
	runErr := root.Run(runCtx)
	elapsed := time.Since(start)

	if shouldCancelRunContextAfterError(runErr) {
		stopSignals()
	}

	if shared.ReportFormat() == shared.ReportFormatJUnit && shared.ReportFile() != "" {
		if reportErr := writeJUnitReport(commandName, runErr, elapsed); reportErr != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to write JUnit report: %v\n", reportErr)
			if runErr == nil {
				return ExitError
			}
		}
	}

	if runErr != nil {
		if _, ok := errors.AsType[shared.ReportedError](runErr); ok {
			return ExitCodeFromError(runErr)
		}
		if errors.Is(runErr, flag.ErrHelp) {
			return ExitUsage
		}
		fmt.Fprint(os.Stderr, errfmt.FormatStderr(runErr))
		return ExitCodeFromError(runErr)
	}

	return ExitSuccess
}

func shouldCancelRunContextAfterError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func isVersionOnlyInvocation(args []string) bool {
	if len(args) != 1 {
		return false
	}
	switch strings.TrimSpace(args[0]) {
	case "--version", "--version=true":
		return true
	default:
		return false
	}
}

func getCommandName(root *ffcli.Command, args []string) string {
	current := root
	path := []string{current.Name}

	if len(args) > 0 && strings.EqualFold(args[0], root.Name) {
		args = args[1:]
	}

	for i := 0; i < len(args); {
		token := args[i]
		if token == "" {
			i++
			continue
		}

		if sub := findDirectSubcommand(current, token); sub != nil {
			path = append(path, sub.Name)
			current = sub
			i++
			continue
		}

		nextIdx, consumed := consumeFlagToken(current.FlagSet, token, args, i)
		if consumed {
			i = nextIdx
			continue
		}
		break
	}

	return strings.Join(path, " ")
}

func findDirectSubcommand(current *ffcli.Command, token string) *ffcli.Command {
	for _, sub := range current.Subcommands {
		if strings.EqualFold(sub.Name, token) {
			return sub
		}
	}
	return nil
}

func consumeFlagToken(fs *flag.FlagSet, token string, args []string, idx int) (int, bool) {
	if fs == nil || token == "" || token == "-" || !strings.HasPrefix(token, "-") {
		return idx, false
	}
	if token == "--" {
		return idx + 1, true
	}
	trimmed := strings.TrimLeft(token, "-")
	if trimmed == "" {
		return idx, false
	}
	name, hasInlineValue := splitFlagToken(trimmed)
	f := fs.Lookup(name)
	if f == nil {
		return idx, false
	}
	if hasInlineValue || isBoolFlag(f) {
		return idx + 1, true
	}
	if idx+1 < len(args) {
		return idx + 2, true
	}
	return idx + 1, true
}

func hasPositionalArgs(fs *flag.FlagSet, args []string) bool {
	for i := 0; i < len(args); {
		token := args[i]
		if token == "" {
			i++
			continue
		}
		if token == "--" {
			return i+1 < len(args)
		}
		nextIdx, consumed := consumeFlagToken(fs, token, args, i)
		if consumed {
			i = nextIdx
			continue
		}
		return true
	}
	return false
}

func splitFlagToken(token string) (name string, hasInlineValue bool) {
	if before, _, ok := strings.Cut(token, "="); ok {
		return before, true
	}
	return token, false
}

func isBoolFlag(f *flag.Flag) bool {
	type boolFlag interface {
		IsBoolFlag() bool
	}
	v, ok := f.Value.(boolFlag)
	return ok && v.IsBoolFlag()
}

// writeJUnitReport writes a JUnit XML report when --report junit --report-file is set.
func writeJUnitReport(commandName string, runErr error, elapsed time.Duration) error {
	reportFile := shared.ReportFile()
	if reportFile == "" {
		return nil
	}

	tc := shared.JUnitTestCase{
		Name:      commandName,
		Classname: commandName,
		Time:      elapsed,
	}
	if runErr != nil {
		tc.Failure = "ERROR"
		tc.Message = runErr.Error()
	}

	report := shared.JUnitReport{
		Tests:     []shared.JUnitTestCase{tc},
		Timestamp: time.Now(),
		Name:      "aso",
	}
	return report.Write(reportFile)
}
