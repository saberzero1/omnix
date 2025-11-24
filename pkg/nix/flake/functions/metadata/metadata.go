package metadata

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/saberzero1/omnix/pkg/nix/flake/functions"
)

// FlakeMetadataFn implements FlakeFn for retrieving flake metadata.
type FlakeMetadataFn struct{}

// FlakeURL returns the flake URL for the metadata function
func (f FlakeMetadataFn) FlakeURL() string {
	return functions.FlakeMetadataURL()
}

// Init is called after reading from Nix build
func (f FlakeMetadataFn) Init(out *Output) {
	// No initialization needed for metadata
}

// Input represents the input to the flake metadata function
type Input struct {
	// Flake is the flake to operate on
	Flake string `json:"flake"`

	// IncludeInputs transitively includes flake inputs in the result
	// NOTE: This makes evaluation more expensive
	IncludeInputs bool `json:"include-inputs"`
}

// Output represents the flake metadata output
type Output struct {
	// Flake is the store path to this flake
	Flake string `json:"flake"`

	// Inputs is the list of flake inputs (only available if IncludeInputs is true)
	Inputs []FlakeInput `json:"inputs,omitempty"`
}

// FlakeInput represents a flake input
type FlakeInput struct {
	// Name is the unique identifier for this input
	Name string `json:"name"`

	// Path is the local path to the input
	Path string `json:"path"`
}

// GetMetadata retrieves the metadata for a flake using the Nix function approach.
//
// This is an alternative to using `nix flake metadata` command. It uses a Nix flake
// to compute the metadata, which allows for more flexibility and can include
// transitive inputs if requested.
//
// Arguments:
//   - ctx: Context for cancellation
//   - input: Input specifying which flake to query and whether to include inputs
//
// Returns:
//   - storePath: The store path of the built metadata JSON
//   - output: The parsed metadata output
//   - error: Any error that occurred
func GetMetadata(ctx context.Context, input Input) (storePath string, output Output, err error) {
	fn := FlakeMetadataFn{}
	return functions.Call(ctx, fn, nil, input)
}

// GetMetadataWithOptions retrieves the metadata for a flake with custom options.
func GetMetadataWithOptions(ctx context.Context, input Input, opts *functions.CallOptions) (storePath string, output Output, err error) {
	fn := FlakeMetadataFn{}
	return functions.Call(ctx, fn, opts, input)
}

// GetFlakePathOnly is a convenience function that only returns the flake store path.
func GetFlakePathOnly(ctx context.Context, flakeURL string) (string, error) {
	input := Input{
		Flake:         flakeURL,
		IncludeInputs: false,
	}
	_, output, err := GetMetadata(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get flake metadata: %w", err)
	}
	return output.Flake, nil
}

// GetFlakeWithInputs retrieves the flake metadata including all transitive inputs.
func GetFlakeWithInputs(ctx context.Context, flakeURL string) (Output, error) {
	input := Input{
		Flake:         flakeURL,
		IncludeInputs: true,
	}
	_, output, err := GetMetadata(ctx, input)
	if err != nil {
		return Output{}, fmt.Errorf("failed to get flake metadata with inputs: %w", err)
	}
	return output, nil
}

// FindInput searches for a specific input by name in the metadata output.
// Returns the input if found, or an error if not found.
func (o *Output) FindInput(name string) (*FlakeInput, error) {
	if o.Inputs == nil {
		return nil, fmt.Errorf("inputs not available (was IncludeInputs set to true?)")
	}

	for i := range o.Inputs {
		if o.Inputs[i].Name == name {
			return &o.Inputs[i], nil
		}
	}

	return nil, fmt.Errorf("input %q not found", name)
}

// InputPaths returns a map of input names to their store paths.
func (o *Output) InputPaths() map[string]string {
	if o.Inputs == nil {
		return nil
	}

	result := make(map[string]string, len(o.Inputs))
	for _, input := range o.Inputs {
		result[input.Name] = input.Path
	}
	return result
}

// AllPaths returns all store paths including the flake itself and all inputs.
func (o *Output) AllPaths() []string {
	paths := []string{o.Flake}
	for _, input := range o.Inputs {
		paths = append(paths, input.Path)
	}
	return paths
}

// FlakeDir returns the directory containing the flake (parent of the flake path).
func (o *Output) FlakeDir() string {
	return filepath.Dir(o.Flake)
}
