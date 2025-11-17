// Package cli provides the command-line interface for omnix
package cli

import (
	"github.com/spf13/cobra"
	
	"github.com/juspay/omnix/pkg/cli/cmd"
)

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "om",
	Short: "omnix - Developer-friendly companion for Nix",
	Long: `omnix (om) is a developer-friendly companion tool for Nix.

It provides various commands to make working with Nix easier:
  - om health: Check the health of your Nix installation
  - om init: Initialize new projects from templates
  - om show: Display flake information
  - om ci: Run CI for Nix projects
  - om develop: Manage development shells`,
	Version: "2.0.0-alpha (Go)",
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	
	// Register subcommands
	rootCmd.AddCommand(cmd.NewHealthCmd())
	rootCmd.AddCommand(cmd.NewInitCmd())
}
