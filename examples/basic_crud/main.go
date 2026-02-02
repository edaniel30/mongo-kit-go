package main

import (
	"context"
	"fmt"
	"log"
	"time"

	mongo_kit "github.com/edaniel30/mongo-kit-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user document in MongoDB
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Email     string             `bson:"email"`
	Age       int                `bson:"age"`
	Active    bool               `bson:"active"`
	CreatedAt time.Time          `bson:"created_at"`
}

func main() {
	// Create a new MongoDB client
	cfg := mongo_kit.DefaultConfig()
	client, err := mongo_kit.New(
		cfg,
		mongo_kit.WithURI("mongodb://localhost:27017"),
		mongo_kit.WithDatabase("myapp"),
		mongo_kit.WithTimeout(10*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Close(context.Background()); err != nil {
			log.Printf("Failed to close client: %v", err)
		}
	}()

	// Create a repository for User documents
	userRepo := mongo_kit.NewRepository[User](client, "users")

	ctx := context.Background()

	fmt.Println("=== Basic CRUD Operations Example ===")

	// 1. CREATE - Insert a single user
	fmt.Println("1. Creating a new user...")
	newUser := User{
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		Active:    true,
		CreatedAt: time.Now(),
	}
	userID, err := userRepo.Create(ctx, newUser)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return
	}
	fmt.Printf("✓ User created with ID: %s\n\n", userID)

	// 2. CREATE MANY - Insert multiple users
	fmt.Println("2. Creating multiple users...")
	users := []User{
		{Name: "Alice Smith", Email: "alice@example.com", Age: 25, Active: true, CreatedAt: time.Now()},
		{Name: "Bob Johnson", Email: "bob@example.com", Age: 35, Active: false, CreatedAt: time.Now()},
		{Name: "Carol White", Email: "carol@example.com", Age: 28, Active: true, CreatedAt: time.Now()},
	}
	ids, err := userRepo.CreateMany(ctx, users)
	if err != nil {
		log.Printf("Failed to create users: %v", err)
		return
	}
	fmt.Printf("✓ Created %d users\n\n", len(ids))

	// 3. READ - Find user by ID
	fmt.Println("3. Finding user by ID...")
	foundUser, err := userRepo.FindByID(ctx, userID)
	if err != nil {
		log.Printf("Failed to find user: %v", err)
		return
	}
	fmt.Printf("✓ Found user: %s (%s)\n\n", foundUser.Name, foundUser.Email)

	// 4. READ - Find all active users
	fmt.Println("4. Finding all active users...")
	activeUsers, err := userRepo.Find(ctx, map[string]any{"active": true})
	if err != nil {
		log.Printf("Failed to find active users: %v", err)
		return
	}
	fmt.Printf("✓ Found %d active users:\n", len(activeUsers))
	for _, user := range activeUsers {
		fmt.Printf("  - %s (%s)\n", user.Name, user.Email)
	}
	fmt.Println()

	// 5. UPDATE - Update user by ID
	fmt.Println("5. Updating user age...")
	update := map[string]any{
		"$set": map[string]any{
			"age": 31,
		},
	}
	result, err := userRepo.UpdateByID(ctx, userID, update)
	if err != nil {
		log.Printf("Failed to update user: %v", err)
		return
	}
	fmt.Printf("✓ Updated %d document(s)\n\n", result.ModifiedCount)

	// 6. UPDATE MANY - Activate all inactive users
	fmt.Println("6. Activating all inactive users...")
	updateMany := map[string]any{
		"$set": map[string]any{
			"active": true,
		},
	}
	resultMany, err := userRepo.UpdateMany(ctx, map[string]any{"active": false}, updateMany)
	if err != nil {
		log.Printf("Failed to update users: %v", err)
		return
	}
	fmt.Printf("✓ Updated %d document(s)\n\n", resultMany.ModifiedCount)

	// 7. UPSERT - Insert or update
	fmt.Println("7. Upserting a user...")
	upsertFilter := map[string]any{"email": "dave@example.com"}
	upsertUpdate := map[string]any{
		"$set": map[string]any{
			"name":       "Dave Brown",
			"email":      "dave@example.com",
			"age":        40,
			"active":     true,
			"created_at": time.Now(),
		},
	}
	upsertResult, err := userRepo.Upsert(ctx, upsertFilter, upsertUpdate)
	if err != nil {
		log.Printf("Failed to upsert user: %v", err)
		return
	}
	if upsertResult.UpsertedID != nil {
		fmt.Printf("✓ User inserted with ID: %v\n\n", upsertResult.UpsertedID)
	} else {
		fmt.Printf("✓ User updated (%d modified)\n\n", upsertResult.ModifiedCount)
	}

	// 8. COUNT - Count all users
	fmt.Println("8. Counting all users...")
	count, err := userRepo.CountAll(ctx)
	if err != nil {
		log.Printf("Failed to count users: %v", err)
		return
	}
	fmt.Printf("✓ Total users: %d\n\n", count)

	// 9. EXISTS - Check if user exists
	fmt.Println("9. Checking if user exists...")
	exists, err := userRepo.ExistsByID(ctx, userID)
	if err != nil {
		log.Printf("Failed to check existence: %v", err)
		return
	}
	fmt.Printf("✓ User exists: %v\n\n", exists)

	// 10. DELETE - Delete user by ID
	fmt.Println("10. Deleting user by ID...")
	deleteResult, err := userRepo.DeleteByID(ctx, userID)
	if err != nil {
		log.Printf("Failed to delete user: %v", err)
		return
	}
	fmt.Printf("✓ Deleted %d document(s)\n\n", deleteResult.DeletedCount)

	// 11. DELETE MANY - Delete inactive users (if any)
	fmt.Println("11. Deleting inactive users...")
	deleteMany, err := userRepo.DeleteMany(ctx, map[string]any{"active": false})
	if err != nil {
		log.Printf("Failed to delete users: %v", err)
		return
	}
	fmt.Printf("✓ Deleted %d document(s)\n\n", deleteMany.DeletedCount)

	// Final count
	finalCount, _ := userRepo.CountAll(ctx)
	fmt.Printf("Final user count: %d\n", finalCount)
}
