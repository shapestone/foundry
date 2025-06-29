# 🧱 Foundry

> Start small. Build fast. Grow smart.

Forge production-grade Go REST APIs in seconds.

## Install

```bash
go install github.com/shapestone/foundry/cmd/foundry@latest
```

Or build from source:
```bash
git clone https://github.com/shapestone/foundry.git
cd foundry
make install
```

## Usage

Create a new REST API:
```bash
foundry new myapp
cd myapp
go mod tidy
go run .
```

Test your API:
```bash
curl http://localhost:8080/
curl http://localhost:8080/health
```

## What You Get

```
myapp/
├── main.go          # Chi router + server setup
├── go.mod           # Go module file
├── README.md        # Project docs
├── .gitignore       # Git ignore file
└── internal/        # Internal packages
    ├── handlers/    # HTTP handlers (ready for your code)
    └── middleware/  # HTTP middleware (ready for your code)
```

## Philosophy

- Start with the minimum viable tool
- Add features as they're needed
- Keep it simple and fast

## Coming Soon

Based on your feedback:
- `foundry add handler <name>`
- `foundry add model <name>`

## License

MIT