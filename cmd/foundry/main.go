package main

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

//go:embed all:templates/*
var templates embed.FS

var version = "0.1.0"

func main() {
	var rootCmd = &cobra.Command{
		Use:   "foundry",
		Short: "Forge production-grade Go services faster",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Version flag
	rootCmd.Version = version

	// Add 'new' command
	var newCmd = &cobra.Command{
		Use:   "new [name]",
		Short: "Create a new Go REST API",
		Args:  cobra.ExactArgs(1),
		Run:   runNew,
	}

	rootCmd.AddCommand(newCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runNew(cmd *cobra.Command, args []string) {
	projectName := args[0]
	
	fmt.Printf("🚀 Creating new project: %s\n", projectName)
	
	// Create project directory
	if err := os.Mkdir(projectName, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}
	
	// Copy templates
	err := filepath.WalkDir("templates/go-rest-api", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Skip the templates directory itself
		if path == "templates/go-rest-api" {
			return nil
		}
		
		// Calculate destination path
		relPath, _ := filepath.Rel("templates/go-rest-api", path)
		destPath := filepath.Join(projectName, relPath)
		
		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}
		
		// Read template from embedded FS
		content, err := templates.ReadFile(path)
		if err != nil {
			return err
		}
		
		// Process .tmpl files
		if strings.HasSuffix(path, ".tmpl") {
			destPath = strings.TrimSuffix(destPath, ".tmpl")
			
			tmpl, err := template.New(filepath.Base(path)).Parse(string(content))
			if err != nil {
				return err
			}
			
			f, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer f.Close()
			
			data := map[string]string{
				"ProjectName": projectName,
			}
			
			return tmpl.Execute(f, data)
		}
		
		// Copy non-template files as-is
		return os.WriteFile(destPath, content, 0644)
	})
	
	if err != nil {
		fmt.Printf("Error creating project: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Project created successfully!\n\n")
	fmt.Printf("Next steps:\n")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Printf("  go mod tidy\n")
	fmt.Printf("  go run .\n")
}
