package flake

// Flake represents all information about a Nix flake.
type Flake struct {
	// URL is the flake URL which this struct represents
	URL string
	// Outputs are the flake outputs
	Outputs *FlakeOutputs
}

// NewFlake creates a new Flake with the given URL and outputs.
func NewFlake(url string, outputs *FlakeOutputs) *Flake {
	return &Flake{
		URL:     url,
		Outputs: outputs,
	}
}

// Note: FromNix method to construct a Flake from a URL using inspect-flake
// can now be implemented using the environment variables available via:
// - GetDefaultFlakeSchemas() - returns path to flake-schemas
// - GetInspectFlake() - returns path to inspect flake
// - HasNixBuildEnvironment() - checks if these are available
//
// These values are injected at build time when building with Nix.
// When building with `go build`, they will be empty and HasNixBuildEnvironment()
// will return false.
//
// Example implementation would:
// 1. Check HasNixBuildEnvironment() - if false, return error
// 2. Use GetInspectFlake() to construct the inspect flake URL
// 3. Use Eval() to call the inspect function with appropriate inputs
// 4. Parse the result into FlakeSchemas and convert to FlakeOutputs
//
// This implementation requires integration with the nix eval functionality
// and proper handling of the flake-schemas inventory format.
