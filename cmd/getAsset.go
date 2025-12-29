package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"automelonloaderinstallergo/internal/releases"

	"github.com/spf13/cobra"
)

var (
	getAssetOwner  string
	getAssetRepo   string
	getAssetTag    string
	getAssetAsset  string
	getAssetOutput string
	getAssetToken  string
)

func newGetAssetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "getAsset",
		Short: "Download a specific release asset by tag from GitHub",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Minute)
			defer cancel()

			token := resolveToken(getAssetToken)
			out := getAssetOutput
			if out == "" {
				out = filepath.Join(".", "downloads", getAssetAsset)
			}

			src := releases.NewGitHubSource()
			if err := src.DownloadAsset(ctx, getAssetOwner, getAssetRepo, getAssetTag, getAssetAsset, out, token); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Downloaded:", out)
			return nil
		},
	}

	cmd.Flags().StringVar(&getAssetOwner, "owner", "", "GitHub repository owner (required)")
	cmd.Flags().StringVar(&getAssetRepo, "repo", "", "GitHub repository name (required)")
	cmd.Flags().StringVar(&getAssetTag, "tag", "", "GitHub release tag (required)")
	cmd.Flags().StringVar(&getAssetAsset, "asset", "", "Release asset filename (required)")
	cmd.Flags().StringVar(&getAssetOutput, "output", "", "Output path (optional; defaults to ./downloads/<asset>)")
	cmd.Flags().StringVar(&getAssetToken, "token", "", "GitHub token (optional; overrides GITHUB_TOKEN)")

	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("tag")
	_ = cmd.MarkFlagRequired("asset")

	return cmd
}

func resolveToken(flagToken string) string {
	if flagToken != "" {
		return flagToken
	}
	return os.Getenv("GITHUB_TOKEN")
}
