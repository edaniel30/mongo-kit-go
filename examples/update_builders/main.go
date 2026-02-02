package main

import (
	"context"
	"fmt"
	"log"
	"time"

	mongo_kit "github.com/edaniel30/mongo-kit-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Article represents a blog article
type Article struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Title       string             `bson:"title"`
	Content     string             `bson:"content"`
	Author      string             `bson:"author"`
	Views       int                `bson:"views"`
	Likes       int                `bson:"likes"`
	Tags        []string           `bson:"tags"`
	Comments    []string           `bson:"comments"`
	Published   bool               `bson:"published"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
	LastViewAt  time.Time          `bson:"last_view_at,omitempty"`
}

func main() {
	// Setup
	cfg := mongo_kit.DefaultConfig()
	client, err := mongo_kit.New(
		cfg,
		mongo_kit.WithURI("mongodb://localhost:27017"),
		mongo_kit.WithDatabase("blog"),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer func() {
		if err := client.Close(context.Background()); err != nil {
			log.Printf("Failed to close client: %v", err)
		}
	}()

	articleRepo := mongo_kit.NewRepository[Article](client, "articles")
	ctx := context.Background()

	// Clean and create sample article
	_ = articleRepo.Drop(ctx)
	article := Article{
		Title:      "Getting Started with Go",
		Content:    "Go is a great programming language...",
		Author:     "John Doe",
		Views:      100,
		Likes:      5,
		Tags:       []string{"go", "programming"},
		Comments:   []string{"Great article!"},
		Published:  true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	articleID, _ := articleRepo.Create(ctx, article)

	fmt.Println("=== Update Builder Examples ===")

	// Example 1: Set single field
	fmt.Println("1. Update article title:")
	ub1 := mongo_kit.NewUpdateBuilder().
		Set("title", "Advanced Go Programming")

	result1, _ := articleRepo.UpdateByID(ctx, articleID, ub1.Build())
	fmt.Printf("  ✓ Modified: %d document(s)\n\n", result1.ModifiedCount)

	// Example 2: Set multiple fields
	fmt.Println("2. Update multiple fields:")
	ub2 := mongo_kit.NewUpdateBuilder().
		Set("content", "Updated content about Go...").
		Set("updated_at", time.Now()).
		Set("author", "Jane Smith")

	result2, _ := articleRepo.UpdateByID(ctx, articleID, ub2.Build())
	fmt.Printf("  ✓ Modified: %d document(s)\n\n", result2.ModifiedCount)

	// Example 3: Increment numeric fields
	fmt.Println("3. Increment views and likes:")
	ub3 := mongo_kit.NewUpdateBuilder().
		Inc("views", 1).
		Inc("likes", 2)

	result3, _ := articleRepo.UpdateByID(ctx, articleID, ub3.Build())
	fmt.Printf("  ✓ Modified: %d document(s)\n\n", result3.ModifiedCount)

	// Example 4: Push to array
	fmt.Println("4. Add new comments:")
	ub4 := mongo_kit.NewUpdateBuilder().
		Push("comments", "Very informative!").
		Push("comments", "Thanks for sharing!")

	result4, _ := articleRepo.UpdateByID(ctx, articleID, ub4.Build())
	fmt.Printf("  ✓ Modified: %d document(s)\n\n", result4.ModifiedCount)

	// Example 5: AddToSet (no duplicates)
	fmt.Println("5. Add tags (avoiding duplicates):")
	ub5 := mongo_kit.NewUpdateBuilder().
		AddToSet("tags", "tutorial").
		AddToSet("tags", "go"). // Already exists, won't be added
		AddToSet("tags", "beginner")

	result5, _ := articleRepo.UpdateByID(ctx, articleID, ub5.Build())
	fmt.Printf("  ✓ Modified: %d document(s)\n\n", result5.ModifiedCount)

	// Example 6: Pull from array
	fmt.Println("6. Remove specific tag:")
	ub6 := mongo_kit.NewUpdateBuilder().
		Pull("tags", "programming")

	result6, _ := articleRepo.UpdateByID(ctx, articleID, ub6.Build())
	fmt.Printf("  ✓ Modified: %d document(s)\n\n", result6.ModifiedCount)

	// Example 7: CurrentDate
	fmt.Println("7. Update last_view_at to current time:")
	ub7 := mongo_kit.NewUpdateBuilder().
		CurrentDate("last_view_at").
		Inc("views", 1)

	result7, _ := articleRepo.UpdateByID(ctx, articleID, ub7.Build())
	fmt.Printf("  ✓ Modified: %d document(s)\n\n", result7.ModifiedCount)

	// Example 8: Multiply
	fmt.Println("8. Apply discount (multiply price by 0.9):")
	// First add price field
	_, _ = articleRepo.UpdateByID(ctx, articleID, bson.M{"$set": bson.M{"price": 100.0}})

	ub8 := mongo_kit.NewUpdateBuilder().
		Mul("price", 0.9)

	result8, _ := articleRepo.UpdateByID(ctx, articleID, ub8.Build())
	fmt.Printf("  ✓ Modified: %d document(s)\n\n", result8.ModifiedCount)

	// Example 9: Min/Max
	fmt.Println("9. Update views (only if new value is greater):")
	ub9 := mongo_kit.NewUpdateBuilder().
		Max("views", 150) // Only updates if 150 > current views

	result9, _ := articleRepo.UpdateByID(ctx, articleID, ub9.Build())
	fmt.Printf("  ✓ Modified: %d document(s)\n\n", result9.ModifiedCount)

	// Example 10: Unset fields
	fmt.Println("10. Remove price field:")
	ub10 := mongo_kit.NewUpdateBuilder().
		Unset("price")

	result10, _ := articleRepo.UpdateByID(ctx, articleID, ub10.Build())
	fmt.Printf("  ✓ Modified: %d document(s)\n\n", result10.ModifiedCount)

	// Example 11: Rename field
	fmt.Println("11. Rename 'content' to 'body':")
	ub11 := mongo_kit.NewUpdateBuilder().
		Rename("content", "body")

	result11, _ := articleRepo.UpdateByID(ctx, articleID, ub11.Build())
	fmt.Printf("  ✓ Modified: %d document(s)\n\n", result11.ModifiedCount)

	// Example 12: Complex update with multiple operations
	fmt.Println("12. Complex update combining multiple operations:")
	ub12 := mongo_kit.NewUpdateBuilder().
		Set("published", true).
		Set("updated_at", time.Now()).
		Inc("views", 5).
		Inc("likes", 1).
		Push("tags", "featured").
		CurrentDate("last_view_at")

	result12, _ := articleRepo.UpdateByID(ctx, articleID, ub12.Build())
	fmt.Printf("  ✓ Modified: %d document(s)\n\n", result12.ModifiedCount)

	// Example 13: Update many documents
	fmt.Println("13. Bulk update - Publish all unpublished articles:")
	// Create some unpublished articles first
	unpublished := []Article{
		{Title: "Draft 1", Content: "...", Author: "Author 1", Published: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Title: "Draft 2", Content: "...", Author: "Author 2", Published: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	_, _ = articleRepo.CreateMany(ctx, unpublished)

	ub13 := mongo_kit.NewUpdateBuilder().
		Set("published", true).
		CurrentDate("updated_at")

	filter := mongo_kit.NewQueryBuilder().Equals("published", false).GetFilter()
	result13, _ := articleRepo.UpdateMany(ctx, filter, ub13.Build())
	fmt.Printf("  ✓ Modified: %d document(s)\n\n", result13.ModifiedCount)

	// Display final article state
	fmt.Println("Final article state:")
	finalArticle, _ := articleRepo.FindByID(ctx, articleID)
	fmt.Printf("  - Title: %s\n", finalArticle.Title)
	fmt.Printf("  - Author: %s\n", finalArticle.Author)
	fmt.Printf("  - Views: %d\n", finalArticle.Views)
	fmt.Printf("  - Likes: %d\n", finalArticle.Likes)
	fmt.Printf("  - Tags: %v\n", finalArticle.Tags)
	fmt.Printf("  - Comments: %d\n", len(finalArticle.Comments))
	fmt.Printf("  - Published: %v\n", finalArticle.Published)
}
