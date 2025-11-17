package main

import (
	"fmt"
	"os"
)

var (
	// Version is the version of the omnix binary
	Version = "dev"
	// Commit is the git commit hash
	Commit = "dev"
)

func main() {
	fmt.Printf("omnix version %s (commit: %s)\n", Version, Commit)
	os.Exit(0)
}
