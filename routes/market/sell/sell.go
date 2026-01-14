package sell

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

// CreateListing allows a farmer to sell/rent out an item
func CreateListing(c *gin.Context) {
	var product market.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	product.ID = primitive.NewObjectID()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	product.IsBlocked = false
	product.IsSponsored = false
	product.Rating = 0
	product.ReviewCount = 0

	// Validate Type
	if product.Type != market.TypeSell && product.Type != market.TypeRent {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid listing type. Must be 'sell' or 'rent'."})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := database.GetCollection("market_products")
	_, err := coll.InsertOne(ctx, product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create listing"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Listing created successfully", "id": product.ID})
}

// ListSellItems returns items for sale by farmers
func ListSellItems(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := database.GetCollection("market_products")
	filter := bson.M{"type": market.TypeSell, "is_blocked": false}

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
		sellGroup.POST("/create", CreateListing)
		sellGroup.GET("/list", ListSellItems)
	}
}
