package mongo_kit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	testhelpers "github.com/edaniel30/mongo-kit-go/testing"
)

type User struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	Name   string             `bson:"name"`
	Email  string             `bson:"email"`
	Age    int                `bson:"age"`
	Active bool               `bson:"active"`
}

func TestRepository_Integration(t *testing.T) {
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

	repo := NewRepository[User](client, "users")
	ctx := context.Background()

	t.Run("NewRepository creates repository", func(t *testing.T) {
		assert.NotNil(t, repo)
		assert.Equal(t, "users", repo.Collection())
		assert.Equal(t, client, repo.Client())
	})

	t.Run("Create inserts document", func(t *testing.T) {
		repo.Drop(ctx)

		user := User{Name: "John", Email: "john@test.com", Age: 30, Active: true}
		id, err := repo.Create(ctx, user)

		require.NoError(t, err)
		require.NotNil(t, id)
	})

	t.Run("CreateMany inserts multiple documents", func(t *testing.T) {
		repo.Drop(ctx)

		users := []User{
			{Name: "Alice", Email: "alice@test.com", Age: 25, Active: true},
			{Name: "Bob", Email: "bob@test.com", Age: 35, Active: false},
		}

		ids, err := repo.CreateMany(ctx, users)
		require.NoError(t, err)
		assert.Len(t, ids, 2)
	})

	t.Run("FindByID returns document", func(t *testing.T) {
		repo.Drop(ctx)

		user := User{Name: "FindMe", Email: "find@test.com", Age: 28, Active: true}
		id, _ := repo.Create(ctx, user)

		found, err := repo.FindByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "FindMe", found.Name)
		assert.Equal(t, "find@test.com", found.Email)
	})

	t.Run("FindByID returns error when not found", func(t *testing.T) {
		repo.Drop(ctx)

		_, err := repo.FindByID(ctx, primitive.NewObjectID())
		assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	})

	t.Run("FindOne returns document", func(t *testing.T) {
		repo.Drop(ctx)
		repo.Create(ctx, User{Name: "FindOne", Email: "one@test.com", Age: 30, Active: true})

		found, err := repo.FindOne(ctx, bson.M{"name": "FindOne"})
		require.NoError(t, err)
		assert.Equal(t, "FindOne", found.Name)
	})

	t.Run("FindOne returns error when not found", func(t *testing.T) {
		repo.Drop(ctx)

		_, err := repo.FindOne(ctx, bson.M{"name": "nonexistent"})
		assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	})

	t.Run("Find returns matching documents", func(t *testing.T) {
		repo.Drop(ctx)
		repo.CreateMany(ctx, []User{
			{Name: "Active1", Email: "a1@test.com", Age: 25, Active: true},
			{Name: "Active2", Email: "a2@test.com", Age: 30, Active: true},
			{Name: "Inactive", Email: "i@test.com", Age: 35, Active: false},
		})

		found, err := repo.Find(ctx, bson.M{"active": true})
		require.NoError(t, err)
		assert.Len(t, found, 2)
	})

	t.Run("Find returns empty slice when no match", func(t *testing.T) {
		repo.Drop(ctx)

		found, err := repo.Find(ctx, bson.M{"name": "nonexistent"})
		require.NoError(t, err)
		assert.Empty(t, found)
	})

	t.Run("FindAll returns all documents", func(t *testing.T) {
		repo.Drop(ctx)
		repo.CreateMany(ctx, []User{
			{Name: "All1", Email: "all1@test.com", Age: 25, Active: true},
			{Name: "All2", Email: "all2@test.com", Age: 30, Active: false},
		})

		found, err := repo.FindAll(ctx)
		require.NoError(t, err)
		assert.Len(t, found, 2)
	})

	t.Run("UpdateByID updates document", func(t *testing.T) {
		repo.Drop(ctx)
		id, _ := repo.Create(ctx, User{Name: "ToUpdate", Email: "update@test.com", Age: 25, Active: true})

		result, err := repo.UpdateByID(ctx, id, bson.M{"$set": bson.M{"age": 26}})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.ModifiedCount)

		updated, _ := repo.FindByID(ctx, id)
		assert.Equal(t, 26, updated.Age)
	})

	t.Run("UpdateOne updates single document", func(t *testing.T) {
		repo.Drop(ctx)
		repo.Create(ctx, User{Name: "UpdateOne", Email: "one@test.com", Age: 25, Active: true})

		result, err := repo.UpdateOne(ctx, bson.M{"name": "UpdateOne"}, bson.M{"$set": bson.M{"age": 30}})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.ModifiedCount)
	})

	t.Run("UpdateMany updates multiple documents", func(t *testing.T) {
		repo.Drop(ctx)
		repo.CreateMany(ctx, []User{
			{Name: "Many1", Email: "m1@test.com", Age: 25, Active: false},
			{Name: "Many2", Email: "m2@test.com", Age: 30, Active: false},
		})

		result, err := repo.UpdateMany(ctx, bson.M{"active": false}, bson.M{"$set": bson.M{"active": true}})
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.ModifiedCount)
	})

	t.Run("Upsert inserts when not exists", func(t *testing.T) {
		repo.Drop(ctx)

		result, err := repo.Upsert(ctx, bson.M{"email": "upsert@test.com"}, bson.M{"$set": bson.M{"name": "Upserted", "age": 40}})
		require.NoError(t, err)
		assert.NotNil(t, result.UpsertedID)
	})

	t.Run("Upsert updates when exists", func(t *testing.T) {
		repo.Drop(ctx)
		repo.Create(ctx, User{Name: "Existing", Email: "existing@test.com", Age: 25, Active: true})

		result, err := repo.Upsert(ctx, bson.M{"email": "existing@test.com"}, bson.M{"$set": bson.M{"age": 30}})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.ModifiedCount)
	})

	t.Run("DeleteByID deletes document", func(t *testing.T) {
		repo.Drop(ctx)
		id, _ := repo.Create(ctx, User{Name: "ToDelete", Email: "delete@test.com", Age: 25, Active: true})

		result, err := repo.DeleteByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.DeletedCount)

		_, err = repo.FindByID(ctx, id)
		assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	})

	t.Run("DeleteOne deletes single document", func(t *testing.T) {
		repo.Drop(ctx)
		repo.Create(ctx, User{Name: "DeleteOne", Email: "d1@test.com", Age: 25, Active: true})

		result, err := repo.DeleteOne(ctx, bson.M{"name": "DeleteOne"})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.DeletedCount)
	})

	t.Run("DeleteMany deletes multiple documents", func(t *testing.T) {
		repo.Drop(ctx)
		repo.CreateMany(ctx, []User{
			{Name: "DelMany1", Email: "dm1@test.com", Age: 25, Active: false},
			{Name: "DelMany2", Email: "dm2@test.com", Age: 30, Active: false},
		})

		result, err := repo.DeleteMany(ctx, bson.M{"active": false})
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.DeletedCount)
	})

	t.Run("Count returns document count", func(t *testing.T) {
		repo.Drop(ctx)
		repo.CreateMany(ctx, []User{
			{Name: "Count1", Email: "c1@test.com", Age: 25, Active: true},
			{Name: "Count2", Email: "c2@test.com", Age: 30, Active: true},
			{Name: "Count3", Email: "c3@test.com", Age: 35, Active: false},
		})

		count, err := repo.Count(ctx, bson.M{"active": true})
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("CountAll returns total count", func(t *testing.T) {
		repo.Drop(ctx)
		repo.CreateMany(ctx, []User{
			{Name: "All1", Email: "all1@test.com", Age: 25, Active: true},
			{Name: "All2", Email: "all2@test.com", Age: 30, Active: false},
		})

		count, err := repo.CountAll(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("EstimatedCount returns count", func(t *testing.T) {
		repo.Drop(ctx)
		repo.CreateMany(ctx, []User{
			{Name: "Est1", Email: "est1@test.com", Age: 25, Active: true},
			{Name: "Est2", Email: "est2@test.com", Age: 30, Active: false},
		})

		count, err := repo.EstimatedCount(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(2))
	})

	t.Run("Exists returns true when document exists", func(t *testing.T) {
		repo.Drop(ctx)
		repo.Create(ctx, User{Name: "Exists", Email: "exists@test.com", Age: 25, Active: true})

		exists, err := repo.Exists(ctx, bson.M{"name": "Exists"})
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Exists returns false when document does not exist", func(t *testing.T) {
		repo.Drop(ctx)

		exists, err := repo.Exists(ctx, bson.M{"name": "nonexistent"})
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("ExistsByID returns true when document exists", func(t *testing.T) {
		repo.Drop(ctx)
		id, _ := repo.Create(ctx, User{Name: "ExistsID", Email: "eid@test.com", Age: 25, Active: true})

		exists, err := repo.ExistsByID(ctx, id)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("ExistsByID returns false when document does not exist", func(t *testing.T) {
		repo.Drop(ctx)

		exists, err := repo.ExistsByID(ctx, primitive.NewObjectID())
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Aggregate executes pipeline", func(t *testing.T) {
		repo.Drop(ctx)
		repo.CreateMany(ctx, []User{
			{Name: "Agg1", Email: "agg1@test.com", Age: 25, Active: true},
			{Name: "Agg2", Email: "agg2@test.com", Age: 30, Active: true},
		})

		pipeline := []bson.M{
			{"$match": bson.M{"active": true}},
			{"$sort": bson.M{"age": 1}},
		}

		results, err := repo.Aggregate(ctx, pipeline)
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, "Agg1", results[0].Name)
	})

	t.Run("Drop removes collection", func(t *testing.T) {
		dropRepo := NewRepository[User](client, "to_drop_repo")
		dropRepo.Create(ctx, User{Name: "DropMe"})

		err := dropRepo.Drop(ctx)
		require.NoError(t, err)

		count, _ := dropRepo.CountAll(ctx)
		assert.Equal(t, int64(0), count)
	})

	t.Run("FindWithBuilder uses QueryBuilder", func(t *testing.T) {
		repo.Drop(ctx)
		repo.CreateMany(ctx, []User{
			{Name: "Builder1", Email: "b1@test.com", Age: 25, Active: true},
			{Name: "Builder2", Email: "b2@test.com", Age: 30, Active: true},
			{Name: "Builder3", Email: "b3@test.com", Age: 35, Active: false},
		})

		qb := NewQueryBuilder().
			Equals("active", true).
			GreaterThan("age", 20).
			Sort("age", false).
			Limit(2)

		results, err := repo.FindWithBuilder(ctx, qb)
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, "Builder2", results[0].Name) // Sorted by age desc
	})

	t.Run("FindOneWithBuilder returns single document", func(t *testing.T) {
		repo.Drop(ctx)
		repo.CreateMany(ctx, []User{
			{Name: "One1", Email: "o1@test.com", Age: 25, Active: true},
			{Name: "One2", Email: "o2@test.com", Age: 30, Active: true},
		})

		qb := NewQueryBuilder().
			Equals("active", true).
			Sort("age", false)

		result, err := repo.FindOneWithBuilder(ctx, qb)
		require.NoError(t, err)
		assert.Equal(t, "One2", result.Name) // Highest age
	})

	t.Run("FindOneWithBuilder returns error when not found", func(t *testing.T) {
		repo.Drop(ctx)

		qb := NewQueryBuilder().Equals("name", "nonexistent")

		_, err := repo.FindOneWithBuilder(ctx, qb)
		assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	})

	t.Run("CountWithBuilder counts matching documents", func(t *testing.T) {
		repo.Drop(ctx)
		repo.CreateMany(ctx, []User{
			{Name: "Count1", Email: "c1@test.com", Age: 25, Active: true},
			{Name: "Count2", Email: "c2@test.com", Age: 30, Active: true},
			{Name: "Count3", Email: "c3@test.com", Age: 35, Active: false},
		})

		qb := NewQueryBuilder().Equals("active", true)

		count, err := repo.CountWithBuilder(ctx, qb)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("ExistsWithBuilder checks existence", func(t *testing.T) {
		repo.Drop(ctx)
		repo.Create(ctx, User{Name: "ExistsBuilder", Email: "eb@test.com", Age: 25, Active: true})

		qb := NewQueryBuilder().Equals("name", "ExistsBuilder")
		exists, err := repo.ExistsWithBuilder(ctx, qb)
		require.NoError(t, err)
		assert.True(t, exists)

		qb2 := NewQueryBuilder().Equals("name", "nonexistent")
		exists, err = repo.ExistsWithBuilder(ctx, qb2)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}
