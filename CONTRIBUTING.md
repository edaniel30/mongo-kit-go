# Contributing to mongo-kit-go

Thank you for considering contributing to mongo-kit-go! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Release Process](#release-process)

## Code of Conduct

This project adheres to a code of conduct that all contributors are expected to follow:

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on constructive feedback
- Assume good intentions

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Create a new branch for your feature or bugfix
4. Make your changes
5. Test your changes
6. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.25 or higher
- MongoDB 4.4 or higher (for running tests)
- Git

### Clone and Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/mongo-kit-go.git
cd mongo-kit-go

# Add upstream remote
git remote add upstream https://github.com/edaniel30/mongo-kit-go.git

# Install dependencies
go mod download

# Verify setup
go build ./...
```

### Running MongoDB for Testing

```bash
# Using Docker
docker run -d -p 27017:27017 --name mongo-test mongo:latest

# Or using Docker Compose
docker-compose up -d
```

## Project Structure

```
mongo-kit-go/
├── errors/              # Error types and functions
├── models/              # Configuration and data models
├── internal/            # Internal packages (not exported)
│   └── helpers/        # Internal helper functions
├── middleware/          # Optional middleware integrations
├── examples/            # Usage examples
│   ├── basic/          # Basic usage examples
│   └── advanced/       # Advanced usage examples
├── mongo.go            # Main client implementation
├── operations.go       # CRUD operations
├── query.go            # Query and update builders
├── exports.go          # Public API exports
├── go.mod              # Go module definition
└── README.md           # Documentation
```

## Coding Standards

### General Guidelines

1. **Follow Go conventions**: Use `gofmt`, `golint`, and `go vet`
2. **Package naming**: Use lowercase, single-word package names
3. **Exported names**: Start with uppercase letter
4. **Unexported names**: Start with lowercase letter
5. **Comments**: Document all exported types, functions, and methods in English
6. **Error handling**: Return errors explicitly, don't panic
7. **Context**: Accept `context.Context` as the first parameter

### Code Style

```go
// Good: Clear, documented, follows conventions
// FindOne finds a single document matching the filter
func (c *Client) FindOne(ctx context.Context, database, collection string, filter interface{}, result interface{}) error {
    c.mu.RLock()
    defer c.mu.RUnlock()

    if c.closed {
        return errors.ErrClientClosed()
    }

    // Implementation...
}

// Bad: No documentation, inconsistent naming
func (c *Client) findone(database string, ctx context.Context, f interface{}) error {
    // Implementation...
}
```

### Documentation Comments

```go
// Package mongo provides a MongoDB client wrapper with convenient methods
// for connection management and database operations.
//
// Basic usage:
//
//     client, err := mongo.New(
//         mongo.DefaultConfig(),
//         mongo.WithURI("mongodb://localhost:27017"),
//     )
//
package mongo

// Client wraps MongoDB client with convenience methods
// It is safe for concurrent use across goroutines
type Client struct { ... }

// New creates a new MongoDB client with the given configuration
// Configuration uses the functional options pattern for flexibility
func New(cfg Config, opts ...Option) (*Client, error) { ... }
```

### Error Handling

```go
// Create custom error types
type operationError struct {
    operation string
    cause     error
}

func (e *operationError) Error() string {
    return fmt.Sprintf("operation failed (%s): %v", e.operation, e.cause)
}

func (e *operationError) Unwrap() error {
    return e.cause
}

// Use error wrapping
if err != nil {
    return errors.NewOperationError("insert", err)
}
```

### Thread Safety

```go
type Client struct {
    config Config
    client *mongo.Client
    mu     sync.RWMutex  // Always use mutex for shared state
    closed bool
}

// Read operation - use RLock
func (c *Client) GetConfig() Config {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.config
}

// Write operation - use Lock
func (c *Client) Close(ctx context.Context) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.closed {
        return nil
    }

    c.closed = true
    return c.client.Disconnect(ctx)
}
```

## Testing

### Writing Tests

```go
func TestClient_FindOne(t *testing.T) {
    // Setup
    client, err := mongo.New(
        mongo.DefaultConfig(),
        mongo.WithURI("mongodb://localhost:27017"),
        mongo.WithDatabase("test"),
    )
    if err != nil {
        t.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close(context.Background())

    // Test
    // ... test implementation

    // Assertions
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...

# Run specific test
go test -run TestClient_FindOne

# Verbose output
go test -v ./...
```

### Test Coverage

Aim for at least 80% test coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Submitting Changes

### Before Submitting

1. **Run tests**: Ensure all tests pass
   ```bash
   go test ./...
   ```

2. **Run linters**: Check code quality
   ```bash
   go vet ./...
   golint ./...
   ```

3. **Format code**: Use gofmt
   ```bash
   gofmt -s -w .
   ```

4. **Update documentation**: Update README.md if needed

5. **Commit messages**: Use clear, descriptive commit messages
   ```
   Add support for aggregation pipelines

   - Implemented AggregationBuilder
   - Added Match, Group, Sort stages
   - Added tests for aggregation operations
   ```

### Pull Request Process

1. **Create a branch**: Use descriptive branch names
   ```bash
   git checkout -b feature/add-aggregation-support
   ```

2. **Make changes**: Implement your feature or fix

3. **Commit changes**: Make atomic commits with clear messages
   ```bash
   git add .
   git commit -m "Add aggregation support"
   ```

4. **Push to fork**: Push your changes to your fork
   ```bash
   git push origin feature/add-aggregation-support
   ```

5. **Create PR**: Open a pull request on GitHub
   - Provide a clear description
   - Reference related issues
   - Include examples if applicable

6. **Address feedback**: Respond to code review comments

7. **Merge**: Once approved, your PR will be merged

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tests pass locally
- [ ] Added new tests
- [ ] Updated existing tests

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] No new warnings generated
```

## Release Process

### Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backwards compatible)
- **PATCH**: Bug fixes (backwards compatible)

### Release Checklist

1. Update version in documentation
2. Update CHANGELOG.md
3. Create git tag
4. Push tag to trigger release

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

## Getting Help

- Open an issue for bugs or feature requests
- Start a discussion for questions
- Check existing issues before creating new ones

## Recognition

Contributors will be recognized in:
- GitHub contributors list
- Release notes
- Project documentation

Thank you for contributing to mongo-kit-go!
