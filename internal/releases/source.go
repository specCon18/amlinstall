package releases

import "context"

// Source abstracts release/tag listing and release asset downloads.
type Source interface {
	ListTags(ctx context.Context, owner, repo, githubToken string) ([]string, error)
	DownloadAsset(ctx context.Context, owner, repo, tag, assetName, outPath, githubToken string) error
}
