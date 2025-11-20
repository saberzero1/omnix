package nix

import (
	"context"
	"fmt"

	"github.com/saberzero1/omnix/pkg/nix/store"
)

// CopyOptions contains options for the nix copy command.
type CopyOptions struct {
	// From is the URI of the store to copy from
	From *store.URI
	
	// To is the URI of the store to copy to
	To *store.URI
	
	// NoCheckSigs disables signature checking
	NoCheckSigs bool
}

// Copy copies store paths to a remote Nix store using `nix copy`.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - cmd: The Nix command executor
//   - options: Copy options (from, to, no-check-sigs)
//   - paths: The paths to copy (should be kept within Unix process argument size limits)
//
// Example:
//
//	cmd := nix.NewCmd()
//	toURI, _ := store.ParseURI("ssh://user@example.com")
//	options := nix.CopyOptions{
//	    To: toURI,
//	    NoCheckSigs: false,
//	}
//	err := nix.Copy(ctx, cmd, options, []string{"/nix/store/abc-foo", "/nix/store/xyz-bar"})
func Copy(ctx context.Context, cmd *Cmd, options CopyOptions, paths []string) error {
	args := []string{"copy", "-v"}
	
	if options.From != nil {
		args = append(args, "--from", options.From.String())
	}
	
	if options.To != nil {
		args = append(args, "--to", options.To.String())
	}
	
	if options.NoCheckSigs {
		args = append(args, "--no-check-sigs")
	}
	
	args = append(args, paths...)
	
	_, err := cmd.Run(ctx, args...)
	if err != nil {
		return fmt.Errorf("nix copy failed: %w", err)
	}
	
	return nil
}

// CopyPath is a convenience function to copy a single store path.
func CopyPath(ctx context.Context, cmd *Cmd, options CopyOptions, path string) error {
	return Copy(ctx, cmd, options, []string{path})
}
