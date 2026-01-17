# Repository Pattern Guide

The Repository pattern provides type-safe, collection-specific database operations using Go generics.

## Overview

The Repository wraps the MongoDB client to provide strongly-typed methods that:
- Eliminate the need to specify collection names repeatedly
- Remove type assertions when reading documents
- Provide a cleaner, more maintainable API
- Work seamlessly with your domain models

## Creating a Repository

```go
type User struct {
    ID    primitive.ObjectID `bson:"_id,omitempty"`
    Name  string             `bson:"name"`
    Email string             `bson:"email"`
    Age   int                `bson:"age"`
}

client, _ := mongo_kit.New(mongo_kit.DefaultConfig())
userRepo := mongo_kit.NewRepository[User](client, "users")
```

## CRUD Operations

### Create

**Create** - Insert a single document
```go
user := User{Name: "John Doe", Email: "john@example.com", Age: 30}
id, err := userRepo.Create(ctx, user)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created user with ID: %v\n", id)
```

**CreateMany** - Insert multiple documents
```go
users := []User{
    {Name: "Alice", Email: "alice@example.com", Age: 25},
    {Name: "Bob", Email: "bob@example.com", Age: 35},
}
ids, err := userRepo.CreateMany(ctx, users)
fmt.Printf("Created %d users\n", len(ids))
```

### Find

**FindOne** - Find a single document
```go
var user User
err := userRepo.FindOne(ctx, bson.M{"email": "john@example.com"}, &user)
if err == mongo.ErrNoDocuments {
    fmt.Println("User not found")
}
```

**Find** - Find multiple documents
```go
var users []User
filter := bson.M{"age": bson.M{"$gte": 18}}
err := userRepo.Find(ctx, filter, &users)
fmt.Printf("Found %d users\n", len(users))
```

**FindAll** - Find all documents in collection
```go
var users []User
err := userRepo.FindAll(ctx, &users)
// No filter needed - returns all documents
```

**FindByID** - Find by ObjectID or string
```go
var user User
err := userRepo.FindByID(ctx, "507f1f77bcf86cd799439011", &user)
// Or with ObjectID
err := userRepo.FindByID(ctx, objectID, &user)
```

### Update

**UpdateOne** - Update a single document
```go
filter := bson.M{"email": "john@example.com"}
update := bson.M{"$set": bson.M{"age": 31}}
result, err := userRepo.UpdateOne(ctx, filter, update)
fmt.Printf("Modified: %d\n", result.ModifiedCount)
```

**UpdateMany** - Update multiple documents
```go
filter := bson.M{"age": bson.M{"$lt": 18}}
update := bson.M{"$set": bson.M{"status": "minor"}}
result, err := userRepo.UpdateMany(ctx, filter, update)
```

**UpdateByID** - Update by ObjectID or string
```go
update := bson.M{"$set": bson.M{"last_login": time.Now()}}
result, err := userRepo.UpdateByID(ctx, userID, update)
```

**Upsert** - Update or insert if not exists
```go
filter := bson.M{"email": "new@example.com"}
update := bson.M{"$set": bson.M{"name": "New User", "email": "new@example.com"}}
result, err := userRepo.Upsert(ctx, filter, update)
if result.UpsertedCount > 0 {
    fmt.Println("Document created:", result.UpsertedID)
}
```

### Delete

**DeleteOne** - Delete a single document
```go
filter := bson.M{"email": "old@example.com"}
result, err := userRepo.DeleteOne(ctx, filter)
fmt.Printf("Deleted: %d\n", result.DeletedCount)
```

**DeleteMany** - Delete multiple documents
```go
filter := bson.M{"status": "inactive"}
result, err := userRepo.DeleteMany(ctx, filter)
```

**DeleteByID** - Delete by ObjectID or string
```go
result, err := userRepo.DeleteByID(ctx, "507f1f77bcf86cd799439011")
```

## Query Operations

### Count

**Count** - Count matching documents (accurate)
```go
count, err := userRepo.Count(ctx, bson.M{"active": true})
fmt.Printf("Active users: %d\n", count)
```

**CountAll** - Count all documents
```go
count, err := userRepo.CountAll(ctx)
fmt.Printf("Total users: %d\n", count)
```

**EstimatedCount** - Fast approximate count
```go
count, err := userRepo.EstimatedCount(ctx)
// Uses collection metadata, no filters
```

### Exists

**Exists** - Check if matching document exists
```go
exists, err := userRepo.Exists(ctx, bson.M{"email": "john@example.com"})
if exists {
    fmt.Println("Email already registered")
}
```

**ExistsByID** - Check if document with ID exists
```go
exists, err := userRepo.ExistsByID(ctx, userID)
```

## QueryBuilder Integration

The Repository works seamlessly with QueryBuilder for complex queries.

### FindWithBuilder

Build complex queries with fluent interface:
```go
qb := mongo_kit.NewQueryBuilder().
    Equals("status", "active").
    GreaterThan("age", 18).
    Sort("name", true).
    Limit(10)

var users []User
err := userRepo.FindWithBuilder(ctx, qb, &users)
```

### FindOneWithBuilder

Find single document with query builder:
```go
qb := mongo_kit.NewQueryBuilder().
    Equals("email", "john@example.com").
    Project(bson.M{"password": 0})  // Exclude password

var user User
err := userRepo.FindOneWithBuilder(ctx, qb, &user)
```

### CountWithBuilder

Count with complex filters:
```go
qb := mongo_kit.NewQueryBuilder().
    Equals("status", "active").
    GreaterThanOrEqual("age", 18).
    In("country", "US", "CA", "UK")

count, err := userRepo.CountWithBuilder(ctx, qb)
```

### ExistsWithBuilder

Check existence with complex conditions:
```go
qb := mongo_kit.NewQueryBuilder().
    Equals("email", "test@example.com").
    Equals("verified", true)

exists, err := userRepo.ExistsWithBuilder(ctx, qb)
```

## Aggregation

Run aggregation pipelines with type-safe results:

```go
pipeline := []bson.M{
    {"$match": bson.M{"status": "active"}},
    {"$group": bson.M{
        "_id": "$country",
        "count": bson.M{"$sum": 1},
        "avgAge": bson.M{"$avg": "$age"},
    }},
    {"$sort": bson.M{"count": -1}},
}

type CountryStats struct {
    Country string  `bson:"_id"`
    Count   int     `bson:"count"`
    AvgAge  float64 `bson:"avgAge"`
}

var stats []CountryStats
err := userRepo.Aggregate(ctx, pipeline, &stats)
```

Using AggregationBuilder:
```go
ab := mongo_kit.NewAggregationBuilder().
    Match(bson.M{"status": "active"}).
    Group("$role", bson.M{"count": bson.M{"$sum": 1}}).
    Sort(bson.M{"count": -1})

type RoleStats struct {
    Role  string `bson:"_id"`
    Count int    `bson:"count"`
}

var stats []RoleStats
err := userRepo.Aggregate(ctx, ab.Build(), &stats)
```

## Utility Methods

### Collection

Get the underlying mongo.Collection:
```go
coll := userRepo.Collection()
// Access native MongoDB driver methods
cursor, err := coll.Watch(ctx, pipeline)
```

### Client

Get the underlying Client:
```go
client := userRepo.Client()
// Access client methods
err := client.Ping(ctx)
```

### Drop

Drop the entire collection:
```go
err := userRepo.Drop(ctx)
// WARNING: Permanently deletes all documents
```

## Complete Examples

### User Service

```go
type UserService struct {
    repo *mongo_kit.Repository[User]
}

func NewUserService(client *mongo_kit.Client) *UserService {
    return &UserService{
        repo: mongo_kit.NewRepository[User](client, "users"),
    }
}

func (s *UserService) Register(ctx context.Context, email, name string) (*User, error) {
    // Check if email exists
    exists, err := s.repo.Exists(ctx, bson.M{"email": email})
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, errors.New("email already registered")
    }

    // Create user
    user := User{
        Name:  name,
        Email: email,
        CreatedAt: time.Now(),
    }
    id, err := s.repo.Create(ctx, user)
    if err != nil {
        return nil, err
    }

    // Fetch created user
    var created User
    err = s.repo.FindByID(ctx, id, &created)
    return &created, err
}

func (s *UserService) GetActiveUsers(ctx context.Context, limit int) ([]User, error) {
    qb := mongo_kit.NewQueryBuilder().
        Equals("status", "active").
        Sort("created_at", false).
        Limit(int64(limit))

    var users []User
    err := s.repo.FindWithBuilder(ctx, qb, &users)
    return users, err
}

func (s *UserService) UpdateLastLogin(ctx context.Context, userID any) error {
    update := bson.M{"$set": bson.M{"last_login": time.Now()}}
    _, err := s.repo.UpdateByID(ctx, userID, update)
    return err
}
```

### CRUD with Error Handling

```go
// Create with validation
func CreateUser(ctx context.Context, repo *mongo_kit.Repository[User], user User) error {
    if user.Email == "" {
        return errors.New("email required")
    }

    exists, err := repo.Exists(ctx, bson.M{"email": user.Email})
    if err != nil {
        return fmt.Errorf("check email: %w", err)
    }
    if exists {
        return errors.New("email already exists")
    }

    _, err = repo.Create(ctx, user)
    return err
}

// Read with not found handling
func GetUser(ctx context.Context, repo *mongo_kit.Repository[User], id any) (*User, error) {
    var user User
    err := repo.FindByID(ctx, id, &user)
    if err == mongo.ErrNoDocuments {
        return nil, fmt.Errorf("user not found: %v", id)
    }
    return &user, err
}

// Update with result check
func ActivateUser(ctx context.Context, repo *mongo_kit.Repository[User], id any) error {
    update := bson.M{"$set": bson.M{"status": "active", "activated_at": time.Now()}}
    result, err := repo.UpdateByID(ctx, id, update)
    if err != nil {
        return err
    }
    if result.MatchedCount == 0 {
        return errors.New("user not found")
    }
    return nil
}

// Delete with confirmation
func DeleteInactiveUsers(ctx context.Context, repo *mongo_kit.Repository[User]) (int64, error) {
    cutoff := time.Now().Add(-90 * 24 * time.Hour)
    filter := bson.M{
        "status": "inactive",
        "last_login": bson.M{"$lt": cutoff},
    }

    result, err := repo.DeleteMany(ctx, filter)
    if err != nil {
        return 0, err
    }
    return result.DeletedCount, nil
}
```

### Pagination

```go
func GetUsersPage(ctx context.Context, repo *mongo_kit.Repository[User], page, pageSize int) ([]User, error) {
    qb := mongo_kit.NewQueryBuilder().
        Equals("status", "active").
        Sort("created_at", false).
        Skip(int64((page - 1) * pageSize)).
        Limit(int64(pageSize))

    var users []User
    err := repo.FindWithBuilder(ctx, qb, &users)
    return users, err
}
```

### Batch Operations

```go
func ImportUsers(ctx context.Context, repo *mongo_kit.Repository[User], users []User) error {
    // Batch insert
    ids, err := repo.CreateMany(ctx, users)
    if err != nil {
        return fmt.Errorf("import failed: %w", err)
    }

    fmt.Printf("Imported %d users\n", len(ids))
    return nil
}

func BulkUpdateStatus(ctx context.Context, repo *mongo_kit.Repository[User], userIDs []string, status string) error {
    filter := bson.M{"_id": bson.M{"$in": userIDs}}
    update := bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}}

    result, err := repo.UpdateMany(ctx, filter, update)
    if err != nil {
        return err
    }

    fmt.Printf("Updated %d users\n", result.ModifiedCount)
    return nil
}
```

## Multiple Repositories

Organize your data layer with multiple repositories:

```go
type DataLayer struct {
    Users    *mongo_kit.Repository[User]
    Orders   *mongo_kit.Repository[Order]
    Products *mongo_kit.Repository[Product]
}

func NewDataLayer(client *mongo_kit.Client) *DataLayer {
    return &DataLayer{
        Users:    mongo_kit.NewRepository[User](client, "users"),
        Orders:   mongo_kit.NewRepository[Order](client, "orders"),
        Products: mongo_kit.NewRepository[Product](client, "products"),
    }
}

// Use in application
func main() {
    client, _ := mongo_kit.New(mongo_kit.DefaultConfig())
    defer client.Close(context.Background())

    data := NewDataLayer(client)

    // Use repositories
    user, _ := data.Users.FindByID(ctx, userID, &user)
    orders, _ := data.Orders.Find(ctx, bson.M{"user_id": userID}, &orders)
    products, _ := data.Products.FindAll(ctx, &products)
}
```

## Best Practices

- **Use generics** for type safety and cleaner code
- **One repository per collection** for clear separation
- **Validate before create** to prevent duplicate data
- **Use QueryBuilder** for complex queries instead of raw bson.M
- **Handle ErrNoDocuments** explicitly when document might not exist
- **Use CountAll over Count** when no filter is needed (more efficient)
- **Use EstimatedCount** for large collections when exact count isn't critical
- **Check ModifiedCount** after updates to verify changes were made
- **Use transactions** via client.WithTransaction() for multi-collection operations
- **Reuse repositories** across your application layer

## Type Safety Benefits

```go
// Without Repository (type assertions needed)
var user User
err := client.FindOne(ctx, "users", filter, &user)

// With Repository (type-safe, no assertions)
var user User
err := userRepo.FindOne(ctx, filter, &user)

// Collection name is implicit, reducing errors
// No need to remember or type "users" every time
```
