package web

import (
	"time"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

const webSpinnerDelay = 200 * time.Millisecond

var webWithSpinnerDelayedFn = shared.WithSpinnerDelayed

func withWebSpinner(label string, fn func() error) error {
	return webWithSpinnerDelayedFn(label, webSpinnerDelay, fn)
}

func withWebSpinnerValue[T any](label string, fn func() (T, error)) (T, error) {
	var result T
	err := withWebSpinner(label, func() error {
		var innerErr error
		result, innerErr = fn()
		return innerErr
	})
	return result, err
}
