package errors

import (
	"fmt"
)

// configError represents a configuration validation error
type configError struct {
	message string
}

func (e *configError) Error() string {
	return fmt.Sprintf("config error: %s", e.message)
}

func ErrInvalidConfig(message string) error {
	return &configError{message: message}
}

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

func NewOperationError(operation string, cause error) error {
	return &operationError{
		operation: operation,
		cause:     cause,
	}
}

func ErrConnectionFailed(cause error) error {
	return &operationError{
		operation: "connect",
		cause:     cause,
	}
}

func ErrClientClosed() error {
	return &operationError{
		operation: "client operation",
		cause:     fmt.Errorf("client is closed"),
	}
}

func ErrInvalidOperation(message string) error {
	return &operationError{
		operation: "validation",
		cause:     fmt.Errorf("%s", message),
	}
}

func ErrDatabaseNotFound(dbName string) error {
	return &operationError{
		operation: "database access",
		cause:     fmt.Errorf("database '%s' not found", dbName),
	}
}

func ErrCollectionNotFound(collName string) error {
	return &operationError{
		operation: "collection access",
		cause:     fmt.Errorf("collection '%s' not found", collName),
	}
}
