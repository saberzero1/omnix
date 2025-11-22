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

// Note: FromNix() method is now implemented in schema.go
// See FromNix() function for constructing a Flake from a URL using inspect-flake.
// This requires the binary to be built with Nix (HasNixBuildEnvironment() must return true).
