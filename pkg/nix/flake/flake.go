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
// is not implemented yet as it requires:
// 1. Environment variables DEFAULT_FLAKE_SCHEMAS and INSPECT_FLAKE (set during Nix build)
// 2. FlakeSchemas implementation with inventory parsing
// 3. Integration with the flake-schemas ecosystem
//
// This will be added in a future migration phase when the Nix build environment
// setup is complete in Go.
