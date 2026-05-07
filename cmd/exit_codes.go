package cmd

import (
	"errors"
	"flag"
	"net/http"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// Exit codes following the CI/CD specification.
const (
	ExitSuccess  = 0
	ExitError    = 1
	ExitUsage    = 2
	ExitAuth     = 3
	ExitNotFound = 4
	ExitConflict = 5

	// HTTP 4xx range: 10 + (status - 400)
	ExitHTTPBadRequest    = 10
	ExitHTTPUnauthorized  = 11
	ExitHTTPForbidden     = 12
	ExitHTTPUnprocessable = 22

	// HTTP 5xx range: 60 + (status - 500)
	ExitHTTPInternalServer     = 60
	ExitHTTPBadGateway         = 62
	ExitHTTPServiceUnavailable = 63
)

// ExitCodeFromError maps an error to the appropriate exit code.
func ExitCodeFromError(err error) int {
	if err == nil {
		return ExitSuccess
	}
	if errors.Is(err, flag.ErrHelp) {
		return ExitUsage
	}
	if errors.Is(err, shared.ErrMissingAuth) {
		return ExitAuth
	}
	return ExitError
}

// HTTPStatusToExitCode maps an HTTP status code to a CLI exit code.
func HTTPStatusToExitCode(status int) int {
	switch {
	case status == http.StatusNotFound:
		return ExitNotFound
	case status == http.StatusConflict:
		return ExitConflict
	case status >= 400 && status < 500:
		code := 10 + (status - 400)
		if code > 59 {
			code = 59
		}
		return code
	case status >= 500 && status < 600:
		code := 60 + (status - 500)
		if code > 99 {
			code = 99
		}
		return code
	default:
		return ExitError
	}
}
