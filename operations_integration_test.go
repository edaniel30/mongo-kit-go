package mongo_kit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	testhelpers "github.com/edaniel30/mongo-kit-go/testing"
)

type testDocument struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	Name   string             `bson:"name"`
	Age    int                `bson:"age"`
	Active bool               `bson:"active"`
	Tags   []string           `bson:"tags,omitempty"`
}

func TestOperations_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	container := testhelpers.SetupMongoContainer(t)
	defer container.Teardown(t)

	cfg := DefaultConfig()
	WithURI(container.URI)(&cfg)
	WithDatabase("testdb")(&cfg)

	client, err := New(cfg)
	require.NoError(t, err)
	defer client.Close(context.Background())

	ctx := context.Background()

	t.Run("InsertOne and FindOne", func(t *testing.T) {
		doc := testDocument{Name: "John", Age: 30, Active: true}

		result, err := client.InsertOne(ctx, "users", doc)
		require.NoError(t, err)
		require.NotNil(t, result.InsertedID)

		var found testDocument
		err = client.FindOne(ctx, "users", bson.M{"_id": result.InsertedID}, &found)
		require.NoError(t, err)
		assert.Equal(t, "John", found.Name)
		assert.Equal(t, 30, found.Age)
	})

	t.Run("InsertMany and Find", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})

		docs := []any{
			testDocument{Name: "Alice", Age: 25, Active: true},
			testDocument{Name: "Bob", Age: 35, Active: false},
			testDocument{Name: "Charlie", Age: 28, Active: true},
		}

		result, err := client.InsertMany(ctx, "users", docs)
		require.NoError(t, err)
		assert.Len(t, result.InsertedIDs, 3)

		var found []testDocument
		err = client.Find(ctx, "users", bson.M{"active": true}, &found)
		require.NoError(t, err)
		assert.Len(t, found, 2)
	})

	t.Run("FindOne returns ErrNoDocuments when not found", func(t *testing.T) {
		var found testDocument
		err := client.FindOne(ctx, "users", bson.M{"name": "nonexistent"}, &found)
		assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	})

	t.Run("UpdateOne", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})
		client.InsertOne(ctx, "users", testDocument{Name: "ToUpdate", Age: 20, Active: false})

		result, err := client.UpdateOne(ctx, "users", bson.M{"name": "ToUpdate"}, bson.M{"$set": bson.M{"age": 21}})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.ModifiedCount)

		var found testDocument
		client.FindOne(ctx, "users", bson.M{"name": "ToUpdate"}, &found)
		assert.Equal(t, 21, found.Age)
	})

	t.Run("UpdateMany", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})
		client.InsertMany(ctx, "users", []any{
			testDocument{Name: "A", Age: 20, Active: false},
			testDocument{Name: "B", Age: 25, Active: false},
		})

		result, err := client.UpdateMany(ctx, "users", bson.M{"active": false}, bson.M{"$set": bson.M{"active": true}})
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.ModifiedCount)
	})

	t.Run("ReplaceOne", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})
		insertResult, _ := client.InsertOne(ctx, "users", testDocument{Name: "Old", Age: 50, Active: true})

		newDoc := testDocument{ID: insertResult.InsertedID.(primitive.ObjectID), Name: "New", Age: 25, Active: false}
		result, err := client.ReplaceOne(ctx, "users", bson.M{"name": "Old"}, newDoc)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.ModifiedCount)

		var found testDocument
		client.FindOne(ctx, "users", bson.M{"_id": insertResult.InsertedID}, &found)
		assert.Equal(t, "New", found.Name)
		assert.Equal(t, 25, found.Age)
	})

	t.Run("DeleteOne", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})
		client.InsertOne(ctx, "users", testDocument{Name: "ToDelete", Age: 30, Active: true})

		result, err := client.DeleteOne(ctx, "users", bson.M{"name": "ToDelete"})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.DeletedCount)
	})

	t.Run("DeleteMany", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})
		client.InsertMany(ctx, "users", []any{
			testDocument{Name: "Del1", Age: 20, Active: true},
			testDocument{Name: "Del2", Age: 25, Active: true},
		})

		result, err := client.DeleteMany(ctx, "users", bson.M{"active": true})
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.DeletedCount)
	})

	t.Run("CountDocuments", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})
		client.InsertMany(ctx, "users", []any{
			testDocument{Name: "C1", Age: 20, Active: true},
			testDocument{Name: "C2", Age: 25, Active: true},
			testDocument{Name: "C3", Age: 30, Active: false},
		})

		count, err := client.CountDocuments(ctx, "users", bson.M{})
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)

		count, err = client.CountDocuments(ctx, "users", bson.M{"active": true})
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("EstimatedDocumentCount", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})
		client.InsertMany(ctx, "users", []any{
			testDocument{Name: "E1", Age: 20, Active: true},
			testDocument{Name: "E2", Age: 25, Active: true},
		})

		count, err := client.EstimatedDocumentCount(ctx, "users")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(2))
	})

	t.Run("FindByID with string", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})
		result, _ := client.InsertOne(ctx, "users", testDocument{Name: "ByID", Age: 40, Active: true})
		id := result.InsertedID.(primitive.ObjectID).Hex()

		var found testDocument
		err := client.FindByID(ctx, "users", id, &found)
		require.NoError(t, err)
		assert.Equal(t, "ByID", found.Name)
	})

	t.Run("FindByID with ObjectID", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})
		result, _ := client.InsertOne(ctx, "users", testDocument{Name: "ByOID", Age: 45, Active: true})
		id := result.InsertedID.(primitive.ObjectID)

		var found testDocument
		err := client.FindByID(ctx, "users", id, &found)
		require.NoError(t, err)
		assert.Equal(t, "ByOID", found.Name)
	})

	t.Run("FindByID with invalid string", func(t *testing.T) {
		var found testDocument
		err := client.FindByID(ctx, "users", "invalid-id", &found)
		require.Error(t, err)
	})

	t.Run("FindByID with zero ObjectID", func(t *testing.T) {
		var found testDocument
		err := client.FindByID(ctx, "users", primitive.NilObjectID, &found)
		require.Error(t, err)
	})

	t.Run("FindByID with invalid type", func(t *testing.T) {
		var found testDocument
		err := client.FindByID(ctx, "users", 12345, &found)
		require.Error(t, err)
	})

	t.Run("UpdateByID", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})
		result, _ := client.InsertOne(ctx, "users", testDocument{Name: "UpdateByID", Age: 30, Active: true})
		id := result.InsertedID.(primitive.ObjectID).Hex()

		updateResult, err := client.UpdateByID(ctx, "users", id, bson.M{"$set": bson.M{"age": 31}})
		require.NoError(t, err)
		assert.Equal(t, int64(1), updateResult.ModifiedCount)
	})

	t.Run("UpdateByID with invalid ID", func(t *testing.T) {
		_, err := client.UpdateByID(ctx, "users", "invalid", bson.M{"$set": bson.M{"age": 31}})
		require.Error(t, err)
	})

	t.Run("UpdateByID with zero ObjectID", func(t *testing.T) {
		_, err := client.UpdateByID(ctx, "users", primitive.NilObjectID, bson.M{"$set": bson.M{"age": 31}})
		require.Error(t, err)
	})

	t.Run("DeleteByID", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})
		result, _ := client.InsertOne(ctx, "users", testDocument{Name: "DeleteByID", Age: 30, Active: true})
		id := result.InsertedID.(primitive.ObjectID)

		deleteResult, err := client.DeleteByID(ctx, "users", id)
		require.NoError(t, err)
		assert.Equal(t, int64(1), deleteResult.DeletedCount)
	})

	t.Run("DeleteByID with invalid ID", func(t *testing.T) {
		_, err := client.DeleteByID(ctx, "users", "invalid")
		require.Error(t, err)
	})

	t.Run("DeleteByID with zero ObjectID", func(t *testing.T) {
		_, err := client.DeleteByID(ctx, "users", primitive.NilObjectID)
		require.Error(t, err)
	})

	t.Run("UpsertOne inserts when not exists", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})

		result, err := client.UpsertOne(ctx, "users", bson.M{"name": "Upserted"}, bson.M{"$set": bson.M{"age": 50}})
		require.NoError(t, err)
		assert.NotNil(t, result.UpsertedID)
	})

	t.Run("UpsertOne updates when exists", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})
		client.InsertOne(ctx, "users", testDocument{Name: "Existing", Age: 30, Active: true})

		result, err := client.UpsertOne(ctx, "users", bson.M{"name": "Existing"}, bson.M{"$set": bson.M{"age": 35}})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.ModifiedCount)
	})

	t.Run("Aggregate", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})
		client.InsertMany(ctx, "users", []any{
			testDocument{Name: "Agg1", Age: 25, Active: true},
			testDocument{Name: "Agg2", Age: 30, Active: true},
			testDocument{Name: "Agg3", Age: 35, Active: false},
		})

		pipeline := []bson.M{
			{"$match": bson.M{"active": true}},
			{"$group": bson.M{"_id": nil, "avgAge": bson.M{"$avg": "$age"}}},
		}

		var results []bson.M
		err := client.Aggregate(ctx, "users", pipeline, &results)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, 27.5, results[0]["avgAge"])
	})

	t.Run("Aggregate with nil pipeline", func(t *testing.T) {
		var results []bson.M
		err := client.Aggregate(ctx, "users", nil, &results)
		require.Error(t, err)
	})

	t.Run("Aggregate with invalid pipeline type", func(t *testing.T) {
		var results []bson.M
		err := client.Aggregate(ctx, "users", "invalid", &results)
		require.Error(t, err)
	})

	t.Run("Distinct", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})
		client.InsertMany(ctx, "users", []any{
			testDocument{Name: "D1", Age: 25, Active: true},
			testDocument{Name: "D2", Age: 25, Active: true},
			testDocument{Name: "D3", Age: 30, Active: false},
		})

		values, err := client.Distinct(ctx, "users", "age", bson.M{})
		require.NoError(t, err)
		assert.Len(t, values, 2)
	})

	t.Run("FindOneAndUpdate", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})
		client.InsertOne(ctx, "users", testDocument{Name: "FindUpdate", Age: 30, Active: true})

		var result testDocument
		err := client.FindOneAndUpdate(ctx, "users", bson.M{"name": "FindUpdate"}, bson.M{"$set": bson.M{"age": 31}}, &result)
		require.NoError(t, err)
		assert.Equal(t, 30, result.Age) // Returns original by default
	})

	t.Run("FindOneAndUpdate not found", func(t *testing.T) {
		var result testDocument
		err := client.FindOneAndUpdate(ctx, "users", bson.M{"name": "nonexistent"}, bson.M{"$set": bson.M{"age": 31}}, &result)
		assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	})

	t.Run("FindOneAndReplace", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})
		insertResult, _ := client.InsertOne(ctx, "users", testDocument{Name: "FindReplace", Age: 30, Active: true})

		replacement := testDocument{ID: insertResult.InsertedID.(primitive.ObjectID), Name: "Replaced", Age: 40, Active: false}
		var result testDocument
		err := client.FindOneAndReplace(ctx, "users", bson.M{"name": "FindReplace"}, replacement, &result)
		require.NoError(t, err)
		assert.Equal(t, "FindReplace", result.Name) // Returns original by default
	})

	t.Run("FindOneAndReplace not found", func(t *testing.T) {
		var result testDocument
		err := client.FindOneAndReplace(ctx, "users", bson.M{"name": "nonexistent"}, testDocument{Name: "New"}, &result)
		assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	})

	t.Run("FindOneAndDelete", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})
		client.InsertOne(ctx, "users", testDocument{Name: "FindDelete", Age: 30, Active: true})

		var result testDocument
		err := client.FindOneAndDelete(ctx, "users", bson.M{"name": "FindDelete"}, &result)
		require.NoError(t, err)
		assert.Equal(t, "FindDelete", result.Name)

		count, _ := client.CountDocuments(ctx, "users", bson.M{"name": "FindDelete"})
		assert.Equal(t, int64(0), count)
	})

	t.Run("FindOneAndDelete not found", func(t *testing.T) {
		var result testDocument
		err := client.FindOneAndDelete(ctx, "users", bson.M{"name": "nonexistent"}, &result)
		assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	})

	t.Run("BulkWrite", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})

		models := []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(testDocument{Name: "Bulk1", Age: 20, Active: true}),
			mongo.NewInsertOneModel().SetDocument(testDocument{Name: "Bulk2", Age: 25, Active: true}),
		}

		result, err := client.BulkWrite(ctx, "users", models)
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.InsertedCount)
	})

	t.Run("CreateIndex", func(t *testing.T) {
		indexName, err := client.CreateIndex(ctx, "users", bson.D{{Key: "name", Value: 1}})
		require.NoError(t, err)
		assert.NotEmpty(t, indexName)
	})

	t.Run("ListIndexes", func(t *testing.T) {
		indexes, err := client.ListIndexes(ctx, "users")
		require.NoError(t, err)
		assert.NotEmpty(t, indexes)
	})

	t.Run("DropIndex", func(t *testing.T) {
		indexName, _ := client.CreateIndex(ctx, "users", bson.D{{Key: "age", Value: 1}})

		err := client.DropIndex(ctx, "users", indexName)
		require.NoError(t, err)
	})

	t.Run("CreateIndexes", func(t *testing.T) {
		indexes := map[string][]mongo.IndexModel{
			"products": {
				{Keys: bson.D{{Key: "sku", Value: 1}}},
				{Keys: bson.D{{Key: "category", Value: 1}}},
			},
		}

		result, err := client.CreateIndexes(ctx, indexes)
		require.NoError(t, err)
		assert.Len(t, result["products"], 2)
	})

	t.Run("ListCollections", func(t *testing.T) {
		collections, err := client.ListCollections(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, collections)
	})

	t.Run("CreateCollection", func(t *testing.T) {
		collName := "test_create_" + primitive.NewObjectID().Hex()
		err := client.CreateCollection(ctx, collName)
		require.NoError(t, err)

		collections, _ := client.ListCollections(ctx)
		assert.Contains(t, collections, collName)
	})

	t.Run("CreateCollections", func(t *testing.T) {
		colls := map[string]*options.CreateCollectionOptions{
			"coll1_" + primitive.NewObjectID().Hex(): nil,
			"coll2_" + primitive.NewObjectID().Hex(): nil,
		}

		err := client.CreateCollections(ctx, colls)
		require.NoError(t, err)

		collections, _ := client.ListCollections(ctx)
		for name := range colls {
			assert.Contains(t, collections, name)
		}
	})

	t.Run("CreateCollections with nil map", func(t *testing.T) {
		err := client.CreateCollections(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("CreateIndex with options", func(t *testing.T) {
		opts := options.Index().SetUnique(true)
		indexName, err := client.CreateIndex(ctx, "idx_test_"+primitive.NewObjectID().Hex(), bson.D{{Key: "email", Value: 1}}, opts)
		require.NoError(t, err)
		assert.NotEmpty(t, indexName)
	})

	t.Run("CreateIndexes with empty models", func(t *testing.T) {
		indexes := map[string][]mongo.IndexModel{
			"empty_test": {},
		}

		result, err := client.CreateIndexes(ctx, indexes)
		require.NoError(t, err)
		assert.Empty(t, result["empty_test"])
	})

	t.Run("Aggregate with bson.D pipeline", func(t *testing.T) {
		client.DeleteMany(ctx, "agg_test", bson.M{})
		client.InsertMany(ctx, "agg_test", []any{
			bson.M{"status": "A", "amount": 10},
			bson.M{"status": "A", "amount": 20},
		})

		pipeline := []bson.D{
			{{Key: "$match", Value: bson.M{"status": "A"}}},
		}

		var results []bson.M
		err := client.Aggregate(ctx, "agg_test", pipeline, &results)
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("Aggregate with bson.A pipeline", func(t *testing.T) {
		pipeline := bson.A{
			bson.M{"$match": bson.M{"status": "A"}},
		}

		var results []bson.M
		err := client.Aggregate(ctx, "agg_test", pipeline, &results)
		require.NoError(t, err)
	})

	t.Run("Aggregate with mongo.Pipeline", func(t *testing.T) {
		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.M{"status": "A"}}},
		}

		var results []bson.M
		err := client.Aggregate(ctx, "agg_test", pipeline, &results)
		require.NoError(t, err)
	})

	t.Run("DropCollection", func(t *testing.T) {
		collName := "to_drop_" + primitive.NewObjectID().Hex()
		client.CreateCollection(ctx, collName)

		err := client.DropCollection(ctx, collName)
		require.NoError(t, err)
	})

	t.Run("DropDatabase", func(t *testing.T) {
		dbName := "db_to_drop_" + primitive.NewObjectID().Hex()
		client.GetDatabase(dbName).Collection("temp").InsertOne(ctx, bson.M{"test": true})

		err := client.DropDatabase(ctx, dbName)
		require.NoError(t, err)
	})

	t.Run("WithTransaction", func(t *testing.T) {
		client.DeleteMany(ctx, "users", bson.M{})

		err := client.WithTransaction(ctx, func(sc mongo.SessionContext) error {
			_, err := client.InsertOne(sc, "users", testDocument{Name: "TxUser1", Age: 30, Active: true})
			if err != nil {
				return err
			}
			_, err = client.InsertOne(sc, "users", testDocument{Name: "TxUser2", Age: 25, Active: true})
			return err
		})
		require.NoError(t, err)

		count, _ := client.CountDocuments(ctx, "users", bson.M{"name": bson.M{"$in": []string{"TxUser1", "TxUser2"}}})
		assert.Equal(t, int64(2), count)
	})

	t.Run("Operations fail on closed client", func(t *testing.T) {
		cfg := DefaultConfig()
		WithURI(container.URI)(&cfg)
		WithDatabase("testdb")(&cfg)

		closedClient, _ := New(cfg)
		closedClient.Close(context.Background())

		_, err := closedClient.InsertOne(ctx, "users", testDocument{Name: "Test"})
		assert.ErrorIs(t, err, ErrClientClosed)

		_, err = closedClient.InsertMany(ctx, "users", []any{testDocument{Name: "Test"}})
		assert.ErrorIs(t, err, ErrClientClosed)

		err = closedClient.FindOne(ctx, "users", bson.M{}, &testDocument{})
		assert.ErrorIs(t, err, ErrClientClosed)

		err = closedClient.Find(ctx, "users", bson.M{}, &[]testDocument{})
		assert.ErrorIs(t, err, ErrClientClosed)

		_, err = closedClient.UpdateOne(ctx, "users", bson.M{}, bson.M{"$set": bson.M{"a": 1}})
		assert.ErrorIs(t, err, ErrClientClosed)

		_, err = closedClient.UpdateMany(ctx, "users", bson.M{}, bson.M{"$set": bson.M{"a": 1}})
		assert.ErrorIs(t, err, ErrClientClosed)

		_, err = closedClient.ReplaceOne(ctx, "users", bson.M{}, testDocument{})
		assert.ErrorIs(t, err, ErrClientClosed)

		_, err = closedClient.DeleteOne(ctx, "users", bson.M{})
		assert.ErrorIs(t, err, ErrClientClosed)

		_, err = closedClient.DeleteMany(ctx, "users", bson.M{})
		assert.ErrorIs(t, err, ErrClientClosed)

		_, err = closedClient.CountDocuments(ctx, "users", bson.M{})
		assert.ErrorIs(t, err, ErrClientClosed)

		_, err = closedClient.EstimatedDocumentCount(ctx, "users")
		assert.ErrorIs(t, err, ErrClientClosed)

		err = closedClient.Aggregate(ctx, "users", []bson.M{}, &[]bson.M{})
		assert.ErrorIs(t, err, ErrClientClosed)

		_, err = closedClient.Distinct(ctx, "users", "name", bson.M{})
		assert.ErrorIs(t, err, ErrClientClosed)

		err = closedClient.FindOneAndUpdate(ctx, "users", bson.M{}, bson.M{"$set": bson.M{"a": 1}}, &testDocument{})
		assert.ErrorIs(t, err, ErrClientClosed)

		err = closedClient.FindOneAndReplace(ctx, "users", bson.M{}, testDocument{}, &testDocument{})
		assert.ErrorIs(t, err, ErrClientClosed)

		err = closedClient.FindOneAndDelete(ctx, "users", bson.M{}, &testDocument{})
		assert.ErrorIs(t, err, ErrClientClosed)

		_, err = closedClient.BulkWrite(ctx, "users", []mongo.WriteModel{})
		assert.ErrorIs(t, err, ErrClientClosed)

		_, err = closedClient.CreateIndex(ctx, "users", bson.D{{Key: "name", Value: 1}})
		assert.ErrorIs(t, err, ErrClientClosed)

		_, err = closedClient.CreateIndexes(ctx, map[string][]mongo.IndexModel{})
		assert.ErrorIs(t, err, ErrClientClosed)

		err = closedClient.DropIndex(ctx, "users", "name_1")
		assert.ErrorIs(t, err, ErrClientClosed)

		_, err = closedClient.ListIndexes(ctx, "users")
		assert.ErrorIs(t, err, ErrClientClosed)

		_, err = closedClient.ListCollections(ctx)
		assert.ErrorIs(t, err, ErrClientClosed)

		err = closedClient.CreateCollection(ctx, "test")
		assert.ErrorIs(t, err, ErrClientClosed)

		err = closedClient.CreateCollections(ctx, map[string]*options.CreateCollectionOptions{})
		assert.ErrorIs(t, err, ErrClientClosed)

		err = closedClient.DropCollection(ctx, "users")
		assert.ErrorIs(t, err, ErrClientClosed)

		err = closedClient.DropDatabase(ctx, "testdb")
		assert.ErrorIs(t, err, ErrClientClosed)

		err = closedClient.WithTransaction(ctx, func(sc mongo.SessionContext) error { return nil })
		assert.ErrorIs(t, err, ErrClientClosed)
	})
}

func TestOperations_Watch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	container := testhelpers.SetupMongoContainer(t)
	defer container.Teardown(t)

	cfg := DefaultConfig()
	WithURI(container.URI)(&cfg)
	WithDatabase("testdb")(&cfg)
	WithTimeout(30 * time.Second)(&cfg)

	client, err := New(cfg)
	require.NoError(t, err)
	defer client.Close(context.Background())

	ctx := context.Background()

	t.Run("Watch collection changes", func(t *testing.T) {
		watchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		stream, err := client.Watch(watchCtx, "watch_test", mongo.Pipeline{})
		require.NoError(t, err)
		defer stream.Close(watchCtx)

		go func() {
			time.Sleep(100 * time.Millisecond)
			client.InsertOne(ctx, "watch_test", bson.M{"name": "watched"})
		}()

		if stream.Next(watchCtx) {
			var event bson.M
			err := stream.Decode(&event)
			require.NoError(t, err)
			assert.Equal(t, "insert", event["operationType"])
		}
	})

	t.Run("Watch with pipeline filter", func(t *testing.T) {
		watchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.M{"operationType": "insert"}}},
		}

		stream, err := client.Watch(watchCtx, "watch_filter_test", pipeline)
		require.NoError(t, err)
		defer stream.Close(watchCtx)

		go func() {
			time.Sleep(100 * time.Millisecond)
			client.InsertOne(ctx, "watch_filter_test", bson.M{"name": "filtered"})
		}()

		if stream.Next(watchCtx) {
			var event bson.M
			err := stream.Decode(&event)
			require.NoError(t, err)
			assert.Equal(t, "insert", event["operationType"])
		}
	})

	t.Run("Watch on closed client", func(t *testing.T) {
		closedCfg := DefaultConfig()
		WithURI(container.URI)(&closedCfg)
		WithDatabase("testdb")(&closedCfg)

		closedClient, _ := New(closedCfg)
		closedClient.Close(ctx)

		_, err := closedClient.Watch(ctx, "users", mongo.Pipeline{})
		assert.ErrorIs(t, err, ErrClientClosed)
	})
}
