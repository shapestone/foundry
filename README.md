# Foundry CLI

A powerful Go project scaffolding and code generation tool that helps developers quickly create production-ready Go applications with best practices built-in.

## Features

- 🚀 **Multiple Project Layouts**: Choose from various architectures (Standard, Microservice, DDD, Hexagonal, etc.)
- 📦 **Smart Code Generation**: Generate handlers, models, middleware, services, and more
- 🔧 **Configurable Templates**: Customize templates or create your own
- 🌐 **Remote Layouts**: Import layouts from URLs or GitHub repositories
- 🎯 **Production Ready**: Generated code includes error handling, logging, and best practices
- 🐳 **Container Support**: Docker and docker-compose files included
- 📝 **Well Documented**: Generated projects include comprehensive documentation

## Installation

```bash
go install github.com/shapestone/foundry/cmd/foundry@latest
```

Or download a pre-built binary from the [releases page](https://github.com/shapestone/foundry/releases).

## Quick Start

### Create a New Project

```bash
# Create a new project with the standard layout
foundry new myproject

# Create a project with a specific layout
foundry new myapi --layout=microservice

# Create a project with custom settings
foundry new myapp --module=github.com/user/myapp --author="John Doe"
```

### Initialize in Current Directory

```bash
# Initialize a project in the current directory
foundry init

# Initialize with a specific layout
foundry init --layout=hexagonal
```

### Add Components

```bash
# Add a new REST API handler
foundry add handler users

# Add a data model
foundry add model product

# Add middleware
foundry add middleware auth

# Add a service
foundry add service payment

# Add a repository
foundry add repository user
```

## Layout System

Foundry uses a flexible layout system that allows you to choose different project architectures:

### Available Layouts

- **standard**: Traditional Go project structure with Chi router
- **microservice**: Microservice with gRPC and HTTP support (coming soon)
- **hexagonal**: Hexagonal architecture (ports and adapters) (coming soon)
- **ddd**: Domain-Driven Design layout (coming soon)
- **clean**: Clean Architecture layout (coming soon)
- **flat**: Minimal flat structure (coming soon)
- **web**: Web application with templates (coming soon)

### Managing Layouts

```bash
# List available layouts
foundry layout list

# Get detailed information about a layout
foundry layout info standard

# Add a remote layout
foundry layout add https://github.com/user/foundry-custom-layout

# Update layout registry
foundry layout update
```

### Creating Custom Layouts

Layouts are defined using a manifest file (`layout.manifest.yaml`):

```yaml
name: "custom"
version: "1.0.0"
description: "My custom layout"

structure:
  directories:
    - path: "cmd/{{.ProjectName}}"
    - path: "internal/handlers"
    
  files:
    - template: "project/main.go.tmpl"
      target: "cmd/{{.ProjectName}}/main.go"
      
components:
  handler:
    template: "components/handler.go.tmpl"
    target_dir: "internal/handlers"
```

## Project Structure

A typical project generated with the standard layout:

```
myproject/
├── cmd/myproject/
│   └── main.go           # Application entrypoint
├── internal/
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP handlers
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Data models
│   ├── services/        # Business logic
│   └── repository/      # Data access layer
├── pkg/                 # Public packages
├── docs/                # Documentation
├── scripts/             # Build and deploy scripts
├── .env.example         # Environment variables example
├── .gitignore          # Git ignore rules
├── Dockerfile          # Container configuration
├── docker-compose.yml  # Local development setup
├── foundry.yaml        # Foundry configuration
├── go.mod              # Go module definition
├── Makefile            # Build automation
└── README.md           # Project documentation
```

## Configuration

Foundry can be configured globally via `~/.foundry/config.yaml`:

```yaml
defaults:
  author: "Your Name"
  license: "MIT"
  github_username: "yourusername"
  
layouts:
  registry_url: "https://registry.foundry.dev"
  cache_dir: "~/.foundry/cache"
  
templates:
  custom_dir: "~/.foundry/templates"
```

## Examples

### Create a Microservice

```bash
# Create a microservice with gRPC support
foundry new payment-service --layout=microservice \
  --module=github.com/mycompany/payment-service \
  --vars="grpc_port=50051,http_port=8080"
```

### Add CRUD Handlers

```bash
# Generate a complete CRUD handler
foundry add handler products

# This creates:
# - GET    /products      (list all)
# - GET    /products/{id} (get one)
# - POST   /products      (create)
# - PUT    /products/{id} (update)
# - DELETE /products/{id} (delete)
```

### Custom Templates

Create your own templates in `~/.foundry/templates/`:

```go
// ~/.foundry/templates/custom-handler.go.tmpl
package handlers

import (
    "net/http"
    "github.com/go-chi/chi/v5"
)

type {{.Name}}Handler struct {
    // Add dependencies
}

func New{{.Name}}Handler() *{{.Name}}Handler {
    return &{{.Name}}Handler{}
}

func (h *{{.Name}}Handler) Routes() chi.Router {
    r := chi.NewRouter()
    // Add routes
    return r
}
```

## Advanced Features

### Layout Variables

Layouts can define custom variables:

```bash
# Set custom variables when creating a project
foundry new myapp --vars="cache_enabled=true,max_connections=100"
```

### Component Options

```bash
# Generate component with custom output directory
foundry add handler users --output=internal/api/handlers

# Dry run to see what would be generated
foundry add model product --dry-run

# Force overwrite existing files
foundry add middleware logging --force
```

### Remote Layouts

```bash
# Add a layout from GitHub
foundry layout add github.com/awesome-layouts/foundry-rails-style

# Add a layout from a URL
foundry layout add https://example.com/layouts/custom.tar.gz

# Use the remote layout
foundry new myapp --layout=rails-style
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/shapestone/foundry.git
cd foundry

# Install dependencies
go mod download

# Run tests
make test

# Build the binary
make build
```

## License

Foundry is released under the MIT License. See [LICENSE](LICENSE) for details.

## Acknowledgments

Foundry is inspired by:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Hugo](https://gohugo.io) - Static site generator with excellent templating
- [Rails](https://rubyonrails.org) - Convention over configuration philosophy
- [Create React App](https://create-react-app.dev) - Simple project bootstrapping

## Support

- 📖 [Documentation](https://foundry.dev/docs)
- 💬 [Discord Community](https://discord.gg/foundry)
- 🐛 [Issue Tracker](https://github.com/shapestone/foundry/issues)
- 📧 [Email Support](mailto:support@foundry.dev)