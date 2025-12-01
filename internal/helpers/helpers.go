package helpers

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ToObjectID(id string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id)
}

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

func IsValidObjectID(id string) bool {
	_, err := primitive.ObjectIDFromHex(id)
	return err == nil
}

func NewObjectID() primitive.ObjectID {
	return primitive.NewObjectID()
}

func ToBSON(data map[string]any) bson.M {
	return bson.M(data)
}

func ToBSONArray(data []map[string]any) []bson.M {
	result := make([]bson.M, len(data))
	for i, item := range data {
		result[i] = bson.M(item)
	}
	return result
}

func MergeBSON(docs ...bson.M) bson.M {
	result := bson.M{}
	for _, doc := range docs {
		for k, v := range doc {
			result[k] = v
		}
	}
	return result
}

func BSONToMap(data bson.M) map[string]any {
	return map[string]any(data)
}
