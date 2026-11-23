# mongo-kit-go

A clean, type-safe MongoDB client library for Go using the Repository pattern with generics.

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## Installation

```bash
go get github.com/edaniel30/mongo-kit-go
```

## Quick Start

```go
import (
    "context"
    mongokit "github.com/edaniel30/mongo-kit-go"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// Define your model
type User struct {
    ID    primitive.ObjectID `bson:"_id,omitempty"`
    Name  string             `bson:"name"`
    Email string             `bson:"email"`
    Age   int                `bson:"age"`
}

// Create client
client, err := mongokit.New(
    mongokit.DefaultConfig(),
    mongokit.WithURI("mongodb://localhost:27017"),
    mongokit.WithDatabase("myapp"),
)
if err != nil {
    log.Fatal(err)
}
defer client.Close(context.Background())

// Create repository
userRepo := mongokit.NewRepository[User](client, "users")

// Use it!
ctx := context.Background()

// Create
user := User{Name: "Alice", Email: "alice@example.com", Age: 25}
id, _ := userRepo.Create(ctx, user)

// Find
user, _ := userRepo.FindByID(ctx, id)
users, _ := userRepo.Find(ctx, map[string]any{"age": map[string]any{"$gte": 18}})

// Update
update := map[string]any{"$set": map[string]any{"age": 26}}
userRepo.UpdateByID(ctx, id, update)

// Delete
userRepo.DeleteByID(ctx, id)
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithURI(uri)` | MongoDB connection URI | `mongodb://localhost:27017` |
| `WithDatabase(name)` | Default database name; `""` uses the database in the URI path | `default` |
| `WithMaxPoolSize(size)` | Max connections | `100` |
| `WithTimeout(duration)` | Operation timeout | `10s` |
| `WithClientOptions(opts)` | Custom driver options | `nil` |
| `WithEncryptionKey(hexKey)` | Enables `Encrypt` / `Decrypt` (AES-256-GCM, 64 hex chars) | disabled |

## Field Encryption

Encrypt sensitive fields before storing them, with one key per database:

```go
client, err := mongokit.New(mongokit.DefaultConfig(),
    mongokit.WithURI(uri),
    mongokit.WithDatabase("myapp"),
    mongokit.WithEncryptionKey(os.Getenv("MONGODB_ENCRYPTION_KEY")), // openssl rand -hex 32
)

doc.Token, err = client.Encrypt(token)   // before saving
token, err := client.Decrypt(doc.Token)  // after reading
```

- Values are base64 of nonce + ciphertext + tag. The same value encrypts differently each time, so encrypted fields cannot be queried by equality.
- **Key check on connect.** The first client stores a known value encrypted with its key in the `_mongokit` collection (`_id: "encryption_check"`); every later client must decrypt it, or `New` fails with `ErrEncryptionKeyMismatch`. A client with the wrong key (e.g. a developer's machine pointed at a shared database) fails at startup instead of writing data nobody else can read. The database user needs write access to create that collection.
- Without a key nothing changes: no `_mongokit` collection, and `Encrypt` / `Decrypt` return `ErrEncryptionDisabled`.
- Never change the key of a database that holds encrypted data: those values can no longer be decrypted.

## Query Builder

Build complex queries with a fluent interface:

```go
qb := mongo_kit.NewQueryBuilder().
    Equals("status", "active").
    GreaterThan("age", 18).
    In("role", "admin", "moderator").
    Sort("name", true).
    Limit(10)

users, _ := userRepo.FindWithBuilder(ctx, qb)
```

**Available operators:** `Equals`, `NotEquals`, `GreaterThan`, `LessThan`, `In`, `NotIn`, `Exists`, `Regex`, `And`, `Or`, `Nor`

## Update Builder

Build complex updates:

```go
ub := mongo_kit.NewUpdateBuilder().
    Set("status", "active").
    Inc("views", 1).
    Push("tags", "featured").
    CurrentDate("updated_at")

userRepo.UpdateByID(ctx, id, ub.Build())
```

**Available operations:** `Set`, `Unset`, `Inc`, `Mul`, `Min`, `Max`, `Push`, `Pull`, `AddToSet`, `Pop`, `CurrentDate`, `Rename`

## Aggregation Builder

Build aggregation pipelines:

```go
ab := mongo_kit.NewAggregationBuilder().
    Match(bson.M{"status": "active"}).
    Group("$category", bson.M{
        "count": bson.M{"$sum": 1},
        "total": bson.M{"$sum": "$amount"},
    }).
    Sort(bson.D{{Key: "total", Value: -1}})

// Use primitive.M for aggregation results
aggRepo := mongo_kit.NewRepository[primitive.M](client, "orders")
results, _ := aggRepo.Aggregate(ctx, ab.Build())
```

## Examples

Complete working examples are available in the [`examples/`](examples/) directory:

| Example | Description |
|---------|-------------|
| [**basic_crud**](examples/basic_crud/) | Create, Read, Update, Delete operations |
| [**query_builders**](examples/query_builders/) | Complex queries with QueryBuilder |
| [**update_builders**](examples/update_builders/) | Update operations with UpdateBuilder |
| [**aggregations**](examples/aggregations/) | Aggregation pipelines with AggregationBuilder |

Run any example:
```bash
go run ./examples/basic_crud/main.go
```

## Documentation

Comprehensive guides for each component:

| Guide | Description |
|-------|-------------|
| [**operations.md**](docs/operations.md) | All repository operations (CRUD, bulk, aggregations) |
| [**query.md**](docs/query.md) | QueryBuilder, UpdateBuilder, AggregationBuilder |
| [**repository.md**](docs/repository.md) | Repository pattern with generics |

## Repository API

### Create Operations
- `Create(ctx, doc)` - Insert single document
- `CreateMany(ctx, docs)` - Insert multiple documents

### Read Operations
- `FindByID(ctx, id)` - Find by ObjectID or string
- `FindOne(ctx, filter, opts...)` - Find single document
- `Find(ctx, filter, opts...)` - Find multiple documents
- `FindAll(ctx, opts...)` - Find all documents
- `FindWithBuilder(ctx, qb)` - Find with QueryBuilder
- `FindOneWithBuilder(ctx, qb)` - Find one with QueryBuilder

### Update Operations
- `UpdateByID(ctx, id, update)` - Update by ID
- `UpdateOne(ctx, filter, update)` - Update single document
- `UpdateMany(ctx, filter, update)` - Update multiple documents
- `Upsert(ctx, filter, update)` - Insert or update

### Delete Operations
- `DeleteByID(ctx, id)` - Delete by ID
- `DeleteOne(ctx, filter)` - Delete single document
- `DeleteMany(ctx, filter)` - Delete multiple documents

### Query Operations
- `Count(ctx, filter)` - Count matching documents
- `CountAll(ctx)` - Count all documents
- `CountWithBuilder(ctx, qb)` - Count with QueryBuilder
- `EstimatedCount(ctx)` - Fast approximate count
- `Exists(ctx, filter)` - Check if document exists
- `ExistsByID(ctx, id)` - Check if ID exists
- `ExistsWithBuilder(ctx, qb)` - Check existence with QueryBuilder

### Other Operations
- `Aggregate(ctx, pipeline, opts...)` - Run aggregation pipeline
- `Drop(ctx)` - Drop entire collection

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

MIT License - see the [LICENSE](LICENSE) file for details.