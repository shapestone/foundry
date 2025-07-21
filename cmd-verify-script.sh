#!/bin/bash

# Foundry Command Verification Script
# Tests commands that should report "not implemented" vs those that should work

echo "🔍 Testing Foundry Commands for 'Not Implemented' Status"
echo "========================================================"

# Check if foundry binary exists, if not build it
if ! command -v foundry &> /dev/null && [ ! -f "./foundry" ]; then
    echo "🔨 Building foundry binary..."
    go build -o foundry ./cmd/foundry/
    if [ $? -ne 0 ]; then
        echo "❌ Failed to build foundry binary"
        exit 1
    fi
    echo "✅ Built foundry binary successfully"
    FOUNDRY_CMD="./foundry"
else
    FOUNDRY_CMD="foundry"
fi

echo "Using foundry command: $FOUNDRY_CMD"

# Create a temporary test directory
TEST_DIR="foundry_test_$(date +%s)"
mkdir "$TEST_DIR"
cd "$TEST_DIR"

echo ""
echo "📁 Testing in: $(pwd)"
echo ""

# Initialize a test project first
echo "🚀 Setting up test project..."
$FOUNDRY_CMD init test-project --force
echo ""

echo "🔍 HIGH PRIORITY TESTS"
echo "======================"

echo ""
echo "1. Layout Management Commands:"
echo "------------------------------"

echo "Testing: $FOUNDRY_CMD layout update"
$FOUNDRY_CMD layout update
echo "Exit code: $?"
echo ""

echo "Testing: $FOUNDRY_CMD layout update specific-layout"
$FOUNDRY_CMD layout update nonexistent-layout
echo "Exit code: $?"
echo ""

echo "Testing: $FOUNDRY_CMD layout remove"
$FOUNDRY_CMD layout remove test-layout
echo "Exit code: $?"
echo ""

echo "Testing: $FOUNDRY_CMD layout info"
$FOUNDRY_CMD layout info standard
echo "Exit code: $?"
echo ""

echo "Testing: $FOUNDRY_CMD layout add (already confirmed broken)"
$FOUNDRY_CMD layout add github.com/test/repo
echo "Exit code: $?"
echo ""

echo "2. Wire Commands:"
echo "-----------------"

echo "Testing: $FOUNDRY_CMD wire handler"
$FOUNDRY_CMD wire handler test
echo "Exit code: $?"
echo ""

echo "Testing: $FOUNDRY_CMD wire middleware"
$FOUNDRY_CMD wire middleware auth
echo "Exit code: $?"
echo ""

echo "🔍 MEDIUM PRIORITY TESTS"
echo "========================"

echo ""
echo "3. Layout System Integration:"
echo "-----------------------------"

# Test in a fresh directory
cd ..
mkdir "${TEST_DIR}_layout_test"
cd "${TEST_DIR}_layout_test"

echo "Testing: $FOUNDRY_CMD new with custom layout"
$FOUNDRY_CMD new test-micro --layout=microservice
echo "Exit code: $?"
echo ""

echo "Testing: $FOUNDRY_CMD new with nonexistent layout"
$FOUNDRY_CMD new test-hex --layout=hexagonal
echo "Exit code: $?"
echo ""

echo "Testing: $FOUNDRY_CMD init with custom layout"
mkdir init_test && cd init_test
$FOUNDRY_CMD init --layout=microservice --force
echo "Exit code: $?"
cd ..
echo ""

echo "4. Auto-Wiring Features:"
echo "------------------------"

cd test-micro 2>/dev/null || echo "Skipping auto-wire tests (no test project)"
if [ -d "test-micro" ]; then
    cd test-micro

    echo "Testing: $FOUNDRY_CMD add handler with auto-wire"
    $FOUNDRY_CMD add handler user --auto-wire
    echo "Exit code: $?"
    echo ""

    echo "Testing: $FOUNDRY_CMD add middleware with auto-wire"
    $FOUNDRY_CMD add middleware auth --auto-wire
    echo "Exit code: $?"
    echo ""

    cd ..
fi

echo "🔍 LOW PRIORITY TESTS"
echo "====================="

echo ""
echo "5. Database Generation:"
echo "-----------------------"

cd test-micro 2>/dev/null || mkdir test_db_temp && cd test_db_temp

echo "Testing: $FOUNDRY_CMD add db with migrations"
$FOUNDRY_CMD add db postgres --with-migrations
echo "Exit code: $?"
echo ""

echo "Testing: $FOUNDRY_CMD add db with docker"
$FOUNDRY_CMD add db mysql --with-docker
echo "Exit code: $?"
echo ""

# Cleanup
cd ..
echo ""
echo "🧹 Cleaning up test directories..."
rm -rf "$TEST_DIR" "${TEST_DIR}_layout_test" "test_db_temp" 2>/dev/null

echo ""
echo "✅ Verification complete!"
echo ""
echo "📝 SUMMARY:"
echo "- Commands that should show 'not implemented': layout update/remove, wire commands"
echo "- Commands that should work: basic add commands, init/new with default layouts"
echo "- Commands to investigate: custom layouts, auto-wiring features"