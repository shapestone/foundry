#!/bin/bash

# Test auto-wiring functionality
set -e

echo "🧪 Testing Foundry auto-wire functionality..."
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
TEST_DIR="$TEST_OUTPUT_DIR/test-autowire-$(date +%s)-$$"

# Determine if we're running from scripts directory or root
if [ -d "scripts" ]; then
    # Running from root
    FOUNDRY_CMD="go run ."
    BASE_DIR="."
else
    # Running from scripts directory
    FOUNDRY_CMD="go run .."
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

# Create test project
echo -e "\n${BLUE}1. Creating test project in $TEST_DIR...${NC}"
./foundry-test new "$TEST_DIR"
cd "$TEST_DIR"

# Get relative path to foundry-test
FOUNDRY_TEST="../../foundry-test"

# Test 1: Add handler with auto-wire
echo -e "\n${BLUE}2. Testing handler with --auto-wire flag...${NC}"
# Use printf to send 'y' followed by newline
printf "y\n" | $FOUNDRY_TEST add handler user --auto-wire

# Check if handler was created
if [ -f "internal/handlers/user.go" ]; then
    echo -e "${GREEN}✓ Handler created successfully${NC}"
else
    echo -e "${RED}✗ Handler creation failed${NC}"
    exit 1
fi

# Check if routes were updated
if grep -q "userHandler" "internal/routes/routes.go"; then
    echo -e "${GREEN}✓ Routes auto-wired successfully${NC}"
else
    echo -e "${RED}✗ Routes auto-wiring failed${NC}"
    exit 1
fi

# Test 2: Add another handler without auto-wire
echo -e "\n${BLUE}3. Testing handler without --auto-wire flag...${NC}"
$FOUNDRY_TEST add handler product

# Check if handler was created
if [ -f "internal/handlers/product.go" ]; then
    echo -e "${GREEN}✓ Handler created successfully${NC}"
else
    echo -e "${RED}✗ Handler creation failed${NC}"
    exit 1
fi

# Check that routes were NOT updated
if grep -q "productHandler" "internal/routes/routes.go"; then
    echo -e "${RED}✗ Routes should not be auto-wired${NC}"
    exit 1
else
    echo -e "${GREEN}✓ Routes correctly not auto-wired${NC}"
fi

# Test 3: Wire the product handler manually
echo -e "\n${BLUE}4. Testing manual wire command...${NC}"
# Use printf to send 'y' followed by newline
printf "y\n" | $FOUNDRY_TEST wire handler product

# Check if routes were updated
if grep -q "productHandler" "internal/routes/routes.go"; then
    echo -e "${GREEN}✓ Handler wired successfully${NC}"
else
    echo -e "${RED}✗ Handler wiring failed${NC}"
    exit 1
fi

# Test 4: Try to wire non-existent handler
echo -e "\n${BLUE}5. Testing wire with non-existent handler...${NC}"
if $FOUNDRY_TEST wire handler nonexistent 2>/dev/null; then
    echo -e "${RED}✗ Should have failed for non-existent handler${NC}"
    exit 1
else
    echo -e "${GREEN}✓ Correctly failed for non-existent handler${NC}"
fi

# Test 5: Add middleware with auto-wire flag
echo -e "\n${BLUE}6. Testing middleware with --auto-wire flag...${NC}"
$FOUNDRY_TEST add middleware auth --auto-wire

# Check if middleware was created
if [ -f "internal/middleware/auth.go" ]; then
    echo -e "${GREEN}✓ Middleware created successfully${NC}"
    echo -e "${BLUE}ℹ  Auto-wire for middleware shows 'coming soon' message${NC}"
else
    echo -e "${RED}✗ Middleware creation failed${NC}"
    exit 1
fi

# Test 6: Test cancellation (user says 'n')
echo -e "\n${BLUE}7. Testing cancellation when user says 'n'...${NC}"
# Create a handler and try to wire it but cancel
$FOUNDRY_TEST add handler order
printf "n\n" | $FOUNDRY_TEST wire handler order

# Check that routes were NOT updated
if grep -q "orderHandler" "internal/routes/routes.go"; then
    echo -e "${RED}✗ Routes should not be wired when cancelled${NC}"
    exit 1
else
    echo -e "${GREEN}✓ Correctly cancelled wire operation${NC}"
fi

# Test 7: Verify syntax of generated files
echo -e "\n${BLUE}8. Testing Go syntax validation...${NC}"
if go mod tidy && go build -o /dev/null .; then
    echo -e "${GREEN}✓ All generated code compiles successfully${NC}"
else
    echo -e "${RED}✗ Generated code has syntax errors${NC}"
    exit 1
fi

# Show final routes.go
echo -e "\n${BLUE}9. Final routes.go content:${NC}"
echo "----------------------------------------"
cat internal/routes/routes.go
echo "----------------------------------------"

# Clean up test binary
cd "$BASE_DIR"
rm -f foundry-test

echo -e "\n${GREEN}✅ All auto-wire tests passed!${NC}"
echo -e "\nTest output saved in: ${BLUE}$TEST_DIR${NC}"
echo -e "\nSummary:"
echo -e "- Handler auto-wiring: ${GREEN}Working${NC}"
echo -e "- Manual wire command: ${GREEN}Working${NC}"
echo -e "- Cancellation handling: ${GREEN}Working${NC}"
echo -e "- Error handling: ${GREEN}Working${NC}"
echo -e "- Middleware auto-wiring: ${BLUE}Planned${NC}"