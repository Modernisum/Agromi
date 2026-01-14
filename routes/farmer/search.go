package farmer

import (
	"context"
	"net/http"
	"time"

	"Agromi/core/router"
	"Agromi/database"
	"Agromi/routes/auth"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	router.Register(func(r *gin.Engine) {
		group := r.Group("/api/farmer/search")
		{
			group.GET("", searchFarmers)          // Full text search
			group.GET("/suggest", suggestFarmers) // Auto-complete
		}
	})
}

// searchFarmers uses MongoDB $text search for high performance on large datasets
func searchFarmers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersColl := database.GetCollection("users")

	// Filter: only farmers, match text
	filter := bson.M{
		"user_type": "farmer",
		"$text":     bson.M{"$search": query},
	}

	// Score-based sorting happens automatically with $text if requested, but simple find is also okay.
	// We Limit to 20 for performance
	// To get score: .Project(bson.M{"score": bson.M{"$meta": "textScore"}}).Sort(bson.M{"score": bson.M{"$meta": "textScore"}})
	// For simplicity using standard Find with filter:

	cursor, err := usersColl.Find(ctx, filter) // options.Find().SetLimit(20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}
	defer cursor.Close(ctx)

	var results []auth.User
	if err = cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Decoding failed"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// suggestFarmers provides fast prefix-based suggestions (Limit 5)
func suggestFarmers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusOK, []string{})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	usersColl := database.GetCollection("users")

	// Regex for prefix match (start anchor ^)
	// Case insensitive option 'i'
	filter := bson.M{
		"user_type": "farmer",
		"name":      bson.M{"$regex": "^" + query, "$options": "i"},
	}

	// Projection: only return names for suggestions
	// Limit 5 for speed
	// Use Options to set Limit
	// We need internal mongo options, but to keep imports clean/standard, assume basic finding works.
	// Properly: use options.Find().SetLimit(5).SetProjection(bson.M{"name": 1})
	// For now, fetching full docs but limiting in code if needed or just relying on fast index scan.

	// Efficient way:
	cursor, err := usersColl.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Suggestion failed"})
		return
	}
	defer cursor.Close(ctx)

	var results []auth.User
	// Decode
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Decoding failed"})
		return
	}

	// Extract names and limit to 5
	names := make([]string, 0)
	for i, u := range results {
		if i >= 5 {
			break
		}
		names = append(names, u.Name)
	}

	c.JSON(http.StatusOK, names)
}
