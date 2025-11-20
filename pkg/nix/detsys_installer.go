package nix

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// DetSysInstaller represents the Determinate Systems nix-installer
// See: https://github.com/DeterminateSystems/nix-installer
type DetSysInstaller struct {
	Version InstallerVersion
}

// InstallerVersion represents the version of the DetSys nix-installer
type InstallerVersion struct {
	Major uint32
	Minor uint32
	Patch uint32
}

// String returns the string representation of the installer version
func (v InstallerVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// String returns a description of the DetSys installer
func (d DetSysInstaller) String() string {
	return fmt.Sprintf("DetSys nix-installer (%s)", d.Version.String())
}

// DetectDetSysInstaller detects if the DetSys nix-installer is installed
func DetectDetSysInstaller() (*DetSysInstaller, error) {
	installerPath := "/nix/nix-installer"
	
	// Check if the installer exists
	if _, err := os.Stat(installerPath); os.IsNotExist(err) {
		return nil, nil // Not installed, but not an error
	} else if err != nil {
		return nil, fmt.Errorf("failed to check installer path: %w", err)
	}

	// Get version
	version, err := getInstallerVersion(installerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get installer version: %w", err)
	}

	return &DetSysInstaller{
		Version: version,
	}, nil
}

// getInstallerVersion retrieves the version from the installer executable
func getInstallerVersion(executablePath string) (InstallerVersion, error) {
	cmd := exec.Command(executablePath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return InstallerVersion{}, fmt.Errorf("failed to execute installer: %w", err)
	}

	versionStr := strings.TrimSpace(string(output))
	return parseInstallerVersion(versionStr)
}

// parseInstallerVersion parses a version string like "0.16.1" or "nix-installer 0.16.1"
func parseInstallerVersion(s string) (InstallerVersion, error) {
	// Match version pattern: digits.digits.digits
	re := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(s)
	
	if len(matches) < 4 {
		return InstallerVersion{}, fmt.Errorf("failed to parse version from: %s", s)
	}

	major, err := strconv.ParseUint(matches[1], 10, 32)
	if err != nil {
		return InstallerVersion{}, fmt.Errorf("invalid major version: %w", err)
	}

	minor, err := strconv.ParseUint(matches[2], 10, 32)
	if err != nil {
		return InstallerVersion{}, fmt.Errorf("invalid minor version: %w", err)
	}

	patch, err := strconv.ParseUint(matches[3], 10, 32)
	if err != nil {
		return InstallerVersion{}, fmt.Errorf("invalid patch version: %w", err)
	}

	return InstallerVersion{
		Major: uint32(major),
		Minor: uint32(minor),
		Patch: uint32(patch),
	}, nil
}
