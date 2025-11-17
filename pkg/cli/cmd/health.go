package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	
	"github.com/juspay/omnix/pkg/health"
	"github.com/juspay/omnix/pkg/nix"
)

var (
	healthJSONOnly bool
)

// NewHealthCmd creates the health command
func NewHealthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check the health of your Nix installation",
		Long: `Check the health of your Nix installation.

This command runs various checks to ensure your Nix environment is properly
configured and meets recommended standards. It checks:
  - Nix version compatibility
  - Flakes are enabled
  - Required caches are configured
  - System-specific requirements (Rosetta on macOS, etc.)
  
The command will exit with code 0 if all required checks pass, or 1 if any
required checks fail. Non-required checks that fail will produce warnings but
won't affect the exit code.`,
		RunE: runHealth,
	}
	
	cmd.Flags().BoolVar(&healthJSONOnly, "json", false, "Output results in JSON format only")
	
	return cmd
}

func runHealth(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	
	// Get Nix installation info
	nixInfo, err := nix.GetInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Nix info: %w", err)
	}
	
	// Create health checks with default configuration
	healthChecks := health.Default()
	
	// Run all checks
	results := healthChecks.RunAllChecks(ctx, nixInfo)
	
	// Print results if not JSON-only mode
	if !healthJSONOnly {
		// Print system info banner
		fmt.Printf("🩺 Checking the health of your Nix setup\n\n")
		fmt.Printf("System: %s\n", nixInfo.Env.OS.String())
		fmt.Printf("Nix Version: %s\n\n", nixInfo.Version.String())
		
		// Print each check result
		for _, result := range results {
			if err := health.PrintCheckResult(result); err != nil {
				return fmt.Errorf("failed to print check result: %w", err)
			}
			fmt.Println()
		}
	}
	
	// Evaluate results and get exit code
	status := health.EvaluateResults(results)
	
	if healthJSONOnly {
		// TODO: Implement JSON output
		fmt.Println("{\"status\": \"not implemented\"}")
	} else {
		fmt.Println(status.SummaryMessage())
	}
	
	// Exit with appropriate code
	os.Exit(status.ExitCode())
	return nil
}
