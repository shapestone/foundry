package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/shapestone/foundry/internal/layout"
	"github.com/spf13/cobra"
)

var layoutCmd = &cobra.Command{
	Use:   "layout",
	Short: "Manage project layouts",
	Long:  `Manage project layouts including listing, adding, updating, and removing layouts.`,
}

var layoutListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available layouts",
	Long:  `List all available layouts from local and remote sources.`,
	RunE:  runLayoutList,
}

var layoutAddCmd = &cobra.Command{
	Use:   "add [URL or GitHub repo]",
	Short: "Add a remote layout",
	Long: `Add a remote layout from a URL or GitHub repository.
	
Examples:
  foundry layout add https://templates.foundry.dev/layouts/microservice
  foundry layout add github.com/user/foundry-hexagonal-layout`,
	Args: cobra.ExactArgs(1),
	RunE: runLayoutAdd,
}

var layoutUpdateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update layout registry",
	Long:  `Update the layout registry by refreshing remote sources. If a layout name is provided, only that layout will be updated.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runLayoutUpdate,
}

var layoutRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a layout",
	Long:  `Remove a layout from the local registry.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runLayoutRemove,
}

var layoutInfoCmd = &cobra.Command{
	Use:   "info [name]",
	Short: "Show layout information",
	Long:  `Display detailed information about a specific layout.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runLayoutInfo,
}

func init() {
	// Add subcommands
	layoutCmd.AddCommand(layoutListCmd)
	layoutCmd.AddCommand(layoutAddCmd)
	layoutCmd.AddCommand(layoutUpdateCmd)
	layoutCmd.AddCommand(layoutRemoveCmd)
	layoutCmd.AddCommand(layoutInfoCmd)

	// Add flags
	layoutListCmd.Flags().BoolP("remote", "r", false, "Show only remote layouts")
	layoutListCmd.Flags().BoolP("local", "l", false, "Show only local layouts")
	layoutListCmd.Flags().BoolP("installed", "i", false, "Show only installed layouts")

	layoutAddCmd.Flags().StringP("name", "n", "", "Custom name for the layout")
	layoutAddCmd.Flags().StringP("ref", "r", "", "Git reference (branch, tag, or commit)")

	// Register with root command
	rootCmd.AddCommand(layoutCmd)
}

func runLayoutList(cmd *cobra.Command, args []string) error {
	// Get flags
	showRemote, _ := cmd.Flags().GetBool("remote")
	showLocal, _ := cmd.Flags().GetBool("local")
	showInstalled, _ := cmd.Flags().GetBool("installed")

	// Create layout manager
	manager, err := createLayoutManager()
	if err != nil {
		return fmt.Errorf("failed to create layout manager: %w", err)
	}

	// Get all layouts
	layouts := manager.ListLayouts()

	// Filter layouts based on flags
	var filtered []layout.LayoutListEntry
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
		fmt.Println("No layouts found.")
		return nil
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
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

func runLayoutAdd(cmd *cobra.Command, args []string) error {
	url := args[0]
	customName, _ := cmd.Flags().GetString("name")
	ref, _ := cmd.Flags().GetString("ref")

	// Create layout manager
	manager, err := createLayoutManager()
	if err != nil {
		return fmt.Errorf("failed to create layout manager: %w", err)
	}

	// Determine source type
	var source layout.LayoutSource
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		source = layout.LayoutSource{
			Type:     "remote",
			Location: url,
			Ref:      ref,
		}
	} else if strings.Contains(url, "github.com/") || strings.Count(url, "/") == 1 {
		// Assume GitHub repository
		repo := strings.TrimPrefix(url, "github.com/")
		source = layout.LayoutSource{
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

	fmt.Printf("Adding layout '%s' from %s...\n", name, url)

	// TODO: Actually fetch and validate the layout before adding
	// For now, we'll add it to the registry and fetch on first use

	if err := manager.AddRemoteLayout(url); err != nil {
		return fmt.Errorf("failed to add layout: %w", err)
	}

	fmt.Printf("Successfully added layout '%s'\n", name)
	fmt.Println("Run 'foundry layout update' to download the layout")
	return nil
}

func runLayoutUpdate(cmd *cobra.Command, args []string) error {
	// Create layout manager
	manager, err := createLayoutManager()
	if err != nil {
		return fmt.Errorf("failed to create layout manager: %w", err)
	}

	if len(args) > 0 {
		// Update specific layout
		name := args[0]
		fmt.Printf("Updating layout '%s'...\n", name)

		// TODO: Implement single layout update
		return fmt.Errorf("updating individual layouts not yet implemented")
	}

	// Update all remote registries
	fmt.Println("Updating layout registry...")
	if err := manager.RefreshLayouts(); err != nil {
		return fmt.Errorf("failed to update registry: %w", err)
	}

	fmt.Println("Layout registry updated successfully")
	return nil
}

func runLayoutRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Create layout manager
	manager, err := createLayoutManager()
	if err != nil {
		return fmt.Errorf("failed to create layout manager: %w", err)
	}

	// TODO: Implement layout removal
	_ = manager

	fmt.Printf("Removing layout '%s'...\n", name)
	return fmt.Errorf("layout removal not yet implemented")
}

func runLayoutInfo(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Create layout manager
	manager, err := createLayoutManager()
	if err != nil {
		return fmt.Errorf("failed to create layout manager: %w", err)
	}

	// Load layout
	ctx := context.Background()
	layout, err := manager.GetLayout(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get layout: %w", err)
	}

	// Display layout information
	fmt.Printf("Layout: %s\n", layout.Name)
	fmt.Printf("Version: %s\n", layout.Version)
	fmt.Printf("Description: %s\n", layout.Manifest.Description)
	if layout.Manifest.Author != "" {
		fmt.Printf("Author: %s\n", layout.Manifest.Author)
	}
	fmt.Printf("Source: %s (%s)\n", layout.Source.Type, layout.Source.Location)

	// Display features
	if len(layout.Manifest.Features) > 0 {
		fmt.Printf("\nFeatures:\n")
		for _, feature := range layout.Manifest.Features {
			fmt.Printf("  - %s\n", feature)
		}
	}

	// Display dependencies
	if len(layout.Manifest.Dependencies) > 0 {
		fmt.Printf("\nDependencies:\n")
		for _, dep := range layout.Manifest.Dependencies {
			fmt.Printf("  - %s\n", dep)
		}
	}

	// Display components
	if len(layout.Manifest.Components) > 0 {
		fmt.Printf("\nAvailable Components:\n")
		for name, comp := range layout.Manifest.Components {
			fmt.Printf("  - %s (target: %s)\n", name, comp.TargetDir)
		}
	}

	// Display variables
	if len(layout.Manifest.Variables) > 0 {
		fmt.Printf("\nConfigurable Variables:\n")
		for _, v := range layout.Manifest.Variables {
			required := ""
			if v.Required {
				required = " (required)"
			}
			fmt.Printf("  - %s%s: %s", v.Name, required, v.Description)
			if v.Default != "" {
				fmt.Printf(" [default: %s]", v.Default)
			}
			fmt.Println()
		}
	}

	// Display directory structure
	fmt.Printf("\nDirectory Structure:\n")
	for _, dir := range layout.Manifest.Structure.Directories {
		fmt.Printf("  %s", dir.Path)
		if dir.Description != "" {
			fmt.Printf(" - %s", dir.Description)
		}
		fmt.Println()
	}

	return nil
}

// createLayoutManager creates a layout manager instance
func createLayoutManager() (*layout.Manager, error) {
	// Determine config path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".foundry", "layouts.yaml")
	return layout.NewManager(configPath)
}
