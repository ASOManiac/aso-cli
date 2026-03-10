package validate

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/validation"
)

type subscriptionImageStatus struct {
	HasImage   bool
	Verified   bool
	SkipReason string
}

var fetchSubscriptionsFn = fetchSubscriptions

func fetchSubscriptions(ctx context.Context, client *asc.Client, appID string) ([]validation.Subscription, error) {
	groupsCtx, groupsCancel := shared.ContextWithTimeout(ctx)
	groupsResp, err := client.GetSubscriptionGroups(groupsCtx, appID, asc.WithSubscriptionGroupsLimit(200))
	groupsCancel()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription groups: %w", err)
	}

	paginatedGroups, err := asc.PaginateAll(ctx, groupsResp, func(_ context.Context, nextURL string) (asc.PaginatedResponse, error) {
		pageCtx, pageCancel := shared.ContextWithTimeout(ctx)
		defer pageCancel()
		return client.GetSubscriptionGroups(pageCtx, appID, asc.WithSubscriptionGroupsNextURL(nextURL))
	})
	if err != nil {
		return nil, fmt.Errorf("paginate subscription groups: %w", err)
	}

	groups, ok := paginatedGroups.(*asc.SubscriptionGroupsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected subscription groups response type %T", paginatedGroups)
	}

	// Pre-fetch group localizations for all groups (needed for diagnostics).
	groupLocalizations := make(map[string][]validation.SubscriptionGroupLocalizationInfo)
	groupNames := make(map[string]string)
	for _, group := range groups.Data {
		groupID := strings.TrimSpace(group.ID)
		if groupID == "" {
			continue
		}
		groupNames[groupID] = strings.TrimSpace(group.Attributes.ReferenceName)

		glCtx, glCancel := shared.ContextWithTimeout(ctx)
		glResp, err := client.GetSubscriptionGroupLocalizations(glCtx, groupID, asc.WithSubscriptionGroupLocalizationsLimit(200))
		glCancel()
		if err != nil {
			// Non-fatal: skip group localization checks on error.
			continue
		}
		for _, loc := range glResp.Data {
			groupLocalizations[groupID] = append(groupLocalizations[groupID], validation.SubscriptionGroupLocalizationInfo{
				Locale: strings.TrimSpace(loc.Attributes.Locale),
				Name:   strings.TrimSpace(loc.Attributes.Name),
			})
		}
	}

	subscriptions := make([]validation.Subscription, 0)
	for _, group := range groups.Data {
		groupID := strings.TrimSpace(group.ID)
		if groupID == "" {
			continue
		}

		subsCtx, subsCancel := shared.ContextWithTimeout(ctx)
		subsResp, err := client.GetSubscriptions(subsCtx, groupID, asc.WithSubscriptionsLimit(200))
		subsCancel()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch subscriptions for group %s: %w", groupID, err)
		}

		paginatedSubs, err := asc.PaginateAll(ctx, subsResp, func(_ context.Context, nextURL string) (asc.PaginatedResponse, error) {
			pageCtx, pageCancel := shared.ContextWithTimeout(ctx)
			defer pageCancel()
			return client.GetSubscriptions(pageCtx, groupID, asc.WithSubscriptionsNextURL(nextURL))
		})
		if err != nil {
			return nil, fmt.Errorf("paginate subscriptions: %w", err)
		}

		subsResult, ok := paginatedSubs.(*asc.SubscriptionsResponse)
		if !ok {
			return nil, fmt.Errorf("unexpected subscriptions response type %T", paginatedSubs)
		}

		for _, sub := range subsResult.Data {
			imageStatus, err := subscriptionHasImage(ctx, client, sub.ID)
			if err != nil {
				return nil, fmt.Errorf("fetch subscription images for %s: %w", strings.TrimSpace(sub.ID), err)
			}

			attrs := sub.Attributes
			valSub := validation.Subscription{
				ID:                   sub.ID,
				Name:                 attrs.Name,
				ProductID:            attrs.ProductID,
				State:                attrs.State,
				GroupID:              groupID,
				GroupName:            groupNames[groupID],
				HasImage:             imageStatus.HasImage,
				ImageCheckSkipped:    !imageStatus.Verified,
				ImageCheckSkipReason: imageStatus.SkipReason,
				GroupLocalizations:   groupLocalizations[groupID],
			}

			// Fetch deep diagnostics only for subscriptions in MISSING_METADATA.
			if strings.EqualFold(strings.TrimSpace(attrs.State), "MISSING_METADATA") {
				valSub.Localizations = fetchSubscriptionLocalizations(ctx, client, sub.ID)
				valSub.PriceCount, valSub.PriceCheckSkipped = fetchSubscriptionPriceCount(ctx, client, sub.ID)
			}

			subscriptions = append(subscriptions, valSub)
		}
	}

	return subscriptions, nil
}

// fetchSubscriptionLocalizations fetches localization info for a subscription.
// Returns nil on error (non-fatal for diagnostics).
func fetchSubscriptionLocalizations(ctx context.Context, client *asc.Client, subscriptionID string) []validation.SubscriptionLocalizationInfo {
	reqCtx, cancel := shared.ContextWithTimeout(ctx)
	defer cancel()

	resp, err := client.GetSubscriptionLocalizations(reqCtx, strings.TrimSpace(subscriptionID), asc.WithSubscriptionLocalizationsLimit(200))
	if err != nil {
		return nil
	}

	locs := make([]validation.SubscriptionLocalizationInfo, 0, len(resp.Data))
	for _, loc := range resp.Data {
		locs = append(locs, validation.SubscriptionLocalizationInfo{
			Locale:      strings.TrimSpace(loc.Attributes.Locale),
			Name:        strings.TrimSpace(loc.Attributes.Name),
			Description: strings.TrimSpace(loc.Attributes.Description),
		})
	}
	return locs
}

// fetchSubscriptionPriceCount checks whether a subscription has any prices set.
// Returns (count, skipped). Skipped is true if the check couldn't be performed.
func fetchSubscriptionPriceCount(ctx context.Context, client *asc.Client, subscriptionID string) (int, bool) {
	reqCtx, cancel := shared.ContextWithTimeout(ctx)
	defer cancel()

	resp, err := client.GetSubscriptionPrices(reqCtx, strings.TrimSpace(subscriptionID), asc.WithSubscriptionPricesLimit(1))
	if err != nil {
		return 0, true
	}
	return len(resp.Data), false
}

func subscriptionHasImage(ctx context.Context, client *asc.Client, subscriptionID string) (subscriptionImageStatus, error) {
	requestCtx, cancel := shared.ContextWithTimeout(ctx)
	defer cancel()

	resp, err := client.GetSubscriptionImages(requestCtx, strings.TrimSpace(subscriptionID), asc.WithSubscriptionImagesLimit(1))
	if err != nil {
		if asc.IsNotFound(err) {
			return subscriptionImageStatus{Verified: true}, nil
		}
		if errors.Is(err, context.Canceled) {
			return subscriptionImageStatus{}, err
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return subscriptionImageStatus{
				Verified:   false,
				SkipReason: "Image verification was skipped because the App Store Connect image endpoint timed out",
			}, nil
		}
		if errors.Is(err, asc.ErrForbidden) || asc.IsUnauthorized(err) {
			return subscriptionImageStatus{
				Verified:   false,
				SkipReason: "Image verification was skipped because this App Store Connect account cannot read subscription image assets",
			}, nil
		}
		if asc.IsRetryable(err) {
			return subscriptionImageStatus{
				Verified:   false,
				SkipReason: "Image verification was skipped because the App Store Connect image endpoint was temporarily unavailable or rate limited",
			}, nil
		}
		var netErr net.Error
		if errors.As(err, &netErr) {
			return subscriptionImageStatus{
				Verified:   false,
				SkipReason: "Image verification was skipped because the App Store Connect image endpoint could not be reached",
			}, nil
		}
		return subscriptionImageStatus{}, err
	}

	return subscriptionImageStatus{
		HasImage: resp != nil && len(resp.Data) > 0,
		Verified: true,
	}, nil
}
