package mongo

import (
	"context"

	"github.com/edaniel30/mongo-kit-go/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertOne inserts a single document into the specified collection.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection where the document will be inserted
//   - document: The document to insert. Can be a struct, bson.M, bson.D, or any serializable type
//
// Returns:
//   - *mongo.InsertOneResult: Contains the InsertedID field with the _id of the inserted document
//   - error: Returns error if insertion fails or client is closed
//
// Example:
//
//	type User struct {
//	    Name  string `bson:"name"`
//	    Email string `bson:"email"`
//	}
//	result, err := client.InsertOne(ctx, "users", User{Name: "John", Email: "john@example.com"})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Inserted ID:", result.InsertedID)
func (c *Client) InsertOne(ctx context.Context, collection string, document any) (*mongo.InsertOneResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	result, err := coll.InsertOne(ctx, document)
	if err != nil {
		return nil, errors.NewOperationError("insert one", err)
	}

	return result, nil
}

// InsertMany inserts multiple documents into the specified collection in a single operation.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection where documents will be inserted
//   - documents: Slice of documents to insert. Each element can be a struct, bson.M, bson.D, or any serializable type
//
// Returns:
//   - *mongo.InsertManyResult: Contains InsertedIDs field with a map of insertion order to _id values
//   - error: Returns error if insertion fails or client is closed
//
// Example:
//
//	users := []any{
//	    User{Name: "John", Email: "john@example.com"},
//	    User{Name: "Jane", Email: "jane@example.com"},
//	}
//	result, err := client.InsertMany(ctx, "users", users)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Inserted %d documents\n", len(result.InsertedIDs))
func (c *Client) InsertMany(ctx context.Context, collection string, documents []any) (*mongo.InsertManyResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	result, err := coll.InsertMany(ctx, documents)
	if err != nil {
		return nil, errors.NewOperationError("insert many", err)
	}

	return result, nil
}

// FindOne finds a single document matching the filter and decodes it into result.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to search in
//   - filter: Query filter to match documents. Can be bson.M, bson.D, or a struct. Use bson.M{} for all documents
//   - result: Pointer to a variable where the found document will be decoded
//   - opts: Optional FindOneOptions for sorting, projection, skip, etc.
//
// Returns:
//   - error: Returns ErrDocumentNotFound if no document matches, ErrClientClosed if client is closed, or operation error
//
// Example:
//
//	var user User
//	err := client.FindOne(ctx, "users", bson.M{"email": "john@example.com"}, &user)
//	if err != nil {
//	    if errors.Is(err, mongo.ErrDocumentNotFound) {
//	        fmt.Println("User not found")
//	    }
//	    return err
//	}
//	fmt.Printf("Found user: %s\n", user.Name)
func (c *Client) FindOne(ctx context.Context, collection string, filter any, result any, opts ...*options.FindOneOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	coll := c.GetCollection(collection)
	err := coll.FindOne(ctx, filter, opts...).Decode(result)
	if err != nil {
		// Return ErrNoDocuments directly for clearer error handling
		if err == mongo.ErrNoDocuments {
			return err
		}
		return errors.NewOperationError("find one", err)
	}

	return nil
}

// Find finds all documents matching the filter and decodes them into results.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to search in
//   - filter: Query filter to match documents. Can be bson.M, bson.D, or a struct. Use bson.M{} for all documents
//   - results: Pointer to a slice where found documents will be decoded (e.g., *[]User)
//   - opts: Optional FindOptions for sorting, projection, limit, skip, etc.
//
// Returns:
//   - error: Returns error if operation fails or client is closed. Empty results is not an error
//
// Example:
//
//	var users []User
//	filter := bson.M{"age": bson.M{"$gte": 18}}
//	opts := options.Find().SetLimit(10).SetSort(bson.M{"name": 1})
//	err := client.Find(ctx, "users", filter, &users, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Found %d users\n", len(users))
func (c *Client) Find(ctx context.Context, collection string, filter any, results any, opts ...*options.FindOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	coll := c.GetCollection(collection)
	cursor, err := coll.Find(ctx, filter, opts...)
	if err != nil {
		return errors.NewOperationError("find", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, results); err != nil {
		return errors.NewOperationError("find decode", err)
	}

	return nil
}

// UpdateOne updates a single document matching the filter.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to update in
//   - filter: Query filter to match the document. Can be bson.M, bson.D, or a struct
//   - update: Update operations to apply. Must use update operators like $set, $inc, etc. (e.g., bson.M{"$set": bson.M{"age": 30}})
//   - opts: Optional UpdateOptions for upsert, bypass validation, etc.
//
// Returns:
//   - *mongo.UpdateResult: Contains MatchedCount, ModifiedCount, UpsertedCount, and UpsertedID
//   - error: Returns error if operation fails or client is closed
//
// Example:
//
//	filter := bson.M{"email": "john@example.com"}
//	update := bson.M{"$set": bson.M{"age": 31}, "$inc": bson.M{"login_count": 1}}
//	result, err := client.UpdateOne(ctx, "users", filter, update)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Modified %d document(s)\n", result.ModifiedCount)
func (c *Client) UpdateOne(ctx context.Context, collection string, filter any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	result, err := coll.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		return nil, errors.NewOperationError("update one", err)
	}

	return result, nil
}

// UpdateMany updates all documents matching the filter.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to update in
//   - filter: Query filter to match documents. Can be bson.M, bson.D, or a struct. Use bson.M{} to update all documents
//   - update: Update operations to apply. Must use update operators like $set, $inc, etc. (e.g., bson.M{"$set": bson.M{"status": "active"}})
//   - opts: Optional UpdateOptions for upsert, bypass validation, etc.
//
// Returns:
//   - *mongo.UpdateResult: Contains MatchedCount, ModifiedCount, UpsertedCount, and UpsertedID
//   - error: Returns error if operation fails or client is closed
//
// Example:
//
//	filter := bson.M{"status": "pending"}
//	update := bson.M{"$set": bson.M{"status": "processed", "processed_at": time.Now()}}
//	result, err := client.UpdateMany(ctx, "orders", filter, update)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Modified %d document(s)\n", result.ModifiedCount)
func (c *Client) UpdateMany(ctx context.Context, collection string, filter any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	result, err := coll.UpdateMany(ctx, filter, update, opts...)
	if err != nil {
		return nil, errors.NewOperationError("update many", err)
	}

	return result, nil
}

// ReplaceOne replaces a single document matching the filter with a new document.
// Unlike UpdateOne, this replaces the entire document (except _id) rather than modifying specific fields.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to replace in
//   - filter: Query filter to match the document. Can be bson.M, bson.D, or a struct
//   - replacement: The new document to replace with. Cannot contain update operators like $set. The _id field will be preserved
//   - opts: Optional ReplaceOptions for upsert, bypass validation, etc.
//
// Returns:
//   - *mongo.UpdateResult: Contains MatchedCount, ModifiedCount, UpsertedCount, and UpsertedID
//   - error: Returns error if operation fails or client is closed
//
// Example:
//
//	filter := bson.M{"_id": userID}
//	newUser := User{Name: "John Doe", Email: "john@example.com", Age: 30}
//	result, err := client.ReplaceOne(ctx, "users", filter, newUser)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Replaced %d document(s)\n", result.ModifiedCount)
func (c *Client) ReplaceOne(ctx context.Context, collection string, filter any, replacement any, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	result, err := coll.ReplaceOne(ctx, filter, replacement, opts...)
	if err != nil {
		return nil, errors.NewOperationError("replace one", err)
	}

	return result, nil
}

// DeleteOne deletes a single document matching the filter.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to delete from
//   - filter: Query filter to match the document to delete. Can be bson.M, bson.D, or a struct
//   - opts: Optional DeleteOptions for collation, hint, etc.
//
// Returns:
//   - *mongo.DeleteResult: Contains DeletedCount with the number of documents deleted (0 or 1)
//   - error: Returns error if operation fails or client is closed
//
// Example:
//
//	filter := bson.M{"email": "user@example.com"}
//	result, err := client.DeleteOne(ctx, "users", filter)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if result.DeletedCount == 0 {
//	    fmt.Println("No document was deleted")
//	}
func (c *Client) DeleteOne(ctx context.Context, collection string, filter any, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	result, err := coll.DeleteOne(ctx, filter, opts...)
	if err != nil {
		return nil, errors.NewOperationError("delete one", err)
	}

	return result, nil
}

// DeleteMany deletes all documents matching the filter.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to delete from
//   - filter: Query filter to match documents to delete. Can be bson.M, bson.D, or a struct. Use bson.M{} to delete all documents
//   - opts: Optional DeleteOptions for collation, hint, etc.
//
// Returns:
//   - *mongo.DeleteResult: Contains DeletedCount with the number of documents deleted
//   - error: Returns error if operation fails or client is closed
//
// Example:
//
//	filter := bson.M{"status": "inactive", "last_login": bson.M{"$lt": time.Now().AddDate(0, -6, 0)}}
//	result, err := client.DeleteMany(ctx, "users", filter)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Deleted %d inactive user(s)\n", result.DeletedCount)
func (c *Client) DeleteMany(ctx context.Context, collection string, filter any, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	result, err := coll.DeleteMany(ctx, filter, opts...)
	if err != nil {
		return nil, errors.NewOperationError("delete many", err)
	}

	return result, nil
}

// CountDocuments counts the number of documents matching the filter.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to count in
//   - filter: Query filter to match documents. Can be bson.M, bson.D, or a struct. Use bson.M{} to count all documents
//   - opts: Optional CountOptions for limit, skip, collation, etc.
//
// Returns:
//   - int64: The number of documents matching the filter
//   - error: Returns error if operation fails or client is closed
//
// Example:
//
//	filter := bson.M{"status": "active"}
//	count, err := client.CountDocuments(ctx, "users", filter)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Found %d active users\n", count)
func (c *Client) CountDocuments(ctx context.Context, collection string, filter any, opts ...*options.CountOptions) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return 0, err
	}

	coll := c.GetCollection(collection)
	count, err := coll.CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, errors.NewOperationError("count documents", err)
	}

	return count, nil
}

// Aggregate executes an aggregation pipeline and decodes the results.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to aggregate on
//   - pipeline: Aggregation pipeline stages as a slice (e.g., []bson.M, mongo.Pipeline)
//   - results: Pointer to a slice where aggregation results will be decoded (e.g., *[]bson.M, *[]User)
//   - opts: Optional AggregateOptions for batch size, collation, etc.
//
// Returns:
//   - error: Returns error if operation fails or client is closed
//
// Example:
//
//	pipeline := []bson.M{
//	    {"$match": bson.M{"status": "active"}},
//	    {"$group": bson.M{"_id": "$country", "count": bson.M{"$sum": 1}}},
//	    {"$sort": bson.M{"count": -1}},
//	}
//	var results []bson.M
//	err := client.Aggregate(ctx, "users", pipeline, &results)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, result := range results {
//	    fmt.Printf("Country: %s, Count: %d\n", result["_id"], result["count"])
//	}
func (c *Client) Aggregate(ctx context.Context, collection string, pipeline any, results any, opts ...*options.AggregateOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	coll := c.GetCollection(collection)
	cursor, err := coll.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return errors.NewOperationError("aggregate", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, results); err != nil {
		return errors.NewOperationError("aggregate decode", err)
	}

	return nil
}

// FindOneAndUpdate finds a single document, updates it, and returns either the original or updated document.
// This is an atomic operation that prevents race conditions.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to search and update in
//   - filter: Query filter to match the document. Can be bson.M, bson.D, or a struct
//   - update: Update operations to apply. Must use update operators like $set, $inc, etc.
//   - result: Pointer to a variable where the document will be decoded (original or updated, depending on options)
//   - opts: Optional FindOneAndUpdateOptions. Use SetReturnDocument(options.After) to return the updated document
//
// Returns:
//   - error: Returns ErrDocumentNotFound if no document matches, or operation error
//
// Example:
//
//	var user User
//	filter := bson.M{"email": "john@example.com"}
//	update := bson.M{"$inc": bson.M{"login_count": 1}, "$set": bson.M{"last_login": time.Now()}}
//	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
//	err := client.FindOneAndUpdate(ctx, "users", filter, update, &user, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("User %s logged in %d times\n", user.Name, user.LoginCount)
func (c *Client) FindOneAndUpdate(ctx context.Context, collection string, filter any, update any, result any, opts ...*options.FindOneAndUpdateOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	coll := c.GetCollection(collection)
	err := coll.FindOneAndUpdate(ctx, filter, update, opts...).Decode(result)
	if err != nil {
		// Return ErrNoDocuments directly for clearer error handling
		if err == mongo.ErrNoDocuments {
			return err
		}
		return errors.NewOperationError("find one and update", err)
	}

	return nil
}

// FindOneAndReplace finds a single document, replaces it, and returns either the original or replaced document.
// This is an atomic operation that prevents race conditions.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to search and replace in
//   - filter: Query filter to match the document. Can be bson.M, bson.D, or a struct
//   - replacement: The new document to replace with. Cannot contain update operators. The _id field will be preserved
//   - result: Pointer to a variable where the document will be decoded (original or replaced, depending on options)
//   - opts: Optional FindOneAndReplaceOptions. Use SetReturnDocument(options.After) to return the replaced document
//
// Returns:
//   - error: Returns ErrDocumentNotFound if no document matches, or operation error
//
// Example:
//
//	var oldUser User
//	filter := bson.M{"_id": userID}
//	newUser := User{Name: "Updated Name", Email: "new@example.com", Status: "active"}
//	opts := options.FindOneAndReplace().SetReturnDocument(options.Before)
//	err := client.FindOneAndReplace(ctx, "users", filter, newUser, &oldUser, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Old user name: %s\n", oldUser.Name)
func (c *Client) FindOneAndReplace(ctx context.Context, collection string, filter any, replacement any, result any, opts ...*options.FindOneAndReplaceOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	coll := c.GetCollection(collection)
	err := coll.FindOneAndReplace(ctx, filter, replacement, opts...).Decode(result)
	if err != nil {
		// Return ErrNoDocuments directly for clearer error handling
		if err == mongo.ErrNoDocuments {
			return err
		}
		return errors.NewOperationError("find one and replace", err)
	}

	return nil
}

// FindOneAndDelete finds a single document, deletes it, and returns the deleted document.
// This is an atomic operation that prevents race conditions.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to search and delete from
//   - filter: Query filter to match the document. Can be bson.M, bson.D, or a struct
//   - result: Pointer to a variable where the deleted document will be decoded
//   - opts: Optional FindOneAndDeleteOptions for sorting, projection, collation, etc.
//
// Returns:
//   - error: Returns ErrDocumentNotFound if no document matches, or operation error
//
// Example:
//
//	var deletedUser User
//	filter := bson.M{"email": "tobedeleted@example.com"}
//	err := client.FindOneAndDelete(ctx, "users", filter, &deletedUser)
//	if err != nil {
//	    if errors.Is(err, mongo.ErrDocumentNotFound) {
//	        fmt.Println("User not found")
//	    }
//	    return err
//	}
//	fmt.Printf("Deleted user: %s\n", deletedUser.Name)
func (c *Client) FindOneAndDelete(ctx context.Context, collection string, filter any, result any, opts ...*options.FindOneAndDeleteOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	coll := c.GetCollection(collection)
	err := coll.FindOneAndDelete(ctx, filter, opts...).Decode(result)
	if err != nil {
		// Return ErrNoDocuments directly for clearer error handling
		if err == mongo.ErrNoDocuments {
			return err
		}
		return errors.NewOperationError("find one and delete", err)
	}

	return nil
}

// BulkWrite executes multiple write operations (insert, update, delete) in a single batch.
// This is more efficient than executing operations individually when performing multiple writes.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to write to
//   - models: Slice of write models (mongo.InsertOneModel, mongo.UpdateOneModel, mongo.DeleteOneModel, etc.)
//   - opts: Optional BulkWriteOptions for ordered execution, bypass validation, etc.
//
// Returns:
//   - *mongo.BulkWriteResult: Contains InsertedCount, MatchedCount, ModifiedCount, DeletedCount, UpsertedCount, and UpsertedIDs
//   - error: Returns error if operation fails or client is closed
//
// Example:
//
//	models := []mongo.WriteModel{
//	    mongo.NewInsertOneModel().SetDocument(User{Name: "Alice", Email: "alice@example.com"}),
//	    mongo.NewUpdateOneModel().SetFilter(bson.M{"name": "Bob"}).SetUpdate(bson.M{"$set": bson.M{"status": "active"}}),
//	    mongo.NewDeleteOneModel().SetFilter(bson.M{"name": "Charlie"}),
//	}
//	result, err := client.BulkWrite(ctx, "users", models)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Inserted: %d, Modified: %d, Deleted: %d\n", result.InsertedCount, result.ModifiedCount, result.DeletedCount)
func (c *Client) BulkWrite(ctx context.Context, collection string, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	result, err := coll.BulkWrite(ctx, models, opts...)
	if err != nil {
		return nil, errors.NewOperationError("bulk write", err)
	}

	return result, nil
}

// Distinct gets all unique values for a specified field across documents matching the filter.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to query
//   - fieldName: The field name to get distinct values from (e.g., "country", "tags", "status")
//   - filter: Query filter to match documents. Can be bson.M, bson.D, or a struct. Use bson.M{} for all documents
//   - opts: Optional DistinctOptions for collation, etc.
//
// Returns:
//   - []any: Slice containing the distinct values found for the field
//   - error: Returns error if operation fails or client is closed
//
// Example:
//
//	filter := bson.M{"status": "active"}
//	countries, err := client.Distinct(ctx, "users", "country", filter)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Active users are from %d different countries\n", len(countries))
//	for _, country := range countries {
//	    fmt.Println(country)
//	}
func (c *Client) Distinct(ctx context.Context, collection string, fieldName string, filter any, opts ...*options.DistinctOptions) ([]any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	results, err := coll.Distinct(ctx, fieldName, filter, opts...)
	if err != nil {
		return nil, errors.NewOperationError("distinct", err)
	}

	return results, nil
}

// CreateIndex creates a new index on the specified collection to improve query performance.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to create the index on
//   - keys: Index keys specification (e.g., bson.M{"email": 1} for ascending, bson.M{"age": -1} for descending)
//   - opts: Optional IndexOptions for unique, sparse, partial filters, TTL, etc.
//
// Returns:
//   - string: The name of the created index
//   - error: Returns error if operation fails or client is closed
//
// Example:
//
//	// Create unique index on email field
//	keys := bson.M{"email": 1}
//	opts := options.Index().SetUnique(true)
//	indexName, err := client.CreateIndex(ctx, "users", keys, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created index: %s\n", indexName)
//
//	// Create compound index on multiple fields
//	compoundKeys := bson.D{{"country", 1}, {"city", 1}, {"created_at", -1}}
//	indexName, err := client.CreateIndex(ctx, "users", compoundKeys)
func (c *Client) CreateIndex(ctx context.Context, collection string, keys any, opts ...*options.IndexOptions) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return "", err
	}

	coll := c.GetCollection(collection)
	indexModel := mongo.IndexModel{Keys: keys}

	if len(opts) > 0 {
		indexModel.Options = opts[0]
	}

	indexName, err := coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return "", errors.NewOperationError("create index", err)
	}

	return indexName, nil
}

// CreateIndexes creates multiple indexes across different collections in a single operation.
// This is useful for setting up database indexes during initialization or migrations.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - indexes: Map where keys are collection names and values are slices of IndexModel to create
//
// Returns:
//   - map[string][]string: Map of collection names to slices of created index names
//   - error: Returns error if any index creation fails or client is closed
//
// Note: If one index fails to create, subsequent indexes will not be attempted.
//
// Example:
//
//	indexes := map[string][]mongo.IndexModel{
//	    "users": {
//	        {
//	            Keys:    bson.M{"email": 1},
//	            Options: options.Index().SetUnique(true).SetName("unique_email"),
//	        },
//	        {
//	            Keys:    bson.D{{"country", 1}, {"city", 1}},
//	            Options: options.Index().SetName("location_idx"),
//	        },
//	    },
//	    "products": {
//	        {
//	            Keys:    bson.M{"sku": 1},
//	            Options: options.Index().SetUnique(true),
//	        },
//	        {
//	            Keys:    bson.M{"category": 1, "price": -1},
//	            Options: options.Index().SetName("category_price_idx"),
//	        },
//	    },
//	}
//	createdIndexes, err := client.CreateIndexes(ctx, indexes)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for collection, indexNames := range createdIndexes {
//	    fmt.Printf("Collection '%s': created %d indexes\n", collection, len(indexNames))
//	}
func (c *Client) CreateIndexes(ctx context.Context, indexes map[string][]mongo.IndexModel) (map[string][]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	result := make(map[string][]string)

	for collectionName, models := range indexes {
		if len(models) == 0 {
			continue
		}

		coll := c.GetCollection(collectionName)
		indexNames, err := coll.Indexes().CreateMany(ctx, models)
		if err != nil {
			return nil, errors.NewOperationError("create indexes on "+collectionName, err)
		}

		result[collectionName] = indexNames
	}

	return result, nil
}

// DropIndex removes an index from the specified collection.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to drop the index from
//   - indexName: Name of the index to drop (can be obtained from ListIndexes or when creating the index)
//
// Returns:
//   - error: Returns error if operation fails, index doesn't exist, or client is closed
//
// Example:
//
//	err := client.DropIndex(ctx, "users", "email_1")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Index dropped successfully")
func (c *Client) DropIndex(ctx context.Context, collection string, indexName string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	coll := c.GetCollection(collection)
	_, err := coll.Indexes().DropOne(ctx, indexName)
	if err != nil {
		return errors.NewOperationError("drop index", err)
	}

	return nil
}

// ListIndexes retrieves information about all indexes on the specified collection.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to list indexes from
//
// Returns:
//   - []bson.M: Slice of index specifications, each containing name, keys, unique, sparse, etc.
//   - error: Returns error if operation fails or client is closed
//
// Example:
//
//	indexes, err := client.ListIndexes(ctx, "users")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, index := range indexes {
//	    fmt.Printf("Index: %s, Keys: %v\n", index["name"], index["key"])
//	}
func (c *Client) ListIndexes(ctx context.Context, collection string) ([]bson.M, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	cursor, err := coll.Indexes().List(ctx)
	if err != nil {
		return nil, errors.NewOperationError("list indexes", err)
	}
	defer cursor.Close(ctx)

	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		return nil, errors.NewOperationError("list indexes decode", err)
	}

	return indexes, nil
}

// ListCollections retrieves the names of all collections in the default database.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns:
//   - []string: Slice of collection names in the database
//   - error: Returns error if operation fails or client is closed
//
// Example:
//
//	collections, err := client.ListCollections(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Found %d collections:\n", len(collections))
//	for _, name := range collections {
//	    fmt.Println("-", name)
//	}
func (c *Client) ListCollections(ctx context.Context) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	db := c.client.Database(c.config.Database)
	cursor, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return nil, errors.NewOperationError("list collections", err)
	}

	return cursor, nil
}

// CreateCollection explicitly creates a new collection in the default database.
// Note: MongoDB creates collections automatically on first insert, but this method is useful
// for creating collections with specific options like validation, capped collections, or time series.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collection: Name of the collection to create
//   - opts: Optional CreateCollectionOptions for validation, capped collections, time series, collation, etc.
//
// Returns:
//   - error: Returns error if collection already exists, operation fails, or client is closed
//
// Example:
//
//	// Create a simple collection
//	err := client.CreateCollection(ctx, "users")
//
//	// Create a capped collection (fixed size, FIFO)
//	cappedOpts := options.CreateCollection().
//	    SetCapped(true).
//	    SetSizeInBytes(1048576). // 1MB
//	    SetMaxDocuments(1000)
//	err := client.CreateCollection(ctx, "logs", cappedOpts)
//
//	// Create collection with schema validation
//	validator := bson.M{
//	    "$jsonSchema": bson.M{
//	        "bsonType": "object",
//	        "required": []string{"name", "email"},
//	        "properties": bson.M{
//	            "name":  bson.M{"bsonType": "string"},
//	            "email": bson.M{"bsonType": "string"},
//	            "age":   bson.M{"bsonType": "int", "minimum": 0},
//	        },
//	    },
//	}
//	validatorOpts := options.CreateCollection().SetValidator(validator)
//	err := client.CreateCollection(ctx, "validated_users", validatorOpts)
func (c *Client) CreateCollection(ctx context.Context, collection string, opts ...*options.CreateCollectionOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	db := c.client.Database(c.config.Database)
	err := db.CreateCollection(ctx, collection, opts...)
	if err != nil {
		return errors.NewOperationError("create collection", err)
	}

	return nil
}

// CreateCollections creates multiple collections in the default database in a single operation.
// This is useful for initializing a database schema with multiple collections at once.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - collections: Map where keys are collection names and values are optional CreateCollectionOptions (can be nil)
//
// Returns:
//   - error: Returns error if any collection creation fails or client is closed
//
// Note: If one collection fails to create, subsequent collections will not be attempted.
// Consider using error handling to manage partial failures if needed.
//
// Example:
//
//	// Create multiple collections with different options
//	collections := map[string]*options.CreateCollectionOptions{
//	    "users":    nil, // Simple collection, no special options
//	    "products": nil,
//	    "logs": options.CreateCollection(). // Capped collection for logs
//	        SetCapped(true).
//	        SetSizeInBytes(5242880). // 5MB
//	        SetMaxDocuments(10000),
//	    "orders": options.CreateCollection(). // Collection with validation
//	        SetValidator(bson.M{
//	            "$jsonSchema": bson.M{
//	                "bsonType": "object",
//	                "required": []string{"user_id", "total"},
//	            },
//	        }),
//	}
//	err := client.CreateCollections(ctx, collections)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("All collections created successfully")
func (c *Client) CreateCollections(ctx context.Context, collections map[string]*options.CreateCollectionOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	db := c.client.Database(c.config.Database)

	for collectionName, opts := range collections {
		// CreateCollection accepts variadic options, so we can pass nil or the options directly
		var createOpts []*options.CreateCollectionOptions
		if opts != nil {
			createOpts = []*options.CreateCollectionOptions{opts}
		}

		if err := db.CreateCollection(ctx, collectionName, createOpts...); err != nil {
			return errors.NewOperationError("create collection "+collectionName, err)
		}
	}

	return nil
}
