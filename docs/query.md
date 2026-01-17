# Query Builders Guide

This document covers the three fluent builder interfaces for constructing MongoDB queries, updates, and aggregations.

## QueryBuilder

Fluent interface for building MongoDB find queries with filters, sorting, and options.

### Basic Usage

```go
qb := mongo_kit.NewQueryBuilder().
    Equals("status", "active").
    GreaterThan("age", 18).
    Limit(10).
    Sort("name", true)

filter, opts := qb.Build()
var users []User
err := client.Find(ctx, "users", filter, &users, opts)
```

### Comparison Operators

**Equals** - Equality filter (alias for Filter)
```go
qb.Equals("status", "active")
// { status: "active" }
```

**NotEquals** - Not equal
```go
qb.NotEquals("status", "banned")
// { status: { $ne: "banned" } }
```

**GreaterThan** - Greater than
```go
qb.GreaterThan("age", 18)
// { age: { $gt: 18 } }
```

**GreaterThanOrEqual** - Greater than or equal
```go
qb.GreaterThanOrEqual("score", 80)
// { score: { $gte: 80 } }
```

**LessThan** - Less than
```go
qb.LessThan("price", 100)
// { price: { $lt: 100 } }
```

**LessThanOrEqual** - Less than or equal
```go
qb.LessThanOrEqual("stock", 10)
// { stock: { $lte: 10 } }
```

### Array Operators

**In** - Value in array
```go
qb.In("status", "active", "pending", "processing")
// { status: { $in: ["active", "pending", "processing"] } }
```

**NotIn** - Value not in array
```go
qb.NotIn("role", "admin", "superuser")
// { role: { $nin: ["admin", "superuser"] } }
```

### Other Operators

**Exists** - Check if field exists
```go
qb.Exists("email", true)  // has email field
qb.Exists("deleted_at", false)  // doesn't have deleted_at
```

**Regex** - Pattern matching
```go
qb.Regex("email", ".*@example\\.com", "i")
// Case-insensitive email domain match
```

**Where** - Raw MongoDB expression
```go
// $expr condition
qb.Where(bson.M{"$expr": bson.M{"$gt": []any{"$spent", "$budget"}}})

// $text search
qb.Where(bson.M{"$text": bson.M{"$search": "mongodb"}})
```

### Logical Operators

**And** - All conditions must match
```go
qb.And(
    bson.D{{"age", bson.M{"$gte": 18}}},
    bson.D{{"verified", true}},
)
```

**Or** - Any condition can match
```go
qb.Or(
    bson.D{{"status", "premium"}},
    bson.D{{"credits", bson.M{"$gt": 100}}},
)
```

**Nor** - None of the conditions should match
```go
qb.Nor(
    bson.D{{"banned", true}},
    bson.D{{"deleted", true}},
)
```

### Query Composition

**AndConditions** - Combine multiple QueryBuilders with AND logic
```go
ageFilter := mongo_kit.NewQueryBuilder().GreaterThan("age", 18)
statusFilter := mongo_kit.NewQueryBuilder().Equals("status", "active")
emailFilter := mongo_kit.NewQueryBuilder().Exists("email", true)

qb := mongo_kit.NewQueryBuilder().AndConditions(ageFilter, statusFilter, emailFilter)
// All conditions must match
```

**OrConditions** - Combine multiple QueryBuilders with OR logic
```go
underageFilter := mongo_kit.NewQueryBuilder().LessThan("age", 18)
seniorFilter := mongo_kit.NewQueryBuilder().GreaterThan("age", 65)

qb := mongo_kit.NewQueryBuilder().OrConditions(underageFilter, seniorFilter)
// Matches age < 18 OR age > 65
```

**NorConditions** - Combine multiple QueryBuilders with NOR logic
```go
bannedFilter := mongo_kit.NewQueryBuilder().Equals("status", "banned")
deletedFilter := mongo_kit.NewQueryBuilder().Equals("deleted", true)

qb := mongo_kit.NewQueryBuilder().NorConditions(bannedFilter, deletedFilter)
// Matches neither banned nor deleted
```

### Query Options

**Limit** - Maximum documents to return
```go
qb.Limit(10)
```

**Skip** - Number of documents to skip
```go
qb.Skip(20)  // Skip first 20, useful for pagination
```

**Sort** - Sort by field (can be called multiple times)
```go
qb.Sort("name", true)     // Ascending
qb.Sort("age", false)     // Descending
// Multiple sorts: first by name ascending, then by age descending
```

**SortBy** - Custom sort specification
```go
qb.SortBy(bson.D{{"priority", -1}, {"created_at", 1}})
// Replaces any previous sorts
```

**Project** - Select specific fields
```go
qb.Project(bson.M{"name": 1, "email": 1, "_id": 0})
// Only return name and email, exclude _id
```

### Building the Query

**GetFilter** - Get only the filter (no options)
```go
qb := mongo_kit.NewQueryBuilder().Equals("status", "active")
filter := qb.GetFilter()
// Use filter with other operations like CountDocuments
count, err := client.CountDocuments(ctx, "users", filter)
```

**Build** - Get filter and options
```go
filter, opts := qb.Build()
err := client.Find(ctx, "users", filter, &users, opts)
```

### Complete Example

```go
qb := mongo_kit.NewQueryBuilder().
    Equals("type", "customer").
    GreaterThanOrEqual("age", 18).
    In("country", "US", "CA", "UK").
    Exists("email", true).
    Sort("created_at", false).
    Limit(50).
    Skip(100)

filter, opts := qb.Build()
var customers []Customer
err := client.Find(ctx, "users", filter, &customers, opts)
```

## UpdateBuilder

Fluent interface for building MongoDB update operations.

### Basic Usage

```go
ub := mongo_kit.NewUpdateBuilder().
    Set("status", "active").
    Inc("login_count", 1).
    CurrentDate("updated_at")

update := ub.Build()
result, err := client.UpdateOne(ctx, "users", filter, update)
```

### Field Update Operators

**Set** - Set field value
```go
ub.Set("status", "active").
   Set("verified", true)
// { $set: { status: "active", verified: true } }
```

**Unset** - Remove fields
```go
ub.Unset("temp_token", "session_id")
// { $unset: { temp_token: "", session_id: "" } }
```

**Inc** - Increment numeric field
```go
ub.Inc("views", 1).
   Inc("credits", -5)  // Can decrement with negative value
// { $inc: { views: 1, credits: -5 } }
```

**Mul** - Multiply numeric field
```go
ub.Mul("price", 1.1)  // Increase price by 10%
// { $mul: { price: 1.1 } }
```

**Min** - Update only if new value is less
```go
ub.Min("lowest_score", 85)
// Updates only if 85 < current value
```

**Max** - Update only if new value is greater
```go
ub.Max("highest_score", 95)
// Updates only if 95 > current value
```

**Rename** - Rename field
```go
ub.Rename("old_name", "new_name")
// { $rename: { old_name: "new_name" } }
```

**CurrentDate** - Set to current date
```go
ub.CurrentDate("updated_at")
// { $currentDate: { updated_at: true } }
```

### Array Update Operators

**Push** - Append to array
```go
ub.Push("tags", "mongodb")
// { $push: { tags: "mongodb" } }
```

**Pull** - Remove all matching values
```go
ub.Pull("tags", "deprecated")
// { $pull: { tags: "deprecated" } }
```

**AddToSet** - Add to array if not present
```go
ub.AddToSet("interests", "coding")
// { $addToSet: { interests: "coding" } }
```

**Pop** - Remove first or last element
```go
ub.Pop("items", true)   // Remove first element
ub.Pop("items", false)  // Remove last element
```

### Complete Example

```go
ub := mongo_kit.NewUpdateBuilder().
    Set("status", "premium").
    Set("upgraded_at", time.Now()).
    Inc("subscription_months", 12).
    Push("features", "advanced_analytics").
    CurrentDate("updated_at")

update := ub.Build()
result, err := client.UpdateOne(ctx, "users", bson.M{"_id": userID}, update)
```

## AggregationBuilder

Fluent interface for building MongoDB aggregation pipelines.

### Basic Usage

```go
ab := mongo_kit.NewAggregationBuilder().
    Match(bson.M{"status": "active"}).
    Group("$country", bson.M{"count": bson.M{"$sum": 1}}).
    Sort(bson.M{"count": -1}).
    Limit(10)

pipeline := ab.Build()
var results []bson.M
err := client.Aggregate(ctx, "users", pipeline, &results)
```

### Pipeline Stages

**Match** - Filter documents
```go
ab.Match(bson.M{"age": bson.M{"$gte": 18}})
// { $match: { age: { $gte: 18 } } }
```

**Group** - Group and aggregate
```go
ab.Group("$country", bson.M{
    "total": bson.M{"$sum": 1},
    "avgAge": bson.M{"$avg": "$age"},
})
// Groups by country, counts total and calculates average age
```

**Sort** - Sort documents
```go
ab.Sort(bson.M{"count": -1, "name": 1})
// Sort by count descending, then name ascending
```

**Limit** - Limit results
```go
ab.Limit(10)
```

**Skip** - Skip documents
```go
ab.Skip(20)
```

**Project** - Select/reshape fields
```go
ab.Project(bson.M{
    "name": 1,
    "email": 1,
    "fullName": bson.M{"$concat": []any{"$firstName", " ", "$lastName"}},
})
```

**Unwind** - Deconstruct array
```go
ab.Unwind("$tags")
// Creates one document per array element
```

**Lookup** - Join collections
```go
ab.Lookup("orders", "user_id", "_id", "user_orders")
// Left outer join: users -> orders
```

**AddStage** - Custom stage
```go
ab.AddStage(bson.D{{
    Key: "$facet",
    Value: bson.M{
        "byCountry": []bson.M{{"$group": bson.M{"_id": "$country", "count": bson.M{"$sum": 1}}}},
        "byAge": []bson.M{{"$group": bson.M{"_id": "$age", "count": bson.M{"$sum": 1}}}},
    },
}})
```

### Complete Examples

**User Statistics by Country**
```go
ab := mongo_kit.NewAggregationBuilder().
    Match(bson.M{"status": "active"}).
    Group("$country", bson.M{
        "totalUsers": bson.M{"$sum": 1},
        "avgAge": bson.M{"$avg": "$age"},
        "totalRevenue": bson.M{"$sum": "$lifetime_value"},
    }).
    Sort(bson.M{"totalRevenue": -1}).
    Limit(20)

pipeline := ab.Build()
var stats []CountryStats
err := client.Aggregate(ctx, "users", pipeline, &stats)
```

**Users with Order Details**
```go
ab := mongo_kit.NewAggregationBuilder().
    Match(bson.M{"status": "active"}).
    Lookup("orders", "_id", "user_id", "orders").
    Project(bson.M{
        "name": 1,
        "email": 1,
        "orderCount": bson.M{"$size": "$orders"},
        "totalSpent": bson.M{"$sum": "$orders.total"},
    }).
    Sort(bson.M{"totalSpent": -1})

pipeline := ab.Build()
var results []UserOrderSummary
err := client.Aggregate(ctx, "users", pipeline, &results)
```

**Tag Frequency Analysis**
```go
ab := mongo_kit.NewAggregationBuilder().
    Unwind("$tags").
    Group("$tags", bson.M{"count": bson.M{"$sum": 1}}).
    Sort(bson.M{"count": -1}).
    Limit(10)

pipeline := ab.Build()
var tagStats []bson.M
err := client.Aggregate(ctx, "articles", pipeline, &tagStats)
```

## Integration with Repository

All builders work seamlessly with the Repository pattern:

```go
userRepo := mongo_kit.NewRepository[User](client, "users")

// QueryBuilder with Repository
qb := mongo_kit.NewQueryBuilder().
    Equals("status", "active").
    GreaterThan("age", 18)
users, err := userRepo.FindWithBuilder(ctx, qb)

// UpdateBuilder with Repository
ub := mongo_kit.NewUpdateBuilder().
    Set("last_login", time.Now()).
    Inc("login_count", 1)
result, err := userRepo.UpdateOne(ctx, filter, ub.Build())

// AggregationBuilder with Repository
ab := mongo_kit.NewAggregationBuilder().
    Match(bson.M{"status": "active"}).
    Group("$role", bson.M{"count": bson.M{"$sum": 1}})
var stats []RoleStats
err := userRepo.Aggregate(ctx, ab.Build(), &stats)
```

## Best Practices

- **Chain methods** for readability
- **Reuse builders** for common queries
- **Use GetFilter()** when you only need the filter without options
- **Combine builders** with logical operators for complex queries
- **Use Where()** for operators not covered by dedicated methods
- **Build once** - Call Build() only when ready to execute the query
