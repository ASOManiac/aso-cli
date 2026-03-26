package cmdtest

import (
	"github.com/peterbourgon/ff/v3/ffcli"

	cmd "github.com/ASOManiac/aso-cli/cmd"
	"github.com/ASOManiac/aso-cli/internal/asc"
	"github.com/ASOManiac/aso-cli/internal/auth"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

func resetCmdtestState() {
	asc.ResetConfigCacheForTest()
	auth.ResetInvalidBypassKeychainWarningsForTest()
	shared.ResetDefaultOutputFormat()
	shared.ResetTierCacheForTest()
}

func RootCommand(version string) *ffcli.Command {
	resetCmdtestState()
	return cmd.RootCommand(version)
}

type ReportedError = shared.ReportedError
