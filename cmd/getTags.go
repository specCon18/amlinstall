package cmd

import (
	"context"
	"fmt"
	"time"

	"automelonloaderinstallergo/internal/releases"

	"github.com/spf13/cobra"
)

var (
	getTagsOwner string
	getTagsRepo  string
	getTagsToken string
)

func newGetTagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "getTags",
		Short: "List tags from a remote GitHub repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			src := releases.NewGitHubSource()
			tags, err := src.ListTags(ctx, getTagsOwner, getTagsRepo, getTagsToken)
			if err != nil {
				return err
			}

			for _, t := range tags {
				fmt.Fprintln(cmd.OutOrStdout(), t)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&getTagsOwner, "owner", "", "GitHub repository owner (required)")
	cmd.Flags().StringVar(&getTagsRepo, "repo", "", "GitHub repository name (required)")
	cmd.Flags().StringVar(&getTagsToken, "token", "", "GitHub token (optional; not used for git ls-remote, provided for symmetry)")

	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")

	return cmd
}
