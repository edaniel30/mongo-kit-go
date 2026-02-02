package mongo_kit

import (
	"errors"
	"fmt"
)

// Public Error Types
// These types are exported so users can use errors.As() to inspect them
// and access their fields for better error handling.

// ConfigError represents a configuration validation error.
// Users can access the Field and Message to understand what configuration is invalid.
type ConfigError struct {
	Field   string // The configuration field that caused the error (optional)
	Message string // Human-readable error message
}

func (e *ConfigError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("mongo: config error [%s]: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("mongo: config error: %s", e.Message)
}

// ConnectionError represents a connection failure to MongoDB.
// The Cause field contains the underlying error from the MongoDB driver.
type ConnectionError struct {
	Cause error // The underlying error that caused the connection failure
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("mongo: connection failed: %v", e.Cause)
}

func (e *ConnectionError) Unwrap() error {
	return e.Cause
}

// OperationError represents an error that occurred during a database operation.
// The Op field identifies which operation failed, and Cause contains the underlying error.
type OperationError struct {
	Op    string // The name of the operation that failed (e.g., "find", "insert", "update")
	Cause error  // The underlying error from MongoDB driver
}

func (e *OperationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("mongo: operation '%s' failed: %v", e.Op, e.Cause)
	}
	return fmt.Sprintf("mongo: operation '%s' failed", e.Op)
}

func (e *OperationError) Unwrap() error {
	return e.Cause
}

// Sentinel Errors
// These are sentinel errors that can be checked using errors.Is().

var (
	// ErrClientClosed is returned when an operation is attempted on a closed client.
	// Use errors.Is(err, mongo.ErrClientClosed) to check for this error.
	ErrClientClosed = errors.New("mongo: client is closed")
)

// Internal constructor functions

// newConfigFieldError creates a configuration error with a specific field.
func newConfigFieldError(field, message string) error {
	return &ConfigError{Field: field, Message: message}
}

// newConnectionError creates a connection error wrapping an underlying cause.
func newConnectionError(cause error) error {
	return &ConnectionError{Cause: cause}
}

// newOperationError creates an operation error for a specific operation and cause.
func newOperationError(operation string, cause error) error {
	return &OperationError{Op: operation, Cause: cause}
}
