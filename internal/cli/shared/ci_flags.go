package shared

import (
	"flag"
	"fmt"
)

// ReportFormatJUnit is the only supported value for --report.
const ReportFormatJUnit = "junit"

var (
	reportFormat string
	reportFile   string
)

// BindCIFlags registers --report and --report-file.
func BindCIFlags(fs *flag.FlagSet) {
	fs.StringVar(&reportFormat, "report", "", "Report format for CI output (e.g., junit)")
	fs.StringVar(&reportFile, "report-file", "", "Path to write CI report file")
}

// ValidateReportFlags validates the CI report flags.
func ValidateReportFlags() error {
	if reportFormat != "" && reportFormat != ReportFormatJUnit {
		return fmt.Errorf("--report must be %q if specified, got %q", ReportFormatJUnit, reportFormat)
	}
	if reportFormat == ReportFormatJUnit && reportFile == "" {
		return fmt.Errorf("--report-file is required when --report is specified")
	}
	if reportFile != "" && reportFormat == "" {
		return fmt.Errorf("--report is required when --report-file is specified")
	}
	return nil
}

// ReportFormat returns the configured report format.
func ReportFormat() string { return reportFormat }

// ReportFile returns the configured report file path.
func ReportFile() string { return reportFile }

// SetReportFormat sets the report format (testing).
func SetReportFormat(v string) { reportFormat = v }

// SetReportFile sets the report file path (testing).
func SetReportFile(v string) { reportFile = v }
