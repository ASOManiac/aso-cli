package aso

import "github.com/peterbourgon/ff/v3/ffcli"

// ASOCommands returns all ASO Maniac subcommands to be registered at the root level.
func ASOCommands() []*ffcli.Command {
	return []*ffcli.Command{
		LoginCommand(),
		LogoutCommand(),
		WhoamiCommand(),
		KeywordsCommand(),
		CompetitorsCommand(),
		RankCommand(),
		TrendsCommand(),
		ExportCommand(),
		DashboardCommand(),
		StorefrontsCommand(),
	}
}
