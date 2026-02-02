# Query Builders Guide

This document covers the three fluent builder interfaces for constructing MongoDB queries, updates, and aggregations with the Repository pattern.

## QueryBuilder

Build complex find queries with filters, sorting, and options using a fluent interface.

### Basic Usage with Repository

```go
// Simple query
qb := mongokit.NewQueryBuilder().
    Equals("status", "active").
    GreaterThan("age", 18).
    Sort("name", true).
    Limit(10)

users, err := userRepo.FindWithBuilder(ctx, qb)
```

### Comparison Operators

**Equals** - Equality filter
```go
qb.Equals("status", "active")
// Generates: { status: "active" }
```

**NotEquals** - Not equal
```go
qb.NotEquals("status", "banned")
// Generates: { status: { $ne: "banned" } }
```

**GreaterThan / GreaterThanOrEqual**
```go
qb.GreaterThan("age", 18)
qb.GreaterThanOrEqual("age", 18)
// Generates: { age: { $gt: 18 } } or { age: { $gte: 18 } }
```

**LessThan / LessThanOrEqual**
```go
qb.LessThan("price", 100)
qb.LessThanOrEqual("price", 100)
// Generates: { price: { $lt: 100 } } or { price: { $lte: 100 } }
```

### Array Operators

**In** - Value in array
```go
qb.In("role", "admin", "moderator", "editor")
// Generates: { role: { $in: ["admin", "moderator", "editor"] } }
```

**NotIn** - Value not in array
```go
qb.NotIn("status", "banned", "suspended")
// Generates: { status: { $nin: ["banned", "suspended"] } }
```

### Field Existence

**Exists** - Check if field exists
```go
qb.Exists("email", true)  // Field must exist
qb.Exists("deleted_at", false)  // Field must not exist
// Generates: { email: { $exists: true } }
```

### Pattern Matching

**Regex** - Regular expression matching
```go
qb.Regex("name", "^John", "i")  // Case-insensitive, starts with "John"
// Generates: { name: { $regex: "^John", $options: "i" } }
```

### Logical Operators

**And** - Combine conditions with AND
```go
condition1 := bson.D{{Key: "age", Value: bson.M{"$gte": 18}}}
condition2 := bson.D{{Key: "age", Value: bson.M{"$lte": 65}}}

qb := mongokit.NewQueryBuilder().
    And(condition1, condition2)
// Generates: { $and: [...] }
```

**Or** - Combine conditions with OR
```go
condition1 := bson.D{{Key: "role", Value: "admin"}}
condition2 := bson.D{{Key: "role", Value: "moderator"}}

qb := mongokit.NewQueryBuilder().
    Or(condition1, condition2)
// Generates: { $or: [...] }
```

**Nor** - Combine conditions with NOR
```go
qb.Nor(condition1, condition2)
// Generates: { $nor: [...] }
```

### Combining Multiple QueryBuilders

**AndConditions** - Combine multiple builders with AND
```go
qb1 := mongokit.NewQueryBuilder().Equals("active", true)
qb2 := mongokit.NewQueryBuilder().GreaterThan("age", 18)

qb := mongokit.NewQueryBuilder().
    AndConditions(qb1, qb2)

users, _ := userRepo.FindWithBuilder(ctx, qb)
```

**OrConditions** - Combine multiple builders with OR
```go
qb1 := mongokit.NewQueryBuilder().Equals("role", "admin")
qb2 := mongokit.NewQueryBuilder().Equals("role", "moderator")

qb := mongokit.NewQueryBuilder().
    OrConditions(qb1, qb2)

users, _ := userRepo.FindWithBuilder(ctx, qb)
```

### Query Options

**Limit** - Limit number of results
```go
qb.Limit(10)
```

**Skip** - Skip number of documents (pagination)
```go
qb.Skip(20).Limit(10)  // Page 3, 10 per page
```

**Sort** - Sort results
```go
qb.Sort("name", true)   // Ascending
qb.Sort("age", false)   // Descending
```

**SortBy** - Custom sort (replaces previous sorts)
```go
qb.SortBy(bson.D{
    {Key: "priority", Value: -1},
    {Key: "created_at", Value: 1},
})
```

**Project** - Select specific fields
```go
qb.Project(bson.M{
    "name": 1,
    "email": 1,
    "_id": 0,
})
```

### Advanced: Raw Expressions

**Where** - Add raw MongoDB expression
```go
qb.Where(bson.M{
    "custom_field": bson.M{"$exists": true},
})

// Also supports bson.D and bson.E
qb.Where(bson.D{{Key: "status", Value: "active"}})
```

### Complete Example

```go
qb := mongokit.NewQueryBuilder().
    Equals("active", true).
    GreaterThanOrEqual("age", 18).
    LessThan("age", 65).
    In("role", "user", "premium").
    Exists("email", true).
    Regex("name", "^[A-Z]", "").  // Starts with capital letter
    Sort("created_at", false).     // Newest first
    Skip(0).
    Limit(20)

users, err := userRepo.FindWithBuilder(ctx, qb)
```

See [examples/query_builders/](../examples/query_builders/) for complete working examples.

---

## UpdateBuilder

Build complex update operations using a fluent interface.

### Basic Usage

```go
ub := mongokit.NewUpdateBuilder().
    Set("status", "active").
    Inc("views", 1).
    CurrentDate("updated_at")

result, err := userRepo.UpdateByID(ctx, id, ub.Build())
```

### Field Update Operators

**Set** - Set field values
```go
ub.Set("name", "Alice").
   Set("email", "alice@example.com")
// Generates: { $set: { name: "Alice", email: "alice@example.com" } }
```

**Unset** - Remove fields
```go
ub.Unset("temp_field", "old_data")
// Generates: { $unset: { temp_field: "", old_data: "" } }
```

### Numeric Operators

**Inc** - Increment value
```go
ub.Inc("views", 1).
   Inc("likes", 5)
// Generates: { $inc: { views: 1, likes: 5 } }
```

**Mul** - Multiply value
```go
ub.Mul("price", 1.1)  // Increase price by 10%
// Generates: { $mul: { price: 1.1 } }
```

**Min** - Update if new value is less
```go
ub.Min("lowest_price", 49.99)
// Only updates if 49.99 < current value
// Generates: { $min: { lowest_price: 49.99 } }
```

**Max** - Update if new value is greater
```go
ub.Max("highest_score", 100)
// Only updates if 100 > current value
// Generates: { $max: { highest_score: 100 } }
```

### Array Operators

**Push** - Add item to array
```go
ub.Push("tags", "featured").
   Push("tags", "trending")
// Generates: { $push: { tags: { $each: ["featured", "trending"] } } }
```

**Pull** - Remove matching items from array
```go
ub.Pull("tags", "deprecated")
// Generates: { $pull: { tags: "deprecated" } }
```

**AddToSet** - Add item only if not present (no duplicates)
```go
ub.AddToSet("categories", "electronics")
// Generates: { $addToSet: { categories: "electronics" } }
```

**Pop** - Remove first or last element
```go
ub.Pop("history", true)  // Remove first
ub.Pop("history", false) // Remove last
// Generates: { $pop: { history: -1 } } or { $pop: { history: 1 } }
```

### Date Operators

**CurrentDate** - Set to current date/time
```go
ub.CurrentDate("updated_at").
   CurrentDate("last_modified")
// Generates: { $currentDate: { updated_at: true, last_modified: true } }
```

### Field Rename

**Rename** - Rename a field
```go
ub.Rename("old_name", "new_name")
// Generates: { $rename: { old_name: "new_name" } }
```

### Complex Update Example

```go
ub := mongokit.NewUpdateBuilder().
    Set("status", "published").
    Set("author", "John Doe").
    Inc("views", 1).
    Push("tags", "featured").
    AddToSet("categories", "tech").
    Pull("flags", "draft").
    CurrentDate("updated_at").
    CurrentDate("published_at")

result, err := articleRepo.UpdateByID(ctx, articleID, ub.Build())
fmt.Printf("Modified %d document(s)\n", result.ModifiedCount)
```

### Update Many Example

```go
ub := mongokit.NewUpdateBuilder().
    Set("verified", true).
    Inc("verification_count", 1).
    CurrentDate("verified_at")

filter := mongokit.NewQueryBuilder().
    Equals("email_confirmed", true).
    Equals("verified", false).
    GetFilter()

result, err := userRepo.UpdateMany(ctx, filter, ub.Build())
```

See [examples/update_builders/](../examples/update_builders/) for complete working examples.

---

## AggregationBuilder

Build aggregation pipelines using a fluent interface.

### Basic Usage

```go
ab := mongokit.NewAggregationBuilder().
    Match(bson.M{"status": "active"}).
    Group("$category", bson.M{
        "count": bson.M{"$sum": 1},
        "total": bson.M{"$sum": "$amount"},
    }).
    Sort(bson.D{{Key: "total", Value: -1}})

// Use primitive.M for aggregation results
aggRepo := mongokit.NewRepository[primitive.M](client, "orders")
results, err := aggRepo.Aggregate(ctx, ab.Build())
```

### Pipeline Stages

**Match** - Filter documents
```go
ab.Match(bson.M{
    "status": "completed",
    "amount": bson.M{"$gte": 100},
})
```

**Group** - Group and aggregate
```go
ab.Group("$customer_id", bson.M{
    "total_orders": bson.M{"$sum": 1},
    "total_spent": bson.M{"$sum": "$amount"},
    "avg_amount": bson.M{"$avg": "$amount"},
})
```

**Sort** - Sort documents
```go
ab.Sort(bson.D{
    {Key: "total", Value: -1},  // Descending
    {Key: "name", Value: 1},    // Ascending
})
```

**Limit** - Limit number of results
```go
ab.Limit(10)
```

**Skip** - Skip documents (pagination)
```go
ab.Skip(20)
```

**Project** - Select/transform fields
```go
ab.Project(bson.M{
    "name": 1,
    "total": 1,
    "discount": bson.M{"$multiply": []any{"$total", 0.1}},
})
```

**Unwind** - Deconstruct array field
```go
ab.Unwind("$items")  // Each array element becomes a document
```

**Lookup** - Join with another collection
```go
ab.Lookup(
    "products",     // from collection
    "product_id",   // local field
    "_id",          // foreign field
    "product_info", // as field name
)
```

**AddStage** - Add custom stage
```go
ab.AddStage(bson.D{{Key: "$count", Value: "total"}})
```

### Complete Aggregation Example

```go
ab := mongokit.NewAggregationBuilder().
    // Filter completed orders
    Match(bson.M{"status": "completed"}).

    // Group by customer
    Group("$customer_id", bson.M{
        "total_orders": bson.M{"$sum": 1},
        "total_spent": bson.M{"$sum": "$total"},
        "avg_order": bson.M{"$avg": "$total"},
    }).

    // Filter customers with >2 orders
    Match(bson.M{"total_orders": bson.M{"$gt": 2}}).

    // Sort by total spent
    Sort(bson.D{{Key: "total_spent", Value: -1}}).

    // Get top 10
    Limit(10).

    // Select fields
    Project(bson.M{
        "customer_id": "$_id",
        "total_spent": 1,
        "total_orders": 1,
        "avg_order": bson.M{"$round": []any{"$avg_order", 2}},
        "_id": 0,
    })

aggRepo := mongokit.NewRepository[primitive.M](client, "orders")
results, err := aggRepo.Aggregate(ctx, ab.Build())

for _, result := range results {
    fmt.Printf("Customer %s: $%.2f (%d orders)\n",
        result["customer_id"],
        result["total_spent"],
        result["total_orders"])
}
```

### Date-Based Aggregation

```go
ab := mongokit.NewAggregationBuilder().
    Match(bson.M{"status": "completed"}).
    Group(bson.M{
        "year": bson.M{"$year": "$order_date"},
        "month": bson.M{"$month": "$order_date"},
    }, bson.M{
        "count": bson.M{"$sum": 1},
        "revenue": bson.M{"$sum": "$total"},
    }).
    Sort(bson.D{
        {Key: "_id.year", Value: -1},
        {Key: "_id.month", Value: -1},
    })
```

See [examples/aggregations/](../examples/aggregations/) for complete working examples.

## Best Practices

1. **Use builders for complex queries** - More readable than raw bson.M
2. **Chain operations** - Fluent interface makes code clean
3. **Type aggregation results as primitive.M** - Aggregations return different structure
4. **Combine with Repository methods** - Use `FindWithBuilder`, `CountWithBuilder`, etc.
5. **Test your queries** - Verify generated BSON matches expectations

## See Also

- [operations.md](operations.md) - All repository operations
- [repository.md](repository.md) - Repository pattern guide
- [examples/](../examples/) - Complete working examples
