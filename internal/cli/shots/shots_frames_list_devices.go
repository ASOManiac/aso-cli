package shots

import (
	"context"
	"flag"

	"github.com/ASOManiac/aso-cli/internal/cli/shared"
	"github.com/ASOManiac/aso-cli/internal/screenshots"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type frameDevicesOutput struct {
	Default string                          `json:"default"`
	Devices []screenshots.FrameDeviceOption `json:"devices"`
}

// ShotsFramesListDevicesCommand returns the screenshots list-frame-devices subcommand.
func ShotsFramesListDevicesCommand() *ffcli.Command {
	fs := flag.NewFlagSet("list-frame-devices", flag.ExitOnError)
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "list-frame-devices",
		ShortUsage: "aso screenshots list-frame-devices [--output json]",
		ShortHelp:  "[experimental] List supported frame devices and the default.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			result := frameDevicesOutput{
				Default: string(screenshots.DefaultFrameDevice()),
				Devices: screenshots.FrameDeviceOptions(),
			}
			return shared.PrintOutput(result, *output.Output, *output.Pretty)
		},
	}
}
