package mongo_kit

import (
	"context"
)

// Context Helpers
//
// This file provides context helper methods to simplify timeout management
// for database operations.
//
// See docs/context.md for detailed usage guide and examples.

// NewContext creates a new context with the default timeout from config.
//
// Use this for standalone operations (CLI tools, scripts, background jobs).
// DO NOT use in web handlers - use the request context instead.
//
// Returns a context with timeout and a cancel function (always call defer cancel()).
func (c *Client) NewContext() (context.Context, context.CancelFunc) {
	c.mu.RLock()
	timeout := c.config.Timeout
	c.mu.RUnlock()

	return context.WithTimeout(context.Background(), timeout)
}

// WithTimeout creates a child context with timeout from an existing parent context.
//
// Use this when you have a parent context (from HTTP request, gRPC) but want to add
// a specific timeout for database operations.
//
// The resulting context will be canceled when the DB timeout expires, the parent is
// canceled, or the returned cancel function is called.
//
// Preserves all values from the parent context (trace IDs, user info, etc.).
func (c *Client) WithTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	c.mu.RLock()
	timeout := c.config.Timeout
	c.mu.RUnlock()

	return context.WithTimeout(parent, timeout)
}

// EnsureTimeout ensures the context has a deadline.
//
// If the context already has a deadline, returns it unchanged.
// If the context has no deadline, adds the default timeout from config.
//
// Use this when you're unsure if the context has a deadline or writing
// library/reusable code that accepts contexts from callers.
func (c *Client) EnsureTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	// Check if context already has a deadline
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		// Context has deadline, return it unchanged with no-op cancel
		return ctx, func() {}
	}

	// No deadline exists, add default timeout from config
	c.mu.RLock()
	timeout := c.config.Timeout
	c.mu.RUnlock()

	return context.WithTimeout(ctx, timeout)
}
