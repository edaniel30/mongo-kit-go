package mongo_kit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestNewQueryBuilder(t *testing.T) {
	qb := NewQueryBuilder()

	require.NotNil(t, qb)
	assert.Empty(t, qb.filter)
	assert.NotNil(t, qb.options)
}

func TestQueryBuilder_FilterOperators(t *testing.T) {
	tests := []struct {
		name        string
		build       func() *QueryBuilder
		expectedKey string
		validateVal func(t *testing.T, val any)
	}{
		{
			name:        "Filter/Equals",
			build:       func() *QueryBuilder { return NewQueryBuilder().Equals("status", "active") },
			expectedKey: "status",
			validateVal: func(t *testing.T, val any) { assert.Equal(t, "active", val) },
		},
		{
			name:        "NotEquals",
			build:       func() *QueryBuilder { return NewQueryBuilder().NotEquals("status", "deleted") },
			expectedKey: "status",
			validateVal: func(t *testing.T, val any) {
				m := val.(bson.M)
				assert.Equal(t, "deleted", m["$ne"])
			},
		},
		{
			name:        "GreaterThan",
			build:       func() *QueryBuilder { return NewQueryBuilder().GreaterThan("age", 18) },
			expectedKey: "age",
			validateVal: func(t *testing.T, val any) {
				m := val.(bson.M)
				assert.Equal(t, 18, m["$gt"])
			},
		},
		{
			name:        "GreaterThanOrEqual",
			build:       func() *QueryBuilder { return NewQueryBuilder().GreaterThanOrEqual("score", 90) },
			expectedKey: "score",
			validateVal: func(t *testing.T, val any) {
				m := val.(bson.M)
				assert.Equal(t, 90, m["$gte"])
			},
		},
		{
			name:        "LessThan",
			build:       func() *QueryBuilder { return NewQueryBuilder().LessThan("price", 100.5) },
			expectedKey: "price",
			validateVal: func(t *testing.T, val any) {
				m := val.(bson.M)
				assert.Equal(t, 100.5, m["$lt"])
			},
		},
		{
			name:        "LessThanOrEqual",
			build:       func() *QueryBuilder { return NewQueryBuilder().LessThanOrEqual("qty", 10) },
			expectedKey: "qty",
			validateVal: func(t *testing.T, val any) {
				m := val.(bson.M)
				assert.Equal(t, 10, m["$lte"])
			},
		},
		{
			name:        "In",
			build:       func() *QueryBuilder { return NewQueryBuilder().In("status", "a", "b", "c") },
			expectedKey: "status",
			validateVal: func(t *testing.T, val any) {
				m := val.(bson.M)
				assert.Equal(t, []any{"a", "b", "c"}, m["$in"])
			},
		},
		{
			name:        "NotIn",
			build:       func() *QueryBuilder { return NewQueryBuilder().NotIn("role", "admin", "super") },
			expectedKey: "role",
			validateVal: func(t *testing.T, val any) {
				m := val.(bson.M)
				assert.Equal(t, []any{"admin", "super"}, m["$nin"])
			},
		},
		{
			name:        "Exists true",
			build:       func() *QueryBuilder { return NewQueryBuilder().Exists("email", true) },
			expectedKey: "email",
			validateVal: func(t *testing.T, val any) {
				m := val.(bson.M)
				assert.Equal(t, true, m["$exists"])
			},
		},
		{
			name:        "Exists false",
			build:       func() *QueryBuilder { return NewQueryBuilder().Exists("deleted", false) },
			expectedKey: "deleted",
			validateVal: func(t *testing.T, val any) {
				m := val.(bson.M)
				assert.Equal(t, false, m["$exists"])
			},
		},
		{
			name:        "Regex",
			build:       func() *QueryBuilder { return NewQueryBuilder().Regex("email", ".*@test\\.com$", "i") },
			expectedKey: "email",
			validateVal: func(t *testing.T, val any) {
				m := val.(bson.M)
				assert.Equal(t, ".*@test\\.com$", m["$regex"])
				assert.Equal(t, "i", m["$options"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := tt.build()
			filter := qb.GetFilter()

			require.Len(t, filter, 1)
			assert.Equal(t, tt.expectedKey, filter[0].Key)
			tt.validateVal(t, filter[0].Value)
		})
	}
}

func TestQueryBuilder_LogicalOperators(t *testing.T) {
	cond1 := bson.D{{Key: "status", Value: "active"}}
	cond2 := bson.D{{Key: "verified", Value: true}}

	tests := []struct {
		name        string
		build       func() *QueryBuilder
		expectedOp  string
		expectEmpty bool
	}{
		{
			name:       "And with conditions",
			build:      func() *QueryBuilder { return NewQueryBuilder().And(cond1, cond2) },
			expectedOp: "$and",
		},
		{
			name:        "And empty",
			build:       func() *QueryBuilder { return NewQueryBuilder().And() },
			expectEmpty: true,
		},
		{
			name:       "Or with conditions",
			build:      func() *QueryBuilder { return NewQueryBuilder().Or(cond1, cond2) },
			expectedOp: "$or",
		},
		{
			name:        "Or empty",
			build:       func() *QueryBuilder { return NewQueryBuilder().Or() },
			expectEmpty: true,
		},
		{
			name:       "Nor with conditions",
			build:      func() *QueryBuilder { return NewQueryBuilder().Nor(cond1, cond2) },
			expectedOp: "$nor",
		},
		{
			name:        "Nor empty",
			build:       func() *QueryBuilder { return NewQueryBuilder().Nor() },
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := tt.build()
			filter := qb.GetFilter()

			if tt.expectEmpty {
				assert.Empty(t, filter)
			} else {
				require.Len(t, filter, 1)
				assert.Equal(t, tt.expectedOp, filter[0].Key)
			}
		})
	}
}

func TestQueryBuilder_ConditionBuilders(t *testing.T) {
	builder1 := NewQueryBuilder().Equals("status", "active")
	builder2 := NewQueryBuilder().GreaterThan("age", 18)

	tests := []struct {
		name        string
		build       func() *QueryBuilder
		expectedOp  string
		expectEmpty bool
	}{
		{
			name:       "AndConditions",
			build:      func() *QueryBuilder { return NewQueryBuilder().AndConditions(builder1, builder2) },
			expectedOp: "$and",
		},
		{
			name:        "AndConditions empty",
			build:       func() *QueryBuilder { return NewQueryBuilder().AndConditions() },
			expectEmpty: true,
		},
		{
			name:       "OrConditions",
			build:      func() *QueryBuilder { return NewQueryBuilder().OrConditions(builder1, builder2) },
			expectedOp: "$or",
		},
		{
			name:        "OrConditions empty",
			build:       func() *QueryBuilder { return NewQueryBuilder().OrConditions() },
			expectEmpty: true,
		},
		{
			name:       "NorConditions",
			build:      func() *QueryBuilder { return NewQueryBuilder().NorConditions(builder1, builder2) },
			expectedOp: "$nor",
		},
		{
			name:        "NorConditions empty",
			build:       func() *QueryBuilder { return NewQueryBuilder().NorConditions() },
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := tt.build()
			filter := qb.GetFilter()

			if tt.expectEmpty {
				assert.Empty(t, filter)
			} else {
				require.Len(t, filter, 1)
				assert.Equal(t, tt.expectedOp, filter[0].Key)
			}
		})
	}
}

func TestQueryBuilder_Options(t *testing.T) {
	t.Run("Limit and Skip", func(t *testing.T) {
		qb := NewQueryBuilder().Limit(10).Skip(20)
		_, opts := qb.Build()

		require.NotNil(t, opts.Limit)
		assert.Equal(t, int64(10), *opts.Limit)
		require.NotNil(t, opts.Skip)
		assert.Equal(t, int64(20), *opts.Skip)
	})

	t.Run("Sort ascending and descending", func(t *testing.T) {
		qb := NewQueryBuilder().Sort("name", true).Sort("createdAt", false)
		_, opts := qb.Build()

		sort := opts.Sort.(bson.D)
		require.Len(t, sort, 2)
		assert.Equal(t, 1, sort[0].Value)
		assert.Equal(t, -1, sort[1].Value)
	})

	t.Run("SortBy replaces previous sort", func(t *testing.T) {
		qb := NewQueryBuilder().Sort("field1", true).SortBy(bson.M{"field2": -1})
		_, opts := qb.Build()

		sortMap := opts.Sort.(bson.M)
		assert.Equal(t, -1, sortMap["field2"])
	})

	t.Run("Project", func(t *testing.T) {
		projection := bson.M{"name": 1, "_id": 0}
		qb := NewQueryBuilder().Project(projection)
		_, opts := qb.Build()

		assert.Equal(t, projection, opts.Projection)
	})
}

func TestQueryBuilder_Where(t *testing.T) {
	tests := []struct {
		name        string
		expr        any
		expectEmpty bool
		expectedKey string
	}{
		{
			name:        "bson.M",
			expr:        bson.M{"$text": bson.M{"$search": "test"}},
			expectedKey: "$text",
		},
		{
			name:        "bson.D",
			expr:        bson.D{{Key: "custom", Value: "value"}},
			expectedKey: "custom",
		},
		{
			name:        "bson.E",
			expr:        bson.E{Key: "single", Value: "element"},
			expectedKey: "single",
		},
		{
			name:        "unsupported type",
			expr:        "invalid",
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder().Where(tt.expr)
			filter := qb.GetFilter()

			if tt.expectEmpty {
				assert.Empty(t, filter)
			} else {
				require.Len(t, filter, 1)
				assert.Equal(t, tt.expectedKey, filter[0].Key)
			}
		})
	}
}

func TestQueryBuilder_ComplexChaining(t *testing.T) {
	qb := NewQueryBuilder().
		Equals("status", "active").
		GreaterThan("age", 18).
		In("role", "user", "premium").
		Limit(50).
		Skip(10).
		Sort("createdAt", false).
		Project(bson.M{"password": 0})

	filter, opts := qb.Build()

	assert.Len(t, filter, 3)
	assert.Equal(t, int64(50), *opts.Limit)
	assert.Equal(t, int64(10), *opts.Skip)
	require.NotNil(t, opts.Sort)
	require.NotNil(t, opts.Projection)
}

func TestNewUpdateBuilder(t *testing.T) {
	ub := NewUpdateBuilder()
	require.NotNil(t, ub)
	assert.Empty(t, ub.update)
}

func TestUpdateBuilder_Operators(t *testing.T) {
	tests := []struct {
		name        string
		build       func() *UpdateBuilder
		expectedOp  string
		validateDoc func(t *testing.T, doc bson.M)
	}{
		{
			name:       "Set",
			build:      func() *UpdateBuilder { return NewUpdateBuilder().Set("name", "John") },
			expectedOp: "$set",
			validateDoc: func(t *testing.T, doc bson.M) {
				assert.Equal(t, "John", doc["name"])
			},
		},
		{
			name:       "Set multiple",
			build:      func() *UpdateBuilder { return NewUpdateBuilder().Set("a", 1).Set("b", 2) },
			expectedOp: "$set",
			validateDoc: func(t *testing.T, doc bson.M) {
				assert.Len(t, doc, 2)
			},
		},
		{
			name:       "Unset",
			build:      func() *UpdateBuilder { return NewUpdateBuilder().Unset("field1", "field2") },
			expectedOp: "$unset",
			validateDoc: func(t *testing.T, doc bson.M) {
				assert.Len(t, doc, 2)
			},
		},
		{
			name:       "Inc",
			build:      func() *UpdateBuilder { return NewUpdateBuilder().Inc("count", 1) },
			expectedOp: "$inc",
			validateDoc: func(t *testing.T, doc bson.M) {
				assert.Equal(t, 1, doc["count"])
			},
		},
		{
			name:       "Mul",
			build:      func() *UpdateBuilder { return NewUpdateBuilder().Mul("price", 1.1) },
			expectedOp: "$mul",
			validateDoc: func(t *testing.T, doc bson.M) {
				assert.Equal(t, 1.1, doc["price"])
			},
		},
		{
			name:       "Min",
			build:      func() *UpdateBuilder { return NewUpdateBuilder().Min("low", 50) },
			expectedOp: "$min",
			validateDoc: func(t *testing.T, doc bson.M) {
				assert.Equal(t, 50, doc["low"])
			},
		},
		{
			name:       "Max",
			build:      func() *UpdateBuilder { return NewUpdateBuilder().Max("high", 100) },
			expectedOp: "$max",
			validateDoc: func(t *testing.T, doc bson.M) {
				assert.Equal(t, 100, doc["high"])
			},
		},
		{
			name:       "Push",
			build:      func() *UpdateBuilder { return NewUpdateBuilder().Push("tags", "go") },
			expectedOp: "$push",
			validateDoc: func(t *testing.T, doc bson.M) {
				assert.Equal(t, "go", doc["tags"])
			},
		},
		{
			name:       "Pull",
			build:      func() *UpdateBuilder { return NewUpdateBuilder().Pull("tags", "old") },
			expectedOp: "$pull",
			validateDoc: func(t *testing.T, doc bson.M) {
				assert.Equal(t, "old", doc["tags"])
			},
		},
		{
			name:       "AddToSet",
			build:      func() *UpdateBuilder { return NewUpdateBuilder().AddToSet("cats", "tech") },
			expectedOp: "$addToSet",
			validateDoc: func(t *testing.T, doc bson.M) {
				assert.Equal(t, "tech", doc["cats"])
			},
		},
		{
			name:       "Pop first",
			build:      func() *UpdateBuilder { return NewUpdateBuilder().Pop("queue", true) },
			expectedOp: "$pop",
			validateDoc: func(t *testing.T, doc bson.M) {
				assert.Equal(t, -1, doc["queue"])
			},
		},
		{
			name:       "Pop last",
			build:      func() *UpdateBuilder { return NewUpdateBuilder().Pop("stack", false) },
			expectedOp: "$pop",
			validateDoc: func(t *testing.T, doc bson.M) {
				assert.Equal(t, 1, doc["stack"])
			},
		},
		{
			name:       "CurrentDate",
			build:      func() *UpdateBuilder { return NewUpdateBuilder().CurrentDate("updatedAt") },
			expectedOp: "$currentDate",
			validateDoc: func(t *testing.T, doc bson.M) {
				assert.Equal(t, true, doc["updatedAt"])
			},
		},
		{
			name:       "Rename",
			build:      func() *UpdateBuilder { return NewUpdateBuilder().Rename("old", "new") },
			expectedOp: "$rename",
			validateDoc: func(t *testing.T, doc bson.M) {
				assert.Equal(t, "new", doc["old"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ub := tt.build()
			update := ub.Build()

			var found bool
			for _, elem := range update {
				if elem.Key == tt.expectedOp {
					found = true
					doc := elem.Value.(bson.M)
					tt.validateDoc(t, doc)
					break
				}
			}
			assert.True(t, found, "operator %s not found", tt.expectedOp)
		})
	}
}

func TestUpdateBuilder_MultipleOperators(t *testing.T) {
	ub := NewUpdateBuilder().
		Set("name", "John").
		Inc("count", 1).
		CurrentDate("updated").
		Unset("temp")

	update := ub.Build()
	assert.Len(t, update, 4)

	ops := make(map[string]bool)
	for _, elem := range update {
		ops[elem.Key] = true
	}

	assert.True(t, ops["$set"])
	assert.True(t, ops["$inc"])
	assert.True(t, ops["$currentDate"])
	assert.True(t, ops["$unset"])
}

func TestNewAggregationBuilder(t *testing.T) {
	ab := NewAggregationBuilder()
	require.NotNil(t, ab)
	assert.Empty(t, ab.pipeline)
}

func TestAggregationBuilder_Stages(t *testing.T) {
	tests := []struct {
		name       string
		build      func() *AggregationBuilder
		expectedOp string
	}{
		{
			name:       "Match",
			build:      func() *AggregationBuilder { return NewAggregationBuilder().Match(bson.M{"active": true}) },
			expectedOp: "$match",
		},
		{
			name: "Group",
			build: func() *AggregationBuilder {
				return NewAggregationBuilder().Group("$cat", bson.M{"total": bson.M{"$sum": 1}})
			},
			expectedOp: "$group",
		},
		{
			name:       "Sort",
			build:      func() *AggregationBuilder { return NewAggregationBuilder().Sort(bson.D{{Key: "x", Value: -1}}) },
			expectedOp: "$sort",
		},
		{
			name:       "Limit",
			build:      func() *AggregationBuilder { return NewAggregationBuilder().Limit(10) },
			expectedOp: "$limit",
		},
		{
			name:       "Skip",
			build:      func() *AggregationBuilder { return NewAggregationBuilder().Skip(5) },
			expectedOp: "$skip",
		},
		{
			name:       "Project",
			build:      func() *AggregationBuilder { return NewAggregationBuilder().Project(bson.M{"name": 1}) },
			expectedOp: "$project",
		},
		{
			name:       "Unwind",
			build:      func() *AggregationBuilder { return NewAggregationBuilder().Unwind("$items") },
			expectedOp: "$unwind",
		},
		{
			name:       "Lookup",
			build:      func() *AggregationBuilder { return NewAggregationBuilder().Lookup("orders", "uid", "_id", "ords") },
			expectedOp: "$lookup",
		},
		{
			name: "AddStage",
			build: func() *AggregationBuilder {
				return NewAggregationBuilder().AddStage(bson.D{{Key: "$out", Value: "out"}})
			},
			expectedOp: "$out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ab := tt.build()
			pipeline := ab.Build()

			require.Len(t, pipeline, 1)
			assert.Equal(t, tt.expectedOp, pipeline[0][0].Key)
		})
	}
}

func TestAggregationBuilder_Lookup_Structure(t *testing.T) {
	ab := NewAggregationBuilder().Lookup("orders", "userId", "_id", "userOrders")
	pipeline := ab.Build()

	lookupDoc := pipeline[0][0].Value.(bson.M)
	assert.Equal(t, "orders", lookupDoc["from"])
	assert.Equal(t, "userId", lookupDoc["localField"])
	assert.Equal(t, "_id", lookupDoc["foreignField"])
	assert.Equal(t, "userOrders", lookupDoc["as"])
}

func TestAggregationBuilder_PipelineOrder(t *testing.T) {
	ab := NewAggregationBuilder().
		Match(bson.M{"active": true}).
		Unwind("$items").
		Group("$category", bson.M{"count": bson.M{"$sum": 1}}).
		Sort(bson.D{{Key: "count", Value: -1}}).
		Limit(10)

	pipeline := ab.Build()

	expectedOrder := []string{"$match", "$unwind", "$group", "$sort", "$limit"}
	require.Len(t, pipeline, len(expectedOrder))

	for i, expected := range expectedOrder {
		assert.Equal(t, expected, pipeline[i][0].Key)
	}
}
