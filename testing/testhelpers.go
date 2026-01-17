package testing

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

type MongoContainer struct {
	*mongodb.MongoDBContainer
	URI string
}

func SetupMongoContainer(t *testing.T) *MongoContainer {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := mongodb.Run(ctx, "mongo:7", mongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatalf("failed to start MongoDB container: %v", err)
	}

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get MongoDB connection string: %v", err)
	}

	// Add directConnection for replica set to work from host
	uri = uri + "&directConnection=true"

	return &MongoContainer{
		MongoDBContainer: container,
		URI:              uri,
	}
}

func (c *MongoContainer) Teardown(t *testing.T) {
	t.Helper()

	if err := testcontainers.TerminateContainer(c.MongoDBContainer); err != nil {
		t.Logf("failed to terminate MongoDB container: %v", err)
	}
}
