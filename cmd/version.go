package cmd

import (
	"fmt"

	"github.com/nicolas-camacho/grim-cli/ui"
	"github.com/spf13/cobra"
)

// version is the current release of grim-cli. Bump this before each release.
const version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.Banner.Render("grim-cli") + "  " + ui.Muted.Render("v"+version))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
