# Operations Guide

This document provides a concise overview of all database operations available in the mongo_kit client.

## CRUD Operations

### Insert

**InsertOne** - Insert a single document
```go
result, err := client.InsertOne(ctx, "users", User{Name: "John", Email: "john@example.com"})
fmt.Println("Inserted ID:", result.InsertedID)
```

**InsertMany** - Insert multiple documents
```go
docs := []any{
    User{Name: "Alice", Email: "alice@example.com"},
    User{Name: "Bob", Email: "bob@example.com"},
}
result, err := client.InsertMany(ctx, "users", docs)
```

### Find

**FindOne** - Find a single document
```go
var user User
err := client.FindOne(ctx, "users", bson.M{"email": "john@example.com"}, &user)
```

**Find** - Find multiple documents
```go
var users []User
err := client.Find(ctx, "users", bson.M{"active": true}, &users)
```

**FindByID** - Find by ObjectID or string ID
```go
var user User
err := client.FindByID(ctx, "users", "507f1f77bcf86cd799439011", &user)
// Also accepts primitive.ObjectID
```

### Update

**UpdateOne** - Update a single document
```go
update := bson.M{"$set": bson.M{"status": "active"}}
result, err := client.UpdateOne(ctx, "users", bson.M{"email": email}, update)
fmt.Println("Modified:", result.ModifiedCount)
```

**UpdateMany** - Update multiple documents
```go
update := bson.M{"$set": bson.M{"verified": true}}
result, err := client.UpdateMany(ctx, "users", bson.M{"age": bson.M{"$gte": 18}}, update)
```

**UpdateByID** - Update by ObjectID or string ID
```go
update := bson.M{"$set": bson.M{"last_login": time.Now()}}
result, err := client.UpdateByID(ctx, "users", userID, update)
```

**ReplaceOne** - Replace entire document (except _id)
```go
newUser := User{Name: "John Doe", Email: "john.doe@example.com"}
result, err := client.ReplaceOne(ctx, "users", bson.M{"email": "john@example.com"}, newUser)
```

**UpsertOne** - Update or insert if not exists
```go
update := bson.M{"$set": bson.M{"name": "John", "email": "john@example.com"}}
result, err := client.UpsertOne(ctx, "users", bson.M{"email": "john@example.com"}, update)
if result.UpsertedCount > 0 {
    fmt.Println("Document created:", result.UpsertedID)
}
```

### Delete

**DeleteOne** - Delete a single document
```go
result, err := client.DeleteOne(ctx, "users", bson.M{"email": "john@example.com"})
fmt.Println("Deleted:", result.DeletedCount)
```

**DeleteMany** - Delete multiple documents
```go
result, err := client.DeleteMany(ctx, "users", bson.M{"active": false})
```

**DeleteByID** - Delete by ObjectID or string ID
```go
result, err := client.DeleteByID(ctx, "users", "507f1f77bcf86cd799439011")
```

## Find and Modify

**FindOneAndUpdate** - Find, update, and return the document
```go
var user User
update := bson.M{"$inc": bson.M{"login_count": 1}}
opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
err := client.FindOneAndUpdate(ctx, "users", bson.M{"email": email}, update, &user, opts)
```

**FindOneAndReplace** - Find, replace, and return the document
```go
var user User
newUser := User{Name: "John Updated", Email: "john@example.com"}
err := client.FindOneAndReplace(ctx, "users", bson.M{"email": email}, newUser, &user)
```

**FindOneAndDelete** - Find, delete, and return the document
```go
var user User
err := client.FindOneAndDelete(ctx, "users", bson.M{"email": email}, &user)
```

## Query Operations

**CountDocuments** - Count matching documents (accurate)
```go
count, err := client.CountDocuments(ctx, "users", bson.M{"active": true})
```

**EstimatedDocumentCount** - Fast approximate count (uses metadata)
```go
count, err := client.EstimatedDocumentCount(ctx, "users")
```

**Distinct** - Get distinct values for a field
```go
values, err := client.Distinct(ctx, "users", "country", bson.M{"active": true})
for _, v := range values {
    fmt.Println(v)
}
```

**Aggregate** - Run aggregation pipeline
```go
pipeline := []bson.M{
    {"$match": bson.M{"active": true}},
    {"$group": bson.M{"_id": "$country", "count": bson.M{"$sum": 1}}},
}
var results []bson.M
err := client.Aggregate(ctx, "users", pipeline, &results)
```

## Bulk Operations

**BulkWrite** - Execute multiple write operations in one call
```go
models := []mongo.WriteModel{
    mongo.NewInsertOneModel().SetDocument(User{Name: "Alice"}),
    mongo.NewUpdateOneModel().SetFilter(bson.M{"name": "Bob"}).SetUpdate(bson.M{"$set": bson.M{"active": true}}),
    mongo.NewDeleteOneModel().SetFilter(bson.M{"name": "Charlie"}),
}
result, err := client.BulkWrite(ctx, "users", models)
fmt.Printf("Inserted: %d, Modified: %d, Deleted: %d\n",
    result.InsertedCount, result.ModifiedCount, result.DeletedCount)
```

## Index Operations

**CreateIndex** - Create a single index
```go
keys := bson.D{{"email", 1}}
opts := options.Index().SetUnique(true)
indexName, err := client.CreateIndex(ctx, "users", keys, opts)
```

**CreateIndexes** - Create multiple indexes across collections
```go
indexes := map[string][]mongo.IndexModel{
    "users": {
        {Keys: bson.D{{"email", 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{"created_at", -1}}},
    },
    "orders": {
        {Keys: bson.D{{"user_id", 1}, {"status", 1}}},
    },
}
results, err := client.CreateIndexes(ctx, indexes)
```

**DropIndex** - Drop an index
```go
err := client.DropIndex(ctx, "users", "email_1")
```

**ListIndexes** - List all indexes in a collection
```go
indexes, err := client.ListIndexes(ctx, "users")
for _, idx := range indexes {
    fmt.Println(idx["name"])
}
```

## Collection Management

**CreateCollection** - Create a collection with options
```go
opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(1024*1024)
err := client.CreateCollection(ctx, "logs", opts)
```

**CreateCollections** - Create multiple collections at once
```go
collections := map[string]*options.CreateCollectionOptions{
    "logs": options.CreateCollection().SetCapped(true).SetSizeInBytes(1024*1024),
    "cache": options.CreateCollection().SetExpireAfterSeconds(3600),
}
err := client.CreateCollections(ctx, collections)
```

**ListCollections** - List all collections in the database
```go
collections, err := client.ListCollections(ctx)
for _, name := range collections {
    fmt.Println(name)
}
```

**DropCollection** - Delete a collection
```go
err := client.DropCollection(ctx, "temp_data")
```

**DropDatabase** - Delete an entire database
```go
err := client.DropDatabase(ctx, "test_database")
```

## Transactions

**WithTransaction** - Execute operations in a transaction
```go
err := client.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
    // All operations here are atomic
    _, err := client.InsertOne(sessCtx, "orders", order)
    if err != nil {
        return err // Transaction will rollback
    }

    update := bson.M{"$inc": bson.M{"stock": -1}}
    _, err = client.UpdateOne(sessCtx, "products", bson.M{"_id": productID}, update)
    if err != nil {
        return err // Transaction will rollback
    }

    return nil // Transaction will commit
})
```

**Important:** Transactions require MongoDB 4.0+ and a replica set or sharded cluster.

## Change Streams

**Watch** - Monitor real-time changes to a collection
```go
pipeline := []bson.M{{"$match": bson.M{"operationType": "insert"}}}
stream, err := client.Watch(ctx, "users", pipeline)
if err != nil {
    log.Fatal(err)
}
defer stream.Close(ctx)

for stream.Next(ctx) {
    var event bson.M
    if err := stream.Decode(&event); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Change detected:", event)
}
```

**Important:** Change streams require MongoDB 3.6+ and a replica set or sharded cluster.

## ID Handling

Methods that accept `id` parameter (FindByID, UpdateByID, DeleteByID) support both:
- **string**: Automatically converted to ObjectID
- **primitive.ObjectID**: Used directly

```go
// Both work
client.FindByID(ctx, "users", "507f1f77bcf86cd799439011", &user)
client.FindByID(ctx, "users", objectID, &user)
```

## Error Handling

All operations return typed errors that can be checked:
```go
err := client.FindOne(ctx, "users", bson.M{"email": email}, &user)
if err != nil {
    if errors.Is(err, mongo.ErrNoDocuments) {
        // Document not found
    } else if mongo.IsDuplicateKeyError(err) {
        // Unique constraint violation
    } else {
        // Other error
    }
}
```
