package mongo_kit

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository Pattern
//
// This file provides a type-safe, generic repository pattern for MongoDB operations.
//
// See docs/repository.md for detailed usage guide and examples.

// Repository provides a type-safe, collection-specific interface for database operations.
type Repository[T any] struct {
	client     *Client
	collection string
}

// NewRepository creates a new type-safe repository for the specified collection.
func NewRepository[T any](client *Client, collection string) *Repository[T] {
	return &Repository[T]{
		client:     client,
		collection: collection,
	}
}

// Create inserts a new document and returns its ID.
func (r *Repository[T]) Create(ctx context.Context, document T) (any, error) {
	result, err := r.client.insertOne(ctx, r.collection, document)
	if err != nil {
		return nil, err
	}
	return result.InsertedID, nil
}

// CreateMany inserts multiple documents and returns their IDs.
func (r *Repository[T]) CreateMany(ctx context.Context, documents []T) ([]any, error) {
	// Convert []T to []any for InsertMany
	docs := make([]any, len(documents))
	for i, doc := range documents {
		docs[i] = doc
	}

	result, err := r.client.insertMany(ctx, r.collection, docs)
	if err != nil {
		return nil, err
	}
	return result.InsertedIDs, nil
}

// FindByID finds a single document by its _id field.
// Returns mongo.ErrNoDocuments if not found.
func (r *Repository[T]) FindByID(ctx context.Context, id any) (*T, error) {
	var result T
	err := r.client.findByID(ctx, r.collection, id, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// FindOne finds a single document matching the filter.
// Returns mongo.ErrNoDocuments if not found.
func (r *Repository[T]) FindOne(ctx context.Context, filter any, opts ...*options.FindOneOptions) (*T, error) {
	var result T
	err := r.client.findOne(ctx, r.collection, filter, &result, opts...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Find finds all documents matching the filter.
func (r *Repository[T]) Find(ctx context.Context, filter any, opts ...*options.FindOptions) ([]T, error) {
	var results []T
	err := r.client.find(ctx, r.collection, filter, &results, opts...)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// FindAll returns all documents in the collection.
func (r *Repository[T]) FindAll(ctx context.Context, opts ...*options.FindOptions) ([]T, error) {
	return r.Find(ctx, bson.M{}, opts...)
}

// UpdateByID updates a single document by its _id field.
func (r *Repository[T]) UpdateByID(ctx context.Context, id any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return r.client.updateByID(ctx, r.collection, id, update, opts...)
}

// UpdateOne updates a single document matching the filter.
func (r *Repository[T]) UpdateOne(ctx context.Context, filter any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return r.client.updateOne(ctx, r.collection, filter, update, opts...)
}

// UpdateMany updates all documents matching the filter.
func (r *Repository[T]) UpdateMany(ctx context.Context, filter any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return r.client.updateMany(ctx, r.collection, filter, update, opts...)
}

// Upsert updates a document if it exists, or inserts it if it doesn't.
func (r *Repository[T]) Upsert(ctx context.Context, filter any, update any) (*mongo.UpdateResult, error) {
	return r.client.upsertOne(ctx, r.collection, filter, update)
}

// DeleteByID deletes a single document by its _id field.
func (r *Repository[T]) DeleteByID(ctx context.Context, id any) (*mongo.DeleteResult, error) {
	return r.client.deleteByID(ctx, r.collection, id)
}

// DeleteOne deletes a single document matching the filter.
func (r *Repository[T]) DeleteOne(ctx context.Context, filter any) (*mongo.DeleteResult, error) {
	return r.client.deleteOne(ctx, r.collection, filter)
}

// DeleteMany deletes all documents matching the filter.
func (r *Repository[T]) DeleteMany(ctx context.Context, filter any) (*mongo.DeleteResult, error) {
	return r.client.deleteMany(ctx, r.collection, filter)
}

// Count returns the number of documents matching the filter.
func (r *Repository[T]) Count(ctx context.Context, filter any, opts ...*options.CountOptions) (int64, error) {
	return r.client.countDocuments(ctx, r.collection, filter, opts...)
}

// CountAll counts all documents in the collection.
func (r *Repository[T]) CountAll(ctx context.Context, opts ...*options.CountOptions) (int64, error) {
	return r.Count(ctx, bson.M{}, opts...)
}

// EstimatedCount returns an estimated count using collection metadata.
// Faster than Count but may be less accurate.
func (r *Repository[T]) EstimatedCount(ctx context.Context, opts ...*options.EstimatedDocumentCountOptions) (int64, error) {
	return r.client.estimatedDocumentCount(ctx, r.collection, opts...)
}

// Exists checks if at least one document matching the filter exists.
func (r *Repository[T]) Exists(ctx context.Context, filter any) (bool, error) {
	count, err := r.client.countDocuments(ctx, r.collection, filter, options.Count().SetLimit(1))
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByID checks if a document with the given _id exists.
func (r *Repository[T]) ExistsByID(ctx context.Context, id any) (bool, error) {
	_, err := r.FindByID(ctx, id)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Aggregate executes an aggregation pipeline and returns typed results.
func (r *Repository[T]) Aggregate(ctx context.Context, pipeline any, opts ...*options.AggregateOptions) ([]T, error) {
	var results []T
	err := r.client.aggregate(ctx, r.collection, pipeline, &results, opts...)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// Drop deletes the entire collection.
// WARNING: This permanently deletes all documents and indexes.
func (r *Repository[T]) Drop(ctx context.Context) error {
	return r.client.dropCollection(ctx, r.collection)
}

// FindWithBuilder finds documents using a QueryBuilder for complex queries.
func (r *Repository[T]) FindWithBuilder(ctx context.Context, qb *QueryBuilder) ([]T, error) {
	filter, opts := qb.Build()
	return r.Find(ctx, filter, opts)
}

// FindOneWithBuilder finds a single document using a QueryBuilder.
// Returns mongo.ErrNoDocuments if not found.
func (r *Repository[T]) FindOneWithBuilder(ctx context.Context, qb *QueryBuilder) (*T, error) {
	filter, opts := qb.Build()

	// Extract FindOneOptions from FindOptions
	findOneOpts := options.FindOne()
	if opts.Sort != nil {
		findOneOpts.SetSort(opts.Sort)
	}
	if opts.Projection != nil {
		findOneOpts.SetProjection(opts.Projection)
	}
	if opts.Skip != nil {
		findOneOpts.SetSkip(*opts.Skip)
	}

	return r.FindOne(ctx, filter, findOneOpts)
}

// CountWithBuilder counts documents using a QueryBuilder filter.
func (r *Repository[T]) CountWithBuilder(ctx context.Context, qb *QueryBuilder) (int64, error) {
	filter := qb.GetFilter()
	return r.Count(ctx, filter)
}

// ExistsWithBuilder checks if at least one document matching the QueryBuilder exists.
func (r *Repository[T]) ExistsWithBuilder(ctx context.Context, qb *QueryBuilder) (bool, error) {
	filter := qb.GetFilter()
	return r.Exists(ctx, filter)
}

// Collection returns the name of the collection this repository operates on.
func (r *Repository[T]) Collection() string {
	return r.collection
}
