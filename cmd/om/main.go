package main

import (
	"os"

	"github.com/juspay/omnix/pkg/cli"
	"github.com/juspay/omnix/pkg/common"
)

var (
	// Version is the version of the omnix binary
	// Set via ldflags: -X main.Version=x.y.z
	Version = "dev"
	// Commit is the git commit hash
	// Set via ldflags: -X main.Commit=abc123
	Commit = "dev"
)

func main() {
	// Setup version information
	cli.SetVersion(Version, Commit)
	
	// Execute the CLI
	if err := cli.Execute(); err != nil {
		// Error is already printed by cobra, just exit
		// Flush any buffered logs before exiting
		_ = common.Sync()
		os.Exit(1)
	}
	
	// Flush logs on clean exit
	_ = common.Sync()
}
