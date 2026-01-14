package buy

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

// ListBuyItems returns sorted buy items
func ListBuyItems(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := database.GetCollection("market_products")

	filter := bson.M{
		"type":       market.TypeBuy,
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
	// Score = 0.4*Rating + 0.3*LowPrice + 0.3*Freshness (simplified)
	type ScoredProduct struct {
		market.Product `json:",inline"`
		Score          float64 `json:"score"`
	}
	var scoredList []ScoredProduct

	for _, p := range products {
		score := 0.0
		// Rating (0-5) -> 40pts
		score += (p.Rating / 5.0) * 40.0

		// Low Price (Inverse, assume max 100k, lower good)
		// If price is 0 (unlikely but possible), strict checking needed
		if p.Price <= 100000 {
			score += ((100000.0 - p.Price) / 100000.0) * 30.0
		}

		// Priority (Admin boost)
		score += float64(p.Priority)

		scoredList = append(scoredList, ScoredProduct{Product: p, Score: score})
	}

	// Sort Descending
	sort.Slice(scoredList, func(i, j int) bool {
		return scoredList[i].Score > scoredList[j].Score
	})

	c.JSON(http.StatusOK, scoredList)
}

func RegisterRoutes(router *gin.RouterGroup) {
	buyGroup := router.Group("/buy")
	{
		buyGroup.GET("/list", ListBuyItems)
	}
}
