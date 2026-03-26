package reviews

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// ReviewCommand returns the review parent command.
func ReviewCommand() *ffcli.Command {
	fs := flag.NewFlagSet("review", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "review",
		ShortUsage: "aso review <subcommand> [flags]",
		ShortHelp:  "Manage App Store review details, attachments, and submissions.",
		LongHelp: `Manage App Store review details, attachments, submissions, and items.

Examples:
  aso review details-get --id "DETAIL_ID"
  aso review details-for-version --version-id "VERSION_ID"
  aso review details-create --version-id "VERSION_ID" --contact-email "dev@example.com"
  aso review details-update --id "DETAIL_ID" --notes "Updated review notes"
  aso review attachments-list --review-detail "DETAIL_ID"
  aso review submissions-list --app "123456789"
  aso review submissions-create --app "123456789" --platform IOS
  aso review submissions-submit --id "SUBMISSION_ID" --confirm
  aso review submissions-update --id "SUBMISSION_ID" --canceled true
  aso review submissions-items-ids --id "SUBMISSION_ID"
  aso review items-get --id "ITEM_ID"
  aso review items-add --submission "SUBMISSION_ID" --item-type appStoreVersions --item-id "VERSION_ID"
  aso review items-update --id "ITEM_ID" --state READY_FOR_REVIEW
  aso review history --app "123456789"`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			ReviewDetailsGetCommand(),
			ReviewDetailsForVersionCommand(),
			ReviewDetailsCreateCommand(),
			ReviewDetailsUpdateCommand(),
			ReviewDetailsAttachmentsListCommand(),
			ReviewDetailsAttachmentsGetCommand(),
			ReviewDetailsAttachmentsUploadCommand(),
			ReviewDetailsAttachmentsDeleteCommand(),
			ReviewHistoryCommand(),
			ReviewSubmissionsListCommand(),
			ReviewSubmissionsGetCommand(),
			ReviewSubmissionsCreateCommand(),
			ReviewSubmissionsSubmitCommand(),
			ReviewSubmissionsCancelCommand(),
			ReviewSubmissionsUpdateCommand(),
			ReviewSubmissionsItemsIDsCommand(),
			ReviewItemsGetCommand(),
			ReviewItemsListCommand(),
			ReviewItemsAddCommand(),
			ReviewItemsUpdateCommand(),
			ReviewItemsRemoveCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
