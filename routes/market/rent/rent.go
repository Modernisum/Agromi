package rent

import (
	"context"
	"net/http"
	"sort"
	"time"

	"Agromi/database"
	market "Agromi/routes/market/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// ListRentItems returns sorted rent items
func ListRentItems(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := database.GetCollection("market_products")

	filter := bson.M{
		"type":       market.TypeRent,
		"is_blocked": false,
	}

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer cursor.Close(ctx)

	var products []market.Product
	if err = cursor.All(ctx, &products); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing products"})
		return
	}

	// Scoring (Low Cost Preference)
	type ScoredProduct struct {
		market.Product `json:",inline"`
		Score          float64 `json:"score"`
	}
	var scoredList []ScoredProduct

	for _, p := range products {
		score := 0.0
		// Rating (0-5) -> 40pts
		score += (p.Rating / 5.0) * 40.0

		// Low Price per unit (Rent is cheaper, say max 5000)
		if p.Price <= 5000 {
			score += ((5000.0 - p.Price) / 5000.0) * 30.0
		}

		// Priority
		score += float64(p.Priority)

		scoredList = append(scoredList, ScoredProduct{Product: p, Score: score})
	}

	sort.Slice(scoredList, func(i, j int) bool {
		return scoredList[i].Score > scoredList[j].Score
	})

	c.JSON(http.StatusOK, scoredList)
}

func RegisterRoutes(router *gin.RouterGroup) {
	rentGroup := router.Group("/rent")
	{
		rentGroup.GET("/list", ListRentItems)
	}
}
