# mongo-kit-go - AI Assistant Context

Concise architectural and development context for the `mongo-kit-go` project.

## Project Overview

**Purpose**: Type-safe MongoDB client library using Repository Pattern with Go generics.

**Key Features**:
- Repository-Only Architecture (private client, public repositories)
- Type-safe operations with generics `Repository[T any]`
- Fluent builders: QueryBuilder, UpdateBuilder, AggregationBuilder
- Thread-safe, production-ready

**Tech Stack**: Go 1.25+, MongoDB Driver v1.17.1, Testcontainers for integration tests

## Quick Start

```go
import mongokit "github.com/edaniel30/mongo-kit-go"

// Create client (private)
client, _ := mongokit.New(
    mongokit.DefaultConfig(),
    mongokit.WithURI("mongodb://localhost:27017"),
    mongokit.WithDatabase("myapp"),
)
defer client.Close(ctx)

// Create repository (public API)
userRepo := mongokit.NewRepository[User](client, "users")

// Use it
user, _ := userRepo.FindByID(ctx, id)
users, _ := userRepo.Find(ctx, filter)
```

## Project Structure

```
mongo-kit-go/
├── client.go           # Private client (unexported)
├── repository.go       # Public Repository[T] (PRIMARY API)
├── operations.go       # Private CRUD operations
├── query.go           # Public builders (Query, Update, Aggregation)
├── config.go          # Configuration with functional options
├── errors.go          # Custom error types
├── docs/              # User documentation
│   ├── operations.md  # All repository operations
│   ├── query.md       # Builder patterns
│   └── repository.md  # Repository guide
├── examples/          # 4 complete working examples
│   ├── basic_crud/
│   ├── query_builders/
│   ├── update_builders/
│   └── aggregations/
└── testing/           # Test helpers (testcontainers)
```

## Architecture

**Repository-Centric Pattern**:
- `Client` type is **exported** (public) for advanced use cases
- Users primarily interact through `Repository[T]` (recommended)
- Direct client access available when needed for custom operations
- Type safety enforced at compile time

**Key Types**:
```go
// Public - connection management (exported for advanced use)
type Client struct { ... }

// Public - primary user-facing API (recommended)
type Repository[T any] struct {
    client     *Client
    collection string
}

// Public - query building
type QueryBuilder struct { ... }
type UpdateBuilder struct { ... }
type AggregationBuilder struct { ... }
```

## Important Files

### Core (Root Package)

**`repository.go`** (228 lines) - **PRIMARY USER API**
- `NewRepository[T](client, collection)` - Constructor
- CRUD: `Create()`, `FindByID()`, `Find()`, `Update*()`, `Delete*()`
- Query: `Count()`, `Exists()`, `Aggregate()`
- Builders: `FindWithBuilder()`, `CountWithBuilder()`

**`client.go`** (134 lines) - **PUBLIC** connection management
- Exported `Client` type (available for advanced use)
- Thread-safe with `sync.RWMutex`
- Connection pooling and lifecycle

**`operations.go`** (334 lines) - **INTERNAL** database operations
- Methods on `*Client` (unexported methods)
- Called by Repository methods
- Error wrapping and state checking

**`query.go`** (410 lines) - **PUBLIC** fluent builders
- `QueryBuilder` - filters, sorting, pagination
- `UpdateBuilder` - update operations
- `AggregationBuilder` - aggregation pipelines

**`config.go`** (142 lines) - Configuration
- Functional options pattern
- `WithURI()`, `WithDatabase()`, `WithTimeout()`

**`errors.go`** (83 lines) - Custom errors
- `ConfigError`, `ConnectionError`, `OperationError`
- Proper error wrapping with `Unwrap()`

## Key Patterns

### 1. Repository Pattern (Primary)
```go
userRepo := mongokit.NewRepository[User](client, "users")
user, err := userRepo.FindByID(ctx, id)
```

### 2. Functional Options
```go
client, _ := mongokit.New(
    mongokit.DefaultConfig(),
    mongokit.WithURI("mongodb://..."),
    mongokit.WithTimeout(30*time.Second),
)
```

### 3. Builder Pattern
```go
qb := mongokit.NewQueryBuilder().
    Equals("status", "active").
    GreaterThan("age", 18).
    Sort("name", true).
    Limit(10)
users, _ := userRepo.FindWithBuilder(ctx, qb)
```

### 4. Error Wrapping
```go
// Library wraps errors
return newOperationError("find one", err)

// Users check with errors.As/Is
if errors.Is(err, mongo.ErrNoDocuments) { ... }
```

## Common Operations

### CRUD
```go
// Create
id, _ := repo.Create(ctx, document)

// Read
doc, _ := repo.FindByID(ctx, id)
docs, _ := repo.Find(ctx, filter)

// Update
result, _ := repo.UpdateByID(ctx, id, update)

// Delete
result, _ := repo.DeleteByID(ctx, id)
```

### Aggregations
```go
// Use primitive.M for aggregation results
aggRepo := mongokit.NewRepository[primitive.M](client, "orders")
ab := mongokit.NewAggregationBuilder().
    Match(bson.M{"status": "completed"}).
    Group("$customer", bson.M{"total": bson.M{"$sum": "$amount"}})
results, _ := aggRepo.Aggregate(ctx, ab.Build())
```

## Testing

**Run tests**: `make test` or `go test ./...`
**Coverage**: `make test-coverage` (threshold: 85%)
**Linting**: `golangci-lint run`
**Pre-commit**: `make pre-commit`

**Integration tests** use testcontainers with real MongoDB:
```go
container := testing.SetupMongoContainer(t)
defer container.Teardown(context.Background())
```

## Code Conventions

### Naming
- **Exported**: PascalCase - `NewRepository`, `FindByID`
- **Unexported**: camelCase - `client`, `checkState`
- **Errors**: `Err` prefix - `ErrClientClosed`

### Error Handling
- **Always return errors**, never panic (except tests)
- **Wrap with context**: `newOperationError(op, cause)`
- **Pass through** `mongo.ErrNoDocuments` unwrapped
- **No logging** in library code (caller's responsibility)

### Thread Safety
- `client` uses `sync.RWMutex` for all operations
- Read lock (`RLock`) for database ops
- Write lock (`Lock`) only for `Close()`
- Builders are **NOT** thread-safe (single use)

## Anti-Patterns to Avoid

❌ **Don't create client per request** - reuse single instance
❌ **Don't use Background context** in HTTP handlers - use `r.Context()`
❌ **Don't reuse builders** across goroutines - not thread-safe
❌ **Don't ignore errors** - always check return values
❌ **Don't use `log.Fatalf`** with deferred cleanup - prevents defer execution

## Documentation Map

- **README.md** - User quick start
- **docs/operations.md** - All repository operations with examples
- **docs/query.md** - QueryBuilder, UpdateBuilder, AggregationBuilder
- **docs/repository.md** - Repository pattern guide
- **examples/** - 4 complete working examples

## Development Commands

```bash
make test                 # Run all tests
make test-coverage        # Coverage report (85% threshold)
make test-race           # Race detector
golangci-lint run        # Linting
make pre-commit          # Run all pre-commit hooks
go run ./examples/basic_crud/main.go  # Run example
```

## Quick Reference

**Most Important Files**:
1. `repository.go` - Primary user API
2. `query.go` - Builders for complex queries
3. `client.go` - Connection management (private)
4. `operations.go` - Database operations (private)
5. `errors.go` - Error handling

**Key Concepts**:
- Repository-Only: All operations through `Repository[T]`
- Private Client: Users never touch `client` directly
- Type Safety: Generics for compile-time checking
- Thread Safe: Safe for concurrent use
- Builders: Fluent interfaces for complex queries

---

**Last Updated**: 2026-02-02
**Coverage**: 85.8%
**Version**: Repository-Centric Architecture (Client exported)
