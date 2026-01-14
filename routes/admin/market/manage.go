package admin_market

import (
	"context"
	"net/http"
	"time"

	"Agromi/database"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateProductFields helper to update specific fields
func updateProduct(c *gin.Context, update bson.M) {
	idHex := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := database.GetCollection("market_products")
	_, err = coll.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}

// BlockProduct
func BlockProduct(c *gin.Context) {
	updateProduct(c, bson.M{"is_blocked": true})
}

// UnblockProduct
func UnblockProduct(c *gin.Context) {
	updateProduct(c, bson.M{"is_blocked": false})
}

// SponsorProduct
func SponsorProduct(c *gin.Context) {
	updateProduct(c, bson.M{"is_sponsored": true})
}

// ChangePriority
func ChangePriority(c *gin.Context) {
	var body struct {
		Priority int `json:"priority"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updateProduct(c, bson.M{"priority": body.Priority})
}

// DeleteProduct
func DeleteProduct(c *gin.Context) {
	idHex := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := database.GetCollection("market_products")
	_, err = coll.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
}

func RegisterManageRoutes(router *gin.RouterGroup) {
	manageGroup := router.Group("/manage")
	{
		manageGroup.PUT("/block/:id", BlockProduct)
		manageGroup.PUT("/unblock/:id", UnblockProduct)
		manageGroup.PUT("/sponsor/:id", SponsorProduct)
		manageGroup.PUT("/priority/:id", ChangePriority)
		manageGroup.DELETE("/delete/:id", DeleteProduct)
	}
}
