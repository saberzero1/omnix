package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/saberzero1/omnix/pkg/nix"
)

// NewShowCmd creates the show command
func NewShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [FLAKE]",
		Short: "Inspect the outputs of a flake",
		Long: `Display information about the outputs of a Nix flake.

This command shows packages, devShells, apps, checks, configurations,
and other outputs available in the flake.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runShow,
	}

	return cmd
}

func runShow(cmd *cobra.Command, args []string) error {
	// Default to current directory if no flake is specified
	flakeURLStr := "."
	if len(args) > 0 {
		flakeURLStr = args[0]
	}

	flakeURL, err := nix.ParseFlakeURL(flakeURLStr)
	if err != nil {
		return fmt.Errorf("invalid flake URL: %w", err)
	}

	ctx := context.Background()
	nixCmd := nix.NewCmd()

	// Get the system configuration
	config, err := nix.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Nix config: %w", err)
	}
	system := config.System.Value

	// Get flake metadata
	metadata, err := nixCmd.FlakeShow(ctx, flakeURL)
	if err != nil {
		return fmt.Errorf("failed to show flake: %w", err)
	}

	if metadata.Outputs == nil {
		fmt.Println("No outputs found in flake")
		return nil
	}

	// Print different output types
	printFlakeOutputTable(cmd.OutOrStdout(), "📦 Packages", metadata.Outputs, []string{"packages", system}, fmt.Sprintf("nix build %s#<name>", flakeURL))
	printFlakeOutputTable(cmd.OutOrStdout(), "🐚 Devshells", metadata.Outputs, []string{"devShells", system}, fmt.Sprintf("nix develop %s#<name>", flakeURL))
	printFlakeOutputTable(cmd.OutOrStdout(), "🚀 Apps", metadata.Outputs, []string{"apps", system}, fmt.Sprintf("nix run %s#<name>", flakeURL))
	printFlakeOutputTable(cmd.OutOrStdout(), "🔍 Checks", metadata.Outputs, []string{"checks", system}, "nix flake check")
	printFlakeOutputTable(cmd.OutOrStdout(), "🐧 NixOS Configurations", metadata.Outputs, []string{"nixosConfigurations"}, fmt.Sprintf("nixos-rebuild switch --flake %s#<name>", flakeURL))
	printFlakeOutputTable(cmd.OutOrStdout(), "🍏 Darwin Configurations", metadata.Outputs, []string{"darwinConfigurations"}, fmt.Sprintf("darwin-rebuild switch --flake %s#<name>", flakeURL))
	printFlakeOutputTable(cmd.OutOrStdout(), "🔧 NixOS Modules", metadata.Outputs, []string{"nixosModules"}, "")
	printFlakeOutputTable(cmd.OutOrStdout(), "🐳 Docker Images", metadata.Outputs, []string{"dockerImages"}, fmt.Sprintf("nix build %s#dockerImages.<name>", flakeURL))
	printFlakeOutputTable(cmd.OutOrStdout(), "🎨 Overlays", metadata.Outputs, []string{"overlays"}, "")
	printFlakeOutputTable(cmd.OutOrStdout(), "📝 Templates", metadata.Outputs, []string{"templates"}, fmt.Sprintf("nix flake init -t %s#<name>", flakeURL))
	printFlakeOutputTable(cmd.OutOrStdout(), "📜 Schemas", metadata.Outputs, []string{"schemas"}, "")

	return nil
}

// printFlakeOutputTable prints a table of flake outputs
func printFlakeOutputTable(w io.Writer, title string, outputs *nix.FlakeOutputs, path []string, command string) {
	// Get the outputs at the specified path
	output := outputs.GetByPath(path...)
	if output == nil {
		return
	}

	// Get all terminal values
	values := output.GetAttrsetOfVal()
	if len(values) == 0 {
		return
	}

	// Print title
	blue := color.New(color.FgBlue, color.Bold)
	green := color.New(color.FgGreen, color.Bold)

	blue.Fprint(w, title)
	if command != "" {
		fmt.Fprint(w, " (")
		green.Fprint(w, command)
		fmt.Fprint(w, ")")
	}
	fmt.Fprintln(w)

	// Create table
	table := tablewriter.NewTable(w)
	table.Header([]string{"Name", "Description"})

	// Add rows
	for _, val := range values {
		desc := val.Val.ShortDescription
		if desc == "" {
			desc = "N/A"
		}
		table.Append([]string{val.Name, desc})
	}

	table.Render()
	fmt.Fprintln(w)
}

// isTerminal checks if the output is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
