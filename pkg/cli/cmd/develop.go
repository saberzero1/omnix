package cmd

import (
	"context"
	"fmt"

	"github.com/saberzero1/omnix/pkg/common"
	"github.com/saberzero1/omnix/pkg/develop"
	"github.com/saberzero1/omnix/pkg/nix"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewDevelopCmd creates the develop command
func NewDevelopCmd() *cobra.Command {
	var developConfigPath string

	cmd := &cobra.Command{
		Use:   "develop [flake-url]",
		Short: "Set up a development environment",
		Long: `Set up a development environment for a Nix flake.

This command performs pre-shell health checks and displays a welcome message
from the project's README.

The develop workflow:
1. Pre-shell: Run health checks to ensure Nix environment is properly configured
2. Post-shell: Display project README as a welcome message

Example:
  om develop
  om develop .
  om develop github:saberzero1/omnix`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			logger := common.Logger()

			// Get flake URL (default to current directory)
			flakeURL := "."
			if len(args) > 0 {
				flakeURL = args[0]
			}

			// Parse flake URL
			flake, err := nix.ParseFlakeURL(flakeURL)
			if err != nil {
				return fmt.Errorf("failed to parse flake URL: %w", err)
			}

			logger.Info("Setting up development environment", zap.String("flake", flake.String()))

			// Load configuration
			var config develop.Config
			if developConfigPath != "" {
				config, err = develop.LoadConfig(developConfigPath)
				if err != nil {
					// If config doesn't exist, use defaults
					logger.Debug("Using default config", zap.Error(err))
					config = develop.DefaultConfig()
				}
			} else {
				config = develop.DefaultConfig()
			}

			// Create project
			project, err := develop.NewProject(ctx, flake, config)
			if err != nil {
				return fmt.Errorf("failed to create project: %w", err)
			}

			// Run develop workflow
			if err := develop.Run(ctx, project); err != nil {
				return fmt.Errorf("develop workflow failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&developConfigPath, "config", "c", "om.yaml", "Path to om.yaml configuration file")

	return cmd
}
