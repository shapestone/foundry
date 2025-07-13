package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shapestone/foundry"
	"github.com/shapestone/foundry/internal/generator"
	"github.com/shapestone/foundry/internal/project"
	"github.com/spf13/cobra"
)

var databaseCmd = &cobra.Command{
	Use:   "db [type]",
	Short: "Add database support to your project",
	Args:  cobra.ExactArgs(1),
	Example: `  foundry add db postgres
  foundry add db mysql
  foundry add db sqlite
  foundry add db mongodb`,
	Run: func(cmd *cobra.Command, args []string) {
		withMigrations, _ := cmd.Flags().GetBool("with-migrations")
		withDocker, _ := cmd.Flags().GetBool("with-docker")
		addDatabase(args[0], withMigrations, withDocker)
	},
}

var supportedDatabases = map[string]struct {
	description string
	driver      string
	defaultPort string
	dockerImage string
}{
	"postgres": {
		description: "PostgreSQL - Advanced open-source relational database",
		driver:      "pgx",
		defaultPort: "5432",
		dockerImage: "postgres:16-alpine",
	},
	"mysql": {
		description: "MySQL - Popular open-source relational database",
		driver:      "mysql",
		defaultPort: "3306",
		dockerImage: "mysql:8",
	},
	"sqlite": {
		description: "SQLite - Lightweight embedded database",
		driver:      "sqlite3",
		defaultPort: "",
		dockerImage: "",
	},
	"mongodb": {
		description: "MongoDB - Document-oriented NoSQL database",
		driver:      "mongo",
		defaultPort: "27017",
		dockerImage: "mongo:7",
	},
}

func init() {
	databaseCmd.Flags().Bool("with-migrations", false, "Include migration setup")
	databaseCmd.Flags().Bool("with-docker", false, "Add docker-compose configuration")
}

func addDatabase(dbType string, withMigrations bool, withDocker bool) {
	// Check if we're in a Go project
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		fmt.Println("❌ Error: go.mod not found. Please run this command from your project root")
		os.Exit(1)
	}

	// Validate database type
	dbInfo, ok := supportedDatabases[dbType]
	if !ok {
		fmt.Printf("❌ Error: unsupported database type '%s'\n", dbType)
		fmt.Println("Supported types:")
		for t, info := range supportedDatabases {
			fmt.Printf("  - %s: %s\n", t, info.description)
		}
		os.Exit(1)
	}

	fmt.Printf("🗄️  Adding database: %s\n", dbType)

	// Create database directory
	dbDir := filepath.Join("internal", "database")
	dbPath := filepath.Join(dbDir, "database.go")

	// Check if database config already exists
	if _, err := os.Stat(dbPath); err == nil {
		fmt.Printf("❌ Error: database configuration already exists at %s\n", dbPath)
		os.Exit(1)
	}

	// Data for templates
	data := struct {
		DBType      string
		DBTypeLower string
		DBTypeTitle string
		Driver      string
		DefaultPort string
		ProjectName string
		Module      string
		DockerImage string
	}{
		DBType:      dbType,
		DBTypeLower: dbType,
		DBTypeTitle: getDBTitle(dbType),
		Driver:      dbInfo.driver,
		DefaultPort: dbInfo.defaultPort,
		ProjectName: project.GetProjectName(),
		Module:      project.GetCurrentModule(),
		DockerImage: dbInfo.dockerImage,
	}

	// Read database template
	templateFile := fmt.Sprintf("templates/database-%s.go.tmpl", dbType)
	tmplContent, err := foundry.Templates.ReadFile(templateFile)
	if err != nil {
		fmt.Printf("❌ Error reading database template: %v\n", err)
		os.Exit(1)
	}

	// Create database configuration
	gen := generator.NewFileGenerator()
	if err := gen.Generate(dbPath, string(tmplContent), data); err != nil {
		fmt.Printf("❌ Error creating database configuration: %v\n", err)
		os.Exit(1)
	}

	// Create config file
	configPath := filepath.Join(dbDir, "config.go")
	configTmpl, err := foundry.Templates.ReadFile("templates/database-config.go.tmpl")
	if err != nil {
		fmt.Printf("❌ Error reading config template: %v\n", err)
		os.Exit(1)
	}

	if err := gen.Generate(configPath, string(configTmpl), data); err != nil {
		fmt.Printf("❌ Error creating config: %v\n", err)
		os.Exit(1)
	}

	// Handle migrations
	if withMigrations && dbType != "mongodb" {
		createMigrationSetup(dbType, data)
	}

	// Handle Docker
	if withDocker && dbInfo.dockerImage != "" {
		createDockerSetup(dbType, data)
	}

	// Create .env.example if it doesn't exist
	createEnvExample(dbType, dbInfo.defaultPort)

	// Show success message
	showDatabaseSuccess(dbType, dbPath, dbInfo, withMigrations, withDocker)
}

func getDBTitle(dbType string) string {
	switch dbType {
	case "postgres":
		return "PostgreSQL"
	case "mysql":
		return "MySQL"
	case "sqlite":
		return "SQLite"
	case "mongodb":
		return "MongoDB"
	default:
		return dbType
	}
}

func createMigrationSetup(dbType string, data interface{}) {
	migrationsDir := filepath.Join("migrations")
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		fmt.Printf("⚠️  Warning: couldn't create migrations directory: %v\n", err)
		return
	}

	// Create README for migrations
	readmePath := filepath.Join(migrationsDir, "README.md")
	readmeTmpl, err := foundry.Templates.ReadFile("templates/migrations-readme.md.tmpl")
	if err == nil {
		gen := generator.NewFileGenerator()
		gen.Generate(readmePath, string(readmeTmpl), data)
	}

	// Create example migration
	examplePath := filepath.Join(migrationsDir, "001_initial_schema.sql")
	exampleTmpl, err := foundry.Templates.ReadFile(fmt.Sprintf("templates/migration-example-%s.sql.tmpl", dbType))
	if err == nil {
		gen := generator.NewFileGenerator()
		gen.Generate(examplePath, string(exampleTmpl), data)
	}
}

func createDockerSetup(dbType string, data interface{}) {
	// Check if docker-compose.yml exists
	dockerPath := "docker-compose.yml"
	if _, err := os.Stat(dockerPath); err == nil {
		fmt.Println("⚠️  docker-compose.yml already exists, skipping Docker setup")
		return
	}

	// Create docker-compose.yml
	dockerTmpl, err := foundry.Templates.ReadFile(fmt.Sprintf("templates/docker-compose-%s.yml.tmpl", dbType))
	if err != nil {
		fmt.Printf("⚠️  Warning: couldn't find Docker template for %s\n", dbType)
		return
	}

	gen := generator.NewFileGenerator()
	if err := gen.Generate(dockerPath, string(dockerTmpl), data); err != nil {
		fmt.Printf("⚠️  Warning: couldn't create docker-compose.yml: %v\n", err)
	}
}

func createEnvExample(dbType string, defaultPort string) {
	envPath := ".env.example"

	// Read existing content if file exists
	existing := ""
	if content, err := os.ReadFile(envPath); err == nil {
		existing = string(content)
	}

	// Prepare database environment variables
	var dbEnvVars string
	switch dbType {
	case "postgres":
		dbEnvVars = `
# PostgreSQL Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable
`
	case "mysql":
		dbEnvVars = `
# MySQL Configuration
DB_HOST=localhost
DB_PORT=3306
DB_NAME=myapp
DB_USER=root
DB_PASSWORD=mysql
`
	case "sqlite":
		dbEnvVars = `
# SQLite Configuration
DB_PATH=./data/app.db
`
	case "mongodb":
		dbEnvVars = `
# MongoDB Configuration
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=myapp
`
	}

	// Append to existing content
	if existing != "" && !strings.Contains(existing, "Database Configuration") {
		existing += "\n" + dbEnvVars
	} else if existing == "" {
		existing = dbEnvVars
	}

	// Write the file
	if err := os.WriteFile(envPath, []byte(existing), 0644); err != nil {
		fmt.Printf("⚠️  Warning: couldn't create .env.example: %v\n", err)
	}
}

func showDatabaseSuccess(dbType string, dbPath string, dbInfo struct {
	description string
	driver      string
	defaultPort string
	dockerImage string
}, withMigrations bool, withDocker bool) {
	moduleName := project.GetCurrentModule()

	migrationInfo := ""
	if withMigrations && dbType != "mongodb" {
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
	if withDocker && dbInfo.dockerImage != "" {
		dockerInfo = fmt.Sprintf(`
🐳 Docker:
  docker-compose.yml created
  
  Start database:
  docker-compose up -d

  Stop database:
  docker-compose down
`)
	}

	var setupSteps string
	switch dbType {
	case "postgres":
		setupSteps = fmt.Sprintf(`
📌 Setup steps:

1. Install dependencies:
   go get github.com/jackc/pgx/v5
   go get github.com/jackc/pgx/v5/pgxpool

2. Update your .env file with database credentials:
   cp .env.example .env
   # Edit .env with your database details

3. Import and initialize the database in main.go:
   import "%s/internal/database"

   // In main():
   db, err := database.NewConnection()
   if err != nil {
       log.Fatal("Failed to connect to database:", err)
   }
   defer db.Close()

4. Use the database in your handlers:
   type UserHandler struct {
       db *database.DB
   }

   func NewUserHandler(db *database.DB) *UserHandler {
       return &UserHandler{db: db}
   }

5. Example query:
   rows, err := h.db.Query(ctx, "SELECT id, name FROM users")
   if err != nil {
       return err
   }
   defer rows.Close()`, moduleName)

	case "mysql":
		setupSteps = fmt.Sprintf(`
📌 Setup steps:

1. Install dependencies:
   go get github.com/go-sql-driver/mysql

2. Update your .env file with database credentials:
   cp .env.example .env
   # Edit .env with your database details

3. Import and initialize the database in main.go:
   import "%s/internal/database"

   // In main():
   db, err := database.NewConnection()
   if err != nil {
       log.Fatal("Failed to connect to database:", err)
   }
   defer db.Close()

4. Use the database in your handlers:
   type UserHandler struct {
       db *database.DB
   }

   func NewUserHandler(db *database.DB) *UserHandler {
       return &UserHandler{db: db}
   }

5. Example query:
   rows, err := h.db.QueryContext(ctx, "SELECT id, name FROM users")
   if err != nil {
       return err
   }
   defer rows.Close()`, moduleName)

	case "sqlite":
		setupSteps = fmt.Sprintf(`
📌 Setup steps:

1. Install dependencies:
   go get github.com/mattn/go-sqlite3

2. Update your .env file with database path:
   cp .env.example .env
   # Edit .env if you want a different path

3. Import and initialize the database in main.go:
   import "%s/internal/database"

   // In main():
   db, err := database.NewConnection()
   if err != nil {
       log.Fatal("Failed to connect to database:", err)
   }
   defer db.Close()

4. Use the database in your handlers:
   type UserHandler struct {
       db *database.DB
   }

   func NewUserHandler(db *database.DB) *UserHandler {
       return &UserHandler{db: db}
   }

5. Example query:
   rows, err := h.db.QueryContext(ctx, "SELECT id, name FROM users")
   if err != nil {
       return err
   }
   defer rows.Close()`, moduleName)

	case "mongodb":
		setupSteps = fmt.Sprintf(`
📌 Setup steps:

1. Install dependencies:
   go get go.mongodb.org/mongo-driver/mongo

2. Update your .env file with MongoDB connection:
   cp .env.example .env
   # Edit .env with your MongoDB details

3. Import and initialize the database in main.go:
   import "%s/internal/database"

   // In main():
   db, err := database.NewConnection()
   if err != nil {
       log.Fatal("Failed to connect to database:", err)
   }
   defer db.Disconnect()

4. Use the database in your handlers:
   type UserHandler struct {
       db *database.DB
   }

   func NewUserHandler(db *database.DB) *UserHandler {
       return &UserHandler{db: db}
   }

5. Example query:
   collection := h.db.Collection("users")
   
   var user User
   err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
   if err != nil {
       return err
   }`, moduleName)
	}

	fmt.Printf(`✅ Database support added successfully!

📁 Created:
  %s
  internal/database/config.go
  .env.example
%s%s
📝 Database: %s
  %s
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
		dbPath,
		migrationInfo,
		dockerInfo,
		dbInfo.description,
		fmt.Sprintf("Driver: %s", dbInfo.driver),
		setupSteps,
		func() string {
			if withMigrations && dbType != "mongodb" {
				return "- Migration support"
			}
			return ""
		}())
}
