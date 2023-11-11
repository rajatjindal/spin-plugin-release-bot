package spin

import "os"

const (
	spinPluginsIndexRepoName  = "spin-plugins"
	spinPluginsIndexRepoOwner = "rajatjindal"
)

// GetSpinPluginsIndexRepoName returns the spin plugins repo name
func GetSpinPluginsIndexRepoName() string {
	override := os.Getenv("UPSTREAM_SPIN_PLUGINS_INDEX_REPO_NAME")
	if override != "" {
		return override
	}

	return spinPluginsIndexRepoName
}

// GetSpinPluginsIndexRepoOwner returns the spin plugins repo owner
func GetSpinPluginsIndexRepoOwner() string {
	override := os.Getenv("UPSTREAM_SPIN_PLUGINS_INDEX_REPO_OWNER")
	if override != "" {
		return override
	}

	return spinPluginsIndexRepoOwner
}
