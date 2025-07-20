package main

import (
	"fmt"
	"os"

	"github.com/shapestone/foundry"
	"github.com/shapestone/foundry/internal/cli"
	"github.com/shapestone/foundry/internal/layout"
)

// Version information - set during build
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Initialize embedded templates for layout system
	layout.SetEmbeddedTemplates(foundry.Templates)

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
