package mongo_kit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient creates a Client for testing context helpers without MongoDB connection.
func newTestClient(timeout time.Duration) *Client {
	return &Client{
		config: Config{Timeout: timeout},
	}
}

func TestClient_NewContext(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{"5 second timeout", 5 * time.Second},
		{"1 second timeout", 1 * time.Second},
		{"30 second timeout", 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(tt.timeout)
			ctx, cancel := client.NewContext()
			defer cancel()

			deadline, hasDeadline := ctx.Deadline()
			require.True(t, hasDeadline)

			expectedDeadline := time.Now().Add(tt.timeout)
			assert.WithinDuration(t, expectedDeadline, deadline, 100*time.Millisecond)
		})
	}

	t.Run("cancel function works", func(t *testing.T) {
		client := newTestClient(5 * time.Second)
		ctx, cancel := client.NewContext()

		cancel()

		select {
		case <-ctx.Done():
			assert.Equal(t, context.Canceled, ctx.Err())
		default:
			t.Fatal("context should be canceled")
		}
	})

	t.Run("timeout expires", func(t *testing.T) {
		client := newTestClient(50 * time.Millisecond)
		ctx, cancel := client.NewContext()
		defer cancel()

		select {
		case <-ctx.Done():
			assert.Equal(t, context.DeadlineExceeded, ctx.Err())
		case <-time.After(200 * time.Millisecond):
			t.Fatal("context should have expired")
		}
	})
}

func TestClient_WithTimeout(t *testing.T) {
	t.Run("creates child context with timeout", func(t *testing.T) {
		timeout := 5 * time.Second
		client := newTestClient(timeout)

		ctx, cancel := client.WithTimeout(context.Background())
		defer cancel()

		deadline, hasDeadline := ctx.Deadline()
		require.True(t, hasDeadline)

		expectedDeadline := time.Now().Add(timeout)
		assert.WithinDuration(t, expectedDeadline, deadline, 100*time.Millisecond)
	})

	t.Run("preserves parent context values", func(t *testing.T) {
		client := newTestClient(5 * time.Second)

		type ctxKey string
		key := ctxKey("traceID")
		parent := context.WithValue(context.Background(), key, "trace-123")

		ctx, cancel := client.WithTimeout(parent)
		defer cancel()

		assert.Equal(t, "trace-123", ctx.Value(key))
	})

	t.Run("canceled when parent canceled", func(t *testing.T) {
		client := newTestClient(5 * time.Second)
		parent, parentCancel := context.WithCancel(context.Background())

		ctx, cancel := client.WithTimeout(parent)
		defer cancel()

		parentCancel()

		select {
		case <-ctx.Done():
			assert.Error(t, ctx.Err())
		case <-time.After(100 * time.Millisecond):
			t.Fatal("child context should be canceled when parent is canceled")
		}
	})
}

func TestClient_EnsureTimeout(t *testing.T) {
	t.Run("returns unchanged when deadline exists", func(t *testing.T) {
		client := newTestClient(5 * time.Second)

		parent, parentCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer parentCancel()
		parentDeadline, _ := parent.Deadline()

		ctx, cancel := client.EnsureTimeout(parent)
		defer cancel()

		deadline, _ := ctx.Deadline()
		assert.Equal(t, parentDeadline, deadline)
	})

	t.Run("adds timeout when no deadline", func(t *testing.T) {
		timeout := 5 * time.Second
		client := newTestClient(timeout)

		ctx, cancel := client.EnsureTimeout(context.Background())
		defer cancel()

		deadline, hasDeadline := ctx.Deadline()
		require.True(t, hasDeadline)

		expectedDeadline := time.Now().Add(timeout)
		assert.WithinDuration(t, expectedDeadline, deadline, 100*time.Millisecond)
	})

	t.Run("no-op cancel when deadline exists", func(t *testing.T) {
		client := newTestClient(5 * time.Second)

		parent, parentCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer parentCancel()

		ctx, cancel := client.EnsureTimeout(parent)
		cancel() // Should be no-op

		select {
		case <-ctx.Done():
			t.Fatal("context should not be done after no-op cancel")
		default:
			// Expected
		}
	})

	t.Run("preserves context values", func(t *testing.T) {
		client := newTestClient(5 * time.Second)

		type ctxKey string
		key := ctxKey("userID")
		parent := context.WithValue(context.Background(), key, "user-456")

		ctx, cancel := client.EnsureTimeout(parent)
		defer cancel()

		assert.Equal(t, "user-456", ctx.Value(key))
	})
}

func TestContextHelpers_Concurrent(t *testing.T) {
	client := newTestClient(5 * time.Second)
	done := make(chan bool, 300)

	// Test all three methods concurrently
	for i := 0; i < 100; i++ {
		go func() {
			ctx, cancel := client.NewContext()
			defer cancel()
			_, hasDeadline := ctx.Deadline()
			assert.True(t, hasDeadline)
			done <- true
		}()

		go func() {
			ctx, cancel := client.WithTimeout(context.Background())
			defer cancel()
			_, hasDeadline := ctx.Deadline()
			assert.True(t, hasDeadline)
			done <- true
		}()

		go func() {
			ctx, cancel := client.EnsureTimeout(context.Background())
			defer cancel()
			_, hasDeadline := ctx.Deadline()
			assert.True(t, hasDeadline)
			done <- true
		}()
	}

	for i := 0; i < 300; i++ {
		<-done
	}
}
