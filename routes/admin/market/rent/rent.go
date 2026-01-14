package admin_rent

import (
	"context"
	"net/http"
	"time"

	"Agromi/database"
	market "Agromi/routes/market/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AddRentItem - Admin adds rental equipment
func AddRentItem(c *gin.Context) {
	var product market.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product.ID = primitive.NewObjectID()
	product.Type = market.TypeRent
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := database.GetCollection("market_products")
	_, err := coll.InsertOne(ctx, product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add rental item"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Rental item added", "id": product.ID})
}

// ManageRentItems - List all rental items
func ManageRentItems(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := database.GetCollection("market_products")
	filter := bson.M{"type": market.TypeRent}

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
	rentGroup := router.Group("/rent")
	{
		rentGroup.POST("/add", AddRentItem)
		rentGroup.GET("/list", ManageRentItems)
	}
}
