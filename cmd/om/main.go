package main

import (
	"fmt"
	"os"

	"github.com/juspay/omnix/pkg/cli"
)

var (
	// Version is the version of the omnix binary
	Version = "dev"
	// Commit is the git commit hash
	Commit = "dev"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
