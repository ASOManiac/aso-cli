package apps

import "github.com/ASOManiac/aso-cli/internal/cli/shared"

func normalizeAppInfoInclude(value string) ([]string, error) {
	return shared.NormalizeSelection(value, appInfoIncludeList(), "--include")
}

func appInfoIncludeList() []string {
	return []string{
		"ageRatingDeclaration",
		"territoryAgeRatings",
		"primaryCategory",
		"primarySubcategoryOne",
		"primarySubcategoryTwo",
		"secondaryCategory",
		"secondarySubcategoryOne",
		"secondarySubcategoryTwo",
	}
}
