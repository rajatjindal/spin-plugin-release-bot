package releaser

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
	"github.com/google/go-github/v66/github"
	"github.com/pkg/errors"
	"github.com/rajatjindal/spin-plugin-release-bot/pkg/spin"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

func getCloneURL(owner, repo string) string {
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}

// TODO: get email, userhandle, name from token
func getUserDetails(token string) (string, string, string) {
	return "spin-plugin-release-bot", "Spin Plugin Release Bot", "rajatjindal83+spinpluginreleasebot@gmail.com"
}

// New returns new releaser object
func New(ctx context.Context, ghToken string) (*Releaser, error) {
	tokenUserHandle, tokenUsername, tokenEmail := getUserDetails(ghToken)
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ghToken})

	spinclient := spinhttp.NewClient()
	tc := oauth2.NewClient(context.WithValue(ctx, oauth2.HTTPClient, spinclient), ts)
	client := github.NewClient(tc)

	upstreamRepo, err := spin.GetSpinPluginsIndexRepoName()
	if err != nil {
		return nil, err
	}

	upstreamRepoOwner, err := spin.GetSpinPluginsIndexRepoOwner()
	if err != nil {
		return nil, err
	}

	return &Releaser{
		Token:                           ghToken,
		TokenEmail:                      tokenEmail,
		TokenUserHandle:                 tokenUserHandle,
		TokenUsername:                   tokenUsername,
		UpstreamSpinPluginsRepo:         upstreamRepo,
		UpstreamSpinPluginsRepoOwner:    upstreamRepoOwner,
		UpstreamSpinPluginsRepoCloneURL: getCloneURL(upstreamRepoOwner, upstreamRepo),
		LocalSpinPluginsRepo:            "spin-plugins",
		LocalSpinPluginsRepoOwner:       "spin-plugin-release-bot",
		LocalSpinPluginsRepoCloneURL:    "https://github.com/spin-plugin-release-bot/spin-plugins.git",

		githubclient: client,
	}, nil
}

// HandleActionWebhook handles requests from github actions
func (r *Releaser) HandleActionWebhook(w http.ResponseWriter, req *http.Request) {
	logrus.Debug("starting HandleActionWebhook")
	releaseRequest := ReleaseRequest{}
	raw, err := io.ReadAll(req.Body)
	if err != nil {
		logrus.Errorf("error when parsing release request %v", err)
		http.Error(w, errors.Wrap(err, "parsing release request").Error(), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(raw, &releaseRequest)
	if err != nil {
		logrus.Errorf("error when marshaling json %v", err)
		http.Error(w, errors.Wrap(err, "json unmarshal parsing release request").Error(), http.StatusInternalServerError)
		return
	}

	decoded, err := base64.StdEncoding.DecodeString(releaseRequest.EncodedProcessedTemplate)
	if err != nil {
		logrus.Errorf("error when base64 decoding template %v", err)
		http.Error(w, errors.Wrap(err, "json unmarshal parsing release request").Error(), http.StatusInternalServerError)
		return
	}
	releaseRequest.ProcessedTemplate = string(decoded)

	pr, err := r.Release(req.Context(), &releaseRequest)
	if err != nil {
		logrus.Errorf("error when opening pr %v", err)
		http.Error(w, errors.Wrap(err, "opening pr").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("PR %q submitted successfully", pr)))
}

// Release releases
func (r *Releaser) Release(ctx context.Context, request *ReleaseRequest) (string, error) {
	logrus.Debug("starting Releaser.Release")

	// create a branch in owned repo
	branch, err := r.createBranch(ctx, request)
	if err != nil {
		logrus.Errorf("error when creating branch %v", err)
		return "", err
	}

	// 2 changes are needed now
	// change the latest to have the current release info
	// add latest as versioned file
	err = r.updateLatestManifest(ctx, branch, request)
	if err != nil {
		logrus.Error("error when updating latest manifest %v", err)
		return "", err
	}

	err = r.makeExistingLatestVersioned(ctx, branch, request)
	logrus.Errorf("error when making existing one versioned %v", err)
	if err != nil {
		return "", err
	}

	// open PR
	return r.submitPR(ctx, request)
}
