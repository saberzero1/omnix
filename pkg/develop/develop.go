package develop

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/juspay/omnix/pkg/common"
	"github.com/juspay/omnix/pkg/health/checks"
	"github.com/juspay/omnix/pkg/nix"
	"go.uber.org/zap"
)

// Run performs the full develop workflow: pre-shell checks and post-shell readme
func Run(ctx context.Context, project *Project) error {
	if err := RunPreShell(ctx, project); err != nil {
		return err
	}

	if err := RunPostShell(ctx, project); err != nil {
		return err
	}

	// Log the warning about shell not being invoked
	logger := common.Logger()
	logger.Warn("🚧 !!!!")
	logger.Warn("🚧 Not invoking Nix devShell (not supported yet). Please use `direnv`!")
	logger.Warn("🚧 !!!!")

	return nil
}

// RunPreShell performs health checks before entering the development shell
func RunPreShell(ctx context.Context, project *Project) error {
	logger := common.Logger()

	// Get Nix info
	info, err := nix.GetInfo(ctx)
	if err != nil {
		return fmt.Errorf("unable to gather nix info: %w", err)
	}

	// Run relevant health checks
	relevantChecks := []checks.Checkable{
		// Always run these checks
		&checks.NixVersion{MinVersion: nix.Version{Major: 2, Minor: 13, Patch: 0}},
		&checks.Rosetta{},
		&checks.MaxJobs{},
	}

	// TODO: Add cache-related checks when health.Config is available
	// For now, we'll keep it simple

	logger.Info("🏥 Running health checks...")

	hasFailures := false
	for _, checkable := range relevantChecks {
		namedChecks := checkable.Check(ctx, info)
		for _, namedCheck := range namedChecks {
			check := namedCheck.Check
			if !check.Result.IsGreen() {
				// Log the check result
				logger.Info(fmt.Sprintf("  %s: %s", check.Title, check.Result.String()))

				if check.Required {
					hasFailures = true
				}
			}
		}
	}

	if hasFailures {
		return fmt.Errorf("ERROR: Your Nix environment is not properly setup. See suggestions above, or run `om health` for details")
	}

	logger.Info("✅ Nix environment is healthy.")
	return nil
}

// RunPostShell displays the README after shell activation
func RunPostShell(ctx context.Context, project *Project) error {
	logger := common.Logger()

	// Get working directory
	dir, err := project.GetWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Get markdown content
	markdown, err := project.Config.Readme.GetMarkdown(dir)
	if err != nil {
		return err
	}

	if markdown == "" {
		// No README to display
		return nil
	}

	// Render markdown
	rendered, err := common.RenderMarkdown(markdown)
	if err != nil {
		logger.Warn("Failed to render README", zap.Error(err))
		// Don't fail the whole command if markdown rendering fails
		fmt.Println(markdown)
		return nil
	}

	fmt.Println()
	fmt.Println(rendered)

	return nil
}

// IsCachixAvailable checks if the cachix command is available
func IsCachixAvailable() bool {
	path := common.WhichStrict("cachix")
	return path != ""
}

// UseCachixCache adds a cachix cache using the `cachix use` command
func UseCachixCache(ctx context.Context, cacheName string) error {
	logger := common.Logger()
	logger.Info(fmt.Sprintf("🐦 Running `cachix use` for %s", cacheName))

	cmd := exec.CommandContext(ctx, "cachix", "use", cacheName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add cachix cache %s: %w", cacheName, err)
	}

	logger.Debug("cachix use output", zap.String("output", string(output)))
	return nil
}
