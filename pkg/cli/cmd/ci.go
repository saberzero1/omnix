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
	"github.com/saberzero1/omnix/pkg/nix/store"
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
		ciRemoteStore    string
		ciCopyInputs     bool
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
			return runCICommand(args, ciRunParams{
				systems:        ciSystems,
				githubOutput:   ciGitHubOutput,
				includeAllDeps: ciIncludeAllDeps,
				configPath:     ciConfigPath,
				outputPath:     ciOutputPath,
				noLink:         ciNoLink,
				remoteHost:     ciRemoteHost,
				remoteStore:    ciRemoteStore,
				copyInputs:     ciCopyInputs,
				parallel:       ciParallel,
				maxConcurrency: ciMaxConcurrency,
				noUI:           ciNoUI,
			})
		},
	}

	cmd.Flags().StringSliceVar(&ciSystems, "systems", nil, "Systems to build for (e.g., x86_64-linux,aarch64-darwin)")
	cmd.Flags().BoolVar(&ciGitHubOutput, "github-output", false, "Print GitHub Actions log groups")
	cmd.Flags().BoolVar(&ciIncludeAllDeps, "include-all-dependencies", false, "Include all dependencies in results")
	cmd.Flags().StringVarP(&ciConfigPath, "config", "c", "om.yaml", "Path to om.yaml configuration file")
	cmd.Flags().StringVarP(&ciOutputPath, "out-link", "o", "result.json", "Path to output results JSON")
	cmd.Flags().BoolVar(&ciNoLink, "no-link", false, "Do not create output results file")
	cmd.Flags().StringVar(&ciRemoteHost, "remote", "", "Remote host for SSH-based builds (e.g., user@host) - uses direct SSH, no caching")
	cmd.Flags().StringVar(&ciRemoteStore, "on", "", "Remote store URI for builds with flake caching (e.g., ssh://user@host)")
	cmd.Flags().BoolVar(&ciCopyInputs, "copy-inputs", false, "Copy all flake inputs to remote store (requires --on)")
	cmd.Flags().BoolVar(&ciParallel, "parallel", false, "Run subflakes in parallel")
	cmd.Flags().IntVar(&ciMaxConcurrency, "max-concurrency", 0, "Maximum number of parallel builds (0 = unlimited)")
	cmd.Flags().BoolVar(&ciNoUI, "no-ui", false, "Disable interactive UI")

	return cmd
}

// ciRunParams holds parameters for running CI
type ciRunParams struct {
	systems        []string
	githubOutput   bool
	includeAllDeps bool
	configPath     string
	outputPath     string
	noLink         bool
	remoteHost     string
	remoteStore    string
	copyInputs     bool
	parallel       bool
	maxConcurrency int
	noUI           bool
}

// runCICommand executes the CI command with the given parameters
func runCICommand(args []string, params ciRunParams) error {
	ctx := context.Background()
	logger := common.Logger()

	// Validate flag combinations
	if params.remoteHost != "" && params.remoteStore != "" {
		return fmt.Errorf("cannot specify both --remote and --on flags")
	}

	if params.copyInputs && params.remoteStore == "" {
		return fmt.Errorf("--copy-inputs requires --on to be specified")
	}

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

	// Extract config name from flake URL attribute
	configName := extractConfigName(flake)

	// Load configuration
	config, err := ci.LoadConfig(params.configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get the specific config by name (defaults to "default" if empty)
	subflakes, err := config.GetConfigByName(configName)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Determine systems to build for
	systems, err := determineSystems(ctx, params.systems)
	if err != nil {
		return err
	}

	// Run CI
	opts := ci.RunOptions{
		Systems:                systems,
		GitHubOutput:           params.githubOutput,
		IncludeAllDependencies: params.includeAllDeps,
		RemoteHost:             params.remoteHost,
		Parallel:               params.parallel,
		MaxConcurrency:         params.maxConcurrency,
	}

	baseFlake := flake.WithoutAttr()
	results, err := executeCIRun(ctx, baseFlake, configName, subflakes, opts, params, logger)
	if err != nil {
		return err
	}

	// Write results to file if requested
	if err := writeResults(ctx, results, params.outputPath, params.noLink, logger); err != nil {
		return err
	}

	// Check if any results failed
	if hasFailures(results) {
		return fmt.Errorf("some CI steps failed")
	}

	// Show success message if not using UI
	if params.noUI || params.githubOutput || !common.IsTerminal() {
		logger.Info("✅ All CI steps passed")
	}
	return nil
}

// extractConfigName extracts the config name from the flake URL attribute
func extractConfigName(flake nix.FlakeURL) string {
	attr := flake.GetAttr()
	attrList := attr.AsList()
	if len(attrList) > 0 {
		return attrList[0]
	}
	return ""
}

// determineSystems determines which systems to build for
func determineSystems(ctx context.Context, systems []string) ([]string, error) {
	if len(systems) > 0 {
		return systems, nil
	}

	// Default to current system
	info, err := nix.GetInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nix info: %w", err)
	}
	return []string{info.Config.System.Value}, nil
}

// executeCIRun executes the CI run based on the configuration
func executeCIRun(
	ctx context.Context,
	baseFlake nix.FlakeURL,
	configName string,
	subflakes map[string]ci.SubflakeConfig,
	opts ci.RunOptions,
	params ciRunParams,
	logger *zap.Logger,
) ([]ci.Result, error) {
	// Check if remote store is specified (new metadata-based remote CI)
	if params.remoteStore != "" {
		return runRemoteStoreCI(ctx, baseFlake, opts, params, logger)
	}

	// Use UI if running in a terminal and not disabled
	if !params.noUI && !params.githubOutput && common.IsTerminal() {
		return ci.RunWithUI(ctx, baseFlake, configName, subflakes, opts)
	}

	// Fall back to non-UI version
	return runLocalCI(ctx, baseFlake, configName, subflakes, opts, params, logger)
}

// runRemoteStoreCI runs CI on a remote store with flake caching
func runRemoteStoreCI(
	ctx context.Context,
	baseFlake nix.FlakeURL,
	opts ci.RunOptions,
	params ciRunParams,
	logger *zap.Logger,
) ([]ci.Result, error) {
	logger.Info("Running CI on remote store with flake caching",
		zap.String("flake", baseFlake.String()),
		zap.String("remote", params.remoteStore),
		zap.Bool("copyInputs", params.copyInputs))

	storeURI, err := store.ParseURI(params.remoteStore)
	if err != nil {
		return nil, fmt.Errorf("failed to parse remote store URI: %w", err)
	}

	remoteOpts := ci.RemoteRunOptions{
		StoreURI:   storeURI,
		CopyInputs: params.copyInputs,
		OutLink:    params.outputPath,
	}
	if params.noLink {
		remoteOpts.OutLink = ""
	}

	results, err := ci.RunOnRemoteStore(ctx, baseFlake, opts, remoteOpts)
	if err != nil {
		return nil, fmt.Errorf("remote CI failed: %w", err)
	}

	return results, nil
}

// runLocalCI runs CI locally without UI
func runLocalCI(
	ctx context.Context,
	baseFlake nix.FlakeURL,
	configName string,
	subflakes map[string]ci.SubflakeConfig,
	opts ci.RunOptions,
	params ciRunParams,
	logger *zap.Logger,
) ([]ci.Result, error) {
	if configName != "" {
		logger.Info("Running CI",
			zap.String("flake", baseFlake.String()),
			zap.String("config", configName),
			zap.Strings("systems", opts.Systems))
	} else {
		logger.Info("Running CI",
			zap.String("flake", baseFlake.String()),
			zap.Strings("systems", opts.Systems))
	}

	results, err := ci.Run(ctx, baseFlake, subflakes, opts)
	if err != nil {
		return nil, fmt.Errorf("CI run failed: %w", err)
	}

	// Log results
	for _, result := range results {
		ci.LogResult(result, logger)
	}

	return results, nil
}

// runResultWrapper wraps CI results in an object structure compatible with addstringcontext.
// The addstringcontext flake.nix uses lib.mapAttrsRecursive which requires an object (attribute set),
// not an array. This structure mirrors the Rust RunResult type.
type runResultWrapper struct {
	// Result contains CI results keyed by subflake name
	Result map[string]ci.Result `json:"result"`
}

// writeResults writes CI results to a file if requested
func writeResults(ctx context.Context, results []ci.Result, outputPath string, noLink bool, logger *zap.Logger) error {
	if noLink || outputPath == "" {
		return nil
	}

	// Convert results slice to a map keyed by subflake name
	// This is required because addstringcontext's flake.nix uses lib.mapAttrsRecursive
	// which only works on attribute sets (objects), not lists (arrays)
	resultMap := make(map[string]ci.Result)
	for _, result := range results {
		resultMap[result.Subflake] = result
	}
	wrapper := runResultWrapper{Result: resultMap}

	data, err := json.MarshalIndent(wrapper, "", "  ")
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
	storePath, err := addstringcontext.AddStringContext(ctx, tmpPath, outputPath)
	if err != nil {
		// If addstringcontext fails, fall back to simple file write
		logger.Warn("Failed to use addstringcontext, writing results directly", zap.Error(err))
		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write results: %w", err)
		}
		logger.Info("Results written", zap.String("path", outputPath))
	} else {
		logger.Info("Results available",
			zap.String("storePath", storePath),
			zap.String("outLink", outputPath))
	}

	return nil
}

// hasFailures checks if any CI results failed
func hasFailures(results []ci.Result) bool {
	for _, result := range results {
		if !result.Success {
			return true
		}
	}
	return false
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
