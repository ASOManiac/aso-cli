package videopreviews

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/assets"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// VideoPreviewsCommand returns the top-level video previews command.
func VideoPreviewsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("video-previews", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "video-previews",
		ShortUsage: "aso video-previews <subcommand> [flags]",
		ShortHelp:  "Manage App Store app preview videos.",
		LongHelp: `Manage App Store app preview videos for a version localization.

Examples:
  aso video-previews list --version-localization "LOC_ID"
  aso video-previews upload --version-localization "LOC_ID" --path "./previews" --device-type "IPHONE_69"
  aso video-previews download --version-localization "LOC_ID" --output-dir "./previews/downloaded"
  aso video-previews delete --id "PREVIEW_ID" --confirm
  aso video-previews set-poster-frame --id "PREVIEW_ID" --time-code "00:00:05:00"`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			assets.AssetsPreviewsListCommand(),
			assets.AssetsPreviewsUploadCommand(),
			assets.AssetsPreviewsDownloadCommand(),
			assets.AssetsPreviewsDeleteCommand(),
			assets.AssetsPreviewsSetPosterFrameCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
