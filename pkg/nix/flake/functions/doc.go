// Package functions provides a framework for calling Nix functions defined in flakes from Go.
//
// This package implements a Go equivalent of the Rust FlakeFn trait, providing FFI-like
// capabilities to interact with Nix flakes as functions. The pattern was originally introduced
// by devour-flake and generalized here.
//
// # Pattern Overview
//
// A Nix flake can act as a "function" where:
//   - Flake inputs are function arguments
//   - The package derivation is the function body
//   - The built output is the function return value
//
// This package provides the Go adapter to work with such Nix functions using a simple API.
// You define your input and output structs in Go, implement the FlakeFn interface, and call it.
//
// # Inspiration
//
//   - devour-flake: https://github.com/srid/devour-flake
//   - inspect: https://github.com/DeterminateSystems/inspect
package functions
