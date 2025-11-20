// Package flake provides types and functions for working with Nix flakes.
//
// This package includes support for:
//   - System types (Linux, Darwin with different architectures)
//   - Flake attributes (output paths like "packages.x86_64-linux.hello")
//
// # System Types
//
// System represents the target platform for Nix derivations.
// Standard systems include Linux and Darwin (macOS) on ARM and x86_64:
//
//	sys := flake.ParseSystem("x86_64-linux")
//	fmt.Println(sys.HumanReadable()) // Output: Linux (Intel)
//	
//	if sys.IsLinux() {
//	    fmt.Println("Building for Linux")
//	}
//
// # Flake Attributes
//
// Attr represents the output attribute path in a flake URL:
//
//	attr := flake.NewAttr("packages.x86_64-linux.hello")
//	parts := attr.AsList() // ["packages", "x86_64-linux", "hello"]
//	
//	// Default attribute
//	none := flake.NoneAttr()
//	fmt.Println(none.GetName()) // Output: default
package flake
