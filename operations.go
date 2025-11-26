package mongo

import (
	"context"

	"github.com/edaniel30/mongo-kit-go/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertOne inserts a single document into the specified collection
func (c *Client) InsertOne(ctx context.Context, database, collection string, document any) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
	result, err := coll.InsertOne(ctx, document)
	if err != nil {
		return nil, errors.NewOperationError("insert one", err)
	}

	return result.InsertedID, nil
}

// InsertMany inserts multiple documents into the specified collection
func (c *Client) InsertMany(ctx context.Context, database, collection string, documents []any) ([]any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
	result, err := coll.InsertMany(ctx, documents)
	if err != nil {
		return nil, errors.NewOperationError("insert many", err)
	}

	return result.InsertedIDs, nil
}

// FindOne finds a single document matching the filter
func (c *Client) FindOne(ctx context.Context, database, collection string, filter any, result any, opts ...*options.FindOneOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
	err := coll.FindOne(ctx, filter, opts...).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.ErrDocumentNotFound()
		}
		return errors.NewOperationError("find one", err)
	}

	return nil
}

// Find finds all documents matching the filter
func (c *Client) Find(ctx context.Context, database, collection string, filter any, results any, opts ...*options.FindOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
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

// UpdateOne updates a single document matching the filter
func (c *Client) UpdateOne(ctx context.Context, database, collection string, filter any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
	result, err := coll.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		return nil, errors.NewOperationError("update one", err)
	}

	return result, nil
}

// UpdateMany updates all documents matching the filter
func (c *Client) UpdateMany(ctx context.Context, database, collection string, filter any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
	result, err := coll.UpdateMany(ctx, filter, update, opts...)
	if err != nil {
		return nil, errors.NewOperationError("update many", err)
	}

	return result, nil
}

// ReplaceOne replaces a single document matching the filter
func (c *Client) ReplaceOne(ctx context.Context, database, collection string, filter any, replacement any, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
	result, err := coll.ReplaceOne(ctx, filter, replacement, opts...)
	if err != nil {
		return nil, errors.NewOperationError("replace one", err)
	}

	return result, nil
}

// DeleteOne deletes a single document matching the filter
func (c *Client) DeleteOne(ctx context.Context, database, collection string, filter any, opts ...*options.DeleteOptions) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
	result, err := coll.DeleteOne(ctx, filter, opts...)
	if err != nil {
		return 0, errors.NewOperationError("delete one", err)
	}

	return result.DeletedCount, nil
}

// DeleteMany deletes all documents matching the filter
func (c *Client) DeleteMany(ctx context.Context, database, collection string, filter any, opts ...*options.DeleteOptions) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
	result, err := coll.DeleteMany(ctx, filter, opts...)
	if err != nil {
		return 0, errors.NewOperationError("delete many", err)
	}

	return result.DeletedCount, nil
}

// CountDocuments counts documents matching the filter
func (c *Client) CountDocuments(ctx context.Context, database, collection string, filter any, opts ...*options.CountOptions) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
	count, err := coll.CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, errors.NewOperationError("count documents", err)
	}

	return count, nil
}

// Aggregate executes an aggregation pipeline
func (c *Client) Aggregate(ctx context.Context, database, collection string, pipeline any, results any, opts ...*options.AggregateOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
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

// FindOneAndUpdate finds a single document and updates it
func (c *Client) FindOneAndUpdate(ctx context.Context, database, collection string, filter any, update any, result any, opts ...*options.FindOneAndUpdateOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
	err := coll.FindOneAndUpdate(ctx, filter, update, opts...).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.ErrDocumentNotFound()
		}
		return errors.NewOperationError("find one and update", err)
	}

	return nil
}

// FindOneAndReplace finds a single document and replaces it
func (c *Client) FindOneAndReplace(ctx context.Context, database, collection string, filter any, replacement any, result any, opts ...*options.FindOneAndReplaceOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
	err := coll.FindOneAndReplace(ctx, filter, replacement, opts...).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.ErrDocumentNotFound()
		}
		return errors.NewOperationError("find one and replace", err)
	}

	return nil
}

// FindOneAndDelete finds a single document and deletes it
func (c *Client) FindOneAndDelete(ctx context.Context, database, collection string, filter any, result any, opts ...*options.FindOneAndDeleteOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
	err := coll.FindOneAndDelete(ctx, filter, opts...).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.ErrDocumentNotFound()
		}
		return errors.NewOperationError("find one and delete", err)
	}

	return nil
}

// BulkWrite executes multiple write operations in bulk
func (c *Client) BulkWrite(ctx context.Context, database, collection string, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
	result, err := coll.BulkWrite(ctx, models, opts...)
	if err != nil {
		return nil, errors.NewOperationError("bulk write", err)
	}

	return result, nil
}

// Distinct gets distinct values for a specified field
func (c *Client) Distinct(ctx context.Context, database, collection string, fieldName string, filter any, opts ...*options.DistinctOptions) ([]any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
	results, err := coll.Distinct(ctx, fieldName, filter, opts...)
	if err != nil {
		return nil, errors.NewOperationError("distinct", err)
	}

	return results, nil
}

// CreateIndex creates a new index on the specified collection
func (c *Client) CreateIndex(ctx context.Context, database, collection string, keys any, opts ...*options.IndexOptions) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return "", errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index(),
	}

	if len(opts) > 0 {
		indexModel.Options = opts[0]
	}

	indexName, err := coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return "", errors.NewOperationError("create index", err)
	}

	return indexName, nil
}

// DropIndex drops an index from the specified collection
func (c *Client) DropIndex(ctx context.Context, database, collection string, indexName string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
	_, err := coll.Indexes().DropOne(ctx, indexName)
	if err != nil {
		return errors.NewOperationError("drop index", err)
	}

	return nil
}

// ListIndexes lists all indexes on the specified collection
func (c *Client) ListIndexes(ctx context.Context, database, collection string) ([]bson.M, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, errors.ErrClientClosed()
	}

	coll := c.GetCollectionFrom(database, collection)
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
