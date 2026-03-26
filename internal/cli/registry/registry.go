package registry

import (
	"context"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/accessibility"
	"github.com/ASOManiac/aso-cli/internal/cli/account"
	asocmd "github.com/ASOManiac/aso-cli/internal/cli/aso"
	"github.com/ASOManiac/aso-cli/internal/cli/actors"
	"github.com/ASOManiac/aso-cli/internal/cli/agerating"
	"github.com/ASOManiac/aso-cli/internal/cli/agreements"
	"github.com/ASOManiac/aso-cli/internal/cli/alternativedistribution"
	"github.com/ASOManiac/aso-cli/internal/cli/analytics"
	"github.com/ASOManiac/aso-cli/internal/cli/androidiosmapping"
	"github.com/ASOManiac/aso-cli/internal/cli/app_events"
	"github.com/ASOManiac/aso-cli/internal/cli/appclips"
	"github.com/ASOManiac/aso-cli/internal/cli/apps"
	"github.com/ASOManiac/aso-cli/internal/cli/auth"
	"github.com/ASOManiac/aso-cli/internal/cli/backgroundassets"
	"github.com/ASOManiac/aso-cli/internal/cli/betaapplocalizations"
	"github.com/ASOManiac/aso-cli/internal/cli/betabuildlocalizations"
	"github.com/ASOManiac/aso-cli/internal/cli/buildbundles"
	"github.com/ASOManiac/aso-cli/internal/cli/buildlocalizations"
	"github.com/ASOManiac/aso-cli/internal/cli/builds"
	"github.com/ASOManiac/aso-cli/internal/cli/bundleids"
	"github.com/ASOManiac/aso-cli/internal/cli/categories"
	"github.com/ASOManiac/aso-cli/internal/cli/certificates"
	"github.com/ASOManiac/aso-cli/internal/cli/completion"
	"github.com/ASOManiac/aso-cli/internal/cli/crashes"
	"github.com/ASOManiac/aso-cli/internal/cli/devices"
	"github.com/ASOManiac/aso-cli/internal/cli/diffcmd"
	"github.com/ASOManiac/aso-cli/internal/cli/docs"
	"github.com/ASOManiac/aso-cli/internal/cli/encryption"
	"github.com/ASOManiac/aso-cli/internal/cli/eula"
	"github.com/ASOManiac/aso-cli/internal/cli/feedback"
	"github.com/ASOManiac/aso-cli/internal/cli/finance"
	"github.com/ASOManiac/aso-cli/internal/cli/gamecenter"
	"github.com/ASOManiac/aso-cli/internal/cli/iap"
	"github.com/ASOManiac/aso-cli/internal/cli/initcmd"
	"github.com/ASOManiac/aso-cli/internal/cli/insights"
	"github.com/ASOManiac/aso-cli/internal/cli/install"
	"github.com/ASOManiac/aso-cli/internal/cli/localizations"
	"github.com/ASOManiac/aso-cli/internal/cli/marketplace"
	"github.com/ASOManiac/aso-cli/internal/cli/merchantids"
	"github.com/ASOManiac/aso-cli/internal/cli/metadata"
	"github.com/ASOManiac/aso-cli/internal/cli/migrate"
	"github.com/ASOManiac/aso-cli/internal/cli/nominations"
	"github.com/ASOManiac/aso-cli/internal/cli/notarization"
	"github.com/ASOManiac/aso-cli/internal/cli/notify"
	"github.com/ASOManiac/aso-cli/internal/cli/passtypeids"
	"github.com/ASOManiac/aso-cli/internal/cli/performance"
	"github.com/ASOManiac/aso-cli/internal/cli/preorders"
	"github.com/ASOManiac/aso-cli/internal/cli/prerelease"
	"github.com/ASOManiac/aso-cli/internal/cli/pricing"
	"github.com/ASOManiac/aso-cli/internal/cli/productpages"
	"github.com/ASOManiac/aso-cli/internal/cli/profiles"
	"github.com/ASOManiac/aso-cli/internal/cli/publish"
	releasecmd "github.com/ASOManiac/aso-cli/internal/cli/release"
	"github.com/ASOManiac/aso-cli/internal/cli/releasenotes"
	"github.com/ASOManiac/aso-cli/internal/cli/reviews"
	"github.com/ASOManiac/aso-cli/internal/cli/routingcoverage"
	"github.com/ASOManiac/aso-cli/internal/cli/sandbox"
	"github.com/ASOManiac/aso-cli/internal/cli/schema"
	"github.com/ASOManiac/aso-cli/internal/cli/screenshots"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
	"github.com/ASOManiac/aso-cli/internal/cli/signing"
	"github.com/ASOManiac/aso-cli/internal/cli/snitch"
	"github.com/ASOManiac/aso-cli/internal/cli/status"
	"github.com/ASOManiac/aso-cli/internal/cli/submit"
	"github.com/ASOManiac/aso-cli/internal/cli/subscriptions"
	"github.com/ASOManiac/aso-cli/internal/cli/testflight"
	"github.com/ASOManiac/aso-cli/internal/cli/users"
	"github.com/ASOManiac/aso-cli/internal/cli/validate"
	"github.com/ASOManiac/aso-cli/internal/cli/versions"
	"github.com/ASOManiac/aso-cli/internal/cli/videopreviews"
	"github.com/ASOManiac/aso-cli/internal/cli/web"
	"github.com/ASOManiac/aso-cli/internal/cli/webhooks"
	"github.com/ASOManiac/aso-cli/internal/cli/workflow"
	"github.com/ASOManiac/aso-cli/internal/cli/xcode"
	"github.com/ASOManiac/aso-cli/internal/cli/xcodecloud"
)

// VersionCommand returns a version subcommand.
func VersionCommand(version string) *ffcli.Command {
	return &ffcli.Command{
		Name:       "version",
		ShortUsage: "aso version",
		ShortHelp:  "Print version information and exit.",
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			fmt.Println(version)
			return nil
		},
	}
}

// Subcommands returns all root subcommands in display order.
func Subcommands(version string) []*ffcli.Command {
	editPaths := map[string]struct{}{
		"aso age-rating set":             {},
		"aso app-setup availability set": {},
		"aso pricing availability set":   {},
	}

	subs := []*ffcli.Command{
		auth.AuthCommand(),
		auth.AuthDoctorCommand(),
		web.WebCommand(),
		account.AccountCommand(),
		install.InstallSkillsCommand(),
		initcmd.InitCommand(),
		docs.DocsCommand(),
		diffcmd.DiffCommand(),
		status.StatusCommand(),
		insights.InsightsCommand(),
		releasenotes.ReleaseNotesCommand(),
		feedback.FeedbackCommand(),
		crashes.CrashesCommand(),
		reviews.ReviewsCommand(),
		reviews.ReviewCommand(),
		analytics.AnalyticsCommand(),
		performance.PerformanceCommand(),
		finance.FinanceCommand(),
		apps.AppsCommand(),
		appclips.AppClipsCommand(),
		androidiosmapping.AndroidIosMappingCommand(),
		apps.AppSetupCommand(),
		apps.AppTagsCommand(),
		marketplace.MarketplaceCommand(),
		alternativedistribution.Command(),
		webhooks.WebhooksCommand(),
		nominations.NominationsCommand(),
		bundleids.BundleIDsCommand(),
		merchantids.MerchantIDsCommand(),
		certificates.CertificatesCommand(),
		passtypeids.PassTypeIDsCommand(),
		profiles.ProfilesCommand(),
		users.UsersCommand(),
		actors.ActorsCommand(),
		devices.DevicesCommand(),
		testflight.TestFlightCommand(),
		builds.BuildsCommand(),
		buildbundles.BuildBundlesCommand(),
		publish.PublishCommand(),
		releasecmd.ReleaseCommand(),
		workflow.WorkflowCommand(),
		xcode.XcodeCommand(),
		versions.VersionsCommand(),
		productpages.ProductPagesCommand(),
		routingcoverage.RoutingCoverageCommand(),
		apps.RemovedAppInfoCommand(),
		apps.RemovedAppInfosCommand(),
		eula.EULACommand(),
		agreements.AgreementsCommand(),
		pricing.PricingCommand(),
		preorders.PreOrdersCommand(),
		prerelease.RemovedPreReleaseVersionsCommand(),
		localizations.LocalizationsCommand(),
		metadata.MetadataCommand(),
		screenshots.ScreenshotsCommand(),
		videopreviews.VideoPreviewsCommand(),
		backgroundassets.BackgroundAssetsCommand(),
		buildlocalizations.BuildLocalizationsCommand(),
		betaapplocalizations.DeprecatedBetaAppLocalizationsCommand(),
		betabuildlocalizations.BetaBuildLocalizationsCommand(),
		sandbox.SandboxCommand(),
		signing.SigningCommand(),
		notarization.NotarizationCommand(),
		iap.IAPCommand(),
		app_events.Command(),
		subscriptions.SubscriptionsCommand(),
		submit.SubmitCommand(),
		validate.ValidateCommand(),
		xcodecloud.XcodeCloudCommand(),
		categories.CategoriesCommand(),
		agerating.AgeRatingCommand(),
		accessibility.AccessibilityCommand(),
		encryption.EncryptionCommand(),
		migrate.MigrateCommand(),
		notify.NotifyCommand(),
		gamecenter.GameCenterCommand(),
		schema.SchemaCommand(),
		snitch.SnitchCommand(version),
		VersionCommand(version),
	}

	// Register ASO Maniac commands at root level (aso login, aso keywords, etc.)
	subs = append(subs, asocmd.ASOCommands()...)

	for i, sub := range subs {
		subs[i] = shared.NormalizeViewEditCommandTree(sub, editPaths)
	}

	subs = append(subs, completion.CompletionCommand(subs))
	return subs
}
