package nix

import "sync"

// GlobalOptions holds global nix command options that apply to all nix invocations.
// These are typically set from CLI flags and affect all nix commands.
type GlobalOptions struct {
	// AcceptFlakeConfig passes --accept-flake-config to nix commands.
	// This trusts flake.nix configuration settings like substituters and public keys.
	AcceptFlakeConfig bool

	// ExtraArgs are additional arguments to pass to all nix commands.
	// These are appended before subcommand-specific arguments.
	ExtraArgs []string
}

var (
	globalOptions = &GlobalOptions{}
	optionsMu     sync.RWMutex
)

// SetGlobalOptions sets the global nix options.
// This should be called early in program initialization (e.g., from CLI PersistentPreRunE).
// If nil is passed, an empty GlobalOptions is set instead.
func SetGlobalOptions(opts *GlobalOptions) {
	optionsMu.Lock()
	defer optionsMu.Unlock()
	if opts == nil {
		globalOptions = &GlobalOptions{}
		return
	}
	globalOptions = opts
}

// GetGlobalOptions returns a deep copy of the current global nix options.
// The returned value is safe to use without synchronization.
func GetGlobalOptions() GlobalOptions {
	optionsMu.RLock()
	defer optionsMu.RUnlock()
	if globalOptions == nil {
		return GlobalOptions{}
	}
	// Deep copy to avoid race conditions on ExtraArgs slice
	copied := *globalOptions
	if len(globalOptions.ExtraArgs) > 0 {
		copied.ExtraArgs = make([]string, len(globalOptions.ExtraArgs))
		copy(copied.ExtraArgs, globalOptions.ExtraArgs)
	}
	return copied
}

// ToArgs converts GlobalOptions to command-line arguments for nix.
func (o *GlobalOptions) ToArgs() []string {
	if o == nil {
		return []string{}
	}

	args := []string{}

	if o.AcceptFlakeConfig {
		args = append(args, "--accept-flake-config")
	}

	args = append(args, o.ExtraArgs...)

	return args
}

// NewCmdWithOptions creates a new Nix command executor with global options applied.
// This is the preferred way to create a Cmd when global options should be respected.
func NewCmdWithOptions() *Cmd {
	opts := GetGlobalOptions()
	return &Cmd{
		ExtraArgs: opts.ToArgs(),
	}
}
