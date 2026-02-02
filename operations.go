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

// insertOne inserts a single document into the specified collection.
// Returns *mongo.InsertOneResult with the InsertedID field.
// This method is unexported and used internally by repositories.
func (c *client) insertOne(ctx context.Context, collection string, document any) (*mongo.InsertOneResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.getCollection(collection)
	result, err := coll.InsertOne(ctx, document)
	if err != nil {
		return nil, newOperationError("insert one", err)
	}

	return result, nil
}

// insertMany inserts multiple documents into the specified collection in a single operation.
// Returns *mongo.InsertManyResult with the InsertedIDs map.
func (c *client) insertMany(ctx context.Context, collection string, documents []any) (*mongo.InsertManyResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.getCollection(collection)
	result, err := coll.InsertMany(ctx, documents)
	if err != nil {
		return nil, newOperationError("insert many", err)
	}

	return result, nil
}

// findOne finds a single document matching the filter and decodes it into result.
// Returns mongo.ErrNoDocuments if no document matches.
func (c *client) findOne(ctx context.Context, collection string, filter any, result any, opts ...*options.FindOneOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	coll := c.getCollection(collection)
	err := coll.FindOne(ctx, filter, opts...).Decode(result)
	if err != nil {
		// Return ErrNoDocuments directly for clearer error handling
		if errors.Is(err, mongo.ErrNoDocuments) {
			return err
		}
		return newOperationError("find one", err)
	}

	return nil
}

// find finds all documents matching the filter and decodes them into results.
// Empty results is not an error.
func (c *client) find(ctx context.Context, collection string, filter any, results any, opts ...*options.FindOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	coll := c.getCollection(collection)
	cursor, err := coll.Find(ctx, filter, opts...)
	if err != nil {
		return newOperationError("find", err)
	}
	defer func() { _ = cursor.Close(ctx) }()

	if err := cursor.All(ctx, results); err != nil {
		return newOperationError("find decode", err)
	}

	return nil
}

// updateOne updates a single document matching the filter.
// Update must use operators like $set, $inc, etc.
func (c *client) updateOne(ctx context.Context, collection string, filter any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.getCollection(collection)
	result, err := coll.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		return nil, newOperationError("update one", err)
	}

	return result, nil
}

// updateMany updates all documents matching the filter.
// Update must use operators like $set, $inc, etc.
func (c *client) updateMany(ctx context.Context, collection string, filter any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.getCollection(collection)
	result, err := coll.UpdateMany(ctx, filter, update, opts...)
	if err != nil {
		return nil, newOperationError("update many", err)
	}

	return result, nil
}

// deleteOne deletes a single document matching the filter.
// Returns *mongo.DeleteResult with DeletedCount (0 or 1).
func (c *client) deleteOne(ctx context.Context, collection string, filter any, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.getCollection(collection)
	result, err := coll.DeleteOne(ctx, filter, opts...)
	if err != nil {
		return nil, newOperationError("delete one", err)
	}

	return result, nil
}

// deleteMany deletes all documents matching the filter.
// Use bson.M{} to delete all documents.
func (c *client) deleteMany(ctx context.Context, collection string, filter any, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	coll := c.getCollection(collection)
	result, err := coll.DeleteMany(ctx, filter, opts...)
	if err != nil {
		return nil, newOperationError("delete many", err)
	}

	return result, nil
}

// countDocuments counts the number of documents matching the filter.
// Accurate but slower than EstimatedDocumentCount.
func (c *client) countDocuments(ctx context.Context, collection string, filter any, opts ...*options.CountOptions) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return 0, err
	}

	coll := c.getCollection(collection)
	count, err := coll.CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, newOperationError("count documents", err)
	}

	return count, nil
}

// aggregate runs an aggregation pipeline and decodes results.
// Pipeline must be []bson.M, []bson.D, mongo.Pipeline, or bson.A.
func (c *client) aggregate(ctx context.Context, collection string, pipeline any, results any, opts ...*options.AggregateOptions) error {
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

	coll := c.getCollection(collection)
	cursor, err := coll.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return newOperationError("aggregate", err)
	}
	defer func() { _ = cursor.Close(ctx) }()

	if err := cursor.All(ctx, results); err != nil {
		return newOperationError("aggregate decode", err)
	}

	return nil
}

// convertToObjectID converts a string or ObjectID to primitive.ObjectID.
// Returns an error if the conversion fails or the ObjectID is invalid.
func convertToObjectID(id any, operation string) (primitive.ObjectID, error) {
	switch v := id.(type) {
	case string:
		objID, err := primitive.ObjectIDFromHex(v)
		if err != nil {
			return primitive.ObjectID{}, newOperationError(operation, err)
		}
		return objID, nil
	case primitive.ObjectID:
		if v.IsZero() {
			return primitive.ObjectID{}, newOperationError(operation, errors.New("ObjectID cannot be zero"))
		}
		return v, nil
	default:
		return primitive.ObjectID{}, newOperationError(operation, mongo.ErrInvalidIndexValue)
	}
}

// findByID finds a single document by its _id field.
// ID can be string or primitive.ObjectID.
func (c *client) findByID(ctx context.Context, collection string, id any, result any) error {
	objID, err := convertToObjectID(id, "find by id")
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}
	return c.findOne(ctx, collection, filter, result)
}

// updateByID updates a single document by its _id field.
// ID can be string or primitive.ObjectID.
func (c *client) updateByID(ctx context.Context, collection string, id any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	objID, err := convertToObjectID(id, "update by id")
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objID}
	return c.updateOne(ctx, collection, filter, update, opts...)
}

// deleteByID deletes a single document by its _id field.
// ID can be string or primitive.ObjectID.
func (c *client) deleteByID(ctx context.Context, collection string, id any) (*mongo.DeleteResult, error) {
	objID, err := convertToObjectID(id, "delete by id")
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objID}
	return c.deleteOne(ctx, collection, filter)
}

// estimatedDocumentCount returns an estimated count using collection metadata.
// Faster than CountDocuments but less accurate. Does not support filters.
func (c *client) estimatedDocumentCount(ctx context.Context, collection string, opts ...*options.EstimatedDocumentCountOptions) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return 0, err
	}

	coll := c.getCollection(collection)
	count, err := coll.EstimatedDocumentCount(ctx, opts...)
	if err != nil {
		return 0, newOperationError("estimated document count", err)
	}

	return count, nil
}

// upsertOne updates a document if it exists, or inserts it if it doesn't.
// Returns UpsertedID if inserted, or MatchedCount/ModifiedCount if updated.
func (c *client) upsertOne(ctx context.Context, collection string, filter any, update any) (*mongo.UpdateResult, error) {
	opts := options.Update().SetUpsert(true)
	return c.updateOne(ctx, collection, filter, update, opts)
}

// dropCollection drops an entire collection from the database.
// WARNING: Permanently deletes all documents and indexes.
func (c *client) dropCollection(ctx context.Context, collection string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	coll := c.getCollection(collection)
	if err := coll.Drop(ctx); err != nil {
		return newOperationError("drop collection", err)
	}

	return nil
}

