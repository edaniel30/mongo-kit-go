package mongo_kit

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	testhelpers "github.com/edaniel30/mongo-kit-go/testing"
)

func TestEncryption_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	container := testhelpers.SetupMongoContainer(t)
	defer container.Teardown(t)
	ctx := context.Background()

	newClient := func(t *testing.T, database, key string) (*Client, error) {
		t.Helper()
		opts := []Option{WithURI(container.URI), WithDatabase(database)}
		if key != "" {
			opts = append(opts, WithEncryptionKey(key))
		}
		return New(DefaultConfig(), opts...)
	}

	t.Run("first client stores the check value", func(t *testing.T) {
		client, err := newClient(t, "enc_first", testKey)
		require.NoError(t, err)
		defer func() { _ = client.Close(ctx) }()

		var doc bson.M
		err = client.defaultDB.Collection(metaCollection).FindOne(ctx, bson.M{"_id": encryptionCheckID}).Decode(&doc)
		require.NoError(t, err)
		assert.NotEmpty(t, doc["value"])
		assert.NotEqual(t, encryptionCheckPlaintext, doc["value"], "stored encrypted")
	})

	t.Run("same key connects and reads what others encrypted", func(t *testing.T) {
		writer, err := newClient(t, "enc_same", testKey)
		require.NoError(t, err)
		defer func() { _ = writer.Close(ctx) }()

		enc, err := writer.Encrypt("shared secret")
		require.NoError(t, err)

		reader, err := newClient(t, "enc_same", testKey)
		require.NoError(t, err)
		defer func() { _ = reader.Close(ctx) }()

		dec, err := reader.Decrypt(enc)
		require.NoError(t, err)
		assert.Equal(t, "shared secret", dec)
	})

	t.Run("other key is rejected", func(t *testing.T) {
		first, err := newClient(t, "enc_mismatch", testKey)
		require.NoError(t, err)
		_ = first.Close(ctx)

		client, err := newClient(t, "enc_mismatch", otherTestKey)
		assert.ErrorIs(t, err, ErrEncryptionKeyMismatch)
		assert.Nil(t, client)
	})

	t.Run("the key is per database", func(t *testing.T) {
		a, err := newClient(t, "enc_db_a", testKey)
		require.NoError(t, err)
		_ = a.Close(ctx)

		b, err := newClient(t, "enc_db_b", otherTestKey)
		require.NoError(t, err)
		_ = b.Close(ctx)
	})

	t.Run("concurrent first starts agree on one value", func(t *testing.T) {
		const n = 8
		var wg sync.WaitGroup
		errs := make([]error, n)
		for i := range n {
			wg.Add(1)
			go func() {
				defer wg.Done()
				client, err := newClient(t, "enc_race", testKey)
				if err == nil {
					_ = client.Close(ctx)
				}
				errs[i] = err
			}()
		}
		wg.Wait()
		for _, err := range errs {
			assert.NoError(t, err)
		}
	})

	t.Run("without a key nothing is created", func(t *testing.T) {
		client, err := newClient(t, "enc_none", "")
		require.NoError(t, err)
		defer func() { _ = client.Close(ctx) }()

		assert.False(t, client.EncryptionEnabled())
		names, err := client.defaultDB.ListCollectionNames(ctx, bson.M{})
		require.NoError(t, err)
		assert.NotContains(t, names, metaCollection)
	})
}
