// Package store provides types and functions for working with the Nix store.
//
// The Nix store is where all build inputs and outputs are kept.
// This package provides types for representing store paths and URIs,
// as well as utilities for working with remote stores over SSH.
//
// # Store Paths
//
// Store paths come in two types:
//   - Derivations (.drv files) which describe how to build a package
//   - Outputs which are the results of building a derivation
//
// Example:
//
//	path := store.NewPath("/nix/store/abc123-hello-2.10")
//	if path.IsOutput() {
//	    fmt.Println("This is a build output")
//	}
//
// # Store URIs
//
// Store URIs specify where a Nix store is located. Currently,
// SSH stores are supported:
//
//	uri, err := store.ParseURI("ssh://user@example.com")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Connecting to: %s\n", uri.GetSSHURI().Host)
//
// Store URIs can include options:
//
//	uri, _ := store.ParseURI("ssh://user@example.com?copy-inputs=true")
//	if uri.GetOptions().CopyInputs {
//	    fmt.Println("Will copy flake inputs recursively")
//	}
package store
