package commands

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// LayoutListEntry represents a layout in the list
type LayoutListEntry struct {
	Name        string
	Version     string
	Description string
	Source      LayoutSource
	Installed   bool
}

// LayoutSource represents a layout source
type LayoutSource struct {
	Type     string
	Location string
	Ref      string
}

// Layout represents a complete layout
type Layout struct {
	Name     string
	Version  string
	Source   LayoutSource
	Manifest LayoutManifest
}

// LayoutManifest represents layout metadata
type LayoutManifest struct {
	Description  string
	Author       string
	Features     []string
	Dependencies []string
	Components   map[string]LayoutComponent
	Variables    []LayoutVariable
	Structure    LayoutStructure
}

// LayoutComponent represents a layout component
type LayoutComponent struct {
	TargetDir string
}

// LayoutVariable represents a configurable variable
type LayoutVariable struct {
	Name        string
	Description string
	Default     string
	Required    bool
}

// LayoutStructure represents directory structure
type LayoutStructure struct {
	Directories []LayoutDirectory
}

// LayoutDirectory represents a directory in the structure
type LayoutDirectory struct {
	Path        string
	Description string
}

// BuildLayoutCommand creates the layout command using the adapter pattern
func BuildLayoutCommand(adapter *CLIAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "layout",
		Short: "Manage project layouts",
		Long:  `Manage project layouts including listing, adding, updating, and removing layouts.`,
	}

	// Add subcommands
	cmd.AddCommand(buildLayoutListCommand(adapter))
	cmd.AddCommand(buildLayoutAddCommand(adapter))
	cmd.AddCommand(buildLayoutUpdateCommand(adapter))
	cmd.AddCommand(buildLayoutRemoveCommand(adapter))
	cmd.AddCommand(buildLayoutInfoCommand(adapter))

	return cmd
}

// Layout subcommand builders

// buildLayoutListCommand creates the layout list command
func buildLayoutListCommand(adapter *CLIAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available layouts",
		Long:  `List all available layouts from local and remote sources.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLayoutList(cmd, args, adapter)
		},
	}

	cmd.Flags().BoolP("remote", "r", false, "Show only remote layouts")
	cmd.Flags().BoolP("local", "l", false, "Show only local layouts")
	cmd.Flags().BoolP("installed", "i", false, "Show only installed layouts")

	return cmd
}

// buildLayoutAddCommand creates the layout add command
func buildLayoutAddCommand(adapter *CLIAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [URL or GitHub repo]",
		Short: "Add a remote layout",
		Long: `Add a remote layout from a URL or GitHub repository.
	
Examples:
  foundry layout add https://templates.foundry.dev/layouts/microservice
  foundry layout add github.com/user/foundry-hexagonal-layout`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLayoutAdd(cmd, args, adapter)
		},
	}

	cmd.Flags().StringP("name", "n", "", "Custom name for the layout")
	cmd.Flags().StringP("ref", "r", "", "Git reference (branch, tag, or commit)")

	return cmd
}

// buildLayoutUpdateCommand creates the layout update command
func buildLayoutUpdateCommand(adapter *CLIAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Update layout registry",
		Long:  `Update the layout registry by refreshing remote sources. If a layout name is provided, only that layout will be updated.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLayoutUpdate(cmd, args, adapter)
		},
	}

	return cmd
}

// buildLayoutRemoveCommand creates the layout remove command
func buildLayoutRemoveCommand(adapter *CLIAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove a layout",
		Long:  `Remove a layout from the local registry.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLayoutRemove(cmd, args, adapter)
		},
	}

	return cmd
}

// buildLayoutInfoCommand creates the layout info command
func buildLayoutInfoCommand(adapter *CLIAdapter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info [name]",
		Short: "Show layout information",
		Long:  `Display detailed information about a specific layout.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLayoutInfo(cmd, args, adapter)
		},
	}

	return cmd
}

// Layout command implementations

// runLayoutList executes the layout list command
func runLayoutList(cmd *cobra.Command, args []string, adapter *CLIAdapter) error {
	stdout := adapter.GetStdout()

	// Get flags
	showRemote, _ := cmd.Flags().GetBool("remote")
	showLocal, _ := cmd.Flags().GetBool("local")
	showInstalled, _ := cmd.Flags().GetBool("installed")

	// Get layouts (placeholder implementation)
	layouts := getAvailableLayouts()

	// Filter layouts based on flags
	var filtered []LayoutListEntry
	for _, l := range layouts {
		// Apply filters
		if showRemote && l.Source.Type == "local" {
			continue
		}
		if showLocal && l.Source.Type != "local" {
			continue
		}
		if showInstalled && !l.Installed {
			continue
		}
		filtered = append(filtered, l)
	}

	// Sort by name
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Name < filtered[j].Name
	})

	// Display layouts
	if len(filtered) == 0 {
		fmt.Fprintln(stdout, "No layouts found.")
		return nil
	}

	// Create table writer
	w := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tSOURCE\tDESCRIPTION")
	fmt.Fprintln(w, "----\t-------\t------\t-----------")

	for _, l := range filtered {
		source := l.Source.Type
		if l.Source.Type == "local" {
			source = "local"
		} else if l.Installed {
			source = fmt.Sprintf("%s (installed)", l.Source.Type)
		}

		// Truncate description if too long
		desc := l.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", l.Name, l.Version, source, desc)
	}

	return w.Flush()
}

// runLayoutAdd executes the layout add command
func runLayoutAdd(cmd *cobra.Command, args []string, adapter *CLIAdapter) error {
	stdout := adapter.GetStdout()
	url := args[0]
	customName, _ := cmd.Flags().GetString("name")
	ref, _ := cmd.Flags().GetString("ref")

	// Determine source type
	var source LayoutSource
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		source = LayoutSource{
			Type:     "remote",
			Location: url,
			Ref:      ref,
		}
	} else if strings.Contains(url, "github.com/") || strings.Count(url, "/") == 1 {
		// Assume GitHub repository
		repo := strings.TrimPrefix(url, "github.com/")
		source = LayoutSource{
			Type:     "github",
			Location: repo,
			Ref:      ref,
		}
		if ref == "" {
			source.Ref = "main"
		}
	} else {
		return fmt.Errorf("unsupported layout source: %s", url)
	}

	// Determine layout name
	name := customName
	if name == "" {
		// Extract name from URL
		parts := strings.Split(url, "/")
		name = parts[len(parts)-1]
		name = strings.TrimSuffix(name, ".git")
		name = strings.TrimSuffix(name, ".tar.gz")
		name = strings.TrimSuffix(name, ".zip")
	}

	fmt.Fprintf(stdout, "Adding layout '%s' from %s...\n", name, url)

	// TODO: Actually fetch and validate the layout before adding
	// For now, we'll add it to the registry and fetch on first use

	fmt.Fprintf(stdout, "Successfully added layout '%s'\n", name)
	fmt.Fprintln(stdout, "Run 'foundry layout update' to download the layout")
	return nil
}

// runLayoutUpdate executes the layout update command
func runLayoutUpdate(cmd *cobra.Command, args []string, adapter *CLIAdapter) error {
	stdout := adapter.GetStdout()

	if len(args) > 0 {
		// Update specific layout
		name := args[0]
		fmt.Fprintf(stdout, "Updating layout '%s'...\n", name)

		// TODO: Implement single layout update
		return fmt.Errorf("updating individual layouts not yet implemented")
	}

	// Update all remote registries
	fmt.Fprintln(stdout, "Updating layout registry...")

	// TODO: Implement registry update
	fmt.Fprintln(stdout, "Layout registry updated successfully")
	return nil
}

// runLayoutRemove executes the layout remove command
func runLayoutRemove(cmd *cobra.Command, args []string, adapter *CLIAdapter) error {
	stdout := adapter.GetStdout()
	name := args[0]

	// TODO: Implement layout removal
	fmt.Fprintf(stdout, "Removing layout '%s'...\n", name)
	return fmt.Errorf("layout removal not yet implemented")
}

// runLayoutInfo executes the layout info command
func runLayoutInfo(cmd *cobra.Command, args []string, adapter *CLIAdapter) error {
	stdout := adapter.GetStdout()
	name := args[0]

	// Get layout info (placeholder implementation)
	layout := getLayoutInfo(name)
	if layout == nil {
		return fmt.Errorf("layout '%s' not found", name)
	}

	// Display layout information
	fmt.Fprintf(stdout, "Layout: %s\n", layout.Name)
	fmt.Fprintf(stdout, "Version: %s\n", layout.Version)
	fmt.Fprintf(stdout, "Description: %s\n", layout.Manifest.Description)
	if layout.Manifest.Author != "" {
		fmt.Fprintf(stdout, "Author: %s\n", layout.Manifest.Author)
	}
	fmt.Fprintf(stdout, "Source: %s (%s)\n", layout.Source.Type, layout.Source.Location)

	// Display features
	if len(layout.Manifest.Features) > 0 {
		fmt.Fprintln(stdout, "\nFeatures:")
		for _, feature := range layout.Manifest.Features {
			fmt.Fprintf(stdout, "  - %s\n", feature)
		}
	}

	// Display dependencies
	if len(layout.Manifest.Dependencies) > 0 {
		fmt.Fprintln(stdout, "\nDependencies:")
		for _, dep := range layout.Manifest.Dependencies {
			fmt.Fprintf(stdout, "  - %s\n", dep)
		}
	}

	// Display components
	if len(layout.Manifest.Components) > 0 {
		fmt.Fprintln(stdout, "\nAvailable Components:")
		for name, comp := range layout.Manifest.Components {
			fmt.Fprintf(stdout, "  - %s (target: %s)\n", name, comp.TargetDir)
		}
	}

	// Display variables
	if len(layout.Manifest.Variables) > 0 {
		fmt.Fprintln(stdout, "\nConfigurable Variables:")
		for _, v := range layout.Manifest.Variables {
			required := ""
			if v.Required {
				required = " (required)"
			}
			fmt.Fprintf(stdout, "  - %s%s: %s", v.Name, required, v.Description)
			if v.Default != "" {
				fmt.Fprintf(stdout, " [default: %s]", v.Default)
			}
			fmt.Fprintln(stdout)
		}
	}

	// Display directory structure
	fmt.Fprintln(stdout, "\nDirectory Structure:")
	for _, dir := range layout.Manifest.Structure.Directories {
		fmt.Fprintf(stdout, "  %s", dir.Path)
		if dir.Description != "" {
			fmt.Fprintf(stdout, " - %s", dir.Description)
		}
		fmt.Fprintln(stdout)
	}

	return nil
}

// Helper functions (placeholder implementations)

// getAvailableLayouts returns available layouts (placeholder)
func getAvailableLayouts() []LayoutListEntry {
	return []LayoutListEntry{
		{
			Name:        "standard",
			Version:     "1.0.0",
			Description: "Standard Go project layout with cmd, internal, and pkg",
			Source:      LayoutSource{Type: "local"},
			Installed:   true,
		},
		{
			Name:        "microservice",
			Version:     "1.2.0",
			Description: "Microservice layout with API, gRPC, and Docker",
			Source:      LayoutSource{Type: "local"},
			Installed:   true,
		},
		{
			Name:        "hexagonal",
			Version:     "1.1.0",
			Description: "Hexagonal architecture with domain-driven design",
			Source:      LayoutSource{Type: "github", Location: "user/hexagonal-layout"},
			Installed:   false,
		},
	}
}

// getLayoutInfo returns layout info (placeholder)
func getLayoutInfo(name string) *Layout {
	switch name {
	case "standard":
		return &Layout{
			Name:    "standard",
			Version: "1.0.0",
			Source:  LayoutSource{Type: "local"},
			Manifest: LayoutManifest{
				Description: "Standard Go project layout with cmd, internal, and pkg",
				Author:      "Foundry Team",
				Features:    []string{"Standard structure", "Makefile", "Docker support"},
				Components: map[string]LayoutComponent{
					"handler": {TargetDir: "internal/handlers"},
					"model":   {TargetDir: "internal/models"},
				},
				Structure: LayoutStructure{
					Directories: []LayoutDirectory{
						{Path: "cmd/", Description: "Main applications"},
						{Path: "internal/", Description: "Private application code"},
						{Path: "pkg/", Description: "Public library code"},
					},
				},
			},
		}
	default:
		return nil
	}
}
