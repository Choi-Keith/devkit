// Package cli defines the devkit command-line interface.
package cli

import (
	"github.com/spf13/cobra"
)

// Version is set via -ldflags "-X .../cli.Version=..." at build time.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "devkit",
	Short: "devkit - project scaffold generator",
	Long: `devkit scaffolds new projects from built-in templates.

Run "devkit list" to see available templates, then "devkit new" to create a project.`,
	Version: Version,
	// Runtime errors are printed by main; dumping usage on every error is noise.
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
