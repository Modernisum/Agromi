package finance_sponsor

import (
	"context"
	"net/http"
	"time"

	"Agromi/database"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateProductScore allows admin to manually set the score/priority of a product
func UpdateProductScore(c *gin.Context) {
	var body struct {
		ProductID string `json:"product_id" binding:"required"`
		Score     int    `json:"score" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objID, err := primitive.ObjectIDFromHex(body.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Product ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := database.GetCollection("market_products")

	// Updating 'priority' field as the score
	_, err = coll.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"priority": body.Score}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product score"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product score updated successfully", "new_score": body.Score})
}

func RegisterRoutes(router *gin.RouterGroup) {
	sponsorGroup := router.Group("/finance/sponsor")
	{
		sponsorGroup.PUT("/score", UpdateProductScore)
	}
}
