# mongo-kit-go

A clean, type-safe MongoDB client library for Go with intuitive API and comprehensive tooling.

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## Installation

```bash
go get github.com/edaniel30/mongo-kit-go
```

## Quick Start

```go
import "github.com/edaniel30/mongo-kit-go"

// Create client
client, _ := mongo_kit.New(
    mongo_kit.DefaultConfig(),
    mongo_kit.WithURI("mongodb://localhost:27017"),
    mongo_kit.WithDatabase("myapp"),
)
defer client.Close(ctx)

// Insert & Find
id, _ := client.InsertOne(ctx, "users", bson.M{"name": "John"})
var user bson.M
client.FindOne(ctx, "users", bson.M{"_id": id}, &user)
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithURI(uri)` | MongoDB connection URI | `mongodb://localhost:27017` |
| `WithDatabase(name)` | Default database name | `default` |
| `WithMaxPoolSize(size)` | Max connections | `100` |
| `WithTimeout(duration)` | Operation timeout | `10s` |
| `WithAppName(name)` | Application identifier | `""` |
| `WithReplicaSet(name)` | Replica set name | `""` |

See `config.go` for all available options.

## Core Concepts

**Client-Based Operations** - Direct database operations via the client
```go
client.InsertOne(ctx, "users", document)
client.Find(ctx, "users", filter, &results)
```

**Repository Pattern** - Type-safe, collection-specific interface
```go
userRepo := mongo_kit.NewRepository[User](client, "users")
users, _ := userRepo.FindAll(ctx)
```

**Fluent Builders** - Build complex queries, updates, and aggregations
```go
qb := mongo_kit.NewQueryBuilder().Equals("status", "active").Limit(10)
users, _ := userRepo.FindWithBuilder(ctx, qb)
```

**Context Helpers** - Simplified timeout management
```go
ctx, cancel := client.NewContext()        // For CLI/scripts
ctx, cancel := client.WithTimeout(ctx)    // For HTTP handlers
```

## Documentation

Comprehensive guides for each component:

| Guide | Description |
|-------|-------------|
| **[operations](docs/operations.md)** | CRUD operations, indexes, transactions, aggregations |
| **[query](docs/query.md)** | QueryBuilder, UpdateBuilder, AggregationBuilder |
| **[repository](docs/repository.md)** | Type-safe repository pattern with generics |
| **[context](docs/context.md)** | Context helpers and timeout management |

### Quick Links by Topic

**Getting Started:**
- [CRUD Operations](docs/operations.md#crud-operations) - Insert, Find, Update, Delete
- [Basic Queries](docs/operations.md#query-operations) - Count, Exists, Distinct

**Advanced Features:**
- [Repository Pattern](docs/repository.md) - Type-safe operations
- [Query Builder](docs/query.md#querybuilder) - Fluent query interface
- [Aggregations](docs/query.md#aggregationbuilder) - Pipeline builder
- [Transactions](docs/operations.md#transactions) - Multi-document ACID
- [Change Streams](docs/operations.md#change-streams) - Real-time monitoring

**Best Practices:**
- [Context Management](docs/context.md) - Timeout strategies
- [Error Handling](docs/operations.md#error-handling) - Type checking
- [Batch Operations](docs/repository.md#batch-operations) - Bulk writes

## Example: Repository Pattern

```go
type User struct {
    ID    primitive.ObjectID `bson:"_id,omitempty"`
    Name  string             `bson:"name"`
    Email string             `bson:"email"`
}

// Create repository
userRepo := mongo_kit.NewRepository[User](client, "users")

// Type-safe operations
user := User{Name: "Alice", Email: "alice@example.com"}
id, _ := userRepo.Create(ctx, user)

// Query with builder
qb := mongo_kit.NewQueryBuilder().
    Equals("status", "active").
    GreaterThan("age", 18).
    Sort("name", true)

users, _ := userRepo.FindWithBuilder(ctx, qb)
```

See [docs/repository.md](docs/repository.md) for complete examples.

## Best Practices

- **Reuse client** - Create once, use across your application
- **Use contexts** - Always pass context with appropriate timeouts
- **Close connections** - Defer `client.Close()` after creation
- **Repository pattern** - Use for type safety and cleaner code
- **Query builders** - Use for complex queries instead of raw bson.M
- **Handle errors** - Check for `mongo.ErrNoDocuments` and operation errors

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For issues, questions, or feature requests, please [open an issue](https://github.com/edaniel30/mongo-kit-go/issues) on GitHub.
