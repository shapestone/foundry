# Test Directory Structure

This directory contains all tests for the Foundry project, organized by package.

## Directory Structure

```
test/
├── scaffolder/                 # Scaffolder package tests
│   ├── handler_scaffolder_test.go    # Handler scaffolding tests
│   ├── helpers_test.go               # Helper function tests  
│   ├── integration_test.go           # Integration tests
│   └── README.md                     # This file
├── parser/                     # Parser package tests (future)
└── templates/                  # Template package tests (future)
```

## Test Categories

### Unit Tests
- **Location**: `test/scaffolder/*_test.go`
- **Purpose**: Test individual functions and methods in isolation
- **Speed**: Fast (< 1 second)
- **Dependencies**: Mocked
- **Run with**: `go test -short ./test/scaffolder/...`

### Integration Tests
- **Location**: `test/scaffolder/integration_test.go`
- **Purpose**: Test components working together with real dependencies
- **Speed**: Slower (1-10 seconds)
- **Dependencies**: Real file system, temporary directories
- **Run with**: `go test -run Integration ./test/scaffolder/...`

### Benchmark Tests
- **Location**: Embedded in test files with `Benchmark` prefix
- **Purpose**: Performance testing and regression detection
- **Run with**: `go test -bench=. ./test/scaffolder/...`

## Running Tests

### Quick Development Workflow
```bash
# Run only fast unit tests
make test-scaffolder-unit

# Run all scaffolder tests
make test-scaffolder

# Run with coverage report
make test-scaffolder-coverage

# Run with race detection
make test-scaffolder-race
```

### Individual Test Commands
```bash
# Run specific test
go test -v ./test/scaffolder -run TestScaffolder_CreateHandler_Success

# Run all tests with coverage
go test -v ./test/scaffolder -cover

# Run benchmarks
go test -v ./test/scaffolder -bench=BenchmarkScaffolder_CreateHandler

# Run integration tests only
go test -v ./test/scaffolder -run Integration
```

## Test Conventions

### File Naming
- `*_test.go` - Standard test files
- `integration_test.go` - Integration tests
- `helpers_test.go` - Test helper functions
- `mock_*.go` - Mock implementations (if needed)

### Test Function Naming
- `TestPackage_Function_Scenario` - Unit tests
- `TestPackage_Integration_Scenario` - Integration tests
- `BenchmarkPackage_Function` - Benchmark tests
- `ExamplePackage_Function` - Example tests

### Test Structure (Given-When-Then)
```go
func TestScaffolder_CreateHandler_Success(t *testing.T) {
    // Given - Set up test data and mocks
    mockFS := newMockFileSystem()
    spec := &scaffolder.HandlerSpec{...}
    
    // When - Execute the function under test
    result, err := scaffolder.CreateHandler(ctx, spec)
    
    // Then - Assert expected outcomes
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    // ... more assertions
}
```

## Mock Implementations

### Available Mocks
- `mockFileSystem` - Mock file system operations
- `mockTemplateRenderer` - Mock template rendering
- `mockProjectAnalyzer` - Mock project analysis
- `mockUserInteraction` - Mock user prompts

### Creating New Mocks
```go
type mockNewInterface struct {
    // Fields for controlling behavior
    returnValue string
    shouldError bool
}

func (m *mockNewInterface) Method(param string) (string, error) {
    if m.shouldError {
        return "", errors.New("mock error")
    }
    return m.returnValue, nil
}
```

## Coverage Expectations

### Current Coverage Targets
- **Unit Tests**: 85-95% line coverage
- **Integration Tests**: 70-80% line coverage
- **Combined**: 80-90% line coverage

### Measuring Coverage
```bash
# Generate coverage report
go test -coverprofile=coverage.out ./test/scaffolder/...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

## Best Practices

### Test Organization
1. **One test file per source file** when possible
2. **Group related tests** in the same file
3. **Use descriptive test names** that explain the scenario
4. **Keep tests focused** - one concept per test

### Test Data Management
1. **Use table-driven tests** for multiple scenarios
2. **Create test fixtures** for complex data
3. **Use temporary directories** for file operations
4. **Clean up resources** in defer statements

### Assertion Guidelines
1. **Test behavior, not implementation** details
2. **Use specific assertions** rather than generic ones
3. **Provide helpful error messages** with context
4. **Test both success and failure paths**

### Performance Considerations
1. **Mark slow tests** with `testing.Short()` checks
2. **Use parallel tests** where appropriate with `t.Parallel()`
3. **Benchmark critical paths** with `Benchmark*` functions
4. **Profile memory usage** with `-benchmem` flag

## Debugging Tests

### Common Issues
1. **Race conditions** - Run with `-race` flag
2. **Resource leaks** - Check temp file cleanup
3. **Flaky tests** - Look for timing dependencies
4. **Mock setup** - Verify mock expectations

### Debugging Commands
```bash
# Run with verbose output
go test -v ./test/scaffolder/...

# Run with race detection
go test -race ./test/scaffolder/...

# Run specific test with debugging
go test -v -run TestSpecificTest ./test/scaffolder/...

# Get test coverage details
go test -cover -coverprofile=coverage.out ./test/scaffolder/...
go tool cover -html=coverage.out
```

## Future Enhancements

### Planned Test Additions
- [ ] Property-based testing with fuzzing
- [ ] Performance regression testing
- [ ] Cross-platform testing
- [ ] Docker-based integration testing

### Test Infrastructure Improvements
- [ ] Shared test utilities package
- [ ] Test data generators
- [ ] Custom assertion helpers
- [ ] Test reporting dashboard