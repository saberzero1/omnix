package flake

import "fmt"

const (
	// OSLinux represents the Linux operating system
	OSLinux = "linux"
	// OSDarwin represents the Darwin (macOS) operating system
	OSDarwin = "darwin"
)

// System represents the system for which a derivation will build.
// The standard systems are Linux and Darwin with different architectures,
// plus a fallback for other systems.
type System struct {
	os   string
	arch Arch
}

// Arch represents the CPU architecture in the system.
type Arch int

const (
	// ArchAarch64 represents ARM64 architecture
	ArchAarch64 Arch = iota
	// ArchX86_64 represents x86-64 (Intel/AMD) architecture
	ArchX86_64
)

// Standard system constants
var (
	// SystemLinuxAarch64 represents aarch64-linux
	SystemLinuxAarch64 = System{os: OSLinux, arch: ArchAarch64}
	// SystemLinuxX86_64 represents x86_64-linux
	SystemLinuxX86_64 = System{os: OSLinux, arch: ArchX86_64}
	// SystemDarwinX86_64 represents x86_64-darwin (Intel Mac)
	SystemDarwinX86_64 = System{os: OSDarwin, arch: ArchX86_64}
	// SystemDarwinAarch64 represents aarch64-darwin (Apple Silicon)
	SystemDarwinAarch64 = System{os: OSDarwin, arch: ArchAarch64}
)

// ParseSystem parses a system string into a System.
func ParseSystem(s string) System {
	switch s {
	case "aarch64-linux":
		return SystemLinuxAarch64
	case "x86_64-linux":
		return SystemLinuxX86_64
	case "x86_64-darwin":
		return SystemDarwinX86_64
	case "aarch64-darwin":
		return SystemDarwinAarch64
	default:
		// Unknown system - store as custom
		return System{os: s, arch: ArchX86_64} // Default to x86_64 for unknown
	}
}

// String returns the string representation of the system (e.g., "x86_64-linux").
func (s System) String() string {
	if s.os == OSLinux || s.os == OSDarwin {
		archStr := "x86_64"
		if s.arch == ArchAarch64 {
			archStr = "aarch64"
		}
		return fmt.Sprintf("%s-%s", archStr, s.os)
	}
	// Custom OS string
	return s.os
}

// HumanReadable returns a human-readable description of the system.
func (s System) HumanReadable() string {
	switch s.os {
	case OSLinux:
		return fmt.Sprintf("Linux (%s)", s.arch.HumanReadable())
	case OSDarwin:
		return fmt.Sprintf("macOS (%s)", s.arch.HumanReadable())
	default:
		return s.os
	}
}

// IsLinux returns true if this is a Linux system.
func (s System) IsLinux() bool {
	return s.os == OSLinux
}

// IsDarwin returns true if this is a Darwin (macOS) system.
func (s System) IsDarwin() bool {
	return s.os == OSDarwin
}

// GetArch returns the architecture of the system.
func (s System) GetArch() Arch {
	return s.arch
}

// HumanReadable returns a human-readable description of the architecture.
func (a Arch) HumanReadable() string {
	switch a {
	case ArchAarch64:
		return "ARM"
	case ArchX86_64:
		return "Intel"
	default:
		return "Unknown"
	}
}

// String returns the string representation of the architecture.
func (a Arch) String() string {
	switch a {
	case ArchAarch64:
		return "aarch64"
	case ArchX86_64:
		return "x86_64"
	default:
		return "unknown"
	}
}
