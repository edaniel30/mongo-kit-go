package mongo

import (
	"github.com/edaniel30/mongo-kit-go/errors"
	"github.com/edaniel30/mongo-kit-go/internal/helpers"
	"github.com/edaniel30/mongo-kit-go/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// DefaultConfig returns a Config with sensible default values
	DefaultConfig = models.DefaultConfig
)

var (
	// WithURI sets the MongoDB connection URI
	WithURI = models.WithURI
	// WithDatabase sets the default database name
	WithDatabase = models.WithDatabase
	// WithMaxPoolSize sets the maximum connection pool size
	WithMaxPoolSize = models.WithMaxPoolSize
	// WithMinPoolSize sets the minimum connection pool size
	WithMinPoolSize = models.WithMinPoolSize
	// WithConnectTimeout sets the connection timeout
	WithConnectTimeout = models.WithConnectTimeout
	// WithServerSelectionTimeout sets the server selection timeout
	WithServerSelectionTimeout = models.WithServerSelectionTimeout
	// WithSocketTimeout sets the socket operation timeout
	WithSocketTimeout = models.WithSocketTimeout
	// WithTimeout sets the default operation timeout
	WithTimeout = models.WithTimeout
	// WithDebug enables or disables debug logging
	WithDebug = models.WithDebug
	// WithRetryWrites enables or disables automatic retry of write operations
	WithRetryWrites = models.WithRetryWrites
	// WithRetryReads enables or disables automatic retry of read operations
	WithRetryReads = models.WithRetryReads
	// WithAppName sets the application name for MongoDB logs
	WithAppName = models.WithAppName
	// WithDirectConnection sets whether to connect directly to a single host
	WithDirectConnection = models.WithDirectConnection
	// WithReplicaSet sets the replica set name
	WithReplicaSet = models.WithReplicaSet
	// WithReadPreference sets the read preference mode
	WithReadPreference = models.WithReadPreference
	// WithMaxConnIdleTime sets the maximum idle time for connections
	WithMaxConnIdleTime = models.WithMaxConnIdleTime
	// WithHeartbeatInterval sets the interval between server heartbeats
	WithHeartbeatInterval = models.WithHeartbeatInterval
)

// Re-export custom error functions
var (
	// ErrInvalidConfig returns a configuration error with the given message
	ErrInvalidConfig = errors.ErrInvalidConfig
	// ErrConnectionFailed returns an error indicating connection failure
	ErrConnectionFailed = errors.ErrConnectionFailed
	// ErrClientClosed returns an error indicating the client is closed
	ErrClientClosed = errors.ErrClientClosed
	// ErrInvalidOperation returns an error indicating an invalid operation
	ErrInvalidOperation = errors.ErrInvalidOperation
	// ErrDatabaseNotFound returns an error indicating database was not found
	ErrDatabaseNotFound = errors.ErrDatabaseNotFound
	// ErrCollectionNotFound returns an error indicating collection was not found
	ErrCollectionNotFound = errors.ErrCollectionNotFound
	// NewOperationError creates a new operation error
	NewOperationError = errors.NewOperationError
)

// Re-export helper functions
var (
	// ToObjectID converts a string to MongoDB ObjectID
	ToObjectID = helpers.ToObjectID
	// ToObjectIDs converts multiple strings to MongoDB ObjectIDs
	ToObjectIDs = helpers.ToObjectIDs
	// IsValidObjectID checks if a string is a valid MongoDB ObjectID
	IsValidObjectID = helpers.IsValidObjectID
	// NewObjectID generates a new MongoDB ObjectID
	NewObjectID = helpers.NewObjectID
	// ToBSON converts a map to BSON document
	ToBSON = helpers.ToBSON
	// ToBSONArray converts a slice of maps to BSON array
	ToBSONArray = helpers.ToBSONArray
	// MergeBSON merges multiple BSON documents
	MergeBSON = helpers.MergeBSON
	// BSONToMap converts a BSON document to a map
	BSONToMap = helpers.BSONToMap
)

// Re-export MongoDB driver sentinel errors for use with errors.Is()
var (
	// ErrNoDocuments is returned when a query that expects a document doesn't find one
	ErrNoDocuments = mongo.ErrNoDocuments
	// ErrNilDocument is returned when a nil document is passed to an insert operation
	ErrNilDocument = mongo.ErrNilDocument
	// ErrNilValue is returned when a nil value is passed where it's not allowed
	ErrNilValue = mongo.ErrNilValue
	// ErrEmptySlice is returned when an empty slice is passed where it's not allowed
	ErrEmptySlice = mongo.ErrEmptySlice
	// ErrUnacknowledgedWrite is returned when attempting to get results from an unacknowledged write
	ErrUnacknowledgedWrite = mongo.ErrUnacknowledgedWrite
	// ErrClientDisconnected is returned when an operation is attempted on a disconnected client
	ErrClientDisconnected = mongo.ErrClientDisconnected
	// ErrInvalidIndexValue is returned when an invalid index value is encountered
	ErrInvalidIndexValue = mongo.ErrInvalidIndexValue
	// ErrInvalidObjectID is returned when an invalid ObjectID hex string is provided
	ErrInvalidObjectID = primitive.ErrInvalidHex
)

// Re-export MongoDB driver error helper functions
var (
	// IsDuplicateKeyError checks if an error is a duplicate key error (E11000)
	IsDuplicateKeyError = mongo.IsDuplicateKeyError
	// IsTimeout checks if an error is a timeout error
	IsTimeout = mongo.IsTimeout
	// IsNetworkError checks if an error is a network error
	IsNetworkError = mongo.IsNetworkError
)

// Options constructors for convenience
var (
	// Index creates a new IndexOptions instance for configuring index creation
	Index = options.Index
	// Find creates a new FindOptions instance for configuring find operations
	Find = options.Find
	// FindOne creates a new FindOneOptions instance for configuring findOne operations
	FindOne = options.FindOne
	// Update creates a new UpdateOptions instance for configuring update operations
	Update = options.Update
	// Replace creates a new ReplaceOptions instance for configuring replace operations
	Replace = options.Replace
	// Delete creates a new DeleteOptions instance for configuring delete operations
	Delete = options.Delete
	// InsertOne creates a new InsertOneOptions instance for configuring insertOne operations
	InsertOne = options.InsertOne
	// InsertMany creates a new InsertManyOptions instance for configuring insertMany operations
	InsertMany = options.InsertMany
	// CreateCollection creates a new CreateCollectionOptions instance for configuring collection creation
	CreateCollection = options.CreateCollection
	// Aggregate creates a new AggregateOptions instance for configuring aggregate operations
	Aggregate = options.Aggregate
	// Count creates a new CountOptions instance for configuring count operations
	Count = options.Count
	// Distinct creates a new DistinctOptions instance for configuring distinct operations
	Distinct = options.Distinct
	// FindOneAndUpdate creates a new FindOneAndUpdateOptions instance
	FindOneAndUpdate = options.FindOneAndUpdate
	// FindOneAndReplace creates a new FindOneAndReplaceOptions instance
	FindOneAndReplace = options.FindOneAndReplace
	// FindOneAndDelete creates a new FindOneAndDeleteOptions instance
	FindOneAndDelete = options.FindOneAndDelete
	// BulkWrite creates a new BulkWriteOptions instance for configuring bulk write operations
	BulkWrite = options.BulkWrite
)
