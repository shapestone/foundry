package generators

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/shapestone/foundry/internal/cli/templates"
)

// DatabaseGenerator handles database file generation
type DatabaseGenerator struct {
	stdout io.Writer
	stderr io.Writer
}

// NewDatabaseGenerator creates a new database generator
func NewDatabaseGenerator(stdout, stderr io.Writer) *DatabaseGenerator {
	return &DatabaseGenerator{
		stdout: stdout,
		stderr: stderr,
	}
}

// DatabaseOptions holds options for database generation
type DatabaseOptions struct {
	Type           string
	WithMigrations bool
	WithDocker     bool
	OutputDir      string
}

// Generate creates database files based on options
func (g *DatabaseGenerator) Generate(options DatabaseOptions) error {
	dbInfo, ok := templates.GetDatabaseInfo(options.Type)
	if !ok {
		return fmt.Errorf("unsupported database type: %s", options.Type)
	}

	// Create database configuration file
	dbPath := filepath.Join(options.OutputDir, "database.go")
	if err := g.createDatabaseFile(dbPath, options.Type); err != nil {
		return fmt.Errorf("failed to create database file: %w", err)
	}

	// Create config file
	configPath := filepath.Join(options.OutputDir, "config.go")
	if err := g.createConfigFile(configPath); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	// Create .env.example
	if err := g.createEnvExample(options.Type, dbInfo.DefaultPort); err != nil {
		fmt.Fprintf(g.stderr, "⚠️  Warning: couldn't create .env.example: %v\n", err)
	}

	// Handle migrations
	if options.WithMigrations && options.Type != "mongodb" {
		if err := g.createMigrationSetup(options.Type); err != nil {
			fmt.Fprintf(g.stderr, "⚠️  Warning: couldn't create migration setup: %v\n", err)
		}
	}

	// Handle Docker
	if options.WithDocker && dbInfo.DockerImage != "" {
		if err := g.createDockerSetup(options.Type); err != nil {
			fmt.Fprintf(g.stderr, "⚠️  Warning: couldn't create Docker setup: %v\n", err)
		}
	}

	// Show success message
	g.showSuccess(options, dbInfo)

	return nil
}

// createDatabaseFile creates the main database configuration file
func (g *DatabaseGenerator) createDatabaseFile(dbPath, dbType string) error {
	template := templates.GetDatabaseTemplate(dbType)
	return writeFile(dbPath, template)
}

// createConfigFile creates the database config file
func (g *DatabaseGenerator) createConfigFile(configPath string) error {
	template := templates.GetConfigTemplate()
	return writeFile(configPath, template)
}

// createEnvExample creates or updates .env.example
func (g *DatabaseGenerator) createEnvExample(dbType, defaultPort string) error {
	envPath := ".env.example"

	// Read existing content if file exists
	existing := ""
	if content, err := os.ReadFile(envPath); err == nil {
		existing = string(content)
	}

	// Get environment variables for this database type
	dbEnvVars := getEnvVars(dbType)

	// Append to existing content if not already present
	if existing != "" && !contains(existing, "Database Configuration") {
		existing += "\n" + dbEnvVars
	} else if existing == "" {
		existing = dbEnvVars
	}

	return os.WriteFile(envPath, []byte(existing), 0644)
}

// createMigrationSetup creates migration directory and files
func (g *DatabaseGenerator) createMigrationSetup(dbType string) error {
	migrationsDir := "migrations"
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return err
	}

	// Create README for migrations
	readmePath := filepath.Join(migrationsDir, "README.md")
	readmeTemplate := templates.GetMigrationReadmeTemplate()
	if err := writeFile(readmePath, readmeTemplate); err != nil {
		return err
	}

	// Create example migration
	examplePath := filepath.Join(migrationsDir, "001_initial_schema.sql")
	exampleTemplate := templates.GetExampleMigrationTemplate(dbType)
	return writeFile(examplePath, exampleTemplate)
}

// createDockerSetup creates docker-compose.yml
func (g *DatabaseGenerator) createDockerSetup(dbType string) error {
	dockerPath := "docker-compose.yml"
	if _, err := os.Stat(dockerPath); err == nil {
		fmt.Fprintln(g.stdout, "⚠️  docker-compose.yml already exists, skipping Docker setup")
		return nil
	}

	template := templates.GetDockerTemplate(dbType)
	return writeFile(dockerPath, template)
}

// showSuccess displays success message with setup instructions
func (g *DatabaseGenerator) showSuccess(options DatabaseOptions, dbInfo templates.DatabaseInfo) {
	migrationInfo := ""
	if options.WithMigrations && options.Type != "mongodb" {
		migrationInfo = `
📁 Migrations:
  migrations/
  ├── README.md
  └── 001_initial_schema.sql

  Run migrations with:
  go run . migrate up
`
	}

	dockerInfo := ""
	if options.WithDocker && dbInfo.DockerImage != "" {
		dockerInfo = `
🐳 Docker:
  docker-compose.yml created
  
  Start database:
  docker-compose up -d

  Stop database:
  docker-compose down
`
	}

	setupSteps := getSetupSteps(options.Type)

	fmt.Fprintf(g.stdout, `✅ Database support added successfully!

📁 Created:
  %s/database.go
  %s/config.go
  .env.example
%s%s
📝 Database: %s
  Driver: %s
%s
🔧 Features included:
  - Connection pooling
  - Context support
  - Graceful shutdown
  - Environment-based configuration
  - Health check endpoint
  - Query timeout handling
  %s

💡 Next steps:
  - Run: go mod tidy
  - Configure your database connection in .env
  - Start coding with your database!
`,
		options.OutputDir,
		options.OutputDir,
		migrationInfo,
		dockerInfo,
		dbInfo.Description,
		dbInfo.Driver,
		setupSteps,
		func() string {
			if options.WithMigrations && options.Type != "mongodb" {
				return "- Migration support"
			}
			return ""
		}())
}

// Helper functions

func getEnvVars(dbType string) string {
	switch dbType {
	case "postgres":
		return `
# PostgreSQL Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable
`
	case "mysql":
		return `
# MySQL Configuration
DB_HOST=localhost
DB_PORT=3306
DB_NAME=myapp
DB_USER=root
DB_PASSWORD=mysql
`
	case "sqlite":
		return `
# SQLite Configuration
DB_PATH=./data/app.db
`
	case "mongodb":
		return `
# MongoDB Configuration
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=myapp
`
	default:
		return ""
	}
}

func getSetupSteps(dbType string) string {
	// Implementation would include database-specific setup steps
	// For brevity, returning a placeholder
	return fmt.Sprintf("Setup steps for %s database", dbType)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		(len(s) > len(substr) && s[1:len(substr)+1] == substr)
}
