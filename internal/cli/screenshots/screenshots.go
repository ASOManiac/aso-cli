package screenshots

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/assets"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
	"github.com/ASOManiac/aso-cli/internal/cli/shots"
)

// ScreenshotsCommand returns the top-level screenshots command.
func ScreenshotsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("screenshots", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "screenshots",
		ShortUsage: "aso screenshots <subcommand> [flags]",
		ShortHelp:  "Upload and manage App Store screenshots; local capture/frame workflow is [experimental].",
		LongHelp: `Manage the full screenshot workflow from local capture to App Store upload.

Local screenshot automation commands are experimental.
If you face issues, please file feedback at:
https://github.com/ASOManiac/aso-cli/issues/new/choose

Local workflow (experimental):
  aso screenshots run --plan .asc/screenshots.json
  aso screenshots capture --bundle-id "com.example.app" --name home
  aso screenshots frame --input ./screenshots/raw/home.png --device iphone-air
  aso screenshots review-generate --framed-dir ./screenshots/framed
  aso screenshots review-open --output-dir ./screenshots/review
  aso screenshots review-approve --all-ready --output-dir ./screenshots/review
  aso screenshots list-frame-devices --output json

App Store workflow:
  aso screenshots list --version-localization "LOC_ID"
  aso screenshots sizes
  aso screenshots sizes --all
  aso screenshots upload --version-localization "LOC_ID" --path "./screenshots/iphone" --device-type "IPHONE_65"
  aso screenshots upload --version-localization "LOC_ID" --path "./screenshots/ipad" --device-type "IPAD_PRO_3GEN_129"
  aso screenshots download --version-localization "LOC_ID" --output-dir "./screenshots/downloaded"
  aso screenshots delete --id "SCREENSHOT_ID" --confirm

For most iOS submissions, one iPhone set (IPHONE_65) and one iPad set
(IPAD_PRO_3GEN_129) are enough. "aso screenshots sizes" focuses on these by
default; use --all only when you need the full matrix.`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			shots.ShotsRunCommand(),
			shots.ShotsCaptureCommand(),
			shots.ShotsFrameCommand(),
			shots.ShotsFramesListDevicesCommand(),
			shots.ShotsReviewGenerateCommand(),
			shots.ShotsReviewOpenCommand(),
			shots.ShotsReviewApproveCommand(),
			assets.AssetsScreenshotsListCommand(),
			assets.AssetsScreenshotsSizesCommand(),
			assets.AssetsScreenshotsUploadCommand(),
			assets.AssetsScreenshotsDownloadCommand(),
			assets.AssetsScreenshotsDeleteCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
