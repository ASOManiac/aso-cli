package cmdtest

import (
	"strings"
	"testing"
)

const (
	feedbackRootDeprecationWarning             = "Warning: `aso feedback` is deprecated. Use `aso testflight feedback list`."
	crashesRootDeprecationWarning              = "Warning: `aso crashes` is deprecated. Use `aso testflight crashes list`."
	betaAppLocalizationsListDeprecationWarning = "Warning: `aso beta-app-localizations list` is deprecated. Use `aso testflight app-localizations list`."
	preReleaseLinksDeprecationWarning          = "Warning: `aso testflight pre-release relationships view` is deprecated. Use `aso testflight pre-release links view`."
)

func requireStderrContainsWarning(t *testing.T, stderr, warning string) {
	t.Helper()
	if !strings.Contains(stderr, warning) {
		t.Fatalf("expected stderr to contain warning %q, got %q", warning, stderr)
	}
}
