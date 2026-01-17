# Context Helpers

The client provides three context helper methods to simplify timeout management for database operations.

## Quick Reference

| Scenario | Helper | Reason |
|----------|--------|---------|
| HTTP/gRPC Handler | Use ctx directly or `WithTimeout(ctx)` | Already has cancellation and deadline |
| CLI Tool / Script | `NewContext()` | Need new standalone context |
| Background Job | `NewContext()` | Need new standalone context |
| Nested Operation | `WithTimeout(parent)` | Add DB timeout to parent |
| Library Function | `EnsureTimeout(ctx)` | Safe, respects existing deadline |
| Uncertain | `EnsureTimeout(ctx)` | Safest option |

## Methods

### NewContext()

Creates a new context with the default timeout from config.

**Use for:** CLI tools, scripts, and background jobs where you don't have a parent context.

**Don't use in:** Web handlers (HTTP, gRPC) - you'll lose the request's cancellation signal, deadline, and propagated values.

```go
// CLI Tool
func main() {
    client, _ := mongo_kit.New(mongo_kit.DefaultConfig())
    ctx, cancel := client.NewContext()
    defer cancel()

    var users []User
    err := client.Find(ctx, "users", bson.M{"active": true}, &users)
}
```

### WithTimeout(parent)

Creates a child context with timeout from an existing parent context.

**Use when:** You have a parent context but want to add a specific timeout for database operations.

The resulting context will be canceled when:
- The database timeout expires
- The parent context is canceled
- The returned cancel function is called

```go
// HTTP Handler
func HandleGetUser(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := client.WithTimeout(r.Context())
    defer cancel()

    var user User
    err := client.FindOne(ctx, "users", bson.M{"email": email}, &user)
}
```

### EnsureTimeout(ctx)

Ensures the context has a deadline.

**Behavior:**
- If the context already has a deadline: Returns it unchanged
- If the context has no deadline: Adds the default timeout from config

**Use when:** You're unsure if the context has a deadline or writing reusable library code.

```go
// Library Function
func (repo *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
    ctx, cancel := repo.client.EnsureTimeout(ctx)
    defer cancel()

    var user User
    err := repo.client.FindOne(ctx, "users", bson.M{"email": email}, &user)
    return &user, err
}
```

## Common Mistakes

- **Don't** use `NewContext()` in HTTP handlers - loses request cancellation
- **Don't** ignore existing deadlines - use `EnsureTimeout()` when unsure
- **Don't** create contexts without cleanup - always `defer cancel()`

- **Do** use request context (`r.Context()`) in HTTP handlers
- **Do** use `WithTimeout()` to add DB-specific timeout to parent context
- **Do** use `NewContext()` only for standalone operations (CLI, scripts)
