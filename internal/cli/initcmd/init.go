package initcmd

import (
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/ASOManiac/aso-cli/internal/cli/docs"
)

// InitCommand returns the root init command.
func InitCommand() *ffcli.Command {
	return docs.NewInitReferenceCommand(
		"init",
		"init",
		"aso init [flags]",
		"Initialize aso helper docs in the current repo.",
		`Initialize aso helper docs in the current repo.

Examples:
  aso init
  aso init --path ./ASC.md
  aso init --force --link=false`,
		"init",
	)
}
