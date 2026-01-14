package admin_buy

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

// AddBuyItem - Admin adds standard "Buy" items (Seeds, etc.)
func AddBuyItem(c *gin.Context) {
	var product market.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product.ID = primitive.NewObjectID()
	product.Type = market.TypeBuy
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := database.GetCollection("market_products")
	_, err := coll.InsertOne(ctx, product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Product added", "id": product.ID})
}

// ManageBuyItems - List all buy items for admin management
func ManageBuyItems(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := database.GetCollection("market_products")
	filter := bson.M{"type": market.TypeBuy}

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
	buyGroup := router.Group("/buy")
	{
		buyGroup.POST("/add", AddBuyItem)
		buyGroup.GET("/list", ManageBuyItems)
	}
}
