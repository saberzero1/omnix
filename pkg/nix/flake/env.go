package flake

// These variables are injected at build time by the Nix build system.
// When building with `nix build`, these will contain the paths to the
// flake-schemas and inspect flakes.
//
// When building outside of Nix (e.g., `go build`), these will be empty strings.
var (
	// defaultFlakeSchemas is the path to the default flake-schemas flake
	// Injected via: -X github.com/saberzero1/omnix/pkg/nix/flake.defaultFlakeSchemas=...
	defaultFlakeSchemas string

	// inspectFlake is the path to the inspect flake for analyzing flake outputs
	// Injected via: -X github.com/saberzero1/omnix/pkg/nix/flake.inspectFlake=...
	inspectFlake string
)

// GetDefaultFlakeSchemas returns the path to the default flake-schemas flake.
// Returns empty string if not built with Nix.
func GetDefaultFlakeSchemas() string {
	return defaultFlakeSchemas
}

// GetInspectFlake returns the path to the inspect flake.
// Returns empty string if not built with Nix.
func GetInspectFlake() string {
	return inspectFlake
}

// HasNixBuildEnvironment returns true if the binary was built with Nix
// and has access to flake-schemas and inspect flake paths.
func HasNixBuildEnvironment() bool {
	return defaultFlakeSchemas != "" && inspectFlake != ""
}
