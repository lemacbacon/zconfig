# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Testing

- `go test ./...` - Run all tests
- `go test -v ./...` - Run all tests with verbose output
- `go test ./... -race` - Run tests with race detection
- `go test -run=TestName` - Run specific test

### Building and Linting

- `go build ./...` - Build all packages
- `go mod tidy` - Clean up module dependencies
- `goimports -w .` - Format code (required by project conventions)
- `go vet ./...` - Static analysis

### Module Management

- `go mod download` - Download dependencies
- `go mod verify` - Verify dependencies

## Architecture

zconfig is a reflection-based configuration and dependency injection library for Go. The architecture centers around three core concepts:

### Core Components

**Processor** (`processor.go`): The central orchestrator that walks struct fields, builds a dependency graph, and executes hooks. It handles:

- Field traversal using reflection
- Dependency resolution with cycle detection
- Hook execution in dependency order
- Help message generation

**Field** (`field.go`): Graph representation of struct fields containing:

- Reflection metadata (Value, Path, Tags)
- Parent/child relationships for traversal
- Configuration state (Key, Configurable, ConfigurationKey)
- Injection metadata

**Repository** (`repository.go`): Configuration source management with:

- Ordered provider chain (Args → Env → custom)
- Parser registry for type conversion
- Thread-safe provider/parser registration

### Key Workflows

**Configuration Flow**: `Configure()` → `Processor.Process()` → field walking → dependency resolution → hook execution (repository configuration + initialization)

**Dependency Injection**: Uses `inject-as` and `inject` tags to share instances between fields. Injection sources are processed before targets in the dependency graph.

**Hook System**: Extensible processing pipeline. Default hooks:

1. `Repository.Hook` - Retrieves and parses configuration values
2. `Initialize` - Calls `Init()` on types implementing `Initializable`

### Provider Priority

1. CLI arguments (`ArgsProvider`) - `--key=value` format
2. Environment variables (`EnvProvider`) - `KEY_NAME` format
3. Dotenv files (`DotenvProvider`) - loads `.env` by default, or specify with `--dotenv=path`
4. Custom providers (e.g., JSON, YAML files)

### Type Support

Built-in parsing for: `encoding.TextUnmarshaller`, `encoding.BinaryUnmarshaller`, integers, floats, strings, string slices, booleans, `time.Duration`, `regexp.Regexp`

### Dotenv Support

The library automatically loads `.env` files from the current directory. Features:

- Default behavior: loads `.env` from current directory if it exists
- Custom path: use `--dotenv=path/to/file.env` to specify a different file
- Format: `KEY=value` with support for quoted values and comments
- Key transformation: `database.url` → `DATABASE_URL` (same as EnvProvider)
- Priority: dotenv values can be overridden by environment variables and CLI arguments

### Key Tags

- `key:"name"` - Configuration key
- `default:"value"` - Default value
- `description:"text"` - Help text
- `inject-as:"key"` - Injection source
- `inject:"key"` - Injection target
