// File: cmd/foundry/main.go
package main

import (
	"fmt"
	"os"

	"github.com/shapestone/foundry/internal/cli"
)

// Version information - set during build
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Create CLI with production dependencies and version info
	app := cli.New(
		cli.WithVersionInfo(version, commit, date),
	)

	// Execute with command line arguments (no global state manipulation)
	if err := app.Execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
