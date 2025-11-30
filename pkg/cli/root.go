// Package cli provides the command-line interface for omnix
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/saberzero1/omnix/pkg/cli/cmd"
	"github.com/saberzero1/omnix/pkg/common"
	"github.com/saberzero1/omnix/pkg/nix"
)

var (
	// verbose flag for logging verbosity
	verbose int
	// version information
	version string
	commit  string
	// nix flags
	acceptFlakeConfig bool
	nixOptions        []string
)

// SetVersion sets the version information for the CLI
func SetVersion(v, c string) {
	version = v
	commit = c
	if commit != "dev" && len(commit) > 7 {
		commit = commit[:7]
	}
	rootCmd.Version = fmt.Sprintf("%s (commit: %s)", version, commit)
}

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
  - om develop: Manage development shells
  - om run: Run tasks from om/ directory
  - om tui: Launch terminal user interface
  - om completion: Generate shell completions`,
	// Version will be set by SetVersion()
	Version: "",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		// Setup logging based on verbosity flag
		var level common.LogLevel
		switch verbose {
		case 0:
			level = common.ErrorLevel
		case 1:
			level = common.WarnLevel
		case 2:
			level = common.InfoLevel
		case 3:
			level = common.DebugLevel
		default:
			level = common.TraceLevel
		}

		if err := common.SetupLogging(level, false); err != nil {
			return err
		}

		// Set global nix options
		nix.SetGlobalOptions(&nix.GlobalOptions{
			AcceptFlakeConfig: acceptFlakeConfig,
			ExtraArgs:         nixOptions,
		})

		return nil
	},
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().IntVarP(&verbose, "verbose", "v", 2, "verbosity level (0=error, 1=warn, 2=info, 3=debug, 4=trace)")

	// Add nix flags
	rootCmd.PersistentFlags().BoolVar(&acceptFlakeConfig, "accept-flake-config", false, "accept nix configuration from flake.nix")
	rootCmd.PersistentFlags().StringArrayVar(&nixOptions, "nix-option", nil, "additional options to pass to nix commands (can be specified multiple times)")

	// Register subcommands
	rootCmd.AddCommand(cmd.NewHealthCmd())
	rootCmd.AddCommand(cmd.NewInitCmd())
	rootCmd.AddCommand(cmd.NewShowCmd())
	rootCmd.AddCommand(cmd.NewCICmd())
	rootCmd.AddCommand(cmd.NewDevelopCmd())
	rootCmd.AddCommand(cmd.NewRunCmd())
	rootCmd.AddCommand(cmd.NewTUICmd())
	rootCmd.AddCommand(cmd.NewCompletionCmd())
}
