package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/saberzero1/omnix/pkg/tui"
)

// NewTUICmd creates the tui command
func NewTUICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch the terminal user interface",
		Long: `Launch an interactive terminal user interface for omnix.

The TUI provides an interactive way to:
  - View Nix system health checks
  - Explore Nix configuration and environment
  - Browse flake outputs
  - Navigate system information

Use keyboard shortcuts to navigate:
  1-4: Switch between views
  r: Refresh current view
  ?: Toggle help
  q: Quit`,
		RunE: runTUI,
	}

	return cmd
}

func runTUI(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	return tui.Run(ctx)
}
