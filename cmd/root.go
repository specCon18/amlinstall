package cmd

import (
	"os"

	"automelonloaderinstallergo/tui"

	"github.com/spf13/cobra"

	"automelonloaderinstallergo/config"
	"automelonloaderinstallergo/internal/logger"
)
var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "A TUI-first MelonLoader Automated Installer with sane Linux packaging.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := tui.Run(); err != nil {
			logger.Log.Error("run tui", "err", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	// NOTE: scaffolded config.Init() returns no value.
	config.Init()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(newGetTagsCmd())
	rootCmd.AddCommand(newGetAssetCmd())
}

