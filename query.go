package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// QueryBuilder provides a fluent interface for building MongoDB queries
type QueryBuilder struct {
	filter  bson.D
	options *options.FindOptions
}

// NewQueryBuilder creates a new QueryBuilder instance
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		filter:  bson.D{},
		options: options.Find(),
	}
}

// Filter adds a filter condition to the query
func (qb *QueryBuilder) Filter(key string, value interface{}) *QueryBuilder {
	qb.filter = append(qb.filter, bson.E{Key: key, Value: value})
	return qb
}

// Equals adds an equality filter
func (qb *QueryBuilder) Equals(key string, value interface{}) *QueryBuilder {
	return qb.Filter(key, value)
}

// NotEquals adds a not equals filter
func (qb *QueryBuilder) NotEquals(key string, value interface{}) *QueryBuilder {
	return qb.Filter(key, bson.M{"$ne": value})
}

// GreaterThan adds a greater than filter
func (qb *QueryBuilder) GreaterThan(key string, value interface{}) *QueryBuilder {
	return qb.Filter(key, bson.M{"$gt": value})
}

// GreaterThanOrEqual adds a greater than or equal filter
func (qb *QueryBuilder) GreaterThanOrEqual(key string, value interface{}) *QueryBuilder {
	return qb.Filter(key, bson.M{"$gte": value})
}

// LessThan adds a less than filter
func (qb *QueryBuilder) LessThan(key string, value interface{}) *QueryBuilder {
	return qb.Filter(key, bson.M{"$lt": value})
}

// LessThanOrEqual adds a less than or equal filter
func (qb *QueryBuilder) LessThanOrEqual(key string, value interface{}) *QueryBuilder {
	return qb.Filter(key, bson.M{"$lte": value})
}

// In adds an in filter
func (qb *QueryBuilder) In(key string, values ...interface{}) *QueryBuilder {
	return qb.Filter(key, bson.M{"$in": values})
}

// NotIn adds a not in filter
func (qb *QueryBuilder) NotIn(key string, values ...interface{}) *QueryBuilder {
	return qb.Filter(key, bson.M{"$nin": values})
}

// Exists adds an exists filter
func (qb *QueryBuilder) Exists(key string, exists bool) *QueryBuilder {
	return qb.Filter(key, bson.M{"$exists": exists})
}

// Regex adds a regex filter
func (qb *QueryBuilder) Regex(key string, pattern string, options string) *QueryBuilder {
	return qb.Filter(key, bson.M{"$regex": pattern, "$options": options})
}

// And adds an and condition
func (qb *QueryBuilder) And(conditions ...bson.D) *QueryBuilder {
	if len(conditions) > 0 {
		qb.filter = append(qb.filter, bson.E{Key: "$and", Value: conditions})
	}
	return qb
}

// Or adds an or condition
func (qb *QueryBuilder) Or(conditions ...bson.D) *QueryBuilder {
	if len(conditions) > 0 {
		qb.filter = append(qb.filter, bson.E{Key: "$or", Value: conditions})
	}
	return qb
}

// Nor adds a nor condition
func (qb *QueryBuilder) Nor(conditions ...bson.D) *QueryBuilder {
	if len(conditions) > 0 {
		qb.filter = append(qb.filter, bson.E{Key: "$nor", Value: conditions})
	}
	return qb
}

// Limit sets the maximum number of documents to return
func (qb *QueryBuilder) Limit(limit int64) *QueryBuilder {
	qb.options.SetLimit(limit)
	return qb
}

// Skip sets the number of documents to skip
func (qb *QueryBuilder) Skip(skip int64) *QueryBuilder {
	qb.options.SetSkip(skip)
	return qb
}

// Sort sets the sort order
func (qb *QueryBuilder) Sort(field string, ascending bool) *QueryBuilder {
	order := 1
	if !ascending {
		order = -1
	}
	qb.options.SetSort(bson.D{{Key: field, Value: order}})
	return qb
}

// SortBy sets custom sort order
func (qb *QueryBuilder) SortBy(sort interface{}) *QueryBuilder {
	qb.options.SetSort(sort)
	return qb
}

// Project sets the projection
func (qb *QueryBuilder) Project(projection interface{}) *QueryBuilder {
	qb.options.SetProjection(projection)
	return qb
}

// Build returns the filter and options
func (qb *QueryBuilder) Build() (bson.D, *options.FindOptions) {
	return qb.filter, qb.options
}

// BuildFilter returns only the filter
func (qb *QueryBuilder) BuildFilter() bson.D {
	return qb.filter
}

// BuildOptions returns only the options
func (qb *QueryBuilder) BuildOptions() *options.FindOptions {
	return qb.options
}

// UpdateBuilder provides a fluent interface for building update operations
type UpdateBuilder struct {
	update bson.D
}

// NewUpdateBuilder creates a new UpdateBuilder instance
func NewUpdateBuilder() *UpdateBuilder {
	return &UpdateBuilder{
		update: bson.D{},
	}
}

// Set sets field values
func (ub *UpdateBuilder) Set(key string, value interface{}) *UpdateBuilder {
	ub.addOperator("$set", key, value)
	return ub
}

// SetMultiple sets multiple field values at once
func (ub *UpdateBuilder) SetMultiple(fields map[string]interface{}) *UpdateBuilder {
	for key, value := range fields {
		ub.Set(key, value)
	}
	return ub
}

// Unset removes fields
func (ub *UpdateBuilder) Unset(keys ...string) *UpdateBuilder {
	for _, key := range keys {
		ub.addOperator("$unset", key, "")
	}
	return ub
}

// Inc increments field values
func (ub *UpdateBuilder) Inc(key string, value interface{}) *UpdateBuilder {
	ub.addOperator("$inc", key, value)
	return ub
}

// Mul multiplies field values
func (ub *UpdateBuilder) Mul(key string, value interface{}) *UpdateBuilder {
	ub.addOperator("$mul", key, value)
	return ub
}

// Min updates field if specified value is less than current value
func (ub *UpdateBuilder) Min(key string, value interface{}) *UpdateBuilder {
	ub.addOperator("$min", key, value)
	return ub
}

// Max updates field if specified value is greater than current value
func (ub *UpdateBuilder) Max(key string, value interface{}) *UpdateBuilder {
	ub.addOperator("$max", key, value)
	return ub
}

// Push appends value to array
func (ub *UpdateBuilder) Push(key string, value interface{}) *UpdateBuilder {
	ub.addOperator("$push", key, value)
	return ub
}

// Pull removes all instances of value from array
func (ub *UpdateBuilder) Pull(key string, value interface{}) *UpdateBuilder {
	ub.addOperator("$pull", key, value)
	return ub
}

// AddToSet adds value to array if not already present
func (ub *UpdateBuilder) AddToSet(key string, value interface{}) *UpdateBuilder {
	ub.addOperator("$addToSet", key, value)
	return ub
}

// Pop removes first or last element from array
func (ub *UpdateBuilder) Pop(key string, first bool) *UpdateBuilder {
	position := 1
	if first {
		position = -1
	}
	ub.addOperator("$pop", key, position)
	return ub
}

// CurrentDate sets field to current date
func (ub *UpdateBuilder) CurrentDate(key string) *UpdateBuilder {
	ub.addOperator("$currentDate", key, true)
	return ub
}

// Rename renames a field
func (ub *UpdateBuilder) Rename(oldName string, newName string) *UpdateBuilder {
	ub.addOperator("$rename", oldName, newName)
	return ub
}

// addOperator is a helper method to add operators to the update document
func (ub *UpdateBuilder) addOperator(operator string, key string, value interface{}) {
	// Find existing operator
	for i, elem := range ub.update {
		if elem.Key == operator {
			// Operator exists, add to it
			if m, ok := elem.Value.(bson.M); ok {
				m[key] = value
				ub.update[i].Value = m
				return
			}
		}
	}

	// Operator doesn't exist, create it
	ub.update = append(ub.update, bson.E{
		Key:   operator,
		Value: bson.M{key: value},
	})
}

// Build returns the update document
func (ub *UpdateBuilder) Build() bson.D {
	return ub.update
}

// AggregationBuilder provides a fluent interface for building aggregation pipelines
type AggregationBuilder struct {
	pipeline []bson.D
}

// NewAggregationBuilder creates a new AggregationBuilder instance
func NewAggregationBuilder() *AggregationBuilder {
	return &AggregationBuilder{
		pipeline: []bson.D{},
	}
}

// Match adds a $match stage
func (ab *AggregationBuilder) Match(filter interface{}) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$match", Value: filter}})
	return ab
}

// Group adds a $group stage
func (ab *AggregationBuilder) Group(id interface{}, fields bson.M) *AggregationBuilder {
	groupDoc := bson.M{"_id": id}
	for k, v := range fields {
		groupDoc[k] = v
	}
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$group", Value: groupDoc}})
	return ab
}

// Sort adds a $sort stage
func (ab *AggregationBuilder) Sort(sort interface{}) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$sort", Value: sort}})
	return ab
}

// Limit adds a $limit stage
func (ab *AggregationBuilder) Limit(limit int64) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$limit", Value: limit}})
	return ab
}

// Skip adds a $skip stage
func (ab *AggregationBuilder) Skip(skip int64) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$skip", Value: skip}})
	return ab
}

// Project adds a $project stage
func (ab *AggregationBuilder) Project(projection interface{}) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$project", Value: projection}})
	return ab
}

// Unwind adds an $unwind stage
func (ab *AggregationBuilder) Unwind(path string) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.D{{Key: "$unwind", Value: path}})
	return ab
}

// Lookup adds a $lookup stage for joins
func (ab *AggregationBuilder) Lookup(from, localField, foreignField, as string) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         from,
			"localField":   localField,
			"foreignField": foreignField,
			"as":           as,
		},
	}})
	return ab
}

// AddStage adds a custom pipeline stage
func (ab *AggregationBuilder) AddStage(stage bson.D) *AggregationBuilder {
	ab.pipeline = append(ab.pipeline, stage)
	return ab
}

// Build returns the aggregation pipeline
func (ab *AggregationBuilder) Build() []bson.D {
	return ab.pipeline
}
