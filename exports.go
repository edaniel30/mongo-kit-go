package mongo

import (
	"github.com/edaniel30/mongo-kit-go/errors"
	"github.com/edaniel30/mongo-kit-go/internal/helpers"
	"github.com/edaniel30/mongo-kit-go/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	// Config holds MongoDB client configuration
	Config = models.Config
	// Option is a function that modifies Config
	Option = models.Option
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

// Re-export error functions
var (
	// ErrInvalidConfig returns a configuration error with the given message
	ErrInvalidConfig = errors.ErrInvalidConfig
	// ErrConnectionFailed returns an error indicating connection failure
	ErrConnectionFailed = errors.ErrConnectionFailed
	// ErrClientClosed returns an error indicating the client is closed
	ErrClientClosed = errors.ErrClientClosed
	// ErrInvalidOperation returns an error indicating an invalid operation
	ErrInvalidOperation = errors.ErrInvalidOperation
	// ErrDocumentNotFound returns an error indicating document was not found
	ErrDocumentNotFound = errors.ErrDocumentNotFound
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

// MongoDB primitive types for convenience
type ObjectID = primitive.ObjectID
