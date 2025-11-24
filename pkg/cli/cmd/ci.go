package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/saberzero1/omnix/pkg/ci"
	"github.com/saberzero1/omnix/pkg/common"
	"github.com/saberzero1/omnix/pkg/nix"
	"github.com/saberzero1/omnix/pkg/nix/flake/functions/addstringcontext"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewCICmd creates the ci command
func NewCICmd() *cobra.Command {
	ciCmd := &cobra.Command{
		Use:   "ci",
		Short: "CI/CD automation for Nix projects",
		Long: `Run CI/CD pipelines for Nix flakes.

The ci command provides comprehensive CI/CD automation including:
- Building all flake outputs
- Checking flake.lock is up to date  
- Running flake checks
- Generating GitHub Actions matrices`,
	}

	// Add subcommands
	ciCmd.AddCommand(newCIRunCmd())
	ciCmd.AddCommand(newCIGHMatrixCmd())

	return ciCmd
}

// newCIRunCmd creates the ci run command
func newCIRunCmd() *cobra.Command {
	var (
		ciSystems        []string
		ciGitHubOutput   bool
		ciIncludeAllDeps bool
		ciConfigPath     string
		ciOutputPath     string
		ciNoLink         bool
		ciRemoteHost     string
		ciParallel       bool
		ciMaxConcurrency int
		ciNoUI           bool
	)

	cmd := &cobra.Command{
		Use:   "run [flake-url]",
		Short: "Run CI steps for a flake",
		Long: `Run all CI steps for a flake.

This command executes the configured CI steps from om.yaml including:
- Build step: Builds all flake outputs
- Lockfile step: Checks flake.lock is up to date
- Flake check step: Runs 'nix flake check'
- Custom steps: Execute custom commands

Example:
  om ci run
  om ci run .
  om ci run github:saberzero1/omnix`,
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

			// Extract config name from flake URL attribute (fragment after #)
			// For example, ".#switch" should use the "switch" config
			// The attribute may have multiple parts like "switch.subflake"
			// We only use the first part as the config name
			attr := flake.GetAttr()
			configName := ""
			attrList := attr.AsList()
			if len(attrList) > 0 {
				configName = attrList[0]
			}

			// Load configuration
			config, err := ci.LoadConfig(ciConfigPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get the specific config by name (defaults to "default" if empty)
			subflakes, err := config.GetConfigByName(configName)
			if err != nil {
				return fmt.Errorf("failed to get config: %w", err)
			}

			// Determine systems to build for
			systems := ciSystems
			if len(systems) == 0 {
				// Default to current system
				info, err := nix.GetInfo(ctx)
				if err != nil {
					return fmt.Errorf("failed to get nix info: %w", err)
				}
				systems = []string{info.Config.System.Value}
			}

			// Run CI
			opts := ci.RunOptions{
				Systems:                systems,
				GitHubOutput:           ciGitHubOutput,
				IncludeAllDependencies: ciIncludeAllDeps,
				RemoteHost:             ciRemoteHost,
				Parallel:               ciParallel,
				MaxConcurrency:         ciMaxConcurrency,
			}

			// We need to pass the base flake URL without the config attribute
			// because the config name is used to select subflakes, not as a flake attribute
			baseFlake := flake.WithoutAttr()

			var results []ci.Result
			// Use UI if running in a terminal and not disabled
			if !ciNoUI && !ciGitHubOutput && common.IsTerminal() {
				results, err = ci.RunWithUI(ctx, baseFlake, configName, subflakes, opts)
			} else {
				// Fall back to non-UI version
				if configName != "" {
					logger.Info("Running CI",
						zap.String("flake", flake.WithoutAttr().String()),
						zap.String("config", configName),
						zap.Strings("systems", systems))
				} else {
					logger.Info("Running CI",
						zap.String("flake", flake.String()),
						zap.Strings("systems", systems))
				}

				results, err = ci.Run(ctx, baseFlake, subflakes, opts)

				// Log results
				for _, result := range results {
					ci.LogResult(result, logger)
				}
			}

			if err != nil {
				return fmt.Errorf("CI run failed: %w", err)
			}

			// Write results to file if requested
			if !ciNoLink && ciOutputPath != "" {
				data, err := json.MarshalIndent(results, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal results: %w", err)
				}

				// Write results to a temporary file first
				tmpFile, err := os.CreateTemp("", "om-ci-results-*.json")
				if err != nil {
					return fmt.Errorf("failed to create temp file: %w", err)
				}
				tmpPath := tmpFile.Name()
				defer func() {
					_ = os.Remove(tmpPath) // Best effort cleanup
				}()

				if _, err := tmpFile.Write(data); err != nil {
					_ = tmpFile.Close()
					return fmt.Errorf("failed to write temp file: %w", err)
				}
				if err := tmpFile.Close(); err != nil {
					return fmt.Errorf("failed to close temp file: %w", err)
				}

				// Use addstringcontext to create a Nix store path that tracks dependencies
				// This ensures that the JSON file properly tracks all the store paths it references
				storePath, err := addstringcontext.AddStringContext(ctx, tmpPath, ciOutputPath)
				if err != nil {
					// If addstringcontext fails (e.g., Nix not available), fall back to simple file write
					logger.Warn("Failed to use addstringcontext, writing results directly", zap.Error(err))
					if err := os.WriteFile(ciOutputPath, data, 0644); err != nil {
						return fmt.Errorf("failed to write results: %w", err)
					}
					logger.Info("Results written", zap.String("path", ciOutputPath))
				} else {
					logger.Info("Results available",
						zap.String("storePath", storePath),
						zap.String("outLink", ciOutputPath))
				}
			}

			// Check if any results failed
			hasFailures := false
			for _, result := range results {
				if !result.Success {
					hasFailures = true
					break
				}
			}

			if hasFailures {
				return fmt.Errorf("some CI steps failed")
			}

			// Show success message if not using UI (UI shows its own success message)
			if ciNoUI || ciGitHubOutput || !common.IsTerminal() {
				logger.Info("✅ All CI steps passed")
			}
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&ciSystems, "systems", nil, "Systems to build for (e.g., x86_64-linux,aarch64-darwin)")
	cmd.Flags().BoolVar(&ciGitHubOutput, "github-output", false, "Print GitHub Actions log groups")
	cmd.Flags().BoolVar(&ciIncludeAllDeps, "include-all-dependencies", false, "Include all dependencies in results")
	cmd.Flags().StringVarP(&ciConfigPath, "config", "c", "om.yaml", "Path to om.yaml configuration file")
	cmd.Flags().StringVarP(&ciOutputPath, "out-link", "o", "result.json", "Path to output results JSON")
	cmd.Flags().BoolVar(&ciNoLink, "no-link", false, "Do not create output results file")
	cmd.Flags().StringVar(&ciRemoteHost, "remote", "", "Remote host for SSH-based builds (e.g., user@host)")
	cmd.Flags().BoolVar(&ciParallel, "parallel", false, "Run subflakes in parallel")
	cmd.Flags().IntVar(&ciMaxConcurrency, "max-concurrency", 0, "Maximum number of parallel builds (0 = unlimited)")
	cmd.Flags().BoolVar(&ciNoUI, "no-ui", false, "Disable interactive UI")

	return cmd
}

// newCIGHMatrixCmd creates the ci gh-matrix command
func newCIGHMatrixCmd() *cobra.Command {
	var (
		ciSystems    []string
		ciConfigPath string
	)

	cmd := &cobra.Command{
		Use:   "gh-matrix",
		Short: "Generate GitHub Actions matrix",
		Long: `Generate a GitHub Actions matrix configuration for multi-platform builds.

The matrix includes all combinations of systems and subflakes that should be built,
taking into account system whitelists and skip flags.

Example:
  om ci gh-matrix
  om ci gh-matrix --systems x86_64-linux,aarch64-darwin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			config, err := ci.LoadConfig(ciConfigPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Generate matrix
			matrix, err := ci.GenerateMatrix(ciSystems, config)
			if err != nil {
				return fmt.Errorf("failed to generate matrix: %w", err)
			}

			// Convert to JSON
			jsonOutput, err := matrix.ToJSON()
			if err != nil {
				return fmt.Errorf("failed to generate JSON: %w", err)
			}

			// Print to stdout
			fmt.Println(jsonOutput)

			// Log summary
			logger := common.Logger()
			logger.Info("Generated matrix",
				zap.Int("rows", matrix.Count()),
				zap.Strings("systems", ciSystems))

			return nil
		},
	}

	cmd.Flags().StringSliceVar(&ciSystems, "systems", []string{"x86_64-linux"}, "Systems to include in matrix")
	cmd.Flags().StringVarP(&ciConfigPath, "config", "c", "om.yaml", "Path to om.yaml configuration file")

	return cmd
}
