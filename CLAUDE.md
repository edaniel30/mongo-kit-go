# mongo-kit-go - Project Context for AI Assistants

This document provides comprehensive architectural and development context for the `mongo-kit-go` project. For usage examples and API details, refer to the documentation in the `docs/` folder and `examples/` directory.

## 1. Project Overview

### Purpose
`mongo-kit-go` is a clean, type-safe MongoDB client library for Go using the **Repository Pattern**:
- **Repository-Only Architecture**: Users interact exclusively through type-safe repositories
- **Private Client**: Internal connection management, not exposed to users
- **Type Safety**: Go generics for compile-time checking
- **Fluent Builders**: QueryBuilder, UpdateBuilder, AggregationBuilder
- **Thread-Safe**: Safe for concurrent use across goroutines
- **Clean API**: No boilerplate, intuitive method names

### Tech Stack
- **Language**: Go 1.25+
- **Database Driver**: `go.mongodb.org/mongo-driver v1.17.1`
- **Testing**:
  - `github.com/stretchr/testify v1.11.1` - Assertions and test helpers
  - `github.com/testcontainers/testcontainers-go v0.40.0` - Integration testing with real MongoDB containers
  - `github.com/testcontainers/testcontainers-go/modules/mongodb v0.40.0` - MongoDB-specific testcontainers
- **Code Quality**:
  - `golangci-lint` - Static analysis and linting
  - `pre-commit` - Git hooks for automated checks

### Entry Points & How to Run

**Installation**:
```bash
go get github.com/edaniel30/mongo-kit-go
```

**Basic Usage** (Repository-Only):
```go
import mongo_kit "github.com/edaniel30/mongo-kit-go"

// Define model
type User struct {
    ID    primitive.ObjectID `bson:"_id,omitempty"`
    Name  string             `bson:"name"`
    Email string             `bson:"email"`
}

// Create client (private, not exposed)
client, err := mongo_kit.New(
    mongo_kit.DefaultConfig(),
    mongo_kit.WithURI("mongodb://localhost:27017"),
    mongo_kit.WithDatabase("myapp"),
)
defer client.Close(ctx)

// Create repository (type-safe public API)
userRepo := mongo_kit.NewRepository[User](client, "users")

// Use repository methods
user := User{Name: "Alice", Email: "alice@example.com"}
id, _ := userRepo.Create(ctx, user)
user, _ = userRepo.FindByID(ctx, id)
```

**Examples**:
Complete working examples are in the `examples/` directory:
- **[examples/basic_crud/](examples/basic_crud/)** - All CRUD operations
- **[examples/query_builders/](examples/query_builders/)** - QueryBuilder usage
- **[examples/update_builders/](examples/update_builders/)** - UpdateBuilder usage
- **[examples/aggregations/](examples/aggregations/)** - AggregationBuilder usage

Run examples from repository root:
```bash
go run ./examples/basic_crud/main.go
```

**Documentation**:
- **[docs/operations.md](docs/operations.md)** - All repository operations
- **[docs/repository.md](docs/repository.md)** - Repository pattern guide
- **[docs/query.md](docs/query.md)** - Query/Update/Aggregation builders
- **[README.md](README.md)** - Quick start and API overview

**Running Tests**:
```bash
make test                    # Run all tests
make test-coverage           # Run tests with coverage report (73.8% current)
make test-coverage-html      # Generate HTML coverage report
make test-race              # Run tests with race detector
```

**Code Quality**:
```bash
golangci-lint run           # Run linters
pre-commit install          # Install git hooks
make setup                  # Install pre-commit hooks
```

### Environment Setup
- **Go Version**: 1.25+ required
- **MongoDB**: Local instance or container for development
- **Docker**: Required for integration tests (testcontainers)
- **Make**: For running common commands
- **golangci-lint**: For code quality checks
- **pre-commit**: For git hooks

## 2. Project Structure

### Directory Layout
```
mongo-kit-go/
├── .github/
│   ├── workflows/          # CI/CD workflows
│   └── release.yml         # Release configuration
├── docs/                   # Comprehensive user documentation
│   ├── operations.md       # Repository operations guide
│   ├── query.md           # Query/Update/Aggregation builders guide
│   └── repository.md      # Repository pattern guide with examples
├── examples/              # Complete working examples
│   ├── basic_crud/        # CRUD operations example
│   ├── query_builders/    # QueryBuilder example
│   ├── update_builders/   # UpdateBuilder example
│   ├── aggregations/      # AggregationBuilder example
│   └── README.md          # Examples guide
├── testing/               # Test helpers and utilities
│   └── testhelpers.go    # MongoDB testcontainer setup
├── client.go             # Core client implementation (PRIVATE)
├── config.go             # Configuration and options
├── errors.go             # Custom error types
├── operations.go         # Database operations (private methods)
├── query.go              # Query/Update/Aggregation builders
├── repository.go         # Generic repository implementation (PUBLIC API)
├── *_test.go             # Unit tests
├── *_integration_test.go # Integration tests with real MongoDB
├── go.mod                # Go module definition
├── Makefile              # Common development commands
├── README.md             # User-facing documentation
├── CLAUDE.md             # This file - architectural context
├── .golangci.yml         # Linter configuration
└── .pre-commit-config.yaml # Pre-commit hooks configuration
```

### Key Files and Responsibilities

#### Core Library Files (Root Package: `mongo_kit`)

**`client.go`** (135 lines) - Private client lifecycle
- `type client` (unexported) - Private MongoDB client wrapper with `sync.RWMutex`
- `New()` - Public constructor returning private `*client`
- Connection methods: `Close()`
- Internal methods: `getCollection()`, `checkState()`
- **NOT exposed to users** - only used internally by Repository

**`config.go`** (143 lines) - Configuration with functional options
- `Config` struct: URI, Database, MaxPoolSize, Timeout, ClientOptions
- `DefaultConfig()`: Sensible defaults (localhost:27017, 100 pool size, 10s timeout)
- Option functions: `WithURI()`, `WithDatabase()`, `WithMaxPoolSize()`, `WithTimeout()`, `WithClientOptions()`
- `validate()`: Configuration validation

**`operations.go`** (473 lines) - Private database operations
- **All methods are unexported (lowercase)**
- **CRUD**: `insertOne/Many()`, `findOne/Find()`, `updateOne/Many()`, `deleteOne/Many()`
- **Convenience**: `findByID()`, `updateByID()`, `deleteByID()`, `upsertOne()`
- **Queries**: `countDocuments()`, `estimatedDocumentCount()`
- **Aggregations**: `aggregate()`
- **Collections**: `dropCollection()`
- All methods: Acquire `c.mu.RLock()` → Check state → Get collection → Operate → Wrap error
- Helper: `convertToObjectID()` for ID conversion deduplication

> **IMPORTANT**: These methods are NOT accessible to users. Only Repository uses them internally.

**`repository.go`** (228 lines) - PUBLIC API - Type-safe repository pattern
- `Repository[T any]` struct: Generic, collection-specific wrapper
- `NewRepository[T]()`: Public constructor
- **Public type-safe CRUD**: `Create()`, `FindByID()`, `FindOne()`, `Find()`, `UpdateByID()`, etc.
- **Builder integration**: `FindWithBuilder()`, `FindOneWithBuilder()`, `CountWithBuilder()`
- **Convenience**: `Exists()`, `ExistsByID()`, `CountAll()`, `FindAll()`
- **This is the ONLY way users interact with the library**

> See [docs/repository.md](docs/repository.md) for complete examples and patterns.

**`query.go`** (411 lines) - Fluent query, update, and aggregation builders
- **`QueryBuilder`**: Fluent filter and query building
  - Filter operators: `Equals()`, `NotEquals()`, `GreaterThan()`, `LessThan()`, `In()`, `NotIn()`, `Exists()`, `Regex()`
  - Logical operators: `And()`, `Or()`, `Nor()`
  - Query options: `Limit()`, `Skip()`, `Sort()`, `Project()`
  - Composition: `AndConditions()`, `OrConditions()`, `Where()`
  - Helper: `combineConditions()` for deduplication
- **`UpdateBuilder`**: Fluent update operations
  - Field updates: `Set()`, `Unset()`, `Inc()`, `Mul()`, `Min()`, `Max()`
  - Array operations: `Push()`, `Pull()`, `AddToSet()`, `Pop()`
  - Utilities: `CurrentDate()`, `Rename()`
- **`AggregationBuilder`**: Fluent aggregation pipeline
  - Stages: `Match()`, `Group()`, `Sort()`, `Limit()`, `Skip()`, `Project()`, `Unwind()`, `Lookup()`

> See [docs/query.md](docs/query.md) for detailed builder usage and examples.

**`errors.go`** (83 lines) - Custom error types
- `ConfigError`: Configuration validation errors with `Field` and `Message`
- `ConnectionError`: Connection failures wrapping underlying driver errors
- `OperationError`: Database operation errors with `Op` (operation name) and `Cause`
- `ErrClientClosed`: Sentinel error for operations on closed client
- All errors implement `Error()` and `Unwrap()` for proper error chain handling

#### Testing Files

**`testing/testhelpers.go`** - Integration test utilities
- `MongoContainer`: Wrapper for testcontainers MongoDB setup
- `SetupMongoContainer()`: Starts MongoDB 7 with replica set for testing
- `Teardown()`: Cleans up container after tests

**`*_test.go`** - Unit tests for each component
- `config_test.go`: Config validation and options
- `query_test.go`: Query/Update/Aggregation builder tests
- `errors_test.go`: Error type tests

**`*_integration_test.go`** - Integration tests with real MongoDB
- `repository_integration_test.go`: Repository pattern functionality with 31 test cases
- Use `testing.Short()` to skip in CI/fast test runs
- Each test creates a fresh MongoDB container for isolation

#### Examples Files

**`examples/`** - Complete working examples (806 total lines)
- **basic_crud/** (155 lines): All CRUD operations
- **query_builders/** (172 lines): QueryBuilder usage patterns
- **update_builders/** (221 lines): UpdateBuilder operations
- **aggregations/** (258 lines): AggregationBuilder pipelines
- **README.md**: Examples guide with instructions

### Module Dependencies

```
mongo_kit (root package - single flat package)
├── Uses: go.mongodb.org/mongo-driver (MongoDB official driver)
├── Testing depends on: testcontainers, testify
└── All files are in single package (no subpackages)

Dependency Flow (Internal):
client.go ← operations.go (private methods on *client)
client.go ← repository.go (Repository wraps *client privately)
config.go → client.go (Config used in New())
errors.go → all files (error types used everywhere)
query.go → repository.go (builders used in repository)

Public API Flow:
User → Repository[T] → (internal) client → (internal) operations → MongoDB Driver
```

## 3. Architecture Decisions

### Architectural Pattern
**Repository-Only Pattern** with key characteristics:
- Users NEVER access the client directly
- All operations through type-safe `Repository[T]`
- Private client manages connections internally
- Builder pattern for complex queries

### Design Philosophy
1. **Type safety over flexibility**: Go generics for compile-time guarantees
2. **Simplicity over power**: Common operations simple, advanced possible
3. **Explicit over implicit**: No magic, clear method names
4. **Safe by default**: Thread-safety, error wrapping built-in
5. **Repository as sole API**: No direct client access

### Layer Separation

#### Layer 1: Configuration (`config.go`)
- **Responsibility**: Define and validate client configuration
- **Pattern**: Functional options pattern for flexible, backward-compatible configuration
- **No dependencies**: Pure configuration logic

#### Layer 2: Private Client (`client.go`)
- **Responsibility**: MongoDB connection lifecycle and state management
- **Thread-safety**: All methods use `sync.RWMutex` for concurrent access
- **State checking**: `checkState()` ensures client not closed before operations
- **Locking strategy**:
  - Read locks (`RLock`) for all database operations (read-only client state)
  - Write lock (`Lock`) only for `Close()` and state modification
  - `checkState()` must be called within locked context
- **Visibility**: Unexported (lowercase `client`), not accessible to users

#### Layer 3: Private Operations (`operations.go`)
- **Responsibility**: Implement database operations as private methods on `*client`
- **Error handling**: Wrap all MongoDB driver errors in custom `OperationError`
- **Consistency**: All methods follow pattern: lock → check state → get collection → operate → wrap error
- **Visibility**: All methods unexported (lowercase), only called by Repository

#### Layer 4: Public Repository (`repository.go`)
- **Responsibility**: Provide type-safe, collection-specific public API
- **Generic Type**: `Repository[T any]` for type safety
- **Delegation**: All methods delegate to underlying private `*client` methods
- **Visibility**: All methods exported (PascalCase), this is THE public API

#### Layer 5: Builder Utilities (`query.go`)
- **Responsibility**: Provide fluent interfaces for constructing queries, updates, aggregations
- **Independence**: Builders are independent, used by repositories
- **Visibility**: All builders and methods exported for public use

### Dependency Injection Approach
**Constructor Injection** - Dependencies passed at creation time:
```go
// Client receives config via constructor (returns private *client)
client, err := mongo_kit.New(cfg, ...options)

// Repository receives private client via constructor
repo := mongo_kit.NewRepository[User](client, "users")
```

**No DI framework**: Simple, explicit dependency passing. No reflection or container.

### Communication Patterns

#### Within Library
- **Method receivers**: Repository methods → Private client methods → MongoDB Driver
- **No interfaces**: Direct calls, no abstraction overhead
- **Context propagation**: All operations accept `context.Context` as first parameter

#### Error Communication
- **Custom types**: `ConfigError`, `ConnectionError`, `OperationError`
- **Error wrapping**: Use `Unwrap()` to preserve error chains
- **Sentinel errors**: `ErrClientClosed` for specific conditions
- **No error swallowing**: All errors returned to caller with context

## 4. Design Patterns Implemented

### 1. Repository Pattern
**Where**: `repository.go` - `Repository[T any]` struct (lines 18-228)

**Why**:
- Type safety: Compile-time checking of document types
- DRY: No need to repeat collection name in every operation
- Clean API: Domain-specific methods instead of generic client
- Single public interface: Consistency across application

**How**:
```go
type Repository[T any] struct {
    client     *client  // Private client
    collection string
}

userRepo := NewRepository[User](client, "users")
user, err := userRepo.FindByID(ctx, id) // Returns *User, not any
```

**This is the CORE pattern - the entire library is built around this.**

> See [docs/repository.md](docs/repository.md) for comprehensive examples.

### 2. Functional Options Pattern
**Where**: `config.go` (lines 42-120) - `Option` type and `With*()` functions

**Why**:
- Backward-compatible API evolution (add new options without breaking changes)
- Clear, self-documenting configuration
- Optional parameters without complex constructors

**How**:
```go
type Option func(*Config)

func WithURI(uri string) Option {
    return func(c *Config) { c.URI = uri }
}

client, err := mongo_kit.New(
    mongo_kit.DefaultConfig(),
    mongo_kit.WithURI("mongodb://localhost:27017"),
    mongo_kit.WithDatabase("mydb"),
)
```

### 3. Builder Pattern (Fluent Interface)
**Where**: `query.go` - `QueryBuilder`, `UpdateBuilder`, `AggregationBuilder` (lines 15-411)

**Why**:
- Avoid raw BSON construction (error-prone, verbose)
- Readable, chainable query construction
- Type-safe query operators

**Implementation detail**: Uses `bson.D` to preserve field order and `bson.M` for operator values.

> See [docs/query.md](docs/query.md) for all builder methods and examples.

### 4. Wrapper/Facade Pattern
**Where**: Entire library wraps `go.mongodb.org/mongo-driver`

**Why**:
- Simplify complex MongoDB driver API
- Provide opinionated defaults (retry, timeout, error handling)
- Hide complexity while preserving power

**How**:
- Private `client` wraps `*mongo.Client`
- All operations delegate to underlying driver
- Escape hatch: `WithClientOptions()` for advanced configuration

### 5. Error Wrapping Pattern
**Where**: `errors.go` - Custom error types implementing `Unwrap()` (lines 1-83)

**Why**:
- Add context to errors (operation name, field name)
- Preserve original error for `errors.Is()` and `errors.As()`
- Distinguish error categories (config, connection, operation)

**How**:
```go
type OperationError struct {
    Op    string
    Cause error
}

func (e *OperationError) Unwrap() error { return e.Cause }

// Usage
err := userRepo.FindOne(ctx, filter)
var opErr *OperationError
if errors.As(err, &opErr) {
    fmt.Printf("Operation %s failed: %v", opErr.Op, opErr.Cause)
}
```

### 6. Singleton-Like Client
**Where**: `client.go` - Private `client` intended for reuse

**Why**:
- Connection pooling efficiency (don't create multiple clients)
- Expensive to create (connection establishment, ping verification)
- Thread-safe for concurrent use

**How**:
- Create once in `main()` or init
- Pass to repository constructors
- Close on shutdown

**Anti-pattern**: Creating new client per request (wastes connections)

## 5. Coding Style & Conventions

### Naming Conventions

#### Files
- **Pattern**: `<component>.go` for implementation, `<component>_test.go` for tests
- **Integration tests**: `<component>_integration_test.go`
- **Examples**: `client.go`, `query_test.go`, `repository_integration_test.go`

#### Variables
- **camelCase** for local variables: `userRepo`, `filter`, `result`
- **Single letter** for common short-lived: `i`, `v`, `ok`
- **Descriptive** for longer scope: `result`, `document`, `collection`

#### Functions
- **Exported (Public)**: PascalCase - `New()`, `NewRepository()`, `FindOne()`, `UpdateBuilder()`
- **Unexported (Private)**: camelCase - `checkState()`, `getCollection()`, `insertOne()`
- **Constructors**: `New()` or `New<Type>()` - `New()`, `NewRepository()`, `NewQueryBuilder()`

#### Types
- **Exported (Public)**: PascalCase - `Repository`, `QueryBuilder`, `Config`
- **Unexported (Private)**: camelCase - `client` (the internal client type)

#### Constants & Errors
- **Exported errors**: PascalCase with `Err` prefix - `ErrClientClosed`

### Code Organization Within Files

**Standard file structure**:
```go
// 1. Package declaration and doc comment
// Package mongo_kit provides a MongoDB client wrapper...
package mongo_kit

// 2. Imports (grouped: stdlib, external, internal)
import (
    "context"
    "fmt"

    "go.mongodb.org/mongo-driver/mongo"
)

// 3. Type definitions
type Repository[T any] struct { ... }

// 4. Constructors
func NewRepository[T any](...) *Repository[T] { ... }

// 5. Public methods (grouped by functionality)
func (r *Repository[T]) Create(...) { ... }
func (r *Repository[T]) FindByID(...) { ... }

// 6. Private helper methods (if any)
func convertToObjectID(...) { ... }
```

### Import Ordering
Follows `goimports` standard with local prefix:
```go
import (
    // 1. Standard library
    "context"
    "errors"
    "fmt"

    // 2. External dependencies
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"

    // 3. Local packages (if any)
    testhelpers "github.com/edaniel30/mongo-kit-go/testing"
)
```

### Comments and Documentation Standards

#### Package Comments
```go
// Package mongo_kit provides a type-safe MongoDB client library using the Repository pattern.
package mongo_kit
```

#### Function Comments
All exported functions must have godoc comments starting with the function name:
```go
// Create inserts a single document into the collection and returns its ID.
// The ID can be a string (converted to ObjectID) or primitive.ObjectID.
//
// Example:
//
//	user := User{Name: "Alice", Email: "alice@example.com"}
//	id, err := userRepo.Create(ctx, user)
//
// Returns an error if the insertion fails.
func (r *Repository[T]) Create(ctx context.Context, doc T) (any, error)
```

**Documentation standards**:
- All exported types, functions, methods must have comments
- Start with type/function name
- Include examples for complex functions
- Document error conditions
- Use godoc formatting (lists, code blocks)

#### Inline Comments
Used sparingly for:
- Critical implementation details
- Non-obvious logic
- Locking requirements
- Edge cases

Example:
```go
// IMPORTANT: This method does NOT acquire any locks. The caller MUST hold c.mu.RLock()
func (c *client) checkState() error { ... }
```

### Linting/Formatting Rules

**Configured in `.golangci.yml`**:

**Enabled linters**:
- `gofmt`, `goimports` - Code formatting
- `govet` - Suspicious constructs
- `staticcheck`, `gosimple` - Code simplification
- `errcheck` - Unchecked errors
- `errorlint` - Error wrapping issues
- `gocritic` - Style, performance, diagnostic
- `ineffassign`, `unused` - Dead code
- `misspell` - Spelling
- `bodyclose` - Unclosed HTTP bodies
- `noctx` - HTTP requests without context
- `tparallel` - Incorrect parallel test usage
- `copyloopvar` - Loop variable capture issues

**Settings**:
- `errcheck.check-blank: true` - Catch `_ = err` (error swallowing)
- `govet.enable-all: true` except `fieldalignment` (overly pedantic)
- US English spelling

**Test exclusions**: Test files exempt from some linters (errcheck, gocritic)

**Running linters**:
```bash
golangci-lint run
```

## 6. Error Handling Strategy

### Error Type Hierarchy

```
error (interface)
├── ConfigError          - Configuration validation errors
├── ConnectionError      - MongoDB connection failures
├── OperationError       - Database operation errors
└── ErrClientClosed      - Sentinel error for closed client

All errors from MongoDB driver are wrapped in OperationError
Special case: mongo.ErrNoDocuments passed through unwrapped for clearer handling
```

### Error Propagation

**Pattern**: Wrap → Return → Inspect at call site

```go
// In library (wrap errors in private operations)
func (c *client) findOne(ctx context.Context, ...) error {
    err := coll.FindOne(ctx, filter, opts...).Decode(result)
    if err != nil {
        // Exception: Pass through ErrNoDocuments for clearer checking
        if errors.Is(err, mongo.ErrNoDocuments) {
            return err
        }
        return newOperationError("find one", err)
    }
    return nil
}

// In user code (inspect errors)
user, err := userRepo.FindOne(ctx, filter)
if errors.Is(err, mongo.ErrNoDocuments) {
    // Handle not found
} else if err != nil {
    var opErr *OperationError
    if errors.As(err, &opErr) {
        log.Printf("Operation %s failed: %v", opErr.Op, opErr.Cause)
    }
}
```

### Logging Approach

**This library does NOT log** - caller decides what to log

**Rationale**:
- Libraries shouldn't make logging decisions
- Errors provide enough context for callers to log
- Avoids dependency on logging framework

**Recommendation for users**:
```go
// Log at call site with appropriate context
err := userRepo.FindOne(ctx, filter)
if err != nil {
    log.Error("failed to find user",
        "error", err,
        "filter", filter,
    )
}
```

## 7. Data Flow

For detailed usage examples, see:
- **[docs/operations.md](docs/operations.md)** - Complete guide to all operations
- **[docs/repository.md](docs/repository.md)** - Type-safe repository pattern
- **[docs/query.md](docs/query.md)** - Query/Update/Aggregation builders
- **[examples/](examples/)** - Complete working examples

### Request/Response Lifecycle

```
User Code
    ↓ userRepo.FindOne(ctx, filter)
Repository.FindOne() (repository.go:69)
    ↓ Call r.client.findOne(ctx, r.collection, filter, &result)
client.findOne() (operations.go:xxx) [PRIVATE]
    ↓ Acquire c.mu.RLock()
    ↓ Check client state (not closed)
    ↓ Get collection: c.getCollection(collection)
    ↓ Call MongoDB driver: coll.FindOne(ctx, filter).Decode(&result)
    ↓ Wrap errors (except ErrNoDocuments)
    ↓ Release lock
    ↓ Return error or nil
Repository.FindOne()
    ↓ Return *T (typed result) or error
User Code
    ↓ Get typed result (no type assertion needed)
```

### Data Validation

**Configuration Validation** - At client creation (`config.go:124`)
- URI, Database, MaxPoolSize, Timeout validated before connection

**Runtime Validation** - Minimal, rely on MongoDB driver
- ObjectID validation in `*ByID()` methods via helper
- Most validation delegated to MongoDB driver

**Document Validation** - Not enforced by library
- Users responsible for struct tags: `bson:"fieldname"`
- MongoDB schema validation can be configured via collection options

## 8. Testing Strategy

### Test Types Implemented

#### Unit Tests (`*_test.go`)
**Purpose**: Test individual components in isolation without MongoDB

**What's tested**:
- Configuration validation logic (`config_test.go`)
- Query/Update/Aggregation builder construction (`query_test.go`)
- Error type behavior (`errors_test.go`)

**No mocks**: Pure logic testing without external dependencies

#### Integration Tests (`*_integration_test.go`)
**Purpose**: Test against real MongoDB using testcontainers

**Files**:
- `repository_integration_test.go` - Repository pattern with real data (31 test cases)

**Setup**: Each test spins up MongoDB 7 container with replica set

**Skip in short mode**:
```go
if testing.Short() {
    t.Skip("skipping integration test in short mode")
}
```

### Testing Frameworks and Tools

**Assertion Library**: `github.com/stretchr/testify`
- `require.*` - Assertions that stop test on failure
- `assert.*` - Assertions that continue test on failure

**Testcontainers**: `github.com/testcontainers/testcontainers-go/modules/mongodb`
- Real MongoDB 7 with replica set
- Each test gets isolated container
- Automatic cleanup

**Setup helper** (`testing/testhelpers.go`):
```go
func SetupMongoContainer(t *testing.T) *MongoContainer {
    container, err := mongodb.Run(ctx, "mongo:7", mongodb.WithReplicaSet("rs0"))
    // Returns container with connection URI
}
```

**No Mocking**: Integration tests use real MongoDB, not mocks

### How to Run Tests

**All tests**:
```bash
make test
# or
go test -v ./...
```

**Unit tests only** (skip slow integration tests):
```bash
go test -v -short ./...
```

**With coverage** (73.8% current):
```bash
make test-coverage
```

**Coverage HTML report**:
```bash
make test-coverage-html  # Opens in browser
```

**Race detector**:
```bash
make test-race
```

**Single test**:
```bash
go test -v -run TestRepository_Integration/Create_inserts_document
```

### Mocking Strategy

**This project does NOT use mocks**

**Rationale**:
- Integration tests with testcontainers provide real behavior testing
- Mocking MongoDB driver would test mock, not actual behavior
- Real MongoDB container startup is fast enough (~2-5 seconds)

**For application code using mongo_kit**:
```go
// In your application - use interfaces
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*User, error)
    Create(ctx context.Context, user User) error
}

// Real implementation
type MongoUserRepository struct {
    repo *mongo_kit.Repository[UserDocument]
}

// Test fake
type FakeUserRepository struct {
    users map[string]*User
}
```

## 9. Important Conventions & Rules

### Business Rules That Affect Code Structure

1. **Repository-Only Access**
   - Users NEVER access client directly
   - All operations through `Repository[T]`
   - Client is unexported (`type client`)

2. **Single Client Instance Per Application**
   - Client maintains connection pool
   - Create once (e.g., in `main()`), reuse everywhere
   - Pass to repository constructors
   - Close on application shutdown

3. **Context Required for All Operations**
   - Every database operation takes `context.Context` as first parameter
   - Enables timeouts, cancellation, tracing

4. **Thread-Safety Guarantee**
   - `client` is safe for concurrent use (uses `sync.RWMutex`)
   - `Repository` is safe for concurrent use
   - Builders (`QueryBuilder`, etc.) are NOT thread-safe (build once, use once)

5. **No Automatic Reconnection After Close**
   - `client.Close()` is terminal
   - Create new client if needed after close

### Security Considerations

**1. Connection String Security**
- May contain credentials: `mongodb://user:pass@host:port/db`
- Never log connection strings
- Store in environment variables or secret management
- Use TLS for production: `mongodb+srv://...` or `?tls=true`

**2. No Input Sanitization**
- Library does NOT sanitize user input
- MongoDB driver handles query injection via BSON encoding
- Safe when using builders or bson.M

**3. Context Timeouts**
- Always use context with timeout
- Prevents resource exhaustion
- Default: 10 seconds (configurable)

### Performance Considerations

**1. Connection Pooling**
- Default pool size: 100 connections
- Adjust via `WithMaxPoolSize()` based on concurrency
- Too small → operations block waiting
- Too large → wasted resources

**2. Index Strategy**
- Users responsible for creating appropriate indexes
- No automatic index creation

**3. Aggregations vs. Queries**
- Use `Find()` for simple queries (faster)
- Use `Aggregate()` for complex transformations

**4. Estimated vs. Accurate Counts**
- `CountDocuments()` - Accurate, slower (scans documents)
- `EstimatedDocumentCount()` - Fast, less accurate (uses metadata)

**5. Limit Results**
- Always use `.Limit()` or pagination for large result sets
- Unbounded queries can cause memory issues

### Things to Avoid (Anti-Patterns)

**1. DON'T Create Client Per Request**
```go
// ❌ WRONG
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    client, _ := mongo_kit.New(...)  // DON'T
    defer client.Close(context.Background())
}

// ✅ CORRECT - Reuse single client
var globalClient *mongo_kit.client
func main() {
    globalClient, _ = mongo_kit.New(...)
    defer globalClient.Close(context.Background())
}
```

**2. DON'T Ignore Errors**
```go
// ❌ WRONG
userRepo.Create(ctx, user)

// ✅ CORRECT
if _, err := userRepo.Create(ctx, user); err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}
```

**3. DON'T Use Background Context in HTTP Handlers**
```go
// ❌ WRONG
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()  // Can't cancel
}

// ✅ CORRECT
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()  // Cancelled when request completes
}
```

**4. DON'T Reuse Builders Across Goroutines**
```go
// ❌ WRONG - Builders are not thread-safe
qb := NewQueryBuilder()
go func() { qb.Equals("field1", value1) }()  // RACE
go func() { qb.Equals("field2", value2) }()  // RACE

// ✅ CORRECT - One builder per goroutine
go func() {
    qb := NewQueryBuilder().Equals("field1", value1)
}()
```

## 10. Common Commands

### Test Commands
```bash
make test                    # Run all tests (unit + integration)
make test-coverage           # Run with coverage (73.8% current)
make test-coverage-html      # Generate HTML coverage report
make test-race              # Run with race detector
go test -short ./...        # Run unit tests only (skip integration)
```

### Lint Commands
```bash
golangci-lint run           # Run all linters
golangci-lint run --fix     # Auto-fix issues where possible
```

### Build Commands
```bash
go build ./...              # Verify no compile errors
go mod tidy                 # Clean up dependencies
go mod download             # Download dependencies
```

### Format Commands
```bash
gofmt -w .                  # Format code
goimports -w .              # Format and organize imports
```

### Examples Commands
```bash
# Run any example from repository root
go run ./examples/basic_crud/main.go
go run ./examples/query_builders/main.go
go run ./examples/update_builders/main.go
go run ./examples/aggregations/main.go
```

## 11. Future Development Guidelines

### How to Add New Features

#### Adding a New Repository Method

1. **Add private method to `operations.go`** (if needed):
```go
func (c *client) yourOperation(ctx context.Context, collection string, ...) error {
    c.mu.RLock()
    defer c.mu.RUnlock()

    if err := c.checkState(); err != nil {
        return err
    }

    coll := c.getCollection(collection)
    // ... operation logic
    if err != nil {
        return newOperationError("operation name", err)
    }
    return nil
}
```

2. **Add public method to `repository.go`**:
```go
// YourMethod performs your operation on the collection.
func (r *Repository[T]) YourMethod(ctx context.Context, ...) error {
    return r.client.yourOperation(ctx, r.collection, ...)
}
```

3. **Add integration test in `repository_integration_test.go`**

4. **Document in `docs/operations.md`**

#### Adding a New Builder Method

1. **Add method to appropriate builder in `query.go`**:
```go
// YourMethod adds your filter/operation.
func (qb *QueryBuilder) YourMethod(key string, value any) *QueryBuilder {
    // Implementation
    return qb  // Always return builder for chaining
}
```

2. **Add unit test in `query_test.go`**

3. **Document in `docs/query.md`**

4. **Add example in `examples/query_builders/` if complex**

### Checklist for New Operations

When adding new database operations:

- [ ] Add private method to `client` in `operations.go`
- [ ] Acquire `c.mu.RLock()` at start, defer unlock
- [ ] Call `c.checkState()` after lock
- [ ] Use `c.getCollection()` to get collection
- [ ] Wrap errors in `newOperationError()` (except `ErrNoDocuments`)
- [ ] Add corresponding public method to `Repository[T]`
- [ ] Write integration test in `repository_integration_test.go`
- [ ] Add godoc comment with example
- [ ] Document in `docs/operations.md`
- [ ] Run `make test-coverage`
- [ ] Run `golangci-lint run`

### Code Review Criteria

**Must have**:
- [ ] All exported types/functions have godoc comments starting with name
- [ ] Error handling: all errors checked and wrapped appropriately
- [ ] Thread-safety: proper locking if accessing shared state
- [ ] Context: all database operations accept `context.Context`
- [ ] Tests: unit tests and/or integration tests
- [ ] Coverage: maintains 70%+ coverage
- [ ] Linters: passes `golangci-lint run` with no errors
- [ ] No breaking changes without major version bump
- [ ] Repository-only: no direct client exposure

**Anti-patterns to reject**:
- [ ] Error swallowing (`_ = err`)
- [ ] Logging in library code
- [ ] Panics (except for programmer errors in tests)
- [ ] Global mutable state (except tests)
- [ ] Missing thread-safety in concurrent code
- [ ] `context.Background()` where parent context should be used
- [ ] Exposing client directly to users

---

## Quick Reference

### Most Important Files
1. **`repository.go`** - PUBLIC API, type-safe operations (THE interface users see)
2. **`client.go`** - Private client, connection management, thread-safety
3. **`operations.go`** - Private database operations (implementation)
4. **`query.go`** - Query/Update/Aggregation builders
5. **`errors.go`** - Error types and handling

### Documentation Map
- **[README.md](README.md)** - User-facing quick start
- **[docs/operations.md](docs/operations.md)** - Complete operations guide
- **[docs/repository.md](docs/repository.md)** - Repository pattern with examples
- **[docs/query.md](docs/query.md)** - Query builders guide
- **[examples/](examples/)** - 4 complete working examples (806 lines)

### Key Patterns
- **Repository Pattern**: THE core pattern - users only access Repository
- **Functional Options**: Configuration in `config.go`
- **Builder Pattern**: Fluent queries in `query.go`
- **Error Wrapping**: Custom errors in `errors.go`
- **Private Client**: Internal implementation, never exposed

### Common Gotchas
- Builders are NOT thread-safe (create new per goroutine)
- Client should be created once and reused
- `ErrNoDocuments` is NOT wrapped (check with `errors.Is()`)
- All operations require locks (via `RLock()` in operations)
- Client type is private - users can't access it directly

### Testing
- Integration tests require Docker (testcontainers)
- Run `go test -short ./...` to skip integration tests
- Coverage: 73.8% current
- Use `make test-coverage-html` for visual coverage report

---

**Last Updated**: 2026-02-02
**Architecture**: Repository-Only Pattern
**Version**: Reflects codebase after Repository-Only refactoring and cleanup
