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

// Order represents an order document
type Order struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	CustomerID string             `bson:"customer_id"`
	Product    string             `bson:"product"`
	Quantity   int                `bson:"quantity"`
	Price      float64            `bson:"price"`
	Total      float64            `bson:"total"`
	Status     string             `bson:"status"`
	OrderDate  time.Time          `bson:"order_date"`
}

// AggregationResult represents aggregation output
type AggregationResult struct {
	ID         any     `bson:"_id"`
	TotalSales float64 `bson:"total_sales"`
	Count      int     `bson:"count"`
	AvgAmount  float64 `bson:"avg_amount,omitempty"`
	MinAmount  float64 `bson:"min_amount,omitempty"`
	MaxAmount  float64 `bson:"max_amount,omitempty"`
}

func main() {
	// Setup
	cfg := mongo_kit.DefaultConfig()
	client, err := mongo_kit.New(
		cfg,
		mongo_kit.WithURI("mongodb://localhost:27017"),
		mongo_kit.WithDatabase("sales"),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer func() {
		if err := client.Close(context.Background()); err != nil {
			log.Printf("Failed to close client: %v", err)
		}
	}()

	orderRepo := mongo_kit.NewRepository[Order](client, "orders")
	// For aggregation results, use primitive.M since results have different structure
	aggRepo := mongo_kit.NewRepository[primitive.M](client, "orders")
	ctx := context.Background()

	// Clean and insert sample orders
	_ = orderRepo.Drop(ctx)
	orders := []Order{
		{CustomerID: "C001", Product: "Laptop", Quantity: 1, Price: 1200, Total: 1200, Status: "completed", OrderDate: time.Now().AddDate(0, -1, 0)},
		{CustomerID: "C002", Product: "Mouse", Quantity: 2, Price: 25, Total: 50, Status: "completed", OrderDate: time.Now().AddDate(0, -1, -5)},
		{CustomerID: "C001", Product: "Keyboard", Quantity: 1, Price: 80, Total: 80, Status: "completed", OrderDate: time.Now().AddDate(0, -1, -3)},
		{CustomerID: "C003", Product: "Monitor", Quantity: 2, Price: 300, Total: 600, Status: "completed", OrderDate: time.Now().AddDate(0, -2, 0)},
		{CustomerID: "C002", Product: "Laptop", Quantity: 1, Price: 1200, Total: 1200, Status: "pending", OrderDate: time.Now()},
		{CustomerID: "C004", Product: "Mouse", Quantity: 5, Price: 25, Total: 125, Status: "completed", OrderDate: time.Now().AddDate(0, -1, -10)},
		{CustomerID: "C003", Product: "Keyboard", Quantity: 2, Price: 80, Total: 160, Status: "completed", OrderDate: time.Now().AddDate(0, -2, -5)},
		{CustomerID: "C001", Product: "Monitor", Quantity: 1, Price: 300, Total: 300, Status: "cancelled", OrderDate: time.Now().AddDate(0, -3, 0)},
	}
	_, _ = orderRepo.CreateMany(ctx, orders)

	fmt.Println("=== Aggregation Builder Examples ===")

	// Example 1: Group by customer and sum totals
	fmt.Println("1. Total sales by customer:")
	ab1 := mongo_kit.NewAggregationBuilder().
		Match(bson.M{"status": "completed"}).
		Group("$customer_id", bson.M{
			"total_sales": bson.M{"$sum": "$total"},
			"order_count": bson.M{"$sum": 1},
		}).
		Sort(bson.D{{Key: "total_sales", Value: -1}}) // descending

	results1, _ := aggRepo.Aggregate(ctx, ab1.Build())
	for _, result := range results1 {
		data := result
		fmt.Printf("  - Customer %s: $%.2f (%d orders)\n",
			data["_id"], data["total_sales"], data["order_count"])
	}
	fmt.Println()

	// Example 2: Group by product and calculate statistics
	fmt.Println("2. Product sales statistics:")
	ab2 := mongo_kit.NewAggregationBuilder().
		Match(bson.M{"status": "completed"}).
		Group("$product", bson.M{
			"total_quantity": bson.M{"$sum": "$quantity"},
			"total_revenue":  bson.M{"$sum": "$total"},
			"avg_price":      bson.M{"$avg": "$price"},
			"min_price":      bson.M{"$min": "$price"},
			"max_price":      bson.M{"$max": "$price"},
			"order_count":    bson.M{"$sum": 1},
		}).
		Sort(bson.D{{Key: "total_revenue", Value: -1}})

	results2, _ := aggRepo.Aggregate(ctx, ab2.Build())
	for _, result := range results2 {
		data := result
		fmt.Printf("  - %s:\n", data["_id"])
		fmt.Printf("    Quantity sold: %d\n", data["total_quantity"])
		fmt.Printf("    Revenue: $%.2f\n", data["total_revenue"])
		fmt.Printf("    Avg price: $%.2f\n", data["avg_price"])
		fmt.Printf("    Orders: %d\n", data["order_count"])
	}
	fmt.Println()

	// Example 3: Filter, group, and limit
	fmt.Println("3. Top 3 customers by spending (completed orders only):")
	ab3 := mongo_kit.NewAggregationBuilder().
		Match(bson.M{"status": "completed"}).
		Group("$customer_id", bson.M{
			"total_spent": bson.M{"$sum": "$total"},
		}).
		Sort(bson.D{{Key: "total_spent", Value: -1}}).
		Limit(3)

	results3, _ := aggRepo.Aggregate(ctx, ab3.Build())
	for i, result := range results3 {
		data := result
		fmt.Printf("  %d. Customer %s: $%.2f\n", i+1, data["_id"], data["total_spent"])
	}
	fmt.Println()

	// Example 4: Group by status and count
	fmt.Println("4. Orders by status:")
	ab4 := mongo_kit.NewAggregationBuilder().
		Group("$status", bson.M{
			"count": bson.M{"$sum": 1},
			"total": bson.M{"$sum": "$total"},
		}).
		Sort(bson.D{{Key: "count", Value: -1}})

	results4, _ := aggRepo.Aggregate(ctx, ab4.Build())
	for _, result := range results4 {
		data := result
		fmt.Printf("  - %s: %d orders ($%.2f)\n",
			data["_id"], data["count"], data["total"])
	}
	fmt.Println()

	// Example 5: Match, project, and sort
	fmt.Println("5. Large orders (>$500) with custom fields:")
	ab5 := mongo_kit.NewAggregationBuilder().
		Match(bson.M{"total": bson.M{"$gte": 500}}).
		Project(bson.M{
			"customer_id": 1,
			"product":     1,
			"total":       1,
			"is_premium":  bson.M{"$gte": []any{"$total", 1000}},
		}).
		Sort(bson.D{{Key: "total", Value: -1}})

	results5, _ := aggRepo.Aggregate(ctx, ab5.Build())
	for _, result := range results5 {
		data := result
		premium := "No"
		if data["is_premium"].(bool) {
			premium = "Yes"
		}
		fmt.Printf("  - Customer %s: %s ($%.2f) - Premium: %s\n",
			data["customer_id"], data["product"], data["total"], premium)
	}
	fmt.Println()

	// Example 6: Date-based aggregation
	fmt.Println("6. Orders per month:")
	ab6 := mongo_kit.NewAggregationBuilder().
		Match(bson.M{"status": "completed"}).
		Group(bson.M{
			"year":  bson.M{"$year": "$order_date"},
			"month": bson.M{"$month": "$order_date"},
		}, bson.M{
			"count":  bson.M{"$sum": 1},
			"total":  bson.M{"$sum": "$total"},
		}).
		Sort(bson.D{
			{Key: "_id.year", Value: -1},
			{Key: "_id.month", Value: -1},
		})

	results6, _ := aggRepo.Aggregate(ctx, ab6.Build())
	for _, result := range results6 {
		data := result
		id := data["_id"].(primitive.M)
		fmt.Printf("  - %d-%02d: %d orders ($%.2f)\n",
			id["year"], id["month"], data["count"], data["total"])
	}
	fmt.Println()

	// Example 7: Skip and limit for pagination
	fmt.Println("7. Orders with pagination (skip 2, limit 3):")
	ab7 := mongo_kit.NewAggregationBuilder().
		Match(bson.M{"status": "completed"}).
		Sort(bson.D{{Key: "total", Value: -1}}).
		Skip(2).
		Limit(3).
		Project(bson.M{
			"customer_id": 1,
			"product":     1,
			"total":       1,
		})

	results7, _ := aggRepo.Aggregate(ctx, ab7.Build())
	for i, result := range results7 {
		data := result
		fmt.Printf("  %d. Customer %s: %s ($%.2f)\n",
			i+1, data["customer_id"], data["product"], data["total"])
	}
	fmt.Println()

	// Example 8: Complex multi-stage aggregation
	fmt.Println("8. Customer insights (multi-stage):")
	ab8 := mongo_kit.NewAggregationBuilder().
		Match(bson.M{"status": "completed"}).
		Group("$customer_id", bson.M{
			"total_spent":    bson.M{"$sum": "$total"},
			"order_count":    bson.M{"$sum": 1},
			"avg_order_size": bson.M{"$avg": "$total"},
		}).
		Match(bson.M{"order_count": bson.M{"$gte": 2}}).
		Project(bson.M{
			"customer_id":    "$_id",
			"total_spent":    1,
			"order_count":    1,
			"avg_order_size": bson.M{"$round": []any{"$avg_order_size", 2}},
			"_id":            0,
		}).
		Sort(bson.D{{Key: "total_spent", Value: -1}})

	results8, _ := aggRepo.Aggregate(ctx, ab8.Build())
	for _, result := range results8 {
		data := result
		fmt.Printf("  - Customer %s:\n", data["customer_id"])
		fmt.Printf("    Total spent: $%.2f\n", data["total_spent"])
		fmt.Printf("    Orders: %d\n", data["order_count"])
		fmt.Printf("    Avg order: $%.2f\n", data["avg_order_size"])
	}
	fmt.Println()

	// Example 9: AddStage for custom stages
	fmt.Println("9. Using custom stage (count all):")
	ab9 := mongo_kit.NewAggregationBuilder().
		Match(bson.M{"status": "completed"}).
		AddStage(bson.D{{Key: "$count", Value: "total"}})

	results9, _ := aggRepo.Aggregate(ctx, ab9.Build())
	if len(results9) > 0 {
		data := results9[0]
		fmt.Printf("  - Total completed orders: %d\n\n", data["total"])
	}
}
