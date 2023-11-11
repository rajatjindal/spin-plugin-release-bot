package releaser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	spinhttp "github.com/fermyon/spin/sdk/go/http"
	"github.com/google/go-github/v56/github"
	"github.com/pkg/errors"
	"github.com/rajatjindal/spin-plugin-release-bot/pkg/spin"
	"golang.org/x/oauth2"
)

func getCloneURL(owner, repo string) string {
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}

// TODO: get email, userhandle, name from token
func getUserDetails(token string) (string, string, string) {
	return "rajatjindal", "Rajat Jindal", "rajatjindal83@gmail.com"
}

// New returns new releaser object
func New(ctx context.Context, ghToken string) *Releaser {
	tokenUserHandle, tokenUsername, tokenEmail := getUserDetails(ghToken)
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ghToken})

	spinclient := spinhttp.NewClient()
	tc := oauth2.NewClient(context.WithValue(ctx, oauth2.HTTPClient, spinclient), ts)
	client := github.NewClient(tc)

	return &Releaser{
		Token:                           ghToken,
		TokenEmail:                      tokenEmail,
		TokenUserHandle:                 tokenUserHandle,
		TokenUsername:                   tokenUsername,
		UpstreamSpinPluginsRepo:         spin.GetSpinPluginsIndexRepoName(),
		UpstreamSpinPluginsRepoOwner:    spin.GetSpinPluginsIndexRepoOwner(),
		UpstreamSpinPluginsRepoCloneURL: getCloneURL(spin.GetSpinPluginsIndexRepoOwner(), spin.GetSpinPluginsIndexRepoName()),
		LocalSpinPluginsRepo:            "spin-plugins",
		LocalSpinPluginsRepoOwner:       "rajatjindal",
		LocalSpinPluginsRepoCloneURL:    "https://github.com/rajatjindal/spin-plugins.git",

		githubclient: client,
	}
}

// HandleActionWebhook handles requests from github actions
func (r *Releaser) HandleActionWebhook(w http.ResponseWriter, req *http.Request) {
	releaseRequest := ReleaseRequest{}
	raw, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, errors.Wrap(err, "parsing release request").Error(), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(raw, &releaseRequest)
	if err != nil {
		http.Error(w, errors.Wrap(err, "json unmarshal parsing release request").Error(), http.StatusInternalServerError)
		return
	}

	pr, err := r.Release(req.Context(), &releaseRequest)
	if err != nil {
		http.Error(w, errors.Wrap(err, "opening pr").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("PR %q submitted successfully", pr)))
}

// Release releases
func (r *Releaser) Release(ctx context.Context, request *ReleaseRequest) (string, error) {
	// create a branch in owned repo
	branch, err := r.createBranch(ctx, request)
	if err != nil {
		return "", err
	}

	// 2 changes are needed now
	// change the latest to have the current release info
	// add latest as versioned file
	err = r.updateLatestManifest(ctx, branch, request)
	if err != nil {
		return "", err
	}

	err = r.makeExistingLatestVersioned(ctx, branch, request)
	if err != nil {
		return "", err
	}

	// open PR
	return r.submitPR(ctx, request)
}
