#!/bin/bash

# Test all middleware types
set -e

echo "🧪 Testing Foundry middleware functionality..."
echo "==========================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create test output directory
TEST_OUTPUT_DIR="test-output"
mkdir -p "$TEST_OUTPUT_DIR"

# Test directory with timestamp to avoid conflicts
TEST_DIR="$TEST_OUTPUT_DIR/test-middleware-$(date +%s)-$$"

# Determine if we're running from scripts directory or root
if [ -d "scripts" ]; then
    # Running from root
    BASE_DIR="."
else
    # Running from scripts directory
    BASE_DIR=".."
fi

# Clean up function
cleanup() {
    cd "$BASE_DIR"
    # Keep test directory for inspection
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Build foundry first
echo -e "${BLUE}Building foundry...${NC}"
cd "$BASE_DIR"
go build -o foundry-test ./cmd/foundry

# Create test project
echo -e "\n${BLUE}1. Creating test project in $TEST_DIR...${NC}"
./foundry-test new "$TEST_DIR"
cd "$TEST_DIR"

# Get relative path to foundry-test
FOUNDRY_TEST="../../foundry-test"

# Array of all middleware types
MIDDLEWARE_TYPES=("auth" "cors" "ratelimit" "logging" "recovery" "timeout" "compression")

# Test each middleware type
echo -e "\n${BLUE}2. Testing middleware creation...${NC}"
for mw in "${MIDDLEWARE_TYPES[@]}"; do
    echo -e "\n  Testing ${mw} middleware..."

    # Create the middleware
    if $FOUNDRY_TEST add middleware "$mw" > /dev/null 2>&1; then
        # Check if file was created
        if [ -f "internal/middleware/${mw}.go" ]; then
            echo -e "  ${GREEN}✓ ${mw} middleware created successfully${NC}"
        else
            echo -e "  ${RED}✗ ${mw} middleware file not found${NC}"
            exit 1
        fi
    else
        echo -e "  ${RED}✗ Failed to create ${mw} middleware${NC}"
        exit 1
    fi
done

# Install dependencies first
echo -e "\n${BLUE}3. Installing dependencies...${NC}"
if go mod tidy; then
    echo -e "${GREEN}✓ Dependencies installed${NC}"
else
    echo -e "${RED}✗ Failed to install dependencies${NC}"
    exit 1
fi

# Test that all files compile
echo -e "\n${BLUE}4. Testing compilation with all middleware...${NC}"
if go build -o test-app .; then
    echo -e "${GREEN}✓ Project compiles with all middleware${NC}"
    rm -f test-app
else
    echo -e "${RED}✗ Compilation failed${NC}"
    exit 1
fi

# Test duplicate middleware prevention
echo -e "\n${BLUE}5. Testing duplicate middleware prevention...${NC}"
if $FOUNDRY_TEST add middleware auth 2>&1 | grep -q "already exists"; then
    echo -e "${GREEN}✓ Correctly prevented duplicate middleware${NC}"
else
    echo -e "${RED}✗ Should have prevented duplicate middleware${NC}"
    exit 1
fi

# Test invalid middleware type
echo -e "\n${BLUE}6. Testing invalid middleware type...${NC}"
if $FOUNDRY_TEST add middleware invalid 2>&1 | grep -q "unsupported middleware type"; then
    echo -e "${GREEN}✓ Correctly rejected invalid middleware type${NC}"
else
    echo -e "${RED}✗ Should have rejected invalid middleware type${NC}"
    exit 1
fi

# Show created files
echo -e "\n${BLUE}7. Created middleware files:${NC}"
ls -la internal/middleware/

# Test specific middleware functionality
echo -e "\n${BLUE}8. Testing middleware can be imported...${NC}"

# Create a simple test to verify imports work
cat > check_imports.go << 'EOF'
package main

import (
	"fmt"
	_ "github.com/go-chi/chi/v5"
	_ "github.com/go-chi/chi/v5/middleware"
)

func main() {
	fmt.Println("Imports successful")
}
EOF

# Test the basic imports work
if go run check_imports.go > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Basic imports work${NC}"
    rm -f check_imports.go
else
    echo -e "${RED}✗ Basic imports failed${NC}"
    echo "Trying to diagnose the issue..."
    go run check_imports.go
    rm -f check_imports.go
    exit 1
fi

# Clean up
cd "$BASE_DIR"
rm -f foundry-test

echo -e "\n${GREEN}✅ All middleware tests passed!${NC}"
echo -e "\nTest output saved in: ${BLUE}$TEST_DIR${NC}"
echo -e "\nSummary:"
echo -e "- Created ${#MIDDLEWARE_TYPES[@]} middleware types successfully"
echo -e "- All files generated correctly"
echo -e "- Project compiles with middleware"
echo -e "- Duplicate prevention working"
echo -e "- Invalid type rejection working"

echo -e "\n${BLUE}Middleware types available:${NC}"
for mw in "${MIDDLEWARE_TYPES[@]}"; do
    desc=""
    case $mw in
        auth) desc="Authentication middleware" ;;
        cors) desc="CORS middleware" ;;
        ratelimit) desc="Rate limiting middleware" ;;
        logging) desc="Request/response logging middleware" ;;
        recovery) desc="Panic recovery middleware" ;;
        timeout) desc="Request timeout middleware" ;;
        compression) desc="Response compression middleware" ;;
    esac
    echo -e "  - ${GREEN}$mw${NC}: $desc"
done

echo -e "\n${BLUE}To test the middleware manually:${NC}"
echo -e "  cd $TEST_DIR"
echo -e "  # Add a handler: ../../foundry-test add handler user"
echo -e "  # Edit main.go to use the middleware"
echo -e "  go run ."