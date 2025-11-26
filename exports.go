package mongo

import (
	"github.com/edaniel30/mongo-kit-go/errors"
	"github.com/edaniel30/mongo-kit-go/internal/helpers"
	"github.com/edaniel30/mongo-kit-go/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Re-export types from models package
type (
	Config = models.Config
	Option = models.Option
)

// Re-export configuration functions
var (
	DefaultConfig = models.DefaultConfig
)

// Re-export configuration option functions
var (
	WithURI                    = models.WithURI
	WithDatabase               = models.WithDatabase
	WithMaxPoolSize            = models.WithMaxPoolSize
	WithMinPoolSize            = models.WithMinPoolSize
	WithConnectTimeout         = models.WithConnectTimeout
	WithServerSelectionTimeout = models.WithServerSelectionTimeout
	WithSocketTimeout          = models.WithSocketTimeout
	WithTimeout                = models.WithTimeout
	WithDebug                  = models.WithDebug
	WithRetryWrites            = models.WithRetryWrites
	WithRetryReads             = models.WithRetryReads
	WithAppName                = models.WithAppName
	WithDirectConnection       = models.WithDirectConnection
	WithReplicaSet             = models.WithReplicaSet
	WithReadPreference         = models.WithReadPreference
	WithMaxConnIdleTime        = models.WithMaxConnIdleTime
	WithHeartbeatInterval      = models.WithHeartbeatInterval
)

// Re-export error functions
var (
	ErrInvalidConfig      = errors.ErrInvalidConfig
	ErrConnectionFailed   = errors.ErrConnectionFailed
	ErrClientClosed       = errors.ErrClientClosed
	ErrInvalidOperation   = errors.ErrInvalidOperation
	ErrDocumentNotFound   = errors.ErrDocumentNotFound
	ErrDatabaseNotFound   = errors.ErrDatabaseNotFound
	ErrCollectionNotFound = errors.ErrCollectionNotFound
	NewOperationError     = errors.NewOperationError
)

// Re-export helper functions
var (
	ToObjectID      = helpers.ToObjectID
	ToObjectIDs     = helpers.ToObjectIDs
	IsValidObjectID = helpers.IsValidObjectID
	NewObjectID     = helpers.NewObjectID
	ToBSON          = helpers.ToBSON
	ToBSONArray     = helpers.ToBSONArray
	MergeBSON       = helpers.MergeBSON
)

// Re-export MongoDB primitive types for convenience
type ObjectID = primitive.ObjectID
