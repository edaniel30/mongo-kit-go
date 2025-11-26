# mongo-kit-go

A comprehensive, thread-safe MongoDB client library for Go with convenient methods for connection management and database operations.

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## Features

- 🚀 **Simple & Intuitive API** - Clean, easy-to-use interface for MongoDB operations
- 🔧 **Functional Options Pattern** - Flexible configuration with sensible defaults
- 🔒 **Thread-Safe** - Safe for concurrent use across goroutines
- ⚡ **Connection Pooling** - Built-in connection pool management
- 🎯 **Generic CRUD Operations** - Comprehensive set of database operations
- 🔍 **Query Builders** - Fluent interface for building complex queries
- 📊 **Aggregation Support** - Full support for MongoDB aggregation pipelines
- 🔄 **Transaction Support** - Session management for multi-document transactions
- ⏱️ **Context Support** - All operations support context for timeouts and cancellation
- 🛡️ **Error Handling** - Rich error types with proper error wrapping

## Installation

```bash
go get github.com/edaniel30/mongo-kit-go
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/edaniel30/mongo-kit-go"
    "go.mongodb.org/mongo-driver/bson"
)

type User struct {
    ID    mongo.ObjectID `bson:"_id,omitempty"`
    Name  string         `bson:"name"`
    Email string         `bson:"email"`
}

func main() {
    // Create client with default configuration
    client, err := mongo.New(
        mongo.DefaultConfig(),
        mongo.WithURI("mongodb://localhost:27017"),
        mongo.WithDatabase("myapp"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close(context.Background())

    ctx := context.Background()

    // Insert a document
    user := User{Name: "John", Email: "john@example.com"}
    id, err := client.InsertOne(ctx, "myapp", "users", user)
    if err != nil {
        log.Fatal(err)
    }

    // Find a document
    var found User
    err = client.FindOne(ctx, "myapp", "users", bson.M{"email": "john@example.com"}, &found)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Configuration

### Default Configuration

```go
cfg := mongo.DefaultConfig()
// Returns:
// {
//     URI: "mongodb://localhost:27017",
//     Database: "default",
//     MaxPoolSize: 100,
//     MinPoolSize: 10,
//     ConnectTimeout: 10 * time.Second,
//     ServerSelectionTimeout: 5 * time.Second,
//     SocketTimeout: 10 * time.Second,
//     Timeout: 10 * time.Second,
//     RetryWrites: true,
//     RetryReads: true,
//     ReadPreference: "primary",
// }
```

### Custom Configuration

```go
client, err := mongo.New(
    mongo.DefaultConfig(),
    mongo.WithURI("mongodb://user:pass@host:port"),
    mongo.WithDatabase("production"),
    mongo.WithMaxPoolSize(200),
    mongo.WithMinPoolSize(20),
    mongo.WithTimeout(30 * time.Second),
    mongo.WithAppName("MyService"),
    mongo.WithReplicaSet("rs0"),
    mongo.WithReadPreference("secondaryPreferred"),
    mongo.WithRetryWrites(true),
    mongo.WithDebug(true),
)
```

## Available Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithURI(uri)` | MongoDB connection URI | `mongodb://localhost:27017` |
| `WithDatabase(name)` | Default database name | `default` |
| `WithMaxPoolSize(size)` | Maximum connection pool size | `100` |
| `WithMinPoolSize(size)` | Minimum connection pool size | `10` |
| `WithConnectTimeout(duration)` | Connection timeout | `10s` |
| `WithServerSelectionTimeout(duration)` | Server selection timeout | `5s` |
| `WithSocketTimeout(duration)` | Socket operation timeout | `10s` |
| `WithTimeout(duration)` | Default operation timeout | `10s` |
| `WithAppName(name)` | Application name for logs | `""` |
| `WithReplicaSet(name)` | Replica set name | `""` |
| `WithReadPreference(pref)` | Read preference mode | `primary` |
| `WithRetryWrites(bool)` | Enable retry writes | `true` |
| `WithRetryReads(bool)` | Enable retry reads | `true` |
| `WithDebug(bool)` | Enable debug logging | `false` |
| `WithDirectConnection(bool)` | Direct connection mode | `false` |

## CRUD Operations

### Insert Operations

```go
// Insert one document
id, err := client.InsertOne(ctx, "db", "collection", document)

// Insert many documents
ids, err := client.InsertMany(ctx, "db", "collection", []interface{}{doc1, doc2})
```

### Find Operations

```go
// Find one document
var user User
err := client.FindOne(ctx, "db", "users", bson.M{"email": "test@example.com"}, &user)

// Find all matching documents
var users []User
err := client.Find(ctx, "db", "users", bson.M{"age": bson.M{"$gte": 18}}, &users)

// Count documents
count, err := client.CountDocuments(ctx, "db", "users", bson.M{"active": true})
```

### Update Operations

```go
// Update one document
result, err := client.UpdateOne(ctx, "db", "users",
    bson.M{"email": "test@example.com"},
    bson.M{"$set": bson.M{"age": 30}})

// Update many documents
result, err := client.UpdateMany(ctx, "db", "users",
    bson.M{"age": bson.M{"$lt": 18}},
    bson.M{"$set": bson.M{"minor": true}})

// Replace one document
result, err := client.ReplaceOne(ctx, "db", "users", filter, newDocument)
```

### Delete Operations

```go
// Delete one document
count, err := client.DeleteOne(ctx, "db", "users", bson.M{"email": "test@example.com"})

// Delete many documents
count, err := client.DeleteMany(ctx, "db", "users", bson.M{"active": false})
```

## Query Builder

Build complex queries with a fluent interface:

```go
qb := mongo.NewQueryBuilder()
filter, opts := qb.
    Equals("category", "Electronics").
    GreaterThan("price", 100).
    In("brand", "Apple", "Samsung", "Sony").
    Regex("name", ".*phone.*", "i").
    Sort("price", false).  // descending
    Limit(10).
    Skip(20).
    Build()

var products []Product
err := client.Find(ctx, "db", "products", filter, &products, opts)
```

### Query Builder Methods

- `Equals(key, value)` - Equality filter
- `NotEquals(key, value)` - Not equals filter
- `GreaterThan(key, value)` - Greater than filter
- `GreaterThanOrEqual(key, value)` - Greater than or equal filter
- `LessThan(key, value)` - Less than filter
- `LessThanOrEqual(key, value)` - Less than or equal filter
- `In(key, ...values)` - In array filter
- `NotIn(key, ...values)` - Not in array filter
- `Exists(key, bool)` - Field exists filter
- `Regex(key, pattern, options)` - Regex filter
- `And(...conditions)` - AND logical operator
- `Or(...conditions)` - OR logical operator
- `Limit(n)` - Limit results
- `Skip(n)` - Skip results
- `Sort(field, ascending)` - Sort results

## Update Builder

Build complex update operations:

```go
ub := mongo.NewUpdateBuilder()
update := ub.
    Set("status", "active").
    Inc("views", 1).
    Push("tags", "featured").
    CurrentDate("updated_at").
    Build()

result, err := client.UpdateOne(ctx, "db", "posts", filter, update)
```

### Update Builder Methods

- `Set(key, value)` - Set field value
- `SetMultiple(map)` - Set multiple fields
- `Unset(...keys)` - Remove fields
- `Inc(key, value)` - Increment value
- `Mul(key, value)` - Multiply value
- `Min(key, value)` - Update if less than current
- `Max(key, value)` - Update if greater than current
- `Push(key, value)` - Append to array
- `Pull(key, value)` - Remove from array
- `AddToSet(key, value)` - Add to array if not exists
- `Pop(key, first)` - Remove first/last array element
- `CurrentDate(key)` - Set to current date
- `Rename(old, new)` - Rename field

## Aggregation Pipeline

Build aggregation pipelines with a fluent interface:

```go
ab := mongo.NewAggregationBuilder()
pipeline := ab.
    Match(bson.M{"status": "active"}).
    Group("$category", bson.M{
        "count": bson.M{"$sum": 1},
        "avgPrice": bson.M{"$avg": "$price"},
    }).
    Sort(bson.M{"count": -1}).
    Limit(10).
    Build()

var results []bson.M
err := client.Aggregate(ctx, "db", "products", pipeline, &results)
```

### Aggregation Methods

- `Match(filter)` - Filter documents
- `Group(id, fields)` - Group documents
- `Sort(sort)` - Sort documents
- `Limit(n)` - Limit results
- `Skip(n)` - Skip documents
- `Project(projection)` - Select fields
- `Unwind(path)` - Deconstruct arrays
- `Lookup(from, localField, foreignField, as)` - Join collections
- `AddStage(stage)` - Add custom stage

## Advanced Operations

### Distinct Values

```go
categories, err := client.Distinct(ctx, "db", "products", "category", bson.M{})
```

### FindOneAndUpdate

```go
var updated User
err := client.FindOneAndUpdate(ctx, "db", "users", filter, update, &updated)
```

### FindOneAndReplace

```go
var replaced User
err := client.FindOneAndReplace(ctx, "db", "users", filter, newDoc, &replaced)
```

### FindOneAndDelete

```go
var deleted User
err := client.FindOneAndDelete(ctx, "db", "users", filter, &deleted)
```

### Bulk Write

```go
models := []mongo.WriteModel{
    mongo.NewInsertOneModel().SetDocument(doc1),
    mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update),
    mongo.NewDeleteOneModel().SetFilter(filter),
}
result, err := client.BulkWrite(ctx, "db", "collection", models)
```

## Index Management

```go
// Create index
indexName, err := client.CreateIndex(ctx, "db", "users",
    bson.D{{Key: "email", Value: 1}})

// Drop index
err := client.DropIndex(ctx, "db", "users", "email_1")

// List indexes
indexes, err := client.ListIndexes(ctx, "db", "users")
```

## Transaction Support

```go
// Using session
err := client.UseSession(ctx, func(sessCtx mongo.SessionContext) error {
    err := sessCtx.StartTransaction()
    if err != nil {
        return err
    }

    // Perform operations
    _, err = client.InsertOne(sessCtx, "db", "users", user1)
    if err != nil {
        sessCtx.AbortTransaction(sessCtx)
        return err
    }

    _, err = client.InsertOne(sessCtx, "db", "users", user2)
    if err != nil {
        sessCtx.AbortTransaction(sessCtx)
        return err
    }

    return sessCtx.CommitTransaction(sessCtx)
})
```

## Helper Functions

```go
// ObjectID helpers
id := mongo.NewObjectID()
oid, err := mongo.ToObjectID("507f1f77bcf86cd799439011")
oids, err := mongo.ToObjectIDs([]string{"id1", "id2"})
valid := mongo.IsValidObjectID("507f1f77bcf86cd799439011")

// BSON helpers
doc := mongo.ToBSON(map[string]interface{}{"key": "value"})
docs := mongo.ToBSONArray([]map[string]interface{}{...})
merged := mongo.MergeBSON(doc1, doc2, doc3)
```

## Error Handling

```go
_, err := client.FindOne(ctx, "db", "users", filter, &user)
if err != nil {
    if errors.Is(err, mongo.ErrDocumentNotFound()) {
        // Handle not found
    } else if errors.Is(err, mongo.ErrClientClosed()) {
        // Handle closed client
    } else {
        // Handle other errors
    }
}
```

## Connection Management

```go
// Check connection status
if !client.IsClosed() {
    if client.IsConnected(ctx) {
        // Client is connected
    }
}

// Get configuration
config := client.GetConfig()

// List databases
dbs, err := client.ListDatabases(ctx, bson.M{})

// List collections
collections, err := client.ListCollections(ctx, "mydb")

// Close connection
err := client.Close(context.Background())
```

## Examples

See the [examples](examples/) directory for complete working examples:

- [Basic Usage](examples/basic/main.go) - Simple CRUD operations
- [Advanced Usage](examples/advanced/main.go) - Query builders, aggregations, indexes

## Best Practices

1. **Use Context**: Always pass context with appropriate timeouts
2. **Defer Close**: Always defer `client.Close()` after creating the client
3. **Error Handling**: Check and handle errors appropriately
4. **Connection Pooling**: Reuse the same client instance across your application
5. **Indexes**: Create indexes for frequently queried fields
6. **Projections**: Use projections to fetch only required fields
7. **Bulk Operations**: Use bulk operations for multiple writes

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For issues, questions, or contributions, please open an issue on GitHub.
