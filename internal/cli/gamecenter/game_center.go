package gamecenter

import (
	"context"
	"flag"
	"slices"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/asc"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// GameCenterCommand returns the game-center command group.
func GameCenterCommand() *ffcli.Command {
	fs := flag.NewFlagSet("game-center", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "game-center",
		ShortUsage: "aso game-center <subcommand> [flags]",
		ShortHelp:  "Manage Game Center resources in App Store Connect.",
		LongHelp: `Manage Game Center resources in App Store Connect.

Examples:
  aso game-center achievements list --app "APP_ID"
  aso game-center achievements create --app "APP_ID" --reference-name "First Win" --vendor-id "com.example.firstwin" --points 10
  aso game-center leaderboards list --app "APP_ID"
  aso game-center leaderboards create --app "APP_ID" --reference-name "High Score" --vendor-id "com.example.highscore" --formatter INTEGER --sort DESC --submission-type BEST_SCORE
  aso game-center leaderboard-sets list --app "APP_ID"
  aso game-center leaderboard-sets create --app "APP_ID" --reference-name "Season 1" --vendor-id "com.example.season1"
  aso game-center challenges list --app "APP_ID"
  aso game-center activities list --app "APP_ID"
  aso game-center groups list --app "APP_ID"
  aso game-center app-versions list --app "APP_ID"
  aso game-center enabled-versions list --app "APP_ID"
  aso game-center enabled-versions compatible-versions --id "ENABLED_VERSION_ID"
  aso game-center details list --app "APP_ID"
  aso game-center details achievements-v2 list --id "DETAILS_ID"
  aso game-center matchmaking queues list`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			GameCenterAchievementsCommand(),
			GameCenterLeaderboardsCommand(),
			GameCenterLeaderboardSetsCommand(),
			GameCenterChallengesCommand(),
			GameCenterActivitiesCommand(),
			GameCenterGroupsCommand(),
			GameCenterAppVersionsCommand(),
			GameCenterEnabledVersionsCommand(),
			GameCenterDetailsCommand(),
			GameCenterMatchmakingCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// isValidLeaderboardFormatter checks if the value is a valid leaderboard formatter.
func isValidLeaderboardFormatter(value string) bool {
	return slices.Contains(asc.ValidLeaderboardFormatters, value)
}

// isValidScoreSortType checks if the value is a valid score sort type.
func isValidScoreSortType(value string) bool {
	return slices.Contains(asc.ValidScoreSortTypes, value)
}

// isValidSubmissionType checks if the value is a valid submission type.
func isValidSubmissionType(value string) bool {
	return slices.Contains(asc.ValidSubmissionTypes, value)
}
