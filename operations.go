package mongo_kit

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database Operations
//
// This file provides all database operations (CRUD, aggregations, indexes, etc.).
//
// See docs/operations.md for detailed usage guide and examples.

// InsertOne inserts a single document into the specified collection.
// Returns *mongo.InsertOneResult with the InsertedID field.
func (c *Client) InsertOne(ctx context.Context, collection string, document any) (*mongo.InsertOneResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	result, err := coll.InsertOne(ctx, document)
	if err != nil {
		return nil, newOperationError("insert one", err)
	}

	return result, nil
}

// InsertMany inserts multiple documents into the specified collection in a single operation.
// Returns *mongo.InsertManyResult with the InsertedIDs map.
func (c *Client) InsertMany(ctx context.Context, collection string, documents []any) (*mongo.InsertManyResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	result, err := coll.InsertMany(ctx, documents)
	if err != nil {
		return nil, newOperationError("insert many", err)
	}

	return result, nil
}

// FindOne finds a single document matching the filter and decodes it into result.
// Returns mongo.ErrNoDocuments if no document matches.
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
		return newOperationError("find one", err)
	}

	return nil
}

// Find finds all documents matching the filter and decodes them into results.
// Empty results is not an error.
func (c *Client) Find(ctx context.Context, collection string, filter any, results any, opts ...*options.FindOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	coll := c.GetCollection(collection)
	cursor, err := coll.Find(ctx, filter, opts...)
	if err != nil {
		return newOperationError("find", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, results); err != nil {
		return newOperationError("find decode", err)
	}

	return nil
}

// UpdateOne updates a single document matching the filter.
// Update must use operators like $set, $inc, etc.
func (c *Client) UpdateOne(ctx context.Context, collection string, filter any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	result, err := coll.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		return nil, newOperationError("update one", err)
	}

	return result, nil
}

// UpdateMany updates all documents matching the filter.
// Update must use operators like $set, $inc, etc.
func (c *Client) UpdateMany(ctx context.Context, collection string, filter any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	result, err := coll.UpdateMany(ctx, filter, update, opts...)
	if err != nil {
		return nil, newOperationError("update many", err)
	}

	return result, nil
}

// ReplaceOne replaces a single document matching the filter with a new document.
// Replaces the entire document (except _id). Cannot use update operators.
func (c *Client) ReplaceOne(ctx context.Context, collection string, filter any, replacement any, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	result, err := coll.ReplaceOne(ctx, filter, replacement, opts...)
	if err != nil {
		return nil, newOperationError("replace one", err)
	}

	return result, nil
}

// DeleteOne deletes a single document matching the filter.
// Returns *mongo.DeleteResult with DeletedCount (0 or 1).
func (c *Client) DeleteOne(ctx context.Context, collection string, filter any, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	result, err := coll.DeleteOne(ctx, filter, opts...)
	if err != nil {
		return nil, newOperationError("delete one", err)
	}

	return result, nil
}

// DeleteMany deletes all documents matching the filter.
// Use bson.M{} to delete all documents.
func (c *Client) DeleteMany(ctx context.Context, collection string, filter any, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	result, err := coll.DeleteMany(ctx, filter, opts...)
	if err != nil {
		return nil, newOperationError("delete many", err)
	}

	return result, nil
}

// CountDocuments counts the number of documents matching the filter.
// Accurate but slower than EstimatedDocumentCount.
func (c *Client) CountDocuments(ctx context.Context, collection string, filter any, opts ...*options.CountOptions) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return 0, err
	}

	coll := c.GetCollection(collection)
	count, err := coll.CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, newOperationError("count documents", err)
	}

	return count, nil
}

// Aggregate runs an aggregation pipeline and decodes results.
// Pipeline must be []bson.M, []bson.D, mongo.Pipeline, or bson.A.
func (c *Client) Aggregate(ctx context.Context, collection string, pipeline any, results any, opts ...*options.AggregateOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	// Validate pipeline type
	switch pipeline.(type) {
	case []bson.M, []bson.D, mongo.Pipeline, bson.A:
		// Valid types - continue
	case nil:
		return newOperationError("aggregate", errors.New("pipeline cannot be nil"))
	default:
		return newOperationError("aggregate", errors.New("pipeline must be []bson.M, []bson.D, mongo.Pipeline, or bson.A"))
	}

	coll := c.GetCollection(collection)
	cursor, err := coll.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return newOperationError("aggregate", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, results); err != nil {
		return newOperationError("aggregate decode", err)
	}

	return nil
}

// FindOneAndUpdate finds, updates, and returns the document atomically.
// Use options.After to return the updated document.
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
		return newOperationError("find one and update", err)
	}

	return nil
}

// FindOneAndReplace finds, replaces, and returns the document atomically.
// Use options.After to return the replaced document.
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
		return newOperationError("find one and replace", err)
	}

	return nil
}

// FindOneAndDelete finds, deletes, and returns the deleted document atomically.
// Returns mongo.ErrNoDocuments if no document matches.
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
		return newOperationError("find one and delete", err)
	}

	return nil
}

// BulkWrite executes multiple write operations in a single batch.
// More efficient than individual operations for bulk changes.
func (c *Client) BulkWrite(ctx context.Context, collection string, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	result, err := coll.BulkWrite(ctx, models, opts...)
	if err != nil {
		return nil, newOperationError("bulk write", err)
	}

	return result, nil
}

// Distinct returns all unique values for a specified field.
// Use bson.M{} filter for all documents.
func (c *Client) Distinct(ctx context.Context, collection string, fieldName string, filter any, opts ...*options.DistinctOptions) ([]any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	results, err := coll.Distinct(ctx, fieldName, filter, opts...)
	if err != nil {
		return nil, newOperationError("distinct", err)
	}

	return results, nil
}

// CreateIndex creates a new index on the specified collection.
// Returns the name of the created index.
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
		return "", newOperationError("create index", err)
	}

	return indexName, nil
}

// CreateIndexes creates multiple indexes across different collections.
// Returns map of collection names to created index names.
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
			return nil, newOperationError("create indexes on "+collectionName, err)
		}

		result[collectionName] = indexNames
	}

	return result, nil
}

// DropIndex removes an index from the specified collection.
// Returns error if index doesn't exist.
func (c *Client) DropIndex(ctx context.Context, collection string, indexName string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	coll := c.GetCollection(collection)
	_, err := coll.Indexes().DropOne(ctx, indexName)
	if err != nil {
		return newOperationError("drop index", err)
	}

	return nil
}

// ListIndexes retrieves information about all indexes on the collection.
// Returns slice of index specifications including name, keys, and options.
func (c *Client) ListIndexes(ctx context.Context, collection string) ([]bson.M, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	cursor, err := coll.Indexes().List(ctx)
	if err != nil {
		return nil, newOperationError("list indexes", err)
	}
	defer cursor.Close(ctx)

	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		return nil, newOperationError("list indexes decode", err)
	}

	return indexes, nil
}

// ListCollections retrieves the names of all collections in the default database.
// Returns slice of collection names.
func (c *Client) ListCollections(ctx context.Context) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	db := c.client.Database(c.config.Database)
	cursor, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return nil, newOperationError("list collections", err)
	}

	return cursor, nil
}

// CreateCollection explicitly creates a new collection in the default database.
// Use options for validation, capped collections, or time series.
func (c *Client) CreateCollection(ctx context.Context, collection string, opts ...*options.CreateCollectionOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	db := c.client.Database(c.config.Database)
	err := db.CreateCollection(ctx, collection, opts...)
	if err != nil {
		return newOperationError("create collection", err)
	}

	return nil
}

// CreateCollections creates multiple collections in the default database.
// Map keys are collection names, values are optional CreateCollectionOptions.
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
			return newOperationError("create collection "+collectionName, err)
		}
	}

	return nil
}

// FindByID finds a single document by its _id field.
// ID can be string or primitive.ObjectID.
func (c *Client) FindByID(ctx context.Context, collection string, id any, result any) error {
	var objID primitive.ObjectID
	var err error

	switch v := id.(type) {
	case string:
		objID, err = primitive.ObjectIDFromHex(v)
		if err != nil {
			return newOperationError("find by id", err)
		}
	case primitive.ObjectID:
		if v.IsZero() {
			return newOperationError("find by id", errors.New("ObjectID cannot be zero"))
		}
		objID = v
	default:
		return newOperationError("find by id", mongo.ErrInvalidIndexValue)
	}

	filter := bson.M{"_id": objID}
	return c.FindOne(ctx, collection, filter, result)
}

// UpdateByID updates a single document by its _id field.
// ID can be string or primitive.ObjectID.
func (c *Client) UpdateByID(ctx context.Context, collection string, id any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	var objID primitive.ObjectID
	var err error

	switch v := id.(type) {
	case string:
		objID, err = primitive.ObjectIDFromHex(v)
		if err != nil {
			return nil, newOperationError("update by id", err)
		}
	case primitive.ObjectID:
		if v.IsZero() {
			return nil, newOperationError("update by id", errors.New("ObjectID cannot be zero"))
		}
		objID = v
	default:
		return nil, newOperationError("update by id", mongo.ErrInvalidIndexValue)
	}

	filter := bson.M{"_id": objID}
	return c.UpdateOne(ctx, collection, filter, update, opts...)
}

// DeleteByID deletes a single document by its _id field.
// ID can be string or primitive.ObjectID.
func (c *Client) DeleteByID(ctx context.Context, collection string, id any) (*mongo.DeleteResult, error) {
	var objID primitive.ObjectID
	var err error

	switch v := id.(type) {
	case string:
		objID, err = primitive.ObjectIDFromHex(v)
		if err != nil {
			return nil, newOperationError("delete by id", err)
		}
	case primitive.ObjectID:
		if v.IsZero() {
			return nil, newOperationError("delete by id", errors.New("ObjectID cannot be zero"))
		}
		objID = v
	default:
		return nil, newOperationError("delete by id", mongo.ErrInvalidIndexValue)
	}

	filter := bson.M{"_id": objID}
	return c.DeleteOne(ctx, collection, filter)
}

// EstimatedDocumentCount returns an estimated count using collection metadata.
// Faster than CountDocuments but less accurate. Does not support filters.
func (c *Client) EstimatedDocumentCount(ctx context.Context, collection string, opts ...*options.EstimatedDocumentCountOptions) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return 0, err
	}

	coll := c.GetCollection(collection)
	count, err := coll.EstimatedDocumentCount(ctx, opts...)
	if err != nil {
		return 0, newOperationError("estimated document count", err)
	}

	return count, nil
}

// UpsertOne updates a document if it exists, or inserts it if it doesn't.
// Returns UpsertedID if inserted, or MatchedCount/ModifiedCount if updated.
func (c *Client) UpsertOne(ctx context.Context, collection string, filter any, update any) (*mongo.UpdateResult, error) {
	opts := options.Update().SetUpsert(true)
	return c.UpdateOne(ctx, collection, filter, update, opts)
}

// DropCollection drops an entire collection from the database.
// WARNING: Permanently deletes all documents and indexes.
func (c *Client) DropCollection(ctx context.Context, collection string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	coll := c.GetCollection(collection)
	if err := coll.Drop(ctx); err != nil {
		return newOperationError("drop collection", err)
	}

	return nil
}

// DropDatabase drops an entire database.
// WARNING: Permanently deletes the database and all its collections.
func (c *Client) DropDatabase(ctx context.Context, database string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	db := c.GetDatabase(database)
	if err := db.Drop(ctx); err != nil {
		return newOperationError("drop database", err)
	}

	return nil
}

// WithTransaction executes a function within a transaction with automatic retry.
// All operations in the function must use the provided SessionContext.
func (c *Client) WithTransaction(ctx context.Context, fn func(mongo.SessionContext) error, opts ...*options.TransactionOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	session, err := c.client.StartSession()
	if err != nil {
		return newOperationError("start transaction session", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		return nil, fn(sessCtx)
	}, opts...)

	if err != nil {
		return newOperationError("transaction", err)
	}

	return nil
}

// Watch opens a change stream to watch for real-time changes on a collection.
// Requires MongoDB replica set or sharded cluster (not standalone).
func (c *Client) Watch(ctx context.Context, collection string, pipeline any, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.GetCollection(collection)
	stream, err := coll.Watch(ctx, pipeline, opts...)
	if err != nil {
		return nil, newOperationError("watch", err)
	}

	return stream, nil
}
