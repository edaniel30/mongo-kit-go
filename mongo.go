// Package mongo provides a MongoDB client wrapper with convenient methods
// for connection management and database operations.
//
// Key features:
//   - Functional options pattern for configuration
//   - Thread-safe operations with connection pooling
//   - Context support for all operations
//   - Generic CRUD operations with dynamic queries
//   - Graceful shutdown with proper cleanup
//
// Basic usage:
//
//	client, err := mongo.New(
//	    mongo.DefaultConfig(),
//	    mongo.WithURI("mongodb://localhost:27017"),
//	    mongo.WithDatabase("myapp"),
//	)
//	if err != nil {
//	    panic(err)
//	}
//	defer client.Close(context.Background())
package mongo

import (
	"context"
	"fmt"
	"sync"

	"github.com/edaniel30/mongo-kit-go/errors"
	"github.com/edaniel30/mongo-kit-go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Client wraps MongoDB client with convenience methods
// It is safe for concurrent use across goroutines
type Client struct {
	config models.Config
	client *mongo.Client
	mu     sync.RWMutex
	closed bool
}

// New creates a new MongoDB client with the given configuration
// Configuration uses the functional options pattern for flexibility
func New(cfg models.Config, opts ...models.Option) (*Client, error) {
	// Apply functional options
	for _, opt := range opts {
		opt(&cfg)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Build client options
	clientOpts := options.Client().ApplyURI(cfg.URI)

	// Connection pool settings
	clientOpts.SetMaxPoolSize(cfg.MaxPoolSize)
	clientOpts.SetMinPoolSize(cfg.MinPoolSize)

	// Timeout settings
	clientOpts.SetConnectTimeout(cfg.ConnectTimeout)
	clientOpts.SetServerSelectionTimeout(cfg.ServerSelectionTimeout)
	clientOpts.SetSocketTimeout(cfg.SocketTimeout)

	// Retry settings
	clientOpts.SetRetryWrites(cfg.RetryWrites)
	clientOpts.SetRetryReads(cfg.RetryReads)

	// Additional settings
	if cfg.AppName != "" {
		clientOpts.SetAppName(cfg.AppName)
	}

	if cfg.DirectConnection {
		clientOpts.SetDirect(cfg.DirectConnection)
	}

	if cfg.ReplicaSet != "" {
		clientOpts.SetReplicaSet(cfg.ReplicaSet)
	}

	if cfg.MaxConnIdleTime > 0 {
		clientOpts.SetMaxConnIdleTime(cfg.MaxConnIdleTime)
	}

	if cfg.HeartbeatInterval > 0 {
		clientOpts.SetHeartbeatInterval(cfg.HeartbeatInterval)
	}

	// Set read preference
	readPref, err := parseReadPreference(cfg.ReadPreference)
	if err != nil {
		return nil, errors.ErrInvalidConfig(fmt.Sprintf("invalid read preference: %v", err))
	}
	clientOpts.SetReadPreference(readPref)

	// Create MongoDB client
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, errors.ErrConnectionFailed(err)
	}

	// Verify connection with ping
	ctx, cancel = context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := mongoClient.Ping(ctx, readPref); err != nil {
		_ = mongoClient.Disconnect(context.Background())
		return nil, errors.ErrConnectionFailed(err)
	}

	return &Client{
		config: cfg,
		client: mongoClient,
		closed: false,
	}, nil
}

// GetDatabase returns a handle to the specified database
// If name is empty, returns the default database from config
func (c *Client) GetDatabase(name string) *mongo.Database {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		panic("mongo: client is closed")
	}

	dbName := name
	if dbName == "" {
		dbName = c.config.Database
	}

	return c.client.Database(dbName)
}

// GetCollection returns a handle to the specified collection in the default database
func (c *Client) GetCollection(collectionName string) *mongo.Collection {
	return c.GetDatabase("").Collection(collectionName)
}

// GetCollectionFrom returns a handle to the specified collection in the specified database
func (c *Client) GetCollectionFrom(databaseName, collectionName string) *mongo.Collection {
	return c.GetDatabase(databaseName).Collection(collectionName)
}

// Ping verifies the connection to MongoDB
// Returns an error if connection verification fails
func (c *Client) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return errors.ErrClientClosed()
	}

	readPref, err := parseReadPreference(c.config.ReadPreference)
	if err != nil {
		return errors.NewOperationError("ping", err)
	}

	return c.client.Ping(ctx, readPref)
}

// IsConnected checks if the client is connected to MongoDB
func (c *Client) IsConnected(ctx context.Context) bool {
	return c.Ping(ctx) == nil
}

// IsClosed returns true if the client has been closed
func (c *Client) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// Close closes the MongoDB client connection gracefully
// After calling Close, the client should not be used
func (c *Client) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return c.client.Disconnect(ctx)
}

// StartSession starts a new session for transaction support
func (c *Client) StartSession(opts ...*options.SessionOptions) (mongo.Session, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, errors.ErrClientClosed()
	}

	return c.client.StartSession(opts...)
}

// UseSession executes a function within a session
func (c *Client) UseSession(ctx context.Context, fn func(mongo.SessionContext) error) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return errors.ErrClientClosed()
	}

	return c.client.UseSession(ctx, fn)
}

// ListDatabases lists all databases on the MongoDB server
func (c *Client) ListDatabases(ctx context.Context, filter interface{}) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, errors.ErrClientClosed()
	}

	result, err := c.client.ListDatabases(ctx, filter)
	if err != nil {
		return nil, errors.NewOperationError("list databases", err)
	}

	databases := make([]string, len(result.Databases))
	for i, db := range result.Databases {
		databases[i] = db.Name
	}

	return databases, nil
}

// ListCollections lists all collections in the specified database
func (c *Client) ListCollections(ctx context.Context, databaseName string) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, errors.ErrClientClosed()
	}

	dbName := databaseName
	if dbName == "" {
		dbName = c.config.Database
	}

	db := c.client.Database(dbName)
	cursor, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return nil, errors.NewOperationError("list collections", err)
	}

	return cursor, nil
}

// GetConfig returns a copy of the client configuration
func (c *Client) GetConfig() models.Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// parseReadPreference converts string preference to readpref.ReadPref
func parseReadPreference(preference string) (*readpref.ReadPref, error) {
	switch preference {
	case "primary":
		return readpref.Primary(), nil
	case "primaryPreferred":
		return readpref.PrimaryPreferred(), nil
	case "secondary":
		return readpref.Secondary(), nil
	case "secondaryPreferred":
		return readpref.SecondaryPreferred(), nil
	case "nearest":
		return readpref.Nearest(), nil
	default:
		return nil, fmt.Errorf("unknown read preference: %s", preference)
	}
}
