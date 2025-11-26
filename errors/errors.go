package errors

import "fmt"

// configError represents a configuration validation error
type configError struct {
	message string
}

func (e *configError) Error() string {
	return fmt.Sprintf("config error: %s", e.message)
}

// ErrInvalidConfig returns a configuration error with the given message
func ErrInvalidConfig(message string) error {
	return &configError{message: message}
}

// operationError represents a runtime operation error
type operationError struct {
	operation string
	cause     error
}

func (e *operationError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("operation failed (%s): %v", e.operation, e.cause)
	}
	return fmt.Sprintf("operation failed (%s)", e.operation)
}

func (e *operationError) Unwrap() error {
	return e.cause
}

// NewOperationError creates a new operation error
func NewOperationError(operation string, cause error) error {
	return &operationError{
		operation: operation,
		cause:     cause,
	}
}

// ErrConnectionFailed returns an error indicating connection failure
func ErrConnectionFailed(cause error) error {
	return &operationError{
		operation: "connect",
		cause:     cause,
	}
}

// ErrClientClosed returns an error indicating the client is closed
func ErrClientClosed() error {
	return &operationError{
		operation: "client operation",
		cause:     fmt.Errorf("client is closed"),
	}
}

// ErrInvalidOperation returns an error indicating an invalid operation
func ErrInvalidOperation(message string) error {
	return &operationError{
		operation: "validation",
		cause:     fmt.Errorf("%s", message),
	}
}

// ErrDocumentNotFound returns an error indicating document was not found
func ErrDocumentNotFound() error {
	return &operationError{
		operation: "find",
		cause:     fmt.Errorf("document not found"),
	}
}

// ErrDatabaseNotFound returns an error indicating database was not found
func ErrDatabaseNotFound(dbName string) error {
	return &operationError{
		operation: "database access",
		cause:     fmt.Errorf("database '%s' not found", dbName),
	}
}

// ErrCollectionNotFound returns an error indicating collection was not found
func ErrCollectionNotFound(collName string) error {
	return &operationError{
		operation: "collection access",
		cause:     fmt.Errorf("collection '%s' not found", collName),
	}
}
