// Package shared contains a slim set of helpers used across the surviving
// subcommands (aso, completion, docs). Anything App Store Connect specific
// has been removed; this package now deals only with output formatting,
// usage rendering, terminal sanitization, and CI report flags.
package shared

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"
	"golang.org/x/term"
)

// ErrMissingAuth is returned by commands that require authentication.
var ErrMissingAuth = errors.New("missing authentication")

// ANSI bold escape codes
var (
	bold  = "\033[1m"
	reset = "\033[22m"
)

var (
	selectedProfileMu sync.RWMutex
	selectedProfile   string
	isTerminal        = term.IsTerminal
)

// SelectedProfile returns the current profile override.
func SelectedProfile() string {
	selectedProfileMu.RLock()
	defer selectedProfileMu.RUnlock()
	return selectedProfile
}

// SetSelectedProfile sets the profile (tests only).
func SetSelectedProfile(value string) {
	selectedProfileMu.Lock()
	defer selectedProfileMu.Unlock()
	selectedProfile = value
}

// BindRootFlags registers root-level flags.
func BindRootFlags(fs *flag.FlagSet) {
	fs.StringVar(&selectedProfile, "profile", "", "Use named authentication profile")
	BindCIFlags(fs)
}

// CleanupTempPrivateKeys is a no-op in the slim build but kept to preserve the
// Run() entrypoint signature.
func CleanupTempPrivateKeys() {}

// Bold wraps the string in ANSI bold codes when the terminal supports color.
func Bold(s string) string {
	if !supportsANSI() {
		return s
	}
	return bold + s + reset
}

func supportsANSI() bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	if strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return false
	}
	return isTerminal(int(os.Stderr.Fd()))
}

// DefaultUsageFunc is the standard help renderer used by survivor commands.
func DefaultUsageFunc(c *ffcli.Command) string {
	var b strings.Builder

	shortHelp := strings.TrimSpace(c.ShortHelp)
	longHelp := strings.TrimSpace(c.LongHelp)
	if shortHelp == "" && longHelp != "" {
		shortHelp = longHelp
		longHelp = ""
	}

	if shortHelp != "" {
		b.WriteString(Bold("DESCRIPTION"))
		b.WriteString("\n  ")
		b.WriteString(shortHelp)
		b.WriteString("\n\n")
	}

	usage := strings.TrimSpace(c.ShortUsage)
	if usage == "" {
		usage = strings.TrimSpace(c.Name)
	}
	if usage != "" {
		b.WriteString(Bold("USAGE"))
		b.WriteString("\n  ")
		b.WriteString(usage)
		b.WriteString("\n\n")
	}

	if longHelp != "" {
		if shortHelp != "" && strings.HasPrefix(longHelp, shortHelp) {
			longHelp = strings.TrimSpace(strings.TrimPrefix(longHelp, shortHelp))
		}
		if longHelp != "" {
			b.WriteString(longHelp)
			b.WriteString("\n\n")
		}
	}

	if len(c.Subcommands) > 0 {
		b.WriteString(Bold("SUBCOMMANDS"))
		b.WriteString("\n")
		tw := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
		for _, sub := range c.Subcommands {
			fmt.Fprintf(tw, "  %s\t%s\n", sub.Name, sub.ShortHelp)
		}
		tw.Flush()
		b.WriteString("\n")
	}

	if c.FlagSet != nil {
		var flags []*flag.Flag
		c.FlagSet.VisitAll(func(f *flag.Flag) {
			flags = append(flags, f)
		})
		if len(flags) > 0 {
			b.WriteString(Bold("FLAGS"))
			b.WriteString("\n")
			tw := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
			for _, f := range flags {
				if f.DefValue != "" {
					fmt.Fprintf(tw, "  --%-12s %s (default: %s)\n", f.Name, f.Usage, f.DefValue)
				} else {
					fmt.Fprintf(tw, "  --%-12s %s\n", f.Name, f.Usage)
				}
			}
			tw.Flush()
			b.WriteString("\n")
		}
	}

	return b.String()
}

// OutputFlags stores pointers to output-related flag values.
type OutputFlags struct {
	Output *string
	Pretty *bool
}

type validatedOutputValue struct {
	value   *string
	pretty  *bool
	allowed []string
}

func (v *validatedOutputValue) String() string {
	if v == nil || v.value == nil {
		return ""
	}
	return *v.value
}

func (v *validatedOutputValue) Set(value string) error {
	if v == nil || v.value == nil {
		return fmt.Errorf("output flag is not initialized")
	}
	*v.value = value
	return nil
}

func (v *validatedOutputValue) Validate() error {
	if v == nil || v.value == nil {
		return nil
	}
	pretty := false
	if v.pretty != nil {
		pretty = *v.pretty
	}
	_, err := validateOutputFormatAllowed(*v.value, pretty, v.allowed...)
	return err
}

// BindOutputFlags registers --output and --pretty.
func BindOutputFlags(fs *flag.FlagSet) OutputFlags {
	outputValue := defaultOutputFormat()
	prettyValue := false
	fs.Var(&validatedOutputValue{
		value:   &outputValue,
		pretty:  &prettyValue,
		allowed: []string{"json", "table", "markdown"},
	}, "output", "Output format: json, table, markdown")
	fs.BoolVar(&prettyValue, "pretty", false, "Pretty-print JSON output")
	return OutputFlags{Output: &outputValue, Pretty: &prettyValue}
}

func defaultOutputFormat() string {
	env := strings.TrimSpace(os.Getenv("ASC_DEFAULT_OUTPUT"))
	switch strings.ToLower(env) {
	case "json", "table", "markdown":
		return strings.ToLower(env)
	case "md":
		return "markdown"
	}
	if isTerminal(int(os.Stdout.Fd())) {
		return "table"
	}
	return "json"
}

func normalizeOutputFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "md":
		return "markdown"
	default:
		return strings.ToLower(strings.TrimSpace(format))
	}
}

func validateOutputFormatAllowed(format string, pretty bool, allowed ...string) (string, error) {
	if len(allowed) == 0 {
		allowed = []string{"json", "table", "markdown"}
	}
	normalized := normalizeOutputFormat(format)
	if normalized == "" {
		normalized = "json"
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, item := range allowed {
		if c := normalizeOutputFormat(item); c != "" {
			allowedSet[c] = struct{}{}
		}
	}
	if _, ok := allowedSet[normalized]; !ok {
		return "", fmt.Errorf("unsupported format: %s", normalized)
	}
	if pretty && normalized != "json" {
		return "", fmt.Errorf("--pretty is only valid with JSON output")
	}
	return normalized, nil
}

// PrintOutput renders data in the requested format. Only "json" is supported
// in the slim build for arbitrary structures; "table" and "markdown" require
// callers to use PrintOutputWithRenderers.
func PrintOutput(data any, format string, pretty bool) error {
	normalized, err := validateOutputFormatAllowed(format, pretty)
	if err != nil {
		return err
	}
	switch normalized {
	case "json":
		return printJSON(data, pretty)
	default:
		return fmt.Errorf("format %q requires a custom renderer", normalized)
	}
}

// PrintOutputWithRenderers prints JSON directly or invokes the supplied
// table/markdown renderer.
func PrintOutputWithRenderers(data any, format string, pretty bool, tableRenderer, markdownRenderer func() error) error {
	normalized, err := validateOutputFormatAllowed(format, pretty)
	if err != nil {
		return err
	}
	switch normalized {
	case "json":
		return printJSON(data, pretty)
	case "table":
		if tableRenderer == nil {
			return fmt.Errorf("table renderer is required")
		}
		return tableRenderer()
	case "markdown":
		if markdownRenderer == nil {
			return fmt.Errorf("markdown renderer is required")
		}
		return markdownRenderer()
	default:
		return fmt.Errorf("unsupported format: %s", normalized)
	}
}

func printJSON(data any, pretty bool) error {
	enc := json.NewEncoder(os.Stdout)
	if pretty {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(data)
}

// ValidateBoundOutputFlags checks the validators registered on flagsets.
func ValidateBoundOutputFlags(fs *flag.FlagSet) error {
	if fs == nil {
		return nil
	}
	var validationErr error
	fs.VisitAll(func(f *flag.Flag) {
		if validationErr != nil {
			return
		}
		if v, ok := f.Value.(interface{ Validate() error }); ok {
			validationErr = v.Validate()
		}
	})
	return validationErr
}

// WrapCommandOutputValidation wraps a command (and its subcommands) so
// invalid output-format flag values fail before Exec runs.
func WrapCommandOutputValidation(cmd *ffcli.Command) {
	wrapCommandOutputValidation(cmd, nil)
}

func wrapCommandOutputValidation(cmd *ffcli.Command, parents []*ffcli.Command) {
	if cmd == nil {
		return
	}
	path := append(append([]*ffcli.Command(nil), parents...), cmd)
	for _, sub := range cmd.Subcommands {
		wrapCommandOutputValidation(sub, path)
	}
	if cmd.Exec == nil {
		return
	}
	original := cmd.Exec
	cmd.Exec = func(ctx context.Context, args []string) error {
		for _, c := range path {
			if err := ValidateBoundOutputFlags(c.FlagSet); err != nil {
				return UsageError(err.Error())
			}
		}
		return original(ctx, args)
	}
}

// ContextWithTimeout returns the context unchanged in the slim build. It
// exists so callers that previously relied on ASC_TIMEOUT can keep their
// signatures.
func ContextWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return ctx, func() {}
}

// SanitizeTerminal strips control characters to prevent escape injection.
func SanitizeTerminal(input string) string {
	if input == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(input))
	for _, r := range input {
		if r < 0x20 || r == 0x7f {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// ReportedError marks errors already printed to the user; the Run loop should
// avoid double-reporting them.
type ReportedError interface {
	error
	Reported() bool
}

type reportedError struct{ err error }

func (e reportedError) Error() string  { return e.err.Error() }
func (e reportedError) Unwrap() error  { return e.err }
func (e reportedError) Reported() bool { return true }

// NewReportedError wraps an already-printed error.
func NewReportedError(err error) error {
	if err == nil {
		return nil
	}
	return reportedError{err: err}
}

// UsageError prints an error to stderr and returns flag.ErrHelp so the run
// loop maps it to the usage exit code.
func UsageError(message string) error {
	if msg := strings.TrimSpace(message); msg != "" {
		fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	}
	return flag.ErrHelp
}

// JUnitTestCase / JUnitReport are kept for the --report junit flow.

// JUnitTestCase represents a single test case in a JUnit report.
type JUnitTestCase struct {
	Name      string
	Classname string
	Time      time.Duration
	Failure   string
	Message   string
	SystemOut string
	SystemErr string
}

// JUnitReport represents a JUnit XML report.
type JUnitReport struct {
	Tests     []JUnitTestCase
	Timestamp time.Time
	Name      string
}

type testCaseXML struct {
	XMLName   xml.Name    `xml:"testcase"`
	Name      string      `xml:"name,attr"`
	Classname string      `xml:"classname,attr"`
	Time      string      `xml:"time,attr"`
	Failure   *failureXML `xml:"failure,omitempty"`
	SystemOut string      `xml:"system-out,omitempty"`
	SystemErr string      `xml:"system-err,omitempty"`
}

type failureXML struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type testsuiteXML struct {
	XMLName   xml.Name      `xml:"testsuite"`
	Name      string        `xml:"name,attr"`
	Tests     int           `xml:"tests,attr"`
	Failures  int           `xml:"failures,attr"`
	Errors    int           `xml:"errors,attr"`
	Time      string        `xml:"time,attr"`
	Timestamp string        `xml:"timestamp,attr,omitempty"`
	TestCases []testCaseXML `xml:"testcase"`
}

// Write writes the JUnit report to a file.
func (r *JUnitReport) Write(path string) error {
	if path == "" {
		return fmt.Errorf("report file path is empty")
	}
	data, err := r.Marshal()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// Marshal renders the report as XML bytes.
func (r *JUnitReport) Marshal() ([]byte, error) {
	name := r.Name
	if name == "" {
		name = "aso"
	}
	tests := len(r.Tests)
	failures := 0
	for _, tc := range r.Tests {
		if tc.Failure != "" {
			failures++
		}
	}
	var cases []testCaseXML
	var total time.Duration
	for _, tc := range r.Tests {
		total += tc.Time
		x := testCaseXML{
			Name:      tc.Name,
			Classname: tc.Classname,
			Time:      fmt.Sprintf("%.3f", tc.Time.Seconds()),
		}
		if tc.Failure != "" {
			x.Failure = &failureXML{Message: tc.Message, Type: tc.Failure}
		}
		if tc.SystemOut != "" {
			x.SystemOut = tc.SystemOut
		}
		if tc.SystemErr != "" {
			x.SystemErr = tc.SystemErr
		}
		cases = append(cases, x)
	}
	ts := testsuiteXML{
		Name:      name,
		Tests:     tests,
		Failures:  failures,
		Time:      fmt.Sprintf("%.3f", total.Seconds()),
		Timestamp: r.Timestamp.Format(time.RFC3339),
		TestCases: cases,
	}
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	enc := xml.NewEncoder(&sb)
	if err := enc.Encode(ts); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return []byte(sb.String()), nil
}
