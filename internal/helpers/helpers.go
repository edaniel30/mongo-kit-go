package helpers

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ToObjectID converts a string to MongoDB ObjectID
func ToObjectID(id string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id)
}

// ToObjectIDs converts multiple strings to MongoDB ObjectIDs
func ToObjectIDs(ids []string) ([]primitive.ObjectID, error) {
	objectIDs := make([]primitive.ObjectID, len(ids))
	for i, id := range ids {
		oid, err := ToObjectID(id)
		if err != nil {
			return nil, err
		}
		objectIDs[i] = oid
	}
	return objectIDs, nil
}

// IsValidObjectID checks if a string is a valid MongoDB ObjectID
func IsValidObjectID(id string) bool {
	_, err := primitive.ObjectIDFromHex(id)
	return err == nil
}

// NewObjectID generates a new MongoDB ObjectID
func NewObjectID() primitive.ObjectID {
	return primitive.NewObjectID()
}

// ToBSON converts a map to BSON document
func ToBSON(data map[string]interface{}) bson.M {
	return bson.M(data)
}

// ToBSONArray converts a slice of maps to BSON array
func ToBSONArray(data []map[string]interface{}) []bson.M {
	result := make([]bson.M, len(data))
	for i, item := range data {
		result[i] = bson.M(item)
	}
	return result
}

// MergeBSON merges multiple BSON documents
func MergeBSON(docs ...bson.M) bson.M {
	result := bson.M{}
	for _, doc := range docs {
		for k, v := range doc {
			result[k] = v
		}
	}
	return result
}
