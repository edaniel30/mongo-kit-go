// Package mongo_kit provides a MongoDB client wrapper with convenient methods
// for connection management and database operations.
package mongo_kit

import (
	"context"
	"fmt"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client wraps the MongoDB driver client with convenience methods.
// It provides a simpler API for common operations while maintaining thread-safety.
// The client is safe for concurrent use across multiple goroutines.
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

// Ping verifies the connection to MongoDB.
// Returns an error if the client is closed or the connection check fails.
//
// Example:
//
//	if err := client.Ping(ctx); err != nil {
//	    log.Println("Connection lost:", err)
//	}
func (c *Client) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	return c.client.Ping(ctx, nil)
}

// IsConnected checks if the client is connected to MongoDB.
// This is a convenience method that calls Ping and returns true if successful.
//
// Example:
//
//	if !client.IsConnected(ctx) {
//	    log.Println("Not connected to MongoDB")
//	}
func (c *Client) IsConnected(ctx context.Context) bool {
	return c.Ping(ctx) == nil
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

// IsClosed returns true if the client has been closed.
//
// Example:
//
//	if client.IsClosed() {
//	    log.Println("Client is closed")
//	}
func (c *Client) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// StartSession starts a new session for transaction support.
// Returns an error if the client is closed.
//
// Example:
//
//	session, err := client.StartSession()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer session.EndSession(context.Background())
func (c *Client) StartSession(opts ...*options.SessionOptions) (mongo.Session, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	return c.client.StartSession(opts...)
}

// UseSession executes a function within a session.
// This is useful for running operations that need to be part of the same session.
//
// Example:
//
//	err := client.UseSession(ctx, func(sessCtx mongo.SessionContext) error {
//	    // Your operations here
//	    return nil
//	})
func (c *Client) UseSession(ctx context.Context, fn func(mongo.SessionContext) error) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	return c.client.UseSession(ctx, fn)
}

// GetDatabase returns a handle to the specified database.
// If name is empty, returns the default database from config (cached, no lock needed).
//
// Example:
//
//	db := client.GetDatabase("myapp")
//	db := client.GetDatabase("") // uses default from config
func (c *Client) GetDatabase(name string) *mongo.Database {
	if name == "" {
		return c.defaultDB
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.client.Database(name)
}

// GetCollection returns a handle to the specified collection in the default database.
// This method does not acquire locks and is safe to call from within locked contexts.
//
// Example:
//
//	coll := client.GetCollection("users")
func (c *Client) GetCollection(collectionName string) *mongo.Collection {
	return c.defaultDB.Collection(collectionName)
}

// GetCollectionFrom returns a handle to the specified collection in the specified database.
//
// Example:
//
//	coll := client.GetCollectionFrom("analytics", "events")
func (c *Client) GetCollectionFrom(databaseName, collectionName string) *mongo.Collection {
	return c.GetDatabase(databaseName).Collection(collectionName)
}

// GetConfig returns a copy of the client configuration.
//
// Example:
//
//	cfg := client.GetConfig()
//	fmt.Println("Connected to:", cfg.URI)
func (c *Client) GetConfig() Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
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
