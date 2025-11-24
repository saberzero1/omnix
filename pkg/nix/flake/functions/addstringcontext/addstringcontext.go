package addstringcontext

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/saberzero1/omnix/pkg/nix/flake/functions"
)

// AddStringContextFn implements FlakeFn for adding string context to JSON files.
type AddStringContextFn struct{}

// FlakeURL returns the flake URL for the addstringcontext function
func (f AddStringContextFn) FlakeURL() string {
	return functions.FlakeAddStringContextURL()
}

// Init is called after reading from Nix build
func (f AddStringContextFn) Init(out *json.RawMessage) {
	// No initialization needed
}

// Input represents the input to the addstringcontext function
type Input struct {
	// JSONFile is the path to the JSON file to process
	// This should be a flake URL (e.g., "path:./file.json")
	JSONFile string `json:"jsonfile"`
}

// AddStringContext adds string context to outPaths in a JSON file.
//
// This transforms a JSON file with Nix store paths such that the resultant JSON
// file path will track those paths as dependencies. This requires use of --impure.
//
// Only values of keys called "outPaths" or "allDeps" in the JSON will be transformed.
//
// See: https://nix.dev/manual/nix/2.23/language/string-context
//
// Arguments:
//   - ctx: Context for cancellation
//   - jsonFile: Path to the JSON file to process
//   - outLink: Optional output link path. If empty, --no-link is used
//
// Returns:
//   - storePath: The store path of the transformed JSON with string context
//   - error: Any error that occurred
func AddStringContext(ctx context.Context, jsonFile string, outLink string) (string, error) {
	// We have to use relative paths to avoid a Nix issue on macOS with /tmp paths.
	jsonFileAbs, err := filepath.Abs(jsonFile)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	jsonFileDir := filepath.Dir(jsonFileAbs)
	jsonFileName := filepath.Base(jsonFileAbs)

	// Convert to relative path from the directory
	input := Input{
		JSONFile: fmt.Sprintf("path:%s", jsonFileName),
	}

	// Set up options
	opts := &functions.CallOptions{
		Impure:  true, // Required because we use builtins.storePath
		WorkDir: jsonFileDir,
	}

	// If outLink is specified, make it absolute
	if outLink != "" {
		currentDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		opts.OutLink = filepath.Join(currentDir, outLink)
	}

	fn := AddStringContextFn{}
	storePath, _, err := functions.Call(ctx, fn, opts, input)
	if err != nil {
		return "", fmt.Errorf("failed to add string context: %w", err)
	}

	return storePath, nil
}

// AddStringContextToData adds string context to JSON data.
//
// This is a convenience wrapper that writes the JSON data to a temporary file,
// processes it with AddStringContext, and returns the result path.
//
// Arguments:
//   - ctx: Context for cancellation
//   - data: The JSON data to process (should contain outPaths or allDeps fields)
//   - outLink: Optional output link path
//
// Returns:
//   - storePath: The store path of the transformed JSON with string context
//   - error: Any error that occurred
func AddStringContextToData(ctx context.Context, data any, outLink string) (string, error) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "addstringcontext-*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	// Write JSON data to file
	encoder := json.NewEncoder(tmpFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		_ = tmpFile.Close()
		return "", fmt.Errorf("failed to encode JSON: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	// Process the file
	return AddStringContext(ctx, tmpPath, outLink)
}

// TransformJSONFile reads a JSON file, transforms it with string context, and returns the new path.
//
// This is the main entry point for the addstringcontext functionality.
func TransformJSONFile(ctx context.Context, jsonPath string) (string, error) {
	return AddStringContext(ctx, jsonPath, "")
}
