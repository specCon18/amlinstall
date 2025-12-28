package ghrel

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// releaseByTagResponse models only the fields of the
// GET /repos/{owner}/{repo}/releases/tags/{tag} response
// required to identify and locate release assets by name.
type releaseByTagResponse struct {
	Assets []struct {
		// Name is the filename of the release asset.
		Name string `json:"name"`

		// BrowserDownloadURL is the public URL for downloading the asset.
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// NewGitHubClient returns an HTTP client configured with a fixed,
// request-wide timeout.
func NewGitHubClient() *http.Client {
	return &http.Client{Timeout: 60 * time.Second}
}

// GetReleaseByTag fetches release metadata for a specific tag from the GitHub Releases API.
// If githubToken is provided, it is used for authentication and rate-limit relief.
func GetReleaseByTag(
	ctx context.Context,
	client *http.Client,
	owner, repo, tag, githubToken string,
) (releaseByTagResponse, error) {
	return getReleaseByTagFromBaseURL(ctx, client, "https://api.github.com", owner, repo, tag, githubToken)
}

// FindAssetDownloadURL returns the browser_download_url for an asset with the given name.
func FindAssetDownloadURL(rel releaseByTagResponse, assetName string) (string, error) {
	for _, a := range rel.Assets {
		if a.Name == assetName {
			if a.BrowserDownloadURL == "" {
				return "", fmt.Errorf("asset %q has empty browser_download_url", assetName)
			}
			return a.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("asset %q not found", assetName)
}

// DownloadToWriter streams the content at downloadURL into w.
// If githubToken is provided, an Authorization header is added to the initial request.
func DownloadToWriter(
	ctx context.Context,
	client *http.Client,
	downloadURL, githubToken string,
	w io.Writer,
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}

	if githubToken != "" {
		req.Header.Set("Authorization", "Bearer "+githubToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		return fmt.Errorf("download asset: status=%s body=%s", resp.Status, string(b))
	}

	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("stream asset: %w", err)
	}

	return nil
}

// WriteFileAtomically writes a file to outPath by writing to a temporary file in the
// destination directory and then renaming it into place.
func WriteFileAtomically(outPath string, write func(f *os.File) error) error {
	if outPath == "" {
		return fmt.Errorf("outPath is empty")
	}

	dir := filepath.Dir(outPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir: %w", err)
		}
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	// Best-effort cleanup: if we fail prior to rename, remove the temp file.
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}()

	if err := write(tmp); err != nil {
		return err
	}

	// Best-effort flush for the file contents.
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpName, outPath); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}

// DownloadReleaseAssetByTag downloads a specific asset from a GitHub release tag and writes it to outPath.
//
// If outPath is empty, assetName is used as the destination filename.
// If githubToken is provided, it is used to authenticate GitHub API requests and may
// provide rate-limit relief. The asset download uses the release's browser_download_url
// and may redirect; an Authorization header may not apply to the final redirected request.
func DownloadReleaseAssetByTag(
	ctx context.Context,
	owner, repo, tag, assetName, outPath string,
	githubToken string,
) error {
	client := NewGitHubClient()
	return downloadReleaseAssetByTagWithClient(ctx, client, "https://api.github.com", owner, repo, tag, assetName, outPath, githubToken)
}

// downloadReleaseAssetByTagWithClient is an internal seam that performs the full operation
// using a provided HTTP client and API base URL.
func downloadReleaseAssetByTagWithClient(
	ctx context.Context,
	client *http.Client,
	apiBaseURL string,
	owner, repo, tag, assetName, outPath string,
	githubToken string,
) error {
	// Validate required persistence destination early.
	// If both are empty, we cannot derive a destination filename.
	if outPath == "" && assetName == "" {
		return fmt.Errorf("outPath is empty")
	}

	// Default destination to the asset name if not provided.
	if outPath == "" {
		outPath = assetName
	}

	rel, err := getReleaseByTagFromBaseURL(ctx, client, apiBaseURL, owner, repo, tag, githubToken)
	if err != nil {
		return err
	}

	downloadURL, err := FindAssetDownloadURL(rel, assetName)
	if err != nil {
		return fmt.Errorf("resolve asset URL: %w", err)
	}

	return WriteFileAtomically(outPath, func(f *os.File) error {
		return DownloadToWriter(ctx, client, downloadURL, githubToken, f)
	})
}

// getReleaseByTagFromBaseURL fetches release metadata for a specific tag from a configurable base URL.
func getReleaseByTagFromBaseURL(
	ctx context.Context,
	client *http.Client,
	baseURL, owner, repo, tag, githubToken string,
) (releaseByTagResponse, error) {
	var rel releaseByTagResponse

	apiURL := fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", strings.TrimRight(baseURL, "/"), owner, repo, tag)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return rel, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	if githubToken != "" {
		req.Header.Set("Authorization", "Bearer "+githubToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return rel, fmt.Errorf("fetch release metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		return rel, fmt.Errorf("fetch release metadata: status=%s body=%s", resp.Status, string(b))
	}

	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return rel, fmt.Errorf("decode release JSON: %w", err)
	}

	return rel, nil
}

// GitRemoteURL returns the canonical HTTPS Git remote URL for owner/repo.
func GitRemoteURL(owner, repo string) string {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}

// GetTagsViaGit retrieves all tag names from a remote Git repository by executing:
//
//	git ls-remote --tags <remoteURL>
//
// Annotated tag dereferences ("^{}") are stripped; the resulting list is de-duplicated
// and returned in sorted order.
func GetTagsViaGit(ctx context.Context, remoteURL string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--tags", remoteURL)

	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git ls-remote failed: %w; stderr=%s", err, string(ee.Stderr))
		}
		return nil, fmt.Errorf("git ls-remote failed: %w", err)
	}

	seen := make(map[string]struct{})
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 {
			continue
		}

		ref := fields[1]
		const prefix = "refs/tags/"
		if !strings.HasPrefix(ref, prefix) {
			continue
		}

		tag := strings.TrimSuffix(strings.TrimPrefix(ref, prefix), "^{}")
		if tag == "" {
			continue
		}
		seen[tag] = struct{}{}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan git output: %w", err)
	}

	tags := make([]string, 0, len(seen))
	for t := range seen {
		tags = append(tags, t)
	}
	sort.Strings(tags)

	return tags, nil
}
