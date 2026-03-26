package subscriptions

import (
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/ASOManiac/aso-cli/internal/cli/promotedpurchases"
	"github.com/ASOManiac/aso-cli/internal/cli/shared"
)

// SubscriptionsPromotedPurchasesCommand returns the canonical nested promoted purchases tree.
func SubscriptionsPromotedPurchasesCommand() *ffcli.Command {
	cmd := shared.RewriteCommandTreePath(
		promotedpurchases.PromotedPurchasesCommand(),
		"aso promoted-purchases",
		"aso subscriptions promoted-purchases",
	)
	if cmd != nil {
		promotedpurchases.ConfigureScopedPromotedPurchasesCommand(cmd, promotedpurchases.ScopedPromotedPurchasesCommandConfig{
			PathPrefix:      "aso subscriptions promoted-purchases",
			ProductType:     "SUBSCRIPTION",
			ProductSingular: "a subscription",
			ProductPlural:   "subscriptions",
			RootShortHelp:   "Manage promoted purchases for subscriptions.",
			RootLongHelp: `Manage promoted purchases for subscriptions.

Only promoted purchases attached to subscriptions are listed or modified.
Link operations preserve any in-app purchase promoted purchases already
attached to the app.

Examples:
  aso subscriptions promoted-purchases list --app "APP_ID"
  aso subscriptions promoted-purchases get --promoted-purchase-id "PROMO_ID"
  aso subscriptions promoted-purchases create --app "APP_ID" --product-id "SUB_ID" --visible-for-all-users true
  aso subscriptions promoted-purchases update --promoted-purchase-id "PROMO_ID" --enabled false
  aso subscriptions promoted-purchases delete --promoted-purchase-id "PROMO_ID" --confirm
  aso subscriptions promoted-purchases link --app "APP_ID" --promoted-purchase-id "PROMO_ID"`,
		})
		configureSubscriptionsPromotedPurchasesCreate(cmd)
	}
	return cmd
}

func configureSubscriptionsPromotedPurchasesCreate(cmd *ffcli.Command) {
	promotedpurchases.ConfigureFixedProductTypeCreateCommand(cmd, promotedpurchases.FixedProductTypeCreateConfig{
		ShortUsage: "aso subscriptions promoted-purchases create --app APP_ID --product-id PRODUCT_ID --visible-for-all-users",
		ShortHelp:  "Create a promoted purchase for a subscription.",
		LongHelp: `Create a promoted purchase for a subscription.

Examples:
  aso subscriptions promoted-purchases create --app "APP_ID" --product-id "SUB_ID" --visible-for-all-users true
  aso subscriptions promoted-purchases create --app "APP_ID" --product-id "SUB_ID" --visible-for-all-users true --enabled true`,
		ProductType:    "SUBSCRIPTION",
		ProductIDUsage: "Subscription ID",
	})
}
