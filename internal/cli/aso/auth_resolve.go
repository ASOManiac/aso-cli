package aso

import (
	"fmt"

	"github.com/ASOManiac/aso-cli/internal/asomaniac"
)

// requireAuth resolves the API key from env var or config file and returns
// an authenticated client. Returns an error if no key is found.
func requireAuth(configPath string) (*asomaniac.Client, error) {
	resolved := asomaniac.ResolveConfig(configPath)
	if !resolved.IsAuthenticated() {
		return nil, fmt.Errorf("not logged in. Run 'aso maniac login' to authenticate")
	}
	return asomaniac.NewClientFromConfig(resolved.Config), nil
}
