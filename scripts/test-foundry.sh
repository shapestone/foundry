#!/bin/bash

# Test foundry basic functionality
set -e

echo "🧪 Testing Foundry basic functionality..."
echo "========================================"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create test output directory
TEST_OUTPUT_DIR="test-output"
mkdir -p "$TEST_OUTPUT_DIR"

# Test directory with timestamp to avoid conflicts
TEST_DIR="$TEST_OUTPUT_DIR/test-foundry-$(date +%s)-$$"

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
    # Optionally remove test directory on success
    # rm -rf "$TEST_DIR"
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Build foundry first
echo -e "${BLUE}Building foundry...${NC}"
cd "$BASE_DIR"
go build -o foundry-test ./cmd/foundry

echo -e "\n${BLUE}1. Creating test project in $TEST_DIR...${NC}"
./foundry-test new "$TEST_DIR"

if [ -d "$TEST_DIR" ]; then
    echo -e "${GREEN}✓ Project created successfully${NC}"
else
    echo -e "${RED}✗ Project creation failed${NC}"
    exit 1
fi

echo -e "\n${BLUE}2. Installing dependencies...${NC}"
cd "$TEST_DIR"
if go mod tidy; then
    echo -e "${GREEN}✓ Dependencies installed${NC}"
else
    echo -e "${RED}✗ Failed to install dependencies${NC}"
    exit 1
fi

# Get relative path to foundry-test
FOUNDRY_TEST="../../foundry-test"

echo -e "\n${BLUE}3. Testing handler creation...${NC}"
$FOUNDRY_TEST add handler user

if [ -f "internal/handlers/user.go" ]; then
    echo -e "${GREEN}✓ Handler created successfully${NC}"
else
    echo -e "${RED}✗ Handler creation failed${NC}"
    exit 1
fi

echo -e "\n${BLUE}4. Testing model creation...${NC}"
$FOUNDRY_TEST add model user

if [ -f "internal/models/user.go" ]; then
    echo -e "${GREEN}✓ Model created successfully${NC}"
else
    echo -e "${RED}✗ Model creation failed${NC}"
    exit 1
fi

echo -e "\n${BLUE}5. Testing middleware creation...${NC}"
$FOUNDRY_TEST add middleware auth

if [ -f "internal/middleware/auth.go" ]; then
    echo -e "${GREEN}✓ Middleware created successfully${NC}"
else
    echo -e "${RED}✗ Middleware creation failed${NC}"
    exit 1
fi

echo -e "\n${BLUE}6. Testing if project compiles...${NC}"
if go build -o test-app .; then
    echo -e "${GREEN}✓ Project compiles successfully${NC}"
    rm -f test-app
else
    echo -e "${RED}✗ Project compilation failed${NC}"
    exit 1
fi

# Clean up
cd "$BASE_DIR"
rm -f foundry-test

echo -e "\n${GREEN}✅ All basic tests passed!${NC}"
echo -e "\nTest output saved in: ${BLUE}$TEST_DIR${NC}"
echo -e "\nNext steps to test manually:"
echo -e "  cd $TEST_DIR"
echo -e "  go run ."
echo -e ""
echo -e "To test the API:"
echo -e "  curl http://localhost:8080/"
echo -e "  curl http://localhost:8080/health"
echo -e "  curl http://localhost:8080/api/v1/users"