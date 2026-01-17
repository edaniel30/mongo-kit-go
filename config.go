package mongo_kit

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config holds the MongoDB client configuration.
// Use DefaultConfig() to get sensible defaults, then customize with Option functions.
type Config struct {
	URI      string // MongoDB connection URI (required)
	Database string // Default database name (required)

	MaxPoolSize   uint64                 // Maximum number of connections in the connection pool (default: 100)
	Timeout       time.Duration          // Default timeout for all operations (default: 10s)
	ClientOptions *options.ClientOptions // Direct access to MongoDB driver options for advanced use cases
}

// DefaultConfig returns a Config with sensible default values.
// This is the recommended starting point for most applications.
//
// Default values:
//   - URI: "mongodb://localhost:27017"
//   - Database: "default"
//   - MaxPoolSize: 100
//   - Timeout: 10 seconds
//
// Example:
//
//	cfg := mongo_kit.DefaultConfig()
//	client, err := mongo_kit.New(cfg, mongo_kit.WithURI("mongodb://prod:27017"))
func DefaultConfig() Config {
	return Config{
		URI:         "mongodb://localhost:27017",
		Database:    "default",
		MaxPoolSize: 100,
		Timeout:     10 * time.Second,
	}
}

// Option is a function that modifies a Config.
// Use Option functions with New() to customize the configuration.
type Option func(*Config)

// WithURI sets the MongoDB connection URI.
//
// Example:
//
//	mongo_kit.WithURI("mongodb://user:pass@localhost:27017")
//	mongo_kit.WithURI("mongodb+srv://cluster.mongodb.net")
func WithURI(uri string) Option {
	return func(c *Config) {
		c.URI = uri
	}
}

// WithDatabase sets the default database name.
// All operations without an explicit database will use this database.
//
// Example:
//
//	mongo_kit.WithDatabase("myapp")
func WithDatabase(database string) Option {
	return func(c *Config) {
		c.Database = database
	}
}

// WithMaxPoolSize sets the maximum number of connections in the connection pool.
// Default is 100. Increase for high-concurrency applications.
//
// Example:
//
//	mongo_kit.WithMaxPoolSize(200)
func WithMaxPoolSize(size uint64) Option {
	return func(c *Config) {
		c.MaxPoolSize = size
	}
}

// WithTimeout sets the default timeout for all database operations.
// This timeout applies to operations that don't specify their own context timeout.
// Default is 10 seconds.
//
// Example:
//
//	mongo_kit.WithTimeout(30 * time.Second)
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithClientOptions allows you to directly configure the underlying MongoDB driver options.
// This is an escape hatch for advanced configurations not covered by the basic options.
//
// Use this when you need fine-grained control over:
//   - Minimum pool size
//   - Retry behavior
//   - TLS/SSL settings
//   - Compression
//   - Read preferences
//   - Write concerns
//   - And more...
//
// Example:
//
//	clientOpts := options.Client()
//	clientOpts.SetMinPoolSize(10)
//	clientOpts.SetRetryWrites(true)
//	clientOpts.SetReadPreference(readpref.Secondary())
//	mongo_kit.WithClientOptions(clientOpts)
//
// Note: Settings applied via WithClientOptions will override basic options like MaxPoolSize.
func WithClientOptions(opts *options.ClientOptions) Option {
	return func(c *Config) {
		c.ClientOptions = opts
	}
}

// Validate checks if the configuration is valid.
// Returns a ConfigError if any required field is missing or invalid.
func (c *Config) validate() error {
	if c.URI == "" {
		return newConfigFieldError("URI", "is required")
	}

	if c.Database == "" {
		return newConfigFieldError("Database", "is required")
	}

	if c.MaxPoolSize == 0 {
		return newConfigFieldError("MaxPoolSize", "must be greater than 0")
	}

	if c.Timeout <= 0 {
		return newConfigFieldError("Timeout", "must be greater than 0")
	}

	return nil
}
