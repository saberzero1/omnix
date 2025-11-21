package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/saberzero1/omnix/pkg/common"
	"github.com/saberzero1/omnix/pkg/nix"
)

var (
	runSystems  string
	runOutLink  string
	runNoLink   bool
	runGHOutput bool
)

// NewRunCmd creates the run command
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [name]",
		Short: "Run tasks from om/ directory",
		Long: `Run tasks from om/ directory with simplified configuration format.

om run loads simplified YAML configuration from the om/ directory and executes
tasks with optimized defaults for quick execution without full CI overhead.

By default, it runs om/default.yaml. You can specify a different task:
  om run update    # runs om/update.yaml
  om run deploy    # runs om/deploy.yaml

The simplified config format disables lockfile, build, and flakeCheck steps
by default for faster execution.

Example om/default.yaml:
  dir: .
  steps:
    activate-configuration:
      type: app
      name: activate
  caches:
    required:
      - https://cache.nixos.org`,
		Args: cobra.MaximumNArgs(1),
		RunE: runRun,
	}

	cmd.Flags().StringVar(&runSystems, "systems", "", "Systems to build for (comma-separated)")
	cmd.Flags().StringVarP(&runOutLink, "out-link", "o", "result", "Symlink to build results (as JSON)")
	cmd.Flags().BoolVar(&runNoLink, "no-link", false, "Do not create a symlink to build results JSON")
	cmd.Flags().BoolVar(&runGHOutput, "github-output", os.Getenv("GITHUB_ACTION") != "", "Print Github Actions log groups")

	return cmd
}

// RunConfig represents the simplified config format for om run
type RunConfig struct {
	Dir            string                 `yaml:"dir"`
	Steps          map[string]interface{} `yaml:"steps"`
	Caches         *CachesConfig          `yaml:"caches,omitempty"`
	OverrideInputs map[string]string      `yaml:"overrideInputs,omitempty"`
	Systems        []string               `yaml:"systems,omitempty"`
}

// CachesConfig represents cache configuration
type CachesConfig struct {
	Required []string `yaml:"required"`
}

func runRun(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logger := common.Logger()

	// Determine task name
	taskName := "default"
	if len(args) > 0 {
		taskName = args[0]
	}

	// Determine flake reference (default to current directory)
	flakeRef := "."

	logger.Info("Reading run config from om/ directory")

	// Get config path
	configPath, err := getConfigPath(flakeRef, taskName)
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Load the simplified config
	runConfig, err := loadRunConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config from %s: %w", configPath, err)
	}

	// Get Nix info
	logger.Info("Gathering NixInfo")
	nixInfo, err := nix.GetInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Nix info: %w", err)
	}

	fmt.Printf("\nSystem: %s\n", nixInfo.Env.OS.String())
	fmt.Printf("Nix Version: %s\n\n", nixInfo.Version.String())

	// Run custom steps
	logger.Info(fmt.Sprintf("Running task '%s' for %s", taskName, flakeRef))

	if runConfig.Steps != nil {
		for stepName, stepConfig := range runConfig.Steps {
			if err := runCustomStep(ctx, stepName, stepConfig, runConfig.Dir, flakeRef); err != nil {
				return fmt.Errorf("step %s failed: %w", stepName, err)
			}
		}
	}

	logger.Info("Success!")
	return nil
}

// runCustomStep executes a single custom step
func runCustomStep(ctx context.Context, name string, stepConfig interface{}, dir, flakeRef string) error {
	logger := common.Logger()
	logger.Info(fmt.Sprintf("Running custom step: %s", name))

	// Parse step config
	stepMap, ok := stepConfig.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid step config for %s", name)
	}

	stepType, ok := stepMap["type"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'type' for step %s", name)
	}

	switch stepType {
	case "app":
		return runAppStep(ctx, name, stepMap, dir, flakeRef)
	case "devshell":
		return runDevshellStep(ctx, name, stepMap, dir, flakeRef)
	default:
		return fmt.Errorf("unknown step type: %s", stepType)
	}
}

// runAppStep runs an app-type step
func runAppStep(ctx context.Context, stepName string, stepMap map[string]interface{}, dir, flakeRef string) error {
	logger := common.Logger()
	appName, ok := stepMap["name"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'name' for app step %s", stepName)
	}

	// Build nix run command
	args := []string{"run", fmt.Sprintf("%s#%s", flakeRef, appName), "--"}

	// Add any custom args
	if argsRaw, ok := stepMap["args"]; ok {
		if argsList, ok := argsRaw.([]interface{}); ok {
			for _, arg := range argsList {
				if argStr, ok := arg.(string); ok {
					args = append(args, argStr)
				}
			}
		}
	}

	logger.Debug(fmt.Sprintf("Running: nix %s", strings.Join(args, " ")))
	cmd := nix.NewCmd()
	_, err := cmd.Run(ctx, args...)
	return err
}

// runDevshellStep runs a devshell-type step
func runDevshellStep(ctx context.Context, stepName string, stepMap map[string]interface{}, dir, flakeRef string) error {
	logger := common.Logger()
	// Get command
	commandRaw, ok := stepMap["command"]
	if !ok {
		return fmt.Errorf("missing 'command' for devshell step %s", stepName)
	}

	commandList, ok := commandRaw.([]interface{})
	if !ok {
		return fmt.Errorf("invalid 'command' for devshell step %s", stepName)
	}

	if len(commandList) == 0 {
		return fmt.Errorf("empty 'command' for devshell step %s", stepName)
	}

	// Convert to string slice
	command := make([]string, 0, len(commandList))
	for _, item := range commandList {
		if str, ok := item.(string); ok {
			command = append(command, str)
		}
	}

	// Build nix develop command
	args := []string{"develop", fmt.Sprintf("%s#default", flakeRef), "-c"}
	args = append(args, command...)

	logger.Debug(fmt.Sprintf("Running: nix %s", strings.Join(args, " ")))
	cmd := nix.NewCmd()
	_, err := cmd.Run(ctx, args...)
	return err
}

// getConfigPath returns the path to the config file
func getConfigPath(flakeRef, taskName string) (string, error) {
	// For local flakes, construct the path
	basePath := flakeRef
	if basePath == "." {
		var err error
		basePath, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	configPath := filepath.Join(basePath, "om", fmt.Sprintf("%s.yaml", taskName))

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("config file not found: %s\nExpected om/%s.yaml to exist", configPath, taskName)
	}

	return configPath, nil
}

// loadRunConfig loads and parses the run config from a YAML file
func loadRunConfig(path string) (*RunConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var config RunConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Set defaults
	if config.Dir == "" {
		config.Dir = "."
	}

	return &config, nil
}
