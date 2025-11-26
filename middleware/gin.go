package middleware

import (
	"github.com/edaniel30/mongo-kit-go"
	"github.com/gin-gonic/gin"
)

const (
	// MongoClientKey is the key used to store the MongoDB client in gin context
	MongoClientKey = "mongo_client"
)

// MongoClient returns a Gin middleware that injects the MongoDB client into the request context
// This allows handlers to access the client via c.MustGet(middleware.MongoClientKey)
func MongoClient(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(MongoClientKey, client)
		c.Next()
	}
}

// GetClient retrieves the MongoDB client from the Gin context
// Returns nil if the client is not found in the context
func GetClient(c *gin.Context) *mongo.Client {
	if client, exists := c.Get(MongoClientKey); exists {
		if mongoClient, ok := client.(*mongo.Client); ok {
			return mongoClient
		}
	}
	return nil
}

// MustGetClient retrieves the MongoDB client from the Gin context
// Panics if the client is not found in the context
func MustGetClient(c *gin.Context) *mongo.Client {
	client := GetClient(c)
	if client == nil {
		panic("MongoDB client not found in context. Did you forget to add the middleware?")
	}
	return client
}
