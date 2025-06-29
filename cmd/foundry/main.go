package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/shapestone/foundry" // Import root package - UPDATE THIS to your module name
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	var rootCmd = &cobra.Command{
		Use:     "foundry",
		Short:   "Forge production-grade Go services faster",
		Version: version,
	}

	var newCmd = &cobra.Command{
		Use:   "new [name]",
		Short: "Create a new Go REST API project",
		Args:  cobra.ExactArgs(1),
		Example: `  foundry new myapp
  foundry new user-service`,
		Run: func(cmd *cobra.Command, args []string) {
			createProject(args[0])
		},
	}

	rootCmd.AddCommand(newCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func createProject(name string) {
	// Validate project name
	if strings.Contains(name, " ") {
		fmt.Println("❌ Error: project name cannot contain spaces")
		os.Exit(1)
	}

	fmt.Printf("🚀 Creating new project: %s\n", name)

	// Create project directory
	if err := os.Mkdir(name, 0755); err != nil {
		if os.IsExist(err) {
			fmt.Printf("❌ Error: directory %s already exists\n", name)
		} else {
			fmt.Printf("❌ Error creating directory: %v\n", err)
		}
		os.Exit(1)
	}

	// Create subdirectories
	dirs := []string{
		filepath.Join(name, "internal", "handlers"),
		filepath.Join(name, "internal", "middleware"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("❌ Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	// Data for templates
	data := struct {
		ProjectName string
		Module      string
	}{
		ProjectName: name,
		Module:      name,
	}

	// Generate files
	files := map[string]string{
		"templates/go.mod.tmpl":    "go.mod",
		"templates/main.go.tmpl":   "main.go",
		"templates/README.md.tmpl": "README.md",
		"templates/gitignore.tmpl": ".gitignore",
	}

	for tmplPath, filename := range files {
		// Read template from embedded FS - using foundry.Templates
		tmplContent, err := foundry.Templates.ReadFile(tmplPath)
		if err != nil {
			fmt.Printf("❌ Error reading template %s: %v\n", tmplPath, err)
			os.RemoveAll(name)
			os.Exit(1)
		}

		outputPath := filepath.Join(name, filename)
		if err := generateFile(outputPath, string(tmplContent), data); err != nil {
			fmt.Printf("❌ Error creating %s: %v\n", filename, err)
			os.RemoveAll(name)
			os.Exit(1)
		}
	}

	// Success message
	fmt.Printf(`✅ Project created successfully!

📁 Structure:
  %s/
  ├── main.go
  ├── go.mod
  ├── README.md
  ├── .gitignore
  └── internal/
      ├── handlers/
      └── middleware/

🚀 Next steps:
  cd %s
  go mod tidy
  go run .

  # Test your API
  curl http://localhost:8080/
  curl http://localhost:8080/health

📚 Coming soon:
  foundry add handler <name>
  foundry add model <name>
`, name, name)
}

func generateFile(path, tmplContent string, data interface{}) error {
	tmpl, err := template.New(filepath.Base(path)).Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}
