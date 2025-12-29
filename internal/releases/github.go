package releases

import (
	"context"

	"automelonloaderinstallergo/internal/ghrel"
)

type gitHubSource struct{}

// NewGitHubSource returns a releases.Source backed by the existing internal/ghrel
// implementation.
func NewGitHubSource() Source {
	return gitHubSource{}
}

func (s gitHubSource) ListTags(ctx context.Context, owner, repo, githubToken string) ([]string, error) {
	// Tag listing uses `git ls-remote`; authentication is not currently applied.
	_ = githubToken
	remote := ghrel.GitRemoteURL(owner, repo)
	return ghrel.GetTagsViaGit(ctx, remote)
}

func (s gitHubSource) DownloadAsset(
	ctx context.Context,
	owner, repo, tag, assetName, outPath, githubToken string,
) error {
	return ghrel.DownloadReleaseAssetByTag(ctx, owner, repo, tag, assetName, outPath, githubToken)
}
