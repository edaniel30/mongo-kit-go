package models

import (
	"time"

	"github.com/edaniel30/mongo-kit-go/errors"
)

type Config struct {
	// Connection settings
	URI      string // MongoDB connection URI (required)
	Database string // Default database name (required)

	// Connection pool settings
	MaxPoolSize uint64 // Maximum number of connections in the pool
	MinPoolSize uint64 // Minimum number of connections in the pool

	// Timeout settings
	ConnectTimeout         time.Duration // Timeout for establishing connection
	ServerSelectionTimeout time.Duration // Timeout for selecting server
	SocketTimeout          time.Duration // Timeout for socket operations
	Timeout                time.Duration // Default timeout for operations

	// Additional settings
	RetryWrites       bool          // Enable automatic retry of write operations
	RetryReads        bool          // Enable automatic retry of read operations
	AppName           string        // Application name for MongoDB logs
	DirectConnection  bool          // Whether to connect directly to a single host
	ReplicaSet        string        // Replica set name (optional)
	ReadPreference    string        // Read preference mode (primary, secondary, etc.)
	MaxConnIdleTime   time.Duration // Maximum idle time for connection
	HeartbeatInterval time.Duration // Interval between server heartbeats
}

func DefaultConfig() Config {
	return Config{
		URI:                    "mongodb://localhost:27017",
		Database:               "default",
		MaxPoolSize:            100,
		MinPoolSize:            10,
		ConnectTimeout:         10 * time.Second,
		ServerSelectionTimeout: 5 * time.Second,
		SocketTimeout:          10 * time.Second,
		Timeout:                10 * time.Second,
		RetryWrites:            true,
		RetryReads:             true,
		AppName:                "",
		DirectConnection:       false,
		ReplicaSet:             "",
		ReadPreference:         "primary",
		MaxConnIdleTime:        10 * time.Minute,
		HeartbeatInterval:      10 * time.Second,
	}
}

type Option func(*Config)

func WithURI(uri string) Option {
	return func(c *Config) {
		c.URI = uri
	}
}

func WithDatabase(database string) Option {
	return func(c *Config) {
		c.Database = database
	}
}

func WithMaxPoolSize(size uint64) Option {
	return func(c *Config) {
		c.MaxPoolSize = size
	}
}

func WithMinPoolSize(size uint64) Option {
	return func(c *Config) {
		c.MinPoolSize = size
	}
}

func WithConnectTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.ConnectTimeout = timeout
	}
}

func WithServerSelectionTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.ServerSelectionTimeout = timeout
	}
}

func WithSocketTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.SocketTimeout = timeout
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

func WithRetryWrites(retry bool) Option {
	return func(c *Config) {
		c.RetryWrites = retry
	}
}

func WithRetryReads(retry bool) Option {
	return func(c *Config) {
		c.RetryReads = retry
	}
}

func WithAppName(appName string) Option {
	return func(c *Config) {
		c.AppName = appName
	}
}

func WithDirectConnection(direct bool) Option {
	return func(c *Config) {
		c.DirectConnection = direct
	}
}

func WithReplicaSet(replicaSet string) Option {
	return func(c *Config) {
		c.ReplicaSet = replicaSet
	}
}

func WithReadPreference(preference string) Option {
	return func(c *Config) {
		c.ReadPreference = preference
	}
}

func WithMaxConnIdleTime(duration time.Duration) Option {
	return func(c *Config) {
		c.MaxConnIdleTime = duration
	}
}

func WithHeartbeatInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.HeartbeatInterval = interval
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.URI == "" {
		return errors.ErrInvalidConfig("URI is required")
	}

	if c.Database == "" {
		return errors.ErrInvalidConfig("Database is required")
	}

	if c.MaxPoolSize < c.MinPoolSize {
		return errors.ErrInvalidConfig("MaxPoolSize must be greater than or equal to MinPoolSize")
	}

	if c.ConnectTimeout <= 0 {
		return errors.ErrInvalidConfig("ConnectTimeout must be greater than 0")
	}

	if c.ServerSelectionTimeout <= 0 {
		return errors.ErrInvalidConfig("ServerSelectionTimeout must be greater than 0")
	}

	if c.Timeout <= 0 {
		return errors.ErrInvalidConfig("Timeout must be greater than 0")
	}

	validPreferences := map[string]bool{
		"primary":            true,
		"primaryPreferred":   true,
		"secondary":          true,
		"secondaryPreferred": true,
		"nearest":            true,
	}
	if !validPreferences[c.ReadPreference] {
		return errors.ErrInvalidConfig("invalid read preference: must be one of [primary, primaryPreferred, secondary, secondaryPreferred, nearest]")
	}

	return nil
}
