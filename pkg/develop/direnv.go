package develop

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/saberzero1/omnix/pkg/common"
	"go.uber.org/zap"
)

// DirenvConfig configures automatic direnv setup
type DirenvConfig struct {
	// Enable controls whether to setup direnv automatically
	Enable bool `yaml:"enable" json:"enable"`
	// AllowAutomatically controls whether to run direnv allow without prompting
	AllowAutomatically bool `yaml:"allow-automatically" json:"allow-automatically"`
}

// SetupDirenv sets up direnv for a project directory
func SetupDirenv(ctx context.Context, dir string, config DirenvConfig) error {
	if !config.Enable {
		return nil
	}

	logger := common.Logger()

	// Check if direnv is installed
	if common.WhichStrict("direnv") == "" {
		logger.Warn("direnv is not installed, skipping setup")
		return nil
	}

	// Create .envrc if it doesn't exist
	envrcPath := filepath.Join(dir, ".envrc")
	if _, err := os.Stat(envrcPath); os.IsNotExist(err) {
		logger.Info("Creating .envrc file")

		envrcContent := `# Use Nix flake
use flake

# Optional: Load local environment variables
if [ -f .env ]; then
  dotenv
fi
`
		if err := os.WriteFile(envrcPath, []byte(envrcContent), 0644); err != nil {
			return fmt.Errorf("failed to create .envrc: %w", err)
		}
	}

	// Run direnv allow if configured
	if config.AllowAutomatically {
		logger.Info("Running direnv allow")
		cmd := exec.CommandContext(ctx, "direnv", "allow", dir)
		if err := cmd.Run(); err != nil {
			logger.Warn("Failed to run direnv allow", zap.Error(err))
			// Don't return error, just warn
		}
	} else {
		logger.Info("Run 'direnv allow' to enable direnv for this project")
	}

	return nil
}

// IsDirenvEnabled checks if direnv is enabled in the current directory
func IsDirenvEnabled(dir string) bool {
	envrcPath := filepath.Join(dir, ".envrc")
	_, err := os.Stat(envrcPath)
	return err == nil
}
