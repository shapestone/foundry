# Foundry Add Command Requirements

## Overview
The `foundry add` command generates boilerplate code for common components (handlers, models, middleware, services, repositories) in a Go project using predefined layout templates.

---

## Command Interface

### Syntax
```bash
foundry add [component-type] [name]
```

### Arguments
- **component-type** (required): Type of component to generate
- **name** (required): Name of the component to create

### Flags
- `--output, -o`: Custom output directory (optional)
- `--force, -f`: Overwrite existing files (optional, default: false)
- `--dry-run`: Show what would be generated without creating files (optional, default: false)

### Examples
```bash
foundry add handler users
foundry add model product
foundry add middleware auth
foundry add service payment
foundry add repository user
foundry add handler users --output custom/path
foundry add model user --force
foundry add service payment --dry-run
```

---

## Functional Requirements

### FR1: Project Detection
**The command MUST detect if it's running in a valid Foundry project**

#### FR1.1: Configuration File Detection
- MUST check for config files in order: `foundry.yaml`, `.foundry.yaml`, `foundry.yml`, `.foundry.yml`
- MUST parse YAML configuration to extract layout information
- MUST default to "standard" layout if config exists but no layout specified

#### FR1.2: Go Project Validation
- MUST verify `go.mod` exists if no Foundry config found
- MUST default to "standard" layout for Go projects without Foundry config
- MUST reject execution if neither Foundry config nor `go.mod` exists

#### FR1.3: Error Messages
- MUST return clear error: "not in a Foundry project directory: no go.mod found"
- MUST return parse errors for invalid YAML config files

### FR2: Component Name Validation
**The command MUST validate component names according to Go conventions**

#### FR2.1: Basic Character Validation
- MUST accept: letters (a-z, A-Z), numbers (0-9), hyphens (-), underscores (_)
- MUST reject: spaces, dots, slashes, special characters (@, #, %, etc.)
- MUST reject: empty strings
- MUST reject: names starting with numbers

#### FR2.2: Go Language Compliance
- SHOULD reject Go reserved keywords (if, for, func, type, var, const, package, import, interface, struct, etc.)
- SHOULD reject problematic names (test, main, error, init)
- MUST generate valid Go identifiers from component names

#### FR2.3: Validation Error Messages
- MUST return descriptive error: "invalid component name: [specific reason]"

### FR3: Component Type Validation
**The command MUST validate component types against available layouts**

#### FR3.1: Layout Loading
- MUST load layout definition from layout manager
- MUST validate that requested component type exists in layout
- MUST handle layout loading failures gracefully

#### FR3.2: Available Component Types
- MUST support at minimum: handler, model, middleware, service, repository
- MUST provide helpful error message listing available types when invalid type requested
- MUST handle layouts with no defined components

#### FR3.3: Error Messages
- MUST return: "unknown component type 'X'. Available types: handler, model, middleware"
- MUST return: "layout 'Y' does not define any components"

### FR4: File Generation
**The command MUST generate component files based on layout templates**

#### FR4.1: File Naming
- MUST convert component names to snake_case for filenames
- MUST append `.go` extension
- MUST place files in component-specific target directories

#### FR4.2: File Content
- MUST use layout-specific templates
- MUST populate template variables:
    - `Name`: original component name
    - `PackageName`: lowercase directory name without underscores
    - `Type`: component type
- MUST generate valid Go code

#### FR4.3: Directory Handling
- MUST create target directories if they don't exist
- MUST respect custom output directory when specified via `--output` flag
- MUST use component's default target directory when no custom output specified

### FR5: File Overwrite Protection
**The command MUST protect against accidental file overwrites**

#### FR5.1: Existence Check
- MUST check if target file already exists before generation
- MUST prevent overwrite unless `--force` flag specified
- MUST skip existence check in `--dry-run` mode

#### FR5.2: Error Messages
- MUST return: "file already exists: [path] (use --force to overwrite)"

### FR6: Dry Run Mode
**The command MUST support preview mode without file modification**

#### FR6.1: Preview Information
- MUST show: component type, name, and target file path
- MUST show: template that would be used
- MUST show: template variables that would be available
- MUST NOT create any files or directories

#### FR6.2: Output Format
```
Would generate handler 'users' at: handlers/users.go

Template that would be used:
  /path/to/handler.tmpl

Variables available in template:
  Name: users
  PackageName: handlers
  Type: handler
```

### FR7: Success Feedback
**The command MUST provide clear feedback on successful completion**

#### FR7.1: Success Message
- MUST display: "✓ Generated [type] '[name]' at: [path]"

#### FR7.2: Next Steps Guidance
- MUST provide component-type-specific next steps
- MUST include recommendations for:
    - **handler**: router configuration, implementation, testing
    - **model**: migrations, custom methods, validation
    - **middleware**: router integration, configuration, testing
    - **service**: dependency injection, implementation, testing
    - **repository**: service injection, data access methods, testing
    - **default**: code review, testing, documentation

---

## Non-Functional Requirements

### NFR1: Performance
- MUST complete execution in under 5 seconds for typical operations
- MUST handle layout loading efficiently

### NFR2: Usability
- MUST provide helpful error messages with actionable guidance
- MUST support common Go naming conventions (camelCase, PascalCase, snake_case, kebab-case)
- MUST integrate smoothly with existing CLI command structure

### NFR3: Reliability
- MUST validate all inputs before making file system changes
- MUST handle file system errors gracefully
- MUST maintain atomic operations (either complete success or clean failure)

### NFR4: Maintainability
- MUST separate validation, file operations, and output formatting concerns
- MUST support extension with new component types
- MUST provide clear interfaces for layout system integration

---

## Edge Cases and Error Handling

### EC1: File System Issues
- MUST handle read-only directories gracefully
- MUST handle insufficient disk space
- MUST handle permission denied errors
- MUST provide clear error messages for file system failures

### EC2: Configuration Issues
- MUST handle malformed YAML configuration files
- MUST handle missing layout definitions
- MUST handle layout manager initialization failures

### EC3: Template Issues
- MUST handle missing template files
- MUST handle template rendering failures
- MUST validate generated code compiles (future enhancement)

### EC4: Input Edge Cases
- MUST handle Unicode characters in component names appropriately
- MUST handle very long component names
- MUST handle component names with unusual but valid character combinations

---

## Dependencies and Constraints

### Internal Dependencies
- Layout manager system for template loading and component generation
- Cobra command framework for CLI interface
- YAML parser for configuration files

### External Dependencies
- Go module system (`go.mod` files)
- File system access for project detection and file generation
- Template system for code generation

### Constraints
- MUST work within existing Foundry CLI architecture
- MUST be compatible with Go 1.19+
- MUST support Windows, macOS, and Linux file systems
- MUST integrate with existing layout and template systems

---

## Acceptance Criteria

### AC1: Basic Functionality
```bash
# Given: I'm in a valid Foundry project
# When: I run `foundry add handler user`
# Then: A file `handlers/user.go` is created with valid Go code
```

### AC2: Validation
```bash
# Given: I'm in a valid Foundry project  
# When: I run `foundry add handler 123invalid`
# Then: I receive an error about invalid component name
```

### AC3: Dry Run
```bash
# Given: I'm in a valid Foundry project
# When: I run `foundry add handler user --dry-run`
# Then: I see what would be generated but no files are created
```

### AC4: Overwrite Protection
```bash
# Given: I'm in a valid Foundry project and `handlers/user.go` exists
# When: I run `foundry add handler user`
# Then: I receive an error about file already existing
# When: I run `foundry add handler user --force`
# Then: The file is overwritten
```

### AC5: Error Handling
```bash
# Given: I'm not in a Foundry project
# When: I run `foundry add handler user`
# Then: I receive a clear error about not being in a Foundry project
```