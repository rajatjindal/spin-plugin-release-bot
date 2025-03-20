package spin

import (
	"github.com/fermyon/spin-go-sdk/variables"
)

const (
	upstreamSpinPluginsIndexRepoName  = "spin-plugins"
	upstreamSpinPluginsIndexRepoOwner = "fermyon"
)

// GetSpinPluginsIndexRepoName returns the spin plugins repo name
func GetSpinPluginsIndexRepoName() (string, error) {
	override, err := variables.Get("upstream_spin_plugins_index_repo_name")
	if err != nil {
		return "", err
	}

	if override != "" {
		return override, nil
	}

	return upstreamSpinPluginsIndexRepoName, nil
}

// GetSpinPluginsIndexRepoOwner returns the spin plugins repo owner
func GetSpinPluginsIndexRepoOwner() (string, error) {
	override, err := variables.Get("upstream_spin_plugins_index_repo_owner")
	if err != nil {
		return "", err
	}
	if override != "" {
		return override, nil
	}

	return upstreamSpinPluginsIndexRepoOwner, nil
}
