package main

import (
	"context"
	"fmt"
	"log"
	"time"

	mongo_kit "github.com/edaniel30/mongo-kit-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product represents a product document
type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Category    string             `bson:"category"`
	Price       float64            `bson:"price"`
	Stock       int                `bson:"stock"`
	Tags        []string           `bson:"tags"`
	Active      bool               `bson:"active"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

func main() {
	// Create client and repository
	cfg := mongo_kit.DefaultConfig()
	client, err := mongo_kit.New(
		cfg,
		mongo_kit.WithURI("mongodb://localhost:27017"),
		mongo_kit.WithDatabase("shop"),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer func() {
		if err := client.Close(context.Background()); err != nil {
			log.Printf("Failed to close client: %v", err)
		}
	}()

	productRepo := mongo_kit.NewRepository[Product](client, "products")
	ctx := context.Background()

	// Clean up and insert sample data
	_ = productRepo.Drop(ctx)
	products := []Product{
		{Name: "Laptop Pro", Category: "electronics", Price: 1299.99, Stock: 15, Tags: []string{"computer", "portable"}, Active: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Name: "Mouse Wireless", Category: "electronics", Price: 29.99, Stock: 50, Tags: []string{"peripheral", "wireless"}, Active: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Name: "Office Chair", Category: "furniture", Price: 249.99, Stock: 8, Tags: []string{"ergonomic", "office"}, Active: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Name: "Desk Lamp", Category: "furniture", Price: 45.99, Stock: 0, Tags: []string{"lighting", "office"}, Active: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Name: "Keyboard Mechanical", Category: "electronics", Price: 149.99, Stock: 25, Tags: []string{"peripheral", "gaming"}, Active: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Name: "Monitor 27\"", Category: "electronics", Price: 399.99, Stock: 12, Tags: []string{"display", "4k"}, Active: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	_, _ = productRepo.CreateMany(ctx, products)

	fmt.Println("=== Query Builder Examples ===")

	// Example 1: Simple equality query
	fmt.Println("1. Find products in 'electronics' category:")
	qb1 := mongo_kit.NewQueryBuilder().
		Equals("category", "electronics")

	results1, _ := productRepo.FindWithBuilder(ctx, qb1)
	for _, p := range results1 {
		fmt.Printf("  - %s: $%.2f\n", p.Name, p.Price)
	}
	fmt.Println()

	// Example 2: Range queries
	fmt.Println("2. Find products priced between $50 and $300:")
	qb2 := mongo_kit.NewQueryBuilder().
		GreaterThanOrEqual("price", 50.0).
		LessThanOrEqual("price", 300.0)

	results2, _ := productRepo.FindWithBuilder(ctx, qb2)
	for _, p := range results2 {
		fmt.Printf("  - %s: $%.2f\n", p.Name, p.Price)
	}
	fmt.Println()

	// Example 3: Combining filters with AND
	fmt.Println("3. Find active electronics with stock > 10:")
	qb3 := mongo_kit.NewQueryBuilder().
		Equals("category", "electronics").
		Equals("active", true).
		GreaterThan("stock", 10)

	results3, _ := productRepo.FindWithBuilder(ctx, qb3)
	for _, p := range results3 {
		fmt.Printf("  - %s (Stock: %d)\n", p.Name, p.Stock)
	}
	fmt.Println()

	// Example 4: Using In operator
	fmt.Println("4. Find products with specific tags:")
	qb4 := mongo_kit.NewQueryBuilder().
		In("tags", "gaming", "wireless")

	results4, _ := productRepo.FindWithBuilder(ctx, qb4)
	for _, p := range results4 {
		fmt.Printf("  - %s (Tags: %v)\n", p.Name, p.Tags)
	}
	fmt.Println()

	// Example 5: Sorting and pagination
	fmt.Println("5. Get top 3 most expensive products:")
	qb5 := mongo_kit.NewQueryBuilder().
		Sort("price", false). // descending
		Limit(3)

	results5, _ := productRepo.FindWithBuilder(ctx, qb5)
	for i, p := range results5 {
		fmt.Printf("  %d. %s: $%.2f\n", i+1, p.Name, p.Price)
	}
	fmt.Println()

	// Example 6: OR conditions
	fmt.Println("6. Find products that are either out of stock OR inactive:")
	condition1 := mongo_kit.NewQueryBuilder().Equals("stock", 0).GetFilter()
	condition2 := mongo_kit.NewQueryBuilder().Equals("active", false).GetFilter()

	qb6 := mongo_kit.NewQueryBuilder().
		Or(condition1, condition2)

	results6, _ := productRepo.FindWithBuilder(ctx, qb6)
	for _, p := range results6 {
		fmt.Printf("  - %s (Stock: %d, Active: %v)\n", p.Name, p.Stock, p.Active)
	}
	fmt.Println()

	// Example 7: Complex query with multiple conditions
	fmt.Println("7. Complex query - Active electronics under $200 with stock:")
	qb7 := mongo_kit.NewQueryBuilder().
		Equals("category", "electronics").
		Equals("active", true).
		LessThan("price", 200.0).
		GreaterThan("stock", 0).
		Sort("price", true). // ascending
		Limit(5)

	results7, _ := productRepo.FindWithBuilder(ctx, qb7)
	for _, p := range results7 {
		fmt.Printf("  - %s: $%.2f (Stock: %d)\n", p.Name, p.Price, p.Stock)
	}
	fmt.Println()

	// Example 8: Count with query builder
	fmt.Println("8. Count products by category:")
	countElectronics, _ := productRepo.CountWithBuilder(ctx,
		mongo_kit.NewQueryBuilder().Equals("category", "electronics"))
	countFurniture, _ := productRepo.CountWithBuilder(ctx,
		mongo_kit.NewQueryBuilder().Equals("category", "furniture"))

	fmt.Printf("  - Electronics: %d\n", countElectronics)
	fmt.Printf("  - Furniture: %d\n", countFurniture)
	fmt.Println()

	// Example 9: Exists check with query builder
	fmt.Println("9. Check if any product has price > $1000:")
	exists, _ := productRepo.ExistsWithBuilder(ctx,
		mongo_kit.NewQueryBuilder().GreaterThan("price", 1000.0))
	fmt.Printf("  - Expensive products exist: %v\n\n", exists)

	// Example 10: Regex search
	fmt.Println("10. Find products with 'Pro' or 'Chair' in name:")
	qb10 := mongo_kit.NewQueryBuilder().
		Regex("name", "Pro|Chair", "i") // case-insensitive

	results10, _ := productRepo.FindWithBuilder(ctx, qb10)
	for _, p := range results10 {
		fmt.Printf("  - %s\n", p.Name)
	}
	fmt.Println()
}
