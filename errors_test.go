package mongo_kit

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ConfigError
		expected string
	}{
		{
			name:     "with field",
			err:      &ConfigError{Field: "URI", Message: "is required"},
			expected: "mongo: config error [URI]: is required",
		},
		{
			name:     "without field",
			err:      &ConfigError{Field: "", Message: "invalid configuration"},
			expected: "mongo: config error: invalid configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestConnectionError(t *testing.T) {
	t.Run("formatting and unwrap", func(t *testing.T) {
		cause := errors.New("connection refused")
		err := &ConnectionError{Cause: cause}

		assert.Equal(t, "mongo: connection failed: connection refused", err.Error())
		assert.Equal(t, cause, err.Unwrap())
		assert.True(t, errors.Is(err, cause))
	})

	t.Run("nil cause", func(t *testing.T) {
		err := &ConnectionError{Cause: nil}
		assert.Equal(t, "mongo: connection failed: <nil>", err.Error())
	})

	t.Run("wrapped error chain", func(t *testing.T) {
		innerErr := errors.New("network unreachable")
		wrappedErr := fmt.Errorf("dial failed: %w", innerErr)
		err := &ConnectionError{Cause: wrappedErr}

		assert.True(t, errors.Is(err, innerErr))
	})
}

func TestOperationError(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		cause    error
		expected string
	}{
		{
			name:     "with cause",
			op:       "find",
			cause:    errors.New("document not found"),
			expected: "mongo: operation 'find' failed: document not found",
		},
		{
			name:     "without cause",
			op:       "insert",
			cause:    nil,
			expected: "mongo: operation 'insert' failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &OperationError{Op: tt.op, Cause: tt.cause}
			assert.Equal(t, tt.expected, err.Error())
			assert.Equal(t, tt.cause, err.Unwrap())
		})
	}
}

func TestErrClientClosed(t *testing.T) {
	assert.Equal(t, "mongo: client is closed", ErrClientClosed.Error())
	assert.True(t, errors.Is(ErrClientClosed, ErrClientClosed))

	wrapped := fmt.Errorf("operation failed: %w", ErrClientClosed)
	assert.True(t, errors.Is(wrapped, ErrClientClosed))
}

func TestErrorConstructors(t *testing.T) {
	t.Run("newConfigFieldError", func(t *testing.T) {
		err := newConfigFieldError("Database", "cannot be empty")
		var configErr *ConfigError
		require.ErrorAs(t, err, &configErr)
		assert.Equal(t, "Database", configErr.Field)
		assert.Equal(t, "cannot be empty", configErr.Message)
	})

	t.Run("newConnectionError", func(t *testing.T) {
		cause := errors.New("dial tcp: connection refused")
		err := newConnectionError(cause)
		var connErr *ConnectionError
		require.ErrorAs(t, err, &connErr)
		assert.Equal(t, cause, connErr.Cause)
	})

	t.Run("newOperationError", func(t *testing.T) {
		cause := errors.New("no documents in result")
		err := newOperationError("findOne", cause)
		var opErr *OperationError
		require.ErrorAs(t, err, &opErr)
		assert.Equal(t, "findOne", opErr.Op)
		assert.Equal(t, cause, opErr.Cause)
	})
}

func TestErrorsAs(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ConfigError", &ConfigError{Field: "test", Message: "msg"}},
		{"ConnectionError", &ConnectionError{Cause: errors.New("test")}},
		{"OperationError", &OperationError{Op: "find", Cause: errors.New("test")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "ConfigError":
				var target *ConfigError
				assert.True(t, errors.As(tt.err, &target))
			case "ConnectionError":
				var target *ConnectionError
				assert.True(t, errors.As(tt.err, &target))
			case "OperationError":
				var target *OperationError
				assert.True(t, errors.As(tt.err, &target))
			}
		})
	}
}
