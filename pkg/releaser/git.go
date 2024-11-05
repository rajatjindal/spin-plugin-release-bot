package releaser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-github/v66/github"
	"github.com/sirupsen/logrus"
)

const (
	main         = "main"
	headsMain    = "heads/" + main
	refHeadsMain = "refs/" + headsMain
)

type Manifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	raw     []byte
	sha     string
}

func (r *Releaser) getLatestManifest(ctx context.Context, name string) (*Manifest, error) {
	c, _, _, err := r.githubclient.Repositories.GetContents(
		ctx,
		r.UpstreamSpinPluginsRepoOwner,
		r.UpstreamSpinPluginsRepo,
		fmt.Sprintf("manifests/%s/%s.json", name, name),
		nil,
	)
	if err != nil {
		return nil, err
	}

	raw, err := c.GetContent()
	if err != nil {
		return nil, err
	}

	manifest := Manifest{}
	err = json.Unmarshal([]byte(raw), &manifest)
	if err != nil {
		return nil, err
	}

	// this is so we can create a versioned file for the old latest manifest
	manifest.raw = []byte(raw)
	manifest.sha = c.GetSHA()

	return &manifest, nil
}

func (r *Releaser) createBranch(ctx context.Context, request *ReleaseRequest) (string, error) {
	logrus.Debug("starting createBranch")
	branch := branchName(request.PluginName, request.TagName)

	// check if branch already exists. if it does, sync with upstream main
	existingRef, _, err := r.githubclient.Git.GetRef(ctx, r.LocalSpinPluginsRepoOwner, r.LocalSpinPluginsRepo, fmt.Sprintf("heads/%s", branch))
	if err != nil && !isNotFoundError(err) {
		return "", err
	}

	if existingRef != nil {
		_, _, err := r.githubclient.Repositories.MergeUpstream(
			ctx,
			r.LocalSpinPluginsRepoOwner,
			r.LocalSpinPluginsRepo,
			&github.RepoMergeUpstreamRequest{
				Branch: github.String(main),
			},
		)
		if err != nil {
			return "", err
		}

		return branch, nil
	}

	// branch don't exist already, lets create it from upstream main
	mainref, _, err := r.githubclient.Git.GetRef(ctx, r.UpstreamSpinPluginsRepoOwner, r.UpstreamSpinPluginsRepo, headsMain)
	if err != nil {
		return "", err
	}

	refName := fmt.Sprintf("refs/heads/%s", branch)
	branchRef := &github.Reference{
		Ref: &refName,
		Object: &github.GitObject{
			SHA: mainref.Object.SHA,
		},
	}

	_, _, err = r.githubclient.Git.CreateRef(ctx, r.LocalSpinPluginsRepoOwner, r.LocalSpinPluginsRepo, branchRef)
	if err != nil {
		return "", err
	}

	return branch, nil
}

func (r *Releaser) updateLatestManifest(ctx context.Context, branch string, req *ReleaseRequest) error {
	logrus.Debug("starting updateLatestManifest")
	// get sha of existing file
	// as this is manifest file for latest version, it should always exist
	// we get the sha here again as this may be a second run for same version of plugin
	// which means there is a possibility that we already updated this file in this branch
	// and to update it with more changes, we need to provide sha in the current branch
	path := fmt.Sprintf("manifests/%s/%s.json", req.PluginName, req.PluginName)
	existingInBranch, _, _, err := r.githubclient.Repositories.GetContents(
		ctx,
		r.LocalSpinPluginsRepoOwner,
		r.LocalSpinPluginsRepo,
		path,
		&github.RepositoryContentGetOptions{
			Ref: fmt.Sprintf("heads/%s", branch),
		},
	)
	if err != nil && !isNotFoundError(err) {
		return err
	}

	opts := &github.RepositoryContentFileOptions{
		Message: github.String("releasing new version of plugin"),
		Content: []byte(req.ProcessedTemplate), //new manifest
		Branch:  &branch,
		SHA:     existingInBranch.SHA,
	}

	//TODO: only add commit if content of file has changed
	_, _, err = r.githubclient.Repositories.CreateFile(ctx, r.LocalSpinPluginsRepoOwner, r.LocalSpinPluginsRepo, path, opts)
	if err != nil {
		return err
	}

	return nil
}

func (r *Releaser) makeExistingLatestVersioned(ctx context.Context, branch string, req *ReleaseRequest) error {
	logrus.Debug("starting makeExistingLatestVersioned")
	path := fmt.Sprintf("manifests/%s/%s.json", req.PluginName, req.PluginName)
	existingInUpstream, _, _, err := r.githubclient.Repositories.GetContents(
		ctx,
		r.UpstreamSpinPluginsRepoOwner,
		r.UpstreamSpinPluginsRepo,
		path,
		&github.RepositoryContentGetOptions{
			Ref: fmt.Sprintf(headsMain),
		},
	)
	if err != nil && !isNotFoundError(err) {
		return err
	}
	existingRawManifest, err := existingInUpstream.GetContent()
	if err != nil {
		return err
	}

	existingManifest := Manifest{}
	err = json.Unmarshal([]byte(existingRawManifest), &existingManifest)
	if err != nil {
		return err
	}

	opts := &github.RepositoryContentFileOptions{
		Message: github.String("backup current release manifest"),
		Content: []byte(existingRawManifest), //old manifest
		Branch:  &branch,
	}

	newpath := fmt.Sprintf("manifests/%s/%s@%s.json", existingManifest.Name, existingManifest.Name, existingManifest.Version)
	existingInLocal, _, _, err := r.githubclient.Repositories.GetContents(
		ctx,
		r.LocalSpinPluginsRepoOwner,
		r.LocalSpinPluginsRepo,
		newpath,
		&github.RepositoryContentGetOptions{
			Ref: fmt.Sprintf("heads/%s", branch),
		},
	)
	if err != nil && !isNotFoundError(err) {
		return err
	}
	if existingInLocal != nil {
		opts.SHA = existingInLocal.SHA
	}

	//TODO: only add commit if content of file has changed
	_, _, err = r.githubclient.Repositories.CreateFile(ctx, r.LocalSpinPluginsRepoOwner, r.LocalSpinPluginsRepo, newpath, opts)
	if err != nil {
		return err
	}

	return nil
}

func branchName(name, version string) string {
	return fmt.Sprintf("%s-%s", name, version)
}

// SubmitPR submits the PR
func (r *Releaser) submitPR(ctx context.Context, request *ReleaseRequest) (string, error) {
	logrus.Debug("starting submitPR")

	prr := &github.NewPullRequest{
		Title: r.getTitle(request),
		Head:  r.getHead(request),
		Base:  github.String(main),
		Body:  r.getPRBody(request),
	}

	fmt.Printf("creating pr with title %q, \nhead %q, \nbase %q, \nbody %q\n",
		github.Stringify(r.getTitle(request)),
		github.Stringify(r.getHead(request)),
		main,
		github.Stringify(r.getPRBody(request)),
	)

	pr, _, err := r.githubclient.PullRequests.Create(
		ctx,
		r.LocalSpinPluginsRepoOwner,
		r.LocalSpinPluginsRepo,
		prr,
	)
	if err != nil {
		return "", err
	}

	fmt.Printf("pr %q opened for releasing new version\n", pr.GetHTMLURL())
	return pr.GetHTMLURL(), nil
}

func (r *Releaser) getTitle(request *ReleaseRequest) *string {
	s := fmt.Sprintf(
		"release new version %s of %s",
		request.TagName,
		request.PluginName,
	)

	return github.String(s)
}

func (r *Releaser) getHead(request *ReleaseRequest) *string {
	branchName := branchName(request.PluginName, request.TagName)
	s := fmt.Sprintf("%s:%s:%s", r.LocalSpinPluginsRepoOwner, r.LocalSpinPluginsRepo, branchName)
	return github.String(s)
}

func (r *Releaser) getPRBody(request *ReleaseRequest) *string {
	prBody := `hey team,

I am [spin-plugin-release-bot](https://github.com/rajatjindal/spin-plugin-release-bot), and I would like to open this PR to publish version %s of %s on behalf of @%s.

Thanks,
@rajatjindal`

	s := fmt.Sprintf(prBody,
		fmt.Sprintf("`%s`", request.TagName),
		fmt.Sprintf("`%s`", request.PluginName),
		request.PluginReleaseActor,
	)

	return github.String(s)
}

func isNotFoundError(err error) bool {
	gherr, ok := err.(*github.ErrorResponse)
	if !ok {
		return false
	}

	return gherr.Response != nil && gherr.Response.StatusCode == http.StatusNotFound
}
