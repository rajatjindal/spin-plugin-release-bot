package releaser

import "github.com/google/go-github/v56/github"

// ReleaseRequest is the release request for new plugin
type ReleaseRequest struct {
	TagName            string `json:"tagName"`
	PluginName         string `json:"pluginName"`
	PluginRepo         string `json:"pluginRepo"`
	PluginOwner        string `json:"pluginOwner"`
	PluginReleaseActor string `json:"pluginReleaseActor"`
	ProcessedTemplate  string `json:"processedTemplate"`
}

// Releaser is what opens PR
type Releaser struct {
	Token                           string
	TokenEmail                      string
	TokenUserHandle                 string
	TokenUsername                   string
	UpstreamSpinPluginsRepo         string
	UpstreamSpinPluginsRepoOwner    string
	UpstreamSpinPluginsRepoCloneURL string
	LocalSpinPluginsRepo            string
	LocalSpinPluginsRepoOwner       string
	LocalSpinPluginsRepoCloneURL    string
	githubclient                    *github.Client
}
