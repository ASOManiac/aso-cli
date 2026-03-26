package versions

import "github.com/ASOManiac/aso-cli/internal/cli/shared"

func normalizeAppStoreVersionInclude(value string) ([]string, error) {
	return shared.NormalizeSelection(value, appStoreVersionIncludeList(), "--include")
}

func appStoreVersionIncludeList() []string {
	return []string{
		"ageRatingDeclaration",
		"appStoreReviewDetail",
		"appClipDefaultExperience",
		"appStoreVersionExperiments",
		"appStoreVersionExperimentsV2",
		"appStoreVersionSubmission",
		"customerReviews",
		"routingAppCoverage",
		"alternativeDistributionPackage",
		"gameCenterAppVersion",
	}
}
