package mongo

import (
	"github.com/edaniel30/mongo-kit-go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// M is a convenient alias for bson.M
type M = bson.M

// D is a convenient alias for bson.D
type D = bson.D

// E is a convenient alias for bson.E
type E = bson.E

// A is a convenient alias for bson.A
type A = bson.A

// FindOptions is a convenient alias for options.FindOptions
type FindOptions = options.FindOptions

// FindOneOptions is a convenient alias for options.FindOneOptions
type FindOneOptions = options.FindOneOptions

// UpdateOptions is a convenient alias for options.UpdateOptions
type UpdateOptions = options.UpdateOptions

// InsertOneOptions is a convenient alias for options.InsertOneOptions
type InsertOneOptions = options.InsertOneOptions

// InsertManyOptions is a convenient alias for options.InsertManyOptions
type InsertManyOptions = options.InsertManyOptions

// DeleteOptions is a convenient alias for options.DeleteOptions
type DeleteOptions = options.DeleteOptions

// ReplaceOptions is a convenient alias for options.ReplaceOptions
type ReplaceOptions = options.ReplaceOptions

// CountOptions is a convenient alias for options.CountOptions
type CountOptions = options.CountOptions

// AggregateOptions is a convenient alias for options.AggregateOptions
type AggregateOptions = options.AggregateOptions

// FindOneAndUpdateOptions is a convenient alias for options.FindOneAndUpdateOptions
type FindOneAndUpdateOptions = options.FindOneAndUpdateOptions

// FindOneAndReplaceOptions is a convenient alias for options.FindOneAndReplaceOptions
type FindOneAndReplaceOptions = options.FindOneAndReplaceOptions

// FindOneAndDeleteOptions is a convenient alias for options.FindOneAndDeleteOptions
type FindOneAndDeleteOptions = options.FindOneAndDeleteOptions

// BulkWriteOptions is a convenient alias for options.BulkWriteOptions
type BulkWriteOptions = options.BulkWriteOptions

// IndexOptions is a convenient alias for options.IndexOptions
type IndexOptions = options.IndexOptions

// DistinctOptions is a convenient alias for options.DistinctOptions
type DistinctOptions = options.DistinctOptions

// CreateCollectionOptions is a convenient alias for options.CreateCollectionOptions
type CreateCollectionOptions = options.CreateCollectionOptions

// IndexModel represents an index to be created
type IndexModel = mongo.IndexModel

// WriteModel is an interface for bulk write operations
type WriteModel = mongo.WriteModel

// Config holds MongoDB client configuration
type Config = models.Config

// Option is a function that modifies Config
type Option = models.Option

// MongoDB primitive types for convenience
type ObjectID = primitive.ObjectID
