// Package cmd contains all Cobra command definitions for grim-cli.
// Each file in this package declares one or more commands and registers
// them via init(), so main.go only needs to call Execute().
package cmd

import (
	"os"

	"github.com/nicolas-camacho/grim-cli/ui"
	"github.com/spf13/cobra"
)

// rootCmd is the base command. Every subcommand (add, list, del, version)
// is attached to this via AddCommand inside each file's init().
var rootCmd = &cobra.Command{
	Use:   "grim",
	Short: ui.Title.Render("grim-cli") + " — your tool description here",
	// Long is shown when the user runs `grim` or `grim --help`.
	Long: ui.Box.Render(
		ui.Title.Render("grim-cli") + "\n\n" +
			ui.Muted.Render("A CLI tool built with Go and the Charm suite."),
	),
}

// Execute is the single entry point called by main.go.
// It runs the root command and exits with code 1 on any error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Hide the auto-generated `completion` subcommand from the help output.
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
