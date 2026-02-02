# Repository operations guide

This document covers all database operations available through the Repository pattern.

## Setup

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
client, _ := mongokit.New(
    mongokit.DefaultConfig(),
    mongokit.WithURI("mongodb://localhost:27017"),
    mongokit.WithDatabase("myapp"),
)
defer client.Close(context.Background())

// Create repository
userRepo := mongokit.NewRepository[User](client, "users")
ctx := context.Background()
```

## Create Operations

### Create - Insert Single Document

```go
user := User{
    Name:  "Alice",
    Email: "alice@example.com",
    Age:   25,
}

id, err := userRepo.Create(ctx, user)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Created ID:", id)
```

### CreateMany - Insert Multiple Documents

```go
users := []User{
    {Name: "Bob", Email: "bob@example.com", Age: 30},
    {Name: "Carol", Email: "carol@example.com", Age: 28},
}

ids, err := userRepo.CreateMany(ctx, users)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created %d users\n", len(ids))
```

## Read Operations

### FindByID - Find by ID

```go
// String ID (automatically converted to ObjectID)
user, err := userRepo.FindByID(ctx, "507f1f77bcf86cd799439011")

// Or ObjectID
objID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
user, err := userRepo.FindByID(ctx, objID)

if errors.Is(err, mongo.ErrNoDocuments) {
    fmt.Println("User not found")
}
```

### FindOne - Find Single Document

```go
filter := map[string]any{"email": "alice@example.com"}
user, err := userRepo.FindOne(ctx, filter)

if err != nil {
    if errors.Is(err, mongo.ErrNoDocuments) {
        // Handle not found
    }
}
```

### Find - Find Multiple Documents

```go
// Simple filter
filter := map[string]any{"age": map[string]any{"$gte": 18}}
users, err := userRepo.Find(ctx, filter)

// With options
opts := options.Find().SetLimit(10).SetSort(bson.D{{Key: "name", Value: 1}})
users, err := userRepo.Find(ctx, filter, opts)
```

### FindAll - Find All Documents

```go
users, err := userRepo.FindAll(ctx)

// With options
opts := options.Find().SetLimit(100)
users, err := userRepo.FindAll(ctx, opts)
```

### FindWithBuilder - Find with QueryBuilder

```go
qb := mongokit.NewQueryBuilder().
    Equals("active", true).
    GreaterThan("age", 18).
    Sort("name", true).
    Limit(10)

users, err := userRepo.FindWithBuilder(ctx, qb)
```

### FindOneWithBuilder - Find One with QueryBuilder

```go
qb := mongokit.NewQueryBuilder().
    Equals("email", "alice@example.com")

user, err := userRepo.FindOneWithBuilder(ctx, qb)
```

## Update Operations

### UpdateByID - Update by ID

```go
update := map[string]any{
    "$set": map[string]any{
        "age": 26,
        "updated_at": time.Now(),
    },
}

result, err := userRepo.UpdateByID(ctx, id, update)
fmt.Printf("Modified %d document(s)\n", result.ModifiedCount)
```

### UpdateOne - Update Single Document

```go
filter := map[string]any{"email": "alice@example.com"}
update := map[string]any{
    "$inc": map[string]any{"age": 1},
}

result, err := userRepo.UpdateOne(ctx, filter, update)
```

### UpdateMany - Update Multiple Documents

```go
filter := map[string]any{"age": map[string]any{"$lt": 18}}
update := map[string]any{
    "$set": map[string]any{"status": "minor"},
}

result, err := userRepo.UpdateMany(ctx, filter, update)
fmt.Printf("Modified %d document(s)\n", result.ModifiedCount)
```

### Upsert - Insert or Update

```go
filter := map[string]any{"email": "dave@example.com"}
update := map[string]any{
    "$set": map[string]any{
        "name": "Dave",
        "age": 35,
    },
}

result, err := userRepo.Upsert(ctx, filter, update)
if result.UpsertedID != nil {
    fmt.Println("Inserted new document")
} else {
    fmt.Println("Updated existing document")
}
```

## Delete Operations

### DeleteByID - Delete by ID

```go
result, err := userRepo.DeleteByID(ctx, id)
if result.DeletedCount == 0 {
    fmt.Println("No document found")
}
```

### DeleteOne - Delete Single Document

```go
filter := map[string]any{"email": "old@example.com"}
result, err := userRepo.DeleteOne(ctx, filter)
```

### DeleteMany - Delete Multiple Documents

```go
filter := map[string]any{"active": false}
result, err := userRepo.DeleteMany(ctx, filter)
fmt.Printf("Deleted %d document(s)\n", result.DeletedCount)
```

## Query Operations

### Count - Count Documents

```go
filter := map[string]any{"age": map[string]any{"$gte": 18}}
count, err := userRepo.Count(ctx, filter)
fmt.Printf("Found %d adults\n", count)
```

### CountAll - Count All Documents

```go
count, err := userRepo.CountAll(ctx)
```

### CountWithBuilder - Count with QueryBuilder

```go
qb := mongokit.NewQueryBuilder().
    Equals("active", true).
    GreaterThan("age", 18)

count, err := userRepo.CountWithBuilder(ctx, qb)
```

### EstimatedCount - Fast Approximate Count

```go
// Fast but approximate (uses collection metadata)
count, err := userRepo.EstimatedCount(ctx)
```

### Exists - Check if Document Exists

```go
filter := map[string]any{"email": "alice@example.com"}
exists, err := userRepo.Exists(ctx, filter)
if exists {
    fmt.Println("User exists")
}
```

### ExistsByID - Check if ID Exists

```go
exists, err := userRepo.ExistsByID(ctx, id)
```

### ExistsWithBuilder - Check Existence with QueryBuilder

```go
qb := mongokit.NewQueryBuilder().
    Equals("email", "alice@example.com")

exists, err := userRepo.ExistsWithBuilder(ctx, qb)
```

## Aggregation Operations

### Aggregate - Run Aggregation Pipeline

```go
// Define pipeline
pipeline := []bson.M{
    {"$match": bson.M{"age": bson.M{"$gte": 18}}},
    {"$group": bson.M{
        "_id": "$status",
        "count": bson.M{"$sum": 1},
        "avgAge": bson.M{"$avg": "$age"},
    }},
    {"$sort": bson.M{"count": -1}},
}

// Create repository for aggregation results (use primitive.M)
aggRepo := mongokit.NewRepository[primitive.M](client, "users")
results, err := aggRepo.Aggregate(ctx, pipeline)

for _, result := range results {
    fmt.Printf("Status: %s, Count: %d, Avg Age: %.1f\n",
        result["_id"], result["count"], result["avgAge"])
}
```

### Aggregate with AggregationBuilder

```go
ab := mongokit.NewAggregationBuilder().
    Match(bson.M{"age": bson.M{"$gte": 18}}).
    Group("$status", bson.M{
        "count": bson.M{"$sum": 1},
        "avgAge": bson.M{"$avg": "$age"},
    }).
    Sort(bson.D{{Key: "count", Value: -1}})

aggRepo := mongokit.NewRepository[primitive.M](client, "users")
results, err := aggRepo.Aggregate(ctx, ab.Build())
```

See [examples/aggregations/](../examples/aggregations/) for complete aggregation examples.

## Collection Operations

### Drop - Drop Collection

**WARNING**: This permanently deletes the entire collection including all documents and indexes.

```go
err := userRepo.Drop(ctx)
if err != nil {
    log.Fatal(err)
}
```

## Error Handling

### Common Error Patterns

```go
// Handle not found
user, err := userRepo.FindByID(ctx, id)
if err != nil {
    if errors.Is(err, mongo.ErrNoDocuments) {
        return nil, fmt.Errorf("user not found")
    }
    return nil, fmt.Errorf("database error: %w", err)
}

// Handle operation errors
result, err := userRepo.UpdateByID(ctx, id, update)
if err != nil {
    var opErr *mongokit.OperationError
    if errors.As(err, &opErr) {
        log.Printf("Operation %s failed: %v", opErr.Op, opErr.Cause)
    }
    return err
}

// Check affected count
if result.ModifiedCount == 0 {
    return fmt.Errorf("no documents were modified")
}
```

## Best Practices

1. **Always use contexts with timeouts**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

2. **Handle ErrNoDocuments explicitly**
```go
if errors.Is(err, mongo.ErrNoDocuments) {
    // This is expected in many cases
}
```

3. **Use UpdateBuilder for complex updates**
```go
ub := mongokit.NewUpdateBuilder().
    Set("status", "active").
    Inc("views", 1).
    CurrentDate("updated_at")

userRepo.UpdateByID(ctx, id, ub.Build())
```

4. **Use QueryBuilder for complex queries**
```go
qb := mongokit.NewQueryBuilder().
    Equals("active", true).
    GreaterThan("age", 18).
    Sort("name", true)

users, _ := userRepo.FindWithBuilder(ctx, qb)
```

5. **Check operation results**
```go
result, err := userRepo.DeleteMany(ctx, filter)
if err != nil {
    return err
}
if result.DeletedCount == 0 {
    log.Println("Warning: No documents were deleted")
}
```

## Complete Examples

For complete working examples with detailed comments, see:
- [examples/basic_crud/](../examples/basic_crud/) - All CRUD operations
- [examples/query_builders/](../examples/query_builders/) - Query building
- [examples/update_builders/](../examples/update_builders/) - Update operations
- [examples/aggregations/](../examples/aggregations/) - Aggregation pipelines
