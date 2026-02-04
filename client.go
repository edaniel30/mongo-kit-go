// Package mongo_kit provides a MongoDB client wrapper with convenient methods
// for connection management and database operations.
package mongo_kit

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client wraps the MongoDB driver client with convenience methods.
// It provides a simpler API for common operations while maintaining thread-safety.
// The client is safe for concurrent use across multiple goroutines.
//
// While Client is exported for advanced use cases, users are encouraged to primarily
// interact with the database through Repository[T] for type-safe operations.
type Client struct {
	config    Config
	client    *mongo.Client
	defaultDB *mongo.Database
	mu        sync.RWMutex
	closed    bool
}

// New creates a new MongoDB client with the given configuration.
// Configuration uses the functional options pattern for flexibility.
//
// The client will automatically:
//   - Validate the configuration
//   - Connect to MongoDB
//   - Verify the connection with a ping
//
// Example:
//
//	client, err := mongo_kit.New(
//	    mongo_kit.DefaultConfig(),
//	    mongo_kit.WithURI("mongodb://localhost:27017"),
//	    mongo_kit.WithDatabase("myapp"),
//	    mongo_kit.WithMaxPoolSize(200),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close(context.Background())
//
// Returns an error if:
//   - Configuration is invalid
//   - Connection to MongoDB fails
//   - Ping verification fails
func New(cfg Config, opts ...Option) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	var clientOpts *options.ClientOptions
	// If user provided custom ClientOptions, use them as base
	if cfg.ClientOptions != nil {
		clientOpts = cfg.ClientOptions
	} else {
		clientOpts = options.Client()
	}

	clientOpts.ApplyURI(cfg.URI)
	clientOpts.SetMaxPoolSize(cfg.MaxPoolSize)
	clientOpts.SetRetryWrites(true)
	clientOpts.SetRetryReads(true)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, newConnectionError(err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := mongoClient.Ping(ctx, nil); err != nil {
		if disconnectErr := mongoClient.Disconnect(context.Background()); disconnectErr != nil {
			return nil, newConnectionError(fmt.Errorf("ping failed: %w, disconnect also failed: %w", err, disconnectErr))
		}
		return nil, newConnectionError(err)
	}

	return &Client{
		config:    cfg,
		client:    mongoClient,
		defaultDB: mongoClient.Database(cfg.Database),
		closed:    false,
	}, nil
}

// Close closes the MongoDB client connection gracefully.
// After calling Close, the client should not be used.
// Calling Close multiple times is safe and will only disconnect once.
//
// Example:
//
//	defer client.Close(context.Background())
func (c *Client) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return c.client.Disconnect(ctx)
}

// getCollection returns a handle to the specified collection in the default database.
// This method does not acquire locks and is safe to call from within locked contexts.
// This method is unexported and used internally by repositories.
func (c *Client) getCollection(collectionName string) *mongo.Collection {
	return c.defaultDB.Collection(collectionName)
}

// checkState verifies that the client is not closed.
// IMPORTANT: This method does NOT acquire any locks. The caller MUST hold c.mu.RLock()
// before calling this method.
// Returns ErrClientClosed if the client has been closed.
func (c *Client) checkState() error {
	if c.closed {
		return ErrClientClosed
	}
	return nil
}

// CreateCollection creates a new collection with optional configuration.
// If the collection already exists, this is a no-op (no error is returned).
//
// Example:
//
//	err := client.CreateCollection(ctx, "users", options.CreateCollection().SetCapped(true).SetSizeInBytes(1000000))
func (c *Client) CreateCollection(ctx context.Context, name string, opts ...*options.CreateCollectionOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return err
	}

	err := c.defaultDB.CreateCollection(ctx, name, opts...)
	if err != nil {
		// Check if collection already exists (MongoDB error code 48: NamespaceExists)
		var cmdErr mongo.CommandError
		if errors.As(err, &cmdErr) && cmdErr.Code == 48 {
			return nil
		}
		return newOperationError("create collection", err)
	}

	return nil
}

// CreateIndexes creates multiple indexes on the specified collection.
// Returns the names of the created indexes.
//
// Example:
//
//	indexes := []mongo.IndexModel{
//	    {
//	        Keys: bson.D{{Key: "email", Value: 1}},
//	        Options: options.Index().SetUnique(true),
//	    },
//	    {
//	        Keys: bson.D{{Key: "created_at", Value: -1}},
//	    },
//	}
//	names, err := client.CreateIndexes(ctx, "users", indexes)
func (c *Client) CreateIndexes(ctx context.Context, collection string, indexes []mongo.IndexModel) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkState(); err != nil {
		return nil, err
	}

	if len(indexes) == 0 {
		return nil, newOperationError("create indexes", errors.New("at least one index model must be provided"))
	}

	coll := c.getCollection(collection)
	names, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return nil, newOperationError("create indexes", err)
	}

	return names, nil
}
