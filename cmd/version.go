package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of automelonlaoderinstallergo",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("automelonlaoderinstallergo ~ 0.0.1")
	},
}
