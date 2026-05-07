// Package errfmt formats errors for stderr output.
package errfmt

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"
)

// FormatStderr returns a user-facing error message suitable for stderr.
// It always ends with a newline.
func FormatStderr(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, flag.ErrHelp) {
		return ""
	}

	msg := err.Error()
	if errors.Is(err, context.Canceled) {
		msg = "operation canceled"
	} else if errors.Is(err, context.DeadlineExceeded) {
		msg = "operation timed out"
	}

	msg = strings.TrimSpace(msg)
	if msg == "" {
		return ""
	}
	return fmt.Sprintf("Error: %s\n", msg)
}
