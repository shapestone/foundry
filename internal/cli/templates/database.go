package templates

// DatabaseInfo holds information about supported databases
type DatabaseInfo struct {
	Type        string
	Description string
	Driver      string
	DefaultPort string
	DockerImage string
}

// GetSupportedDatabases returns all supported database types
func GetSupportedDatabases() []DatabaseInfo {
	return []DatabaseInfo{
		{
			Type:        "postgres",
			Description: "PostgreSQL - Advanced open-source relational database",
			Driver:      "pgx",
			DefaultPort: "5432",
			DockerImage: "postgres:16-alpine",
		},
		{
			Type:        "mysql",
			Description: "MySQL - Popular open-source relational database",
			Driver:      "mysql",
			DefaultPort: "3306",
			DockerImage: "mysql:8",
		},
		{
			Type:        "sqlite",
			Description: "SQLite - Lightweight embedded database",
			Driver:      "sqlite3",
			DefaultPort: "",
			DockerImage: "",
		},
		{
			Type:        "mongodb",
			Description: "MongoDB - Document-oriented NoSQL database",
			Driver:      "mongo",
			DefaultPort: "27017",
			DockerImage: "mongo:7",
		},
	}
}

// IsSupportedDatabase checks if a database type is supported
func IsSupportedDatabase(dbType string) bool {
	for _, db := range GetSupportedDatabases() {
		if db.Type == dbType {
			return true
		}
	}
	return false
}

// GetDatabaseInfo returns info for a specific database type
func GetDatabaseInfo(dbType string) (DatabaseInfo, bool) {
	for _, db := range GetSupportedDatabases() {
		if db.Type == dbType {
			return db, true
		}
	}
	return DatabaseInfo{}, false
}

// GetDatabaseTemplate returns the Go template for a database type
func GetDatabaseTemplate(dbType string) string {
	switch dbType {
	case "postgres":
		return PostgreSQLTemplate
	case "mysql":
		return MySQLTemplate
	case "sqlite":
		return SQLiteTemplate
	case "mongodb":
		return MongoDBTemplate
	default:
		return DefaultDatabaseTemplate
	}
}

// GetConfigTemplate returns the config template for databases
func GetConfigTemplate() string {
	return DatabaseConfigTemplate
}

// GetMigrationReadmeTemplate returns the migration README template
func GetMigrationReadmeTemplate() string {
	return MigrationReadmeTemplate
}

// GetExampleMigrationTemplate returns an example migration for a database type
func GetExampleMigrationTemplate(dbType string) string {
	switch dbType {
	case "postgres", "mysql":
		return PostgreSQLMigrationTemplate
	case "sqlite":
		return SQLiteMigrationTemplate
	default:
		return DefaultMigrationTemplate
	}
}

// GetDockerTemplate returns the docker-compose template for a database type
func GetDockerTemplate(dbType string) string {
	switch dbType {
	case "postgres":
		return PostgreSQLDockerTemplate
	case "mysql":
		return MySQLDockerTemplate
	case "mongodb":
		return MongoDBDockerTemplate
	default:
		return DefaultDockerTemplate
	}
}
