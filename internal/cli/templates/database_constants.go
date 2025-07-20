package templates

// Database Templates

const PostgreSQLTemplate = `package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps the PostgreSQL connection pool
type DB struct {
	*pgxpool.Pool
}

// NewConnection creates a new PostgreSQL connection pool
func NewConnection() (*DB, error) {
	cfg := LoadConfig()

	// Build connection string
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)

	// Configure connection pool
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set pool configuration
	config.MaxConns = cfg.MaxConnections
	config.MinConns = cfg.MinConnections
	config.MaxConnLifetime = cfg.MaxConnLifetime
	config.MaxConnIdleTime = cfg.MaxConnIdleTime

	// Set connection timeout
	config.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	// Create connection pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	db.Pool.Close()
}

// Health checks the database connection
func (db *DB) Health(ctx context.Context) error {
	return db.Ping(ctx)
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("transaction failed: %v, rollback failed: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// QueryRow is a convenience wrapper that adds query logging
func (db *DB) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	start := time.Now()
	row := db.Pool.QueryRow(ctx, query, args...)

	if os.Getenv("DB_DEBUG") == "true" {
		fmt.Printf("[DB] Query (%v): %s\n", time.Since(start), query)
	}

	return row
}

// Query is a convenience wrapper that adds query logging
func (db *DB) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	start := time.Now()
	rows, err := db.Pool.Query(ctx, query, args...)

	if os.Getenv("DB_DEBUG") == "true" {
		fmt.Printf("[DB] Query (%v): %s\n", time.Since(start), query)
	}

	return rows, err
}

// Exec is a convenience wrapper that adds query logging
func (db *DB) Exec(ctx context.Context, query string, args ...interface{}) (pgx.CommandTag, error) {
	start := time.Now()
	tag, err := db.Pool.Exec(ctx, query, args...)

	if os.Getenv("DB_DEBUG") == "true" {
		fmt.Printf("[DB] Exec (%v): %s\n", time.Since(start), query)
	}

	return tag, err
}

// Migrate runs database migrations (basic implementation)
func (db *DB) Migrate(ctx context.Context) error {
	// Create migrations table if not exists
	_, err := db.Exec(ctx, ` + "`" + `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	` + "`" + `)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// TODO: Implement migration logic
	// This is a placeholder for migration functionality
	// Consider using a migration library like golang-migrate/migrate

	return nil
}
`

const MySQLTemplate = `package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DB wraps the MySQL connection
type DB struct {
	*sql.DB
}

// NewConnection creates a new MySQL connection
func NewConnection() (*DB, error) {
	cfg := LoadConfig()

	// Build connection string
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.ConnectTimeout,
	)

	// Open database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool configuration
	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MinConnections)
	db.SetConnMaxLifetime(cfg.MaxConnLifetime)
	db.SetConnMaxIdleTime(cfg.MaxConnIdleTime)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Health checks the database connection
func (db *DB) Health(ctx context.Context) error {
	return db.PingContext(ctx)
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction failed: %v, rollback failed: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// QueryRowContext is a convenience wrapper that adds query logging
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := db.DB.QueryRowContext(ctx, query, args...)

	if os.Getenv("DB_DEBUG") == "true" {
		fmt.Printf("[DB] Query (%v): %s\n", time.Since(start), query)
	}

	return row
}

// QueryContext is a convenience wrapper that adds query logging
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := db.DB.QueryContext(ctx, query, args...)

	if os.Getenv("DB_DEBUG") == "true" {
		fmt.Printf("[DB] Query (%v): %s\n", time.Since(start), query)
	}

	return rows, err
}

// ExecContext is a convenience wrapper that adds query logging
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := db.DB.ExecContext(ctx, query, args...)

	if os.Getenv("DB_DEBUG") == "true" {
		fmt.Printf("[DB] Exec (%v): %s\n", time.Since(start), query)
	}

	return result, err
}

// Migrate runs database migrations (basic implementation)
func (db *DB) Migrate(ctx context.Context) error {
	// Create migrations table if not exists
	_, err := db.ExecContext(ctx, ` + "`" + `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	` + "`" + `)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// TODO: Implement migration logic
	// This is a placeholder for migration functionality
	// Consider using a migration library like golang-migrate/migrate

	return nil
}
`

const SQLiteTemplate = `package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQLite connection
type DB struct {
	*sql.DB
}

// NewConnection creates a new SQLite connection
func NewConnection() (*DB, error) {
	cfg := LoadConfig()

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(cfg.Path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// SQLite specific configurations
	db.SetMaxOpenConns(1) // SQLite doesn't support multiple concurrent writers
	db.SetMaxIdleConns(1)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return &DB{DB: db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Health checks the database connection
func (db *DB) Health(ctx context.Context) error {
	return db.PingContext(ctx)
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction failed: %v, rollback failed: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// QueryRowContext is a convenience wrapper that adds query logging
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := db.DB.QueryRowContext(ctx, query, args...)

	if os.Getenv("DB_DEBUG") == "true" {
		fmt.Printf("[DB] Query (%v): %s\n", time.Since(start), query)
	}

	return row
}

// QueryContext is a convenience wrapper that adds query logging
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := db.DB.QueryContext(ctx, query, args...)

	if os.Getenv("DB_DEBUG") == "true" {
		fmt.Printf("[DB] Query (%v): %s\n", time.Since(start), query)
	}

	return rows, err
}

// ExecContext is a convenience wrapper that adds query logging
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := db.DB.ExecContext(ctx, query, args...)

	if os.Getenv("DB_DEBUG") == "true" {
		fmt.Printf("[DB] Exec (%v): %s\n", time.Since(start), query)
	}

	return result, err
}

// Migrate runs database migrations (basic implementation)
func (db *DB) Migrate(ctx context.Context) error {
	// Create migrations table if not exists
	_, err := db.ExecContext(ctx, ` + "`" + `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	` + "`" + `)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// TODO: Implement migration logic
	// This is a placeholder for migration functionality
	// Consider using a migration library like golang-migrate/migrate

	return nil
}
`

const MongoDBTemplate = `package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB wraps the MongoDB client
type DB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// NewConnection creates a new MongoDB connection
func NewConnection() (*DB, error) {
	cfg := LoadConfig()

	// Configure client options
	clientOptions := options.Client().ApplyURI(cfg.URI)
	clientOptions.SetMaxPoolSize(uint64(cfg.MaxConnections))
	clientOptions.SetMinPoolSize(uint64(cfg.MinConnections))
	clientOptions.SetMaxConnIdleTime(cfg.MaxConnIdleTime)
	clientOptions.SetConnectTimeout(cfg.ConnectTimeout)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Get database instance
	database := client.Database(cfg.DatabaseName)

	return &DB{
		Client:   client,
		Database: database,
	}, nil
}

// Close closes the MongoDB connection
func (db *DB) Disconnect(ctx context.Context) error {
	return db.Client.Disconnect(ctx)
}

// Health checks the database connection
func (db *DB) Health(ctx context.Context) error {
	return db.Client.Ping(ctx, nil)
}

// Collection returns a MongoDB collection
func (db *DB) Collection(name string) *mongo.Collection {
	return db.Database.Collection(name)
}

// Transaction executes a function within a MongoDB transaction
func (db *DB) Transaction(ctx context.Context, fn func(mongo.SessionContext) error) error {
	session, err := db.Client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	}

	_, err = session.WithTransaction(ctx, callback)
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}

// RunCommand executes a database command
func (db *DB) RunCommand(ctx context.Context, command interface{}) *mongo.SingleResult {
	start := time.Now()
	result := db.Database.RunCommand(ctx, command)

	if os.Getenv("DB_DEBUG") == "true" {
		fmt.Printf("[DB] Command (%v): %+v\n", time.Since(start), command)
	}

	return result
}

// Migrate runs database migrations (basic implementation)
func (db *DB) Migrate(ctx context.Context) error {
	// Create migrations collection if not exists
	migrationsColl := db.Collection("schema_migrations")

	// Create index on version field
	indexModel := mongo.IndexModel{
		Keys: map[string]int{"version": 1},
		Options: options.Index().SetUnique(true),
	}

	_, err := migrationsColl.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create migrations index: %w", err)
	}

	// TODO: Implement migration logic
	// This is a placeholder for migration functionality
	// Consider using a migration library

	return nil
}
`

const DefaultDatabaseTemplate = `package database

import (
	"context"
	"fmt"
)

// DB is a placeholder database interface
type DB struct {
	// Add your database connection here
}

// NewConnection creates a new database connection
func NewConnection() (*DB, error) {
	// TODO: Implement your database connection logic
	return &DB{}, fmt.Errorf("database connection not implemented")
}

// Close closes the database connection
func (db *DB) Close() error {
	// TODO: Implement close logic
	return nil
}

// Health checks the database connection
func (db *DB) Health(ctx context.Context) error {
	// TODO: Implement health check
	return fmt.Errorf("health check not implemented")
}
`

const DatabaseConfigTemplate = `package database

import (
	"os"
	"strconv"
	"time"
)

// Config holds database configuration
type Config struct {
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	SSLMode         string
	Path            string
	URI             string
	DatabaseName    string
	MaxConnections  int32
	MinConnections  int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	ConnectTimeout  time.Duration
}

// LoadConfig loads database configuration from environment variables
func LoadConfig() *Config {
	cfg := &Config{}

	// Common configuration
	cfg.Host = getEnv("DB_HOST", "localhost")
	cfg.Port = getEnv("DB_PORT", "5432")
	cfg.Name = getEnv("DB_NAME", "myapp")
	cfg.User = getEnv("DB_USER", "postgres")
	cfg.Password = getEnv("DB_PASSWORD", "postgres")
	cfg.SSLMode = getEnv("DB_SSLMODE", "disable")
	cfg.Path = getEnv("DB_PATH", "./data/app.db")
	cfg.URI = getEnv("MONGO_URI", "mongodb://localhost:27017")
	cfg.DatabaseName = getEnv("MONGO_DATABASE", "myapp")

	cfg.MaxConnections = int32(getEnvAsInt("DB_MAX_CONNECTIONS", 25))
	cfg.MinConnections = int32(getEnvAsInt("DB_MIN_CONNECTIONS", 5))
	cfg.MaxConnLifetime = getEnvAsDuration("DB_MAX_CONN_LIFETIME", 30*time.Minute)
	cfg.MaxConnIdleTime = getEnvAsDuration("DB_MAX_CONN_IDLE_TIME", 10*time.Minute)
	cfg.ConnectTimeout = getEnvAsDuration("DB_CONNECT_TIMEOUT", 10*time.Second)

	return cfg
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	strValue := getEnv(key, "")
	if strValue == "" {
		return defaultValue
	}

	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}

	return defaultValue
}

// getEnvAsDuration gets an environment variable as a duration with a default value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	strValue := getEnv(key, "")
	if strValue == "" {
		return defaultValue
	}

	if value, err := time.ParseDuration(strValue); err == nil {
		return value
	}

	return defaultValue
}
`

const MigrationReadmeTemplate = `# Database Migrations

This directory contains database migration files.

## Migration Naming Convention

Migration files should follow this naming pattern:
` + "```" + `
XXX_description.sql
` + "```" + `

Where:
- ` + "`XXX`" + ` is a sequential number (001, 002, 003, etc.)
- ` + "`description`" + ` is a brief description of what the migration does

Examples:
- ` + "`001_initial_schema.sql`" + `
- ` + "`002_add_users_table.sql`" + `
- ` + "`003_add_email_index.sql`" + `

## Running Migrations

### Manual Migration

Apply migrations manually by running each SQL file in order.

### Using a Migration Tool

For production use, consider using a migration tool:

1. **golang-migrate/migrate**
   ` + "```bash" + `
   # Install
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   
   # Run migrations
   migrate -path migrations -database "postgres://..." up
   ` + "```" + `

2. **pressly/goose**
   ` + "```bash" + `
   # Install
   go install github.com/pressly/goose/v3/cmd/goose@latest
   
   # Run migrations
   goose postgres "connection-string" up
   ` + "```" + `

## Creating Migrations

When creating a new migration:

1. Use the next sequential number
2. Keep migrations small and focused
3. Always include both UP and DOWN migrations (if using a tool)
4. Test migrations on a development database first
5. Never modify existing migrations that have been applied

## Best Practices

- **Backward Compatibility**: Ensure migrations don't break existing code
- **Data Safety**: Always backup before running migrations in production
- **Idempotency**: Make migrations safe to run multiple times
- **Performance**: Consider the impact of migrations on large tables
- **Documentation**: Document complex migrations with comments

## Migration Status

Track which migrations have been applied:

` + "```sql" + `
SELECT * FROM schema_migrations ORDER BY version;
` + "```" + `
`

const PostgreSQLMigrationTemplate = `-- Migration: Initial Schema
-- Description: Create initial database schema for PostgreSQL

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Create trigger to update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add any additional tables here
`

const SQLiteMigrationTemplate = `-- Migration: Initial Schema
-- Description: Create initial database schema for SQLite

-- Enable foreign key constraints
PRAGMA foreign_keys = ON;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Create trigger to update updated_at
CREATE TRIGGER IF NOT EXISTS update_users_updated_at 
    AFTER UPDATE ON users
    FOR EACH ROW
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Add any additional tables here
`

const DefaultMigrationTemplate = `-- Migration: Initial Schema
-- Description: Create initial database schema

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Add any additional tables here
`

const PostgreSQLDockerTemplate = `version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: {{ .ProjectName }}_postgres
    environment:
      POSTGRES_DB: {{ .ProjectName }}
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d:ro
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
`

const MySQLDockerTemplate = `version: '3.8'

services:
  mysql:
    image: mysql:8
    container_name: {{ .ProjectName }}_mysql
    environment:
      MYSQL_DATABASE: {{ .ProjectName }}
      MYSQL_ROOT_PASSWORD: mysql
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./migrations:/docker-entrypoint-initdb.d:ro
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  mysql_data:
`

const MongoDBDockerTemplate = `version: '3.8'

services:
  mongodb:
    image: mongo:7
    container_name: {{ .ProjectName }}_mongodb
    environment:
      MONGO_INITDB_DATABASE: {{ .ProjectName }}
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db
    restart: unless-stopped
    healthcheck:
      test: echo 'db.runCommand("ping").ok' | mongosh localhost:27017/test --quiet
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  mongodb_data:
`

const DefaultDockerTemplate = `version: '3.8'

services:
  database:
    # TODO: Configure your database service
    # image: your-database-image
    # container_name: {{ .ProjectName }}_db
    # environment:
    #   - DATABASE_ENV_VARS=value
    # ports:
    #   - "port:port"
    # volumes:
    #   - db_data:/data
    restart: unless-stopped

volumes:
  db_data:
`
