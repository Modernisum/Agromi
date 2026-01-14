package admin_sell

import (
	"context"
	"net/http"
	"time"

	"Agromi/database"
	market "Agromi/routes/market/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// ManageSellItems - Monitor user listings
func ManageSellItems(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := database.GetCollection("market_products")
	filter := bson.M{"type": market.TypeSell}

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

	c.JSON(http.StatusOK, products)
}

func RegisterRoutes(router *gin.RouterGroup) {
	sellGroup := router.Group("/sell")
	{
		sellGroup.GET("/list", ManageSellItems)
	}
}
