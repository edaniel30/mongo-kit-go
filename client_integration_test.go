package mongo_kit

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"

	testhelpers "github.com/edaniel30/mongo-kit-go/testing"
)

func TestClient_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	container := testhelpers.SetupMongoContainer(t)
	defer container.Teardown(t)

	t.Run("New creates client and connects", func(t *testing.T) {
		cfg := DefaultConfig()
		WithURI(container.URI)(&cfg)
		WithDatabase("testdb")(&cfg)

		client, err := New(cfg)
		require.NoError(t, err)
		require.NotNil(t, client)
		defer client.Close(context.Background())

		assert.False(t, client.IsClosed())
	})

	t.Run("New fails with invalid URI", func(t *testing.T) {
		cfg := DefaultConfig()
		WithURI("mongodb://invalid:12345")(&cfg)
		WithTimeout(2 * time.Second)(&cfg)

		_, err := New(cfg)
		require.Error(t, err)

		var connErr *ConnectionError
		assert.ErrorAs(t, err, &connErr)
	})

	t.Run("Ping succeeds on connected client", func(t *testing.T) {
		cfg := DefaultConfig()
		WithURI(container.URI)(&cfg)
		WithDatabase("testdb")(&cfg)

		client, err := New(cfg)
		require.NoError(t, err)
		defer client.Close(context.Background())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = client.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("IsConnected returns true when connected", func(t *testing.T) {
		cfg := DefaultConfig()
		WithURI(container.URI)(&cfg)
		WithDatabase("testdb")(&cfg)

		client, err := New(cfg)
		require.NoError(t, err)
		defer client.Close(context.Background())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		assert.True(t, client.IsConnected(ctx))
	})

	t.Run("Close closes client", func(t *testing.T) {
		cfg := DefaultConfig()
		WithURI(container.URI)(&cfg)
		WithDatabase("testdb")(&cfg)

		client, err := New(cfg)
		require.NoError(t, err)

		err = client.Close(context.Background())
		require.NoError(t, err)

		assert.True(t, client.IsClosed())
	})

	t.Run("Close is idempotent", func(t *testing.T) {
		cfg := DefaultConfig()
		WithURI(container.URI)(&cfg)
		WithDatabase("testdb")(&cfg)

		client, err := New(cfg)
		require.NoError(t, err)

		err = client.Close(context.Background())
		require.NoError(t, err)

		err = client.Close(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Operations fail on closed client", func(t *testing.T) {
		cfg := DefaultConfig()
		WithURI(container.URI)(&cfg)
		WithDatabase("testdb")(&cfg)

		client, err := New(cfg)
		require.NoError(t, err)
		client.Close(context.Background())

		ctx := context.Background()
		err = client.Ping(ctx)
		assert.ErrorIs(t, err, ErrClientClosed)
	})

	t.Run("GetDatabase returns database handle", func(t *testing.T) {
		cfg := DefaultConfig()
		WithURI(container.URI)(&cfg)
		WithDatabase("testdb")(&cfg)

		client, err := New(cfg)
		require.NoError(t, err)
		defer client.Close(context.Background())

		db := client.GetDatabase("")
		assert.NotNil(t, db)
		assert.Equal(t, "testdb", db.Name())

		db2 := client.GetDatabase("otherdb")
		assert.NotNil(t, db2)
		assert.Equal(t, "otherdb", db2.Name())
	})

	t.Run("GetCollection returns collection handle", func(t *testing.T) {
		cfg := DefaultConfig()
		WithURI(container.URI)(&cfg)
		WithDatabase("testdb")(&cfg)

		client, err := New(cfg)
		require.NoError(t, err)
		defer client.Close(context.Background())

		coll := client.GetCollection("users")
		assert.NotNil(t, coll)
		assert.Equal(t, "users", coll.Name())
	})

	t.Run("GetCollectionFrom returns collection from specific database", func(t *testing.T) {
		cfg := DefaultConfig()
		WithURI(container.URI)(&cfg)
		WithDatabase("testdb")(&cfg)

		client, err := New(cfg)
		require.NoError(t, err)
		defer client.Close(context.Background())

		coll := client.GetCollectionFrom("otherdb", "events")
		assert.NotNil(t, coll)
		assert.Equal(t, "events", coll.Name())
	})

	t.Run("GetConfig returns config copy", func(t *testing.T) {
		cfg := DefaultConfig()
		WithURI(container.URI)(&cfg)
		WithDatabase("testdb")(&cfg)
		WithMaxPoolSize(50)(&cfg)

		client, err := New(cfg)
		require.NoError(t, err)
		defer client.Close(context.Background())

		returnedCfg := client.GetConfig()
		assert.Equal(t, "testdb", returnedCfg.Database)
		assert.Equal(t, uint64(50), returnedCfg.MaxPoolSize)
	})

	t.Run("Thread safety with concurrent operations", func(t *testing.T) {
		cfg := DefaultConfig()
		WithURI(container.URI)(&cfg)
		WithDatabase("testdb")(&cfg)

		client, err := New(cfg)
		require.NoError(t, err)
		defer client.Close(context.Background())

		var wg sync.WaitGroup
		errors := make(chan error, 100)

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := client.Ping(ctx); err != nil {
					errors <- err
				}
			}()
		}

		wg.Wait()
		close(errors)

		for err := range errors {
			t.Errorf("concurrent ping failed: %v", err)
		}
	})

	t.Run("StartSession creates session", func(t *testing.T) {
		cfg := DefaultConfig()
		WithURI(container.URI)(&cfg)
		WithDatabase("testdb")(&cfg)

		client, err := New(cfg)
		require.NoError(t, err)
		defer client.Close(context.Background())

		session, err := client.StartSession()
		require.NoError(t, err)
		require.NotNil(t, session)
		defer session.EndSession(context.Background())
	})

	t.Run("StartSession fails on closed client", func(t *testing.T) {
		cfg := DefaultConfig()
		WithURI(container.URI)(&cfg)
		WithDatabase("testdb")(&cfg)

		client, err := New(cfg)
		require.NoError(t, err)
		client.Close(context.Background())

		_, err = client.StartSession()
		assert.ErrorIs(t, err, ErrClientClosed)
	})

	t.Run("UseSession executes function", func(t *testing.T) {
		cfg := DefaultConfig()
		WithURI(container.URI)(&cfg)
		WithDatabase("testdb")(&cfg)

		client, err := New(cfg)
		require.NoError(t, err)
		defer client.Close(context.Background())

		called := false
		err = client.UseSession(context.Background(), func(sc mongo.SessionContext) error {
			called = true
			return nil
		})

		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("UseSession fails on closed client", func(t *testing.T) {
		cfg := DefaultConfig()
		WithURI(container.URI)(&cfg)
		WithDatabase("testdb")(&cfg)

		client, err := New(cfg)
		require.NoError(t, err)
		client.Close(context.Background())

		err = client.UseSession(context.Background(), func(sc mongo.SessionContext) error {
			return nil
		})

		assert.ErrorIs(t, err, ErrClientClosed)
	})
}
