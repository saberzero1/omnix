package common

import (
	"errors"
	"os/exec"
)

// NixInstalled checks if Nix is installed on the system
func NixInstalled() bool {
	return WhichStrict("nix") != ""
}

// WhichStrict checks if a binary is available in the system's PATH and returns its path.
// Returns empty string if the binary is not found.
// Panics on unexpected errors.
func WhichStrict(binary string) string {
	path, err := exec.LookPath(binary)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return ""
		}
		panic("Unexpected error while searching for binary '" + binary + "': " + err.Error())
	}
	return path
}
