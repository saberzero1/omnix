package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	templatepkg "github.com/saberzero1/omnix/pkg/init"
)

var (
	initTemplatePath   string
	initParams         map[string]string
	initNonInteractive bool
)

// NewInitCmd creates the init command
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [output-directory]",
		Short: "Initialize a new project from a template",
		Long: `Initialize a new project from a Nix flake template.

This command scaffolds a new project by copying a template directory and
applying parameter substitutions. You can specify parameters via command-line
flags or interactively.

Examples:
  # Initialize from a template path
  om init --template ./my-template my-project
  
  # Initialize with parameter substitution
  om init --template ./template --param name=my-app output/
  
  # Non-interactive mode (all params must be provided)
  om init --template ./template --non-interactive --param name=app output/`,
		Args: cobra.ExactArgs(1),
		RunE: runInit,
	}

	cmd.Flags().StringVar(&initTemplatePath, "template", "", "Path to the template directory (required)")
	cmd.Flags().StringToStringVar(&initParams, "param", map[string]string{}, "Template parameters (key=value)")
	cmd.Flags().BoolVar(&initNonInteractive, "non-interactive", false, "Disable interactive prompts")

	_ = cmd.MarkFlagRequired("template")

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	outputDir := args[0]

	// Check if output directory already exists
	if _, err := os.Stat(outputDir); err == nil {
		return fmt.Errorf("output directory already exists: %s", outputDir)
	}

	// Check if template path exists
	if _, err := os.Stat(initTemplatePath); os.IsNotExist(err) {
		return fmt.Errorf("template directory does not exist: %s", initTemplatePath)
	}

	// Create template
	// Note: In a full implementation, we would load template metadata from
	// the template directory (e.g., from a flake.nix or template.yaml file)
	// For now, we create a basic template
	template := &templatepkg.Template{
		Path: initTemplatePath,
	}

	// Set parameter values from command line
	if len(initParams) > 0 {
		params := make(map[string]interface{})
		for k, v := range initParams {
			params[k] = v
		}
		template.SetParamValues(params)
	}

	// TODO: In a full implementation:
	// - Load template metadata
	// - Prompt for missing parameters if interactive mode
	// - Validate all required parameters are set in non-interactive mode

	fmt.Printf("🏗️  Initializing project from template: %s\n", initTemplatePath)

	// Scaffold the template
	outPath, err := template.ScaffoldAt(ctx, outputDir)
	if err != nil {
		return fmt.Errorf("failed to scaffold template: %w", err)
	}

	absPath, _ := filepath.Abs(outPath)
	fmt.Printf("\n✅ Project initialized at: %s\n", absPath)

	// Print welcome text if available
	if template.WelcomeText != nil {
		fmt.Printf("\n%s\n", *template.WelcomeText)
	}

	return nil
}
