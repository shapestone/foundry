#!/bin/bash

# Integration test for database connections
set -e

echo "🧪 Testing Foundry database integration..."
echo "========================================"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Create test output directory
TEST_OUTPUT_DIR="test-output"
mkdir -p "$TEST_OUTPUT_DIR"

# Test directory with timestamp
TEST_DIR="$TEST_OUTPUT_DIR/test-db-integration-$(date +%s)-$$"

# Determine base directory
if [ -d "scripts" ]; then
    BASE_DIR="."
else
    BASE_DIR=".."
fi

# Clean up function
cleanup() {
    cd "$BASE_DIR"
    # Stop any running containers
    if command -v docker &> /dev/null; then
        docker-compose -f "$TEST_DIR/test-postgres-docker/docker-compose.yml" down 2>/dev/null || true
        docker-compose -f "$TEST_DIR/test-mysql-docker/docker-compose.yml" down 2>/dev/null || true
        docker-compose -f "$TEST_DIR/test-mongodb-docker/docker-compose.yml" down 2>/dev/null || true
    fi
}

trap cleanup EXIT

# Build foundry
echo -e "${BLUE}Building foundry...${NC}"
cd "$BASE_DIR"
go build -o foundry-test ./cmd/foundry

# Create base test project
echo -e "\n${BLUE}Creating test project...${NC}"
./foundry-test new "$TEST_DIR"
cd "$TEST_DIR"

FOUNDRY_TEST="../../foundry-test"

# Test SQLite (no external dependencies needed)
echo -e "\n${BLUE}Testing SQLite integration...${NC}"
mkdir test-sqlite-integration
cd test-sqlite-integration
cp ../go.mod .

echo -e "  Creating SQLite database configuration..."
if $FOUNDRY_TEST add db sqlite --with-migrations; then
    echo -e "  ${GREEN}✓ SQLite configuration created${NC}"
else
    echo -e "  ${RED}✗ Failed to create SQLite configuration${NC}"
    exit 1
fi

# Create a test program
cat > test_db.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "test-sqlite-integration/internal/database"
)

func main() {
    // Connect to database
    db, err := database.NewConnection()
    if err != nil {
        log.Fatal("Connection failed:", err)
    }
    defer db.Close()

    // Test health check
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := db.Health(ctx); err != nil {
        log.Fatal("Health check failed:", err)
    }

    // Run migration
    if err := db.Migrate(ctx); err != nil {
        log.Fatal("Migration failed:", err)
    }

    // Test a simple query
    var result int
    err = db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
    if err != nil {
        log.Fatal("Query failed:", err)
    }

    if result != 1 {
        log.Fatal("Unexpected result:", result)
    }

    fmt.Println("✓ SQLite test passed")
}
EOF

# Update go.mod with proper module name
go mod edit -module test-sqlite-integration

# Install dependencies and run test
echo -e "  Installing SQLite dependencies..."
go get github.com/mattn/go-sqlite3

echo -e "  Running SQLite test..."
if go run test_db.go; then
    echo -e "${GREEN}✓ SQLite integration test passed${NC}"
else
    echo -e "${RED}✗ SQLite integration test failed${NC}"
    exit 1
fi

cd ..

# Test with Docker if available
if command -v docker &> /dev/null && docker info > /dev/null 2>&1; then
    echo -e "\n${BLUE}Docker detected - testing containerized databases...${NC}"

    # Test PostgreSQL with Docker
    echo -e "\n${BLUE}Testing PostgreSQL with Docker...${NC}"
    mkdir test-postgres-docker
    cd test-postgres-docker
    cp ../go.mod .

    echo -e "  Creating PostgreSQL configuration..."
    if $FOUNDRY_TEST add db postgres --with-docker; then
        echo -e "  ${GREEN}✓ PostgreSQL configuration created${NC}"
    else
        echo -e "  ${RED}✗ Failed to create PostgreSQL configuration${NC}"
        exit 1
    fi

    # Update go.mod with proper module name
    go mod edit -module test-postgres-docker

    # Start PostgreSQL
    echo -e "  Starting PostgreSQL container..."
    docker-compose up -d

    # Wait for PostgreSQL to be ready
    echo -e "  Waiting for PostgreSQL to be ready..."
    for i in {1..30}; do
        if docker-compose exec -T postgres pg_isready -U postgres > /dev/null 2>&1; then
            break
        fi
        if [ $i -eq 30 ]; then
            echo -e "${RED}✗ PostgreSQL failed to start${NC}"
            exit 1
        fi
        sleep 1
    done

    # Create test program
    cat > test_pg.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "test-postgres-docker/internal/database"
)

func main() {
    // Set environment variables
    os.Setenv("DB_HOST", "localhost")
    os.Setenv("DB_PORT", "5432")
    os.Setenv("DB_NAME", "testapp")
    os.Setenv("DB_USER", "postgres")
    os.Setenv("DB_PASSWORD", "postgres")
    os.Setenv("DB_SSLMODE", "disable")

    // Connect to database
    db, err := database.NewConnection()
    if err != nil {
        log.Fatal("Connection failed:", err)
    }
    defer db.Close()

    // Test health check
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := db.Health(ctx); err != nil {
        log.Fatal("Health check failed:", err)
    }

    // Test a simple query
    var result int
    err = db.QueryRow(ctx, "SELECT 1").Scan(&result)
    if err != nil {
        log.Fatal("Query failed:", err)
    }

    if result != 1 {
        log.Fatal("Unexpected result:", result)
    }

    fmt.Println("✓ PostgreSQL test passed")
}
EOF

    # Install dependencies and run test
    echo -e "  Installing PostgreSQL dependencies..."
    go get github.com/jackc/pgx/v5
    go get github.com/jackc/pgx/v5/pgxpool

    echo -e "  Running PostgreSQL test..."
    if go run test_pg.go; then
        echo -e "${GREEN}✓ PostgreSQL integration test passed${NC}"
    else
        echo -e "${RED}✗ PostgreSQL integration test failed${NC}"
        docker-compose logs postgres
        exit 1
    fi

    # Stop PostgreSQL
    docker-compose down
    cd ..

else
    echo -e "\n${YELLOW}⚠ Docker not available - skipping containerized database tests${NC}"
fi

# Summary
echo -e "\n${GREEN}✅ Database integration tests completed!${NC}"
echo -e "\nTest output saved in: ${BLUE}$TEST_DIR${NC}"

# Clean up
cd "$BASE_DIR"
rm -f foundry-test

echo -e "\n${BLUE}Next steps:${NC}"
echo -e "1. Run the basic test: ${GREEN}./scripts/test-database.sh${NC}"
echo -e "2. Run this integration test: ${GREEN}./scripts/test-database-integration.sh${NC}"
echo -e "3. Test with your own database setup"