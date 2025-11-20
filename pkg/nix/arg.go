package nix

import "strings"

// Args represents all arguments you can pass to the `nix` command.
// This provides a way to manage common Nix arguments in a structured way.
type Args struct {
	// ExtraExperimentalFeatures appends to the experimental-features setting of Nix
	ExtraExperimentalFeatures []string

	// ExtraAccessTokens appends to the access-tokens setting of Nix
	ExtraAccessTokens []string

	// ExtraNixArgs are additional arguments to pass through to `nix`
	// Note: Arguments irrelevant to a nix subcommand will automatically be ignored
	ExtraNixArgs []string
}

// NewArgs creates a new Args with default values
func NewArgs() *Args {
	return &Args{
		ExtraExperimentalFeatures: []string{},
		ExtraAccessTokens:         []string{},
		ExtraNixArgs:              []string{},
	}
}

// ToArgs converts this Args configuration into a list of arguments for exec.Command.
// The subcommands parameter is used to filter out nonsense arguments for specific subcommands.
func (a *Args) ToArgs(subcommands ...string) []string {
	args := []string{}

	if len(a.ExtraExperimentalFeatures) > 0 {
		args = append(args, "--extra-experimental-features")
		args = append(args, strings.Join(a.ExtraExperimentalFeatures, " "))
	}

	if len(a.ExtraAccessTokens) > 0 {
		args = append(args, "--extra-access-tokens")
		args = append(args, strings.Join(a.ExtraAccessTokens, " "))
	}

	// Clone extra args to avoid modifying the original
	extraArgs := make([]string, len(a.ExtraNixArgs))
	copy(extraArgs, a.ExtraNixArgs)

	// Remove nonsense arguments when using specific subcommands
	removeNonsenseArgs(subcommands, &extraArgs)
	args = append(args, extraArgs...)

	return args
}

// WithFlakes enables flakes on this Args configuration
func (a *Args) WithFlakes() *Args {
	a.ExtraExperimentalFeatures = append(a.ExtraExperimentalFeatures, "nix-command", "flakes")
	return a
}

// WithNixCommand enables nix-command on this Args configuration
func (a *Args) WithNixCommand() *Args {
	a.ExtraExperimentalFeatures = append(a.ExtraExperimentalFeatures, "nix-command")
	return a
}

// removeNonsenseArgs removes certain options that are not supported by all subcommands.
// For example, --rebuild is not supported by `nix develop`.
func removeNonsenseArgs(subcommands []string, args *[]string) {
	unsupported := getNonsenseOptions(subcommands)
	for option, argCount := range unsupported {
		removeArgument(args, option, argCount)
	}
}

// getNonsenseOptions returns a map of unsupported options for specific subcommands.
// The key is the option name, the value is the number of additional arguments it takes.
func getNonsenseOptions(subcommands []string) map[string]int {
	rebuildAndOverride := map[string]int{
		"--rebuild":        0,
		"--override-input": 2,
	}
	rebuild := map[string]int{"--rebuild": 0}

	key := strings.Join(subcommands, " ")
	switch key {
	case "eval":
		return rebuildAndOverride
	case "flake lock":
		return rebuildAndOverride
	case "flake check":
		return rebuild
	case "develop":
		return rebuild
	case "run":
		return rebuild
	default:
		return map[string]int{}
	}
}

// removeArgument removes all occurrences of the given argument and its following arguments.
// argCount is the number of additional arguments the option takes (e.g., 2 for --override-input).
func removeArgument(args *[]string, arg string, argCount int) {
	i := 0
	for i < len(*args) {
		if (*args)[i] == arg && i+argCount < len(*args) {
			// Remove the argument and its following arguments
			*args = append((*args)[:i], (*args)[i+argCount+1:]...)
		} else {
			i++
		}
	}
}
