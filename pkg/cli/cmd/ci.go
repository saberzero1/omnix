package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/juspay/omnix/pkg/ci"
	"github.com/juspay/omnix/pkg/common"
	"github.com/juspay/omnix/pkg/nix"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	ciSystems              []string
	ciGitHubOutput         bool
	ciIncludeAllDeps       bool
	ciConfigPath           string
	ciOutputPath           string
	ciNoLink               bool
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
  om ci run github:juspay/omnix`,
		Args: cobra.MaximumNArgs(1),
		RunE: runCIRun,
	}
	
	cmd.Flags().StringSliceVar(&ciSystems, "systems", nil, "Systems to build for (e.g., x86_64-linux,aarch64-darwin)")
	cmd.Flags().BoolVar(&ciGitHubOutput, "github-output", false, "Print GitHub Actions log groups")
	cmd.Flags().BoolVar(&ciIncludeAllDeps, "include-all-dependencies", false, "Include all dependencies in results")
	cmd.Flags().StringVarP(&ciConfigPath, "config", "c", "om.yaml", "Path to om.yaml configuration file")
	cmd.Flags().StringVarP(&ciOutputPath, "out-link", "o", "result.json", "Path to output results JSON")
	cmd.Flags().BoolVar(&ciNoLink, "no-link", false, "Do not create output results file")
	
	return cmd
}

// newCIGHMatrixCmd creates the ci gh-matrix command
func newCIGHMatrixCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gh-matrix",
		Short: "Generate GitHub Actions matrix",
		Long: `Generate a GitHub Actions matrix configuration for multi-platform builds.

The matrix includes all combinations of systems and subflakes that should be built,
taking into account system whitelists and skip flags.

Example:
  om ci gh-matrix
  om ci gh-matrix --systems x86_64-linux,aarch64-darwin`,
		RunE: runCIGHMatrix,
	}
	
	cmd.Flags().StringSliceVar(&ciSystems, "systems", []string{"x86_64-linux"}, "Systems to include in matrix")
	cmd.Flags().StringVarP(&ciConfigPath, "config", "c", "om.yaml", "Path to om.yaml configuration file")
	
	return cmd
}

func runCIRun(cmd *cobra.Command, args []string) error {
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
	
	// Load configuration
	config, err := ci.LoadConfig(ciConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
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
	
	logger.Info("Running CI", zap.String("flake", flake.String()), zap.Strings("systems", systems))
	
	// Run CI
	opts := ci.RunOptions{
		Systems:                systems,
		GitHubOutput:          ciGitHubOutput,
		IncludeAllDependencies: ciIncludeAllDeps,
	}
	
	results, err := ci.Run(ctx, flake, config, opts)
	if err != nil {
		return fmt.Errorf("CI run failed: %w", err)
	}
	
	// Log results
	for _, result := range results {
		ci.LogResult(result, logger)
	}
	
	// Write results to file if requested
	if !ciNoLink && ciOutputPath != "" {
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal results: %w", err)
		}
		
		if err := os.WriteFile(ciOutputPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write results: %w", err)
		}
		
		logger.Info("Results written", zap.String("path", ciOutputPath))
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
	
	logger.Info("✅ All CI steps passed")
	return nil
}

func runCIGHMatrix(cmd *cobra.Command, args []string) error {
	// Load configuration
	config, err := ci.LoadConfig(ciConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	// Generate matrix
	matrix := ci.GenerateMatrix(ciSystems, config)
	
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
}
