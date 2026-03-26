//go:build darwin || linux || freebsd || netbsd || openbsd || dragonfly

package asc

import (
	"os"

	"github.com/ASOManiac/aso-cli/internal/secureopen"
)

func openExistingNoFollow(path string) (*os.File, error) {
	return secureopen.OpenExistingNoFollow(path)
}
