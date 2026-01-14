package finance_routes

import (
	"context"
	"net/http"
	"time"

	"Agromi/database"
	"Agromi/routes/consultant/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// VerifyEntity allows admin to verify a Farmer, Consultant, or Product
func VerifyEntity(c *gin.Context) {
	var body struct {
		ID         string `json:"id" binding:"required"`
		Type       string `json:"type" binding:"required"` // "consultant", "farmer", "product"
		IsVerified bool   `json:"is_verified"`             // true/false for status
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objID, err := primitive.ObjectIDFromHex(body.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var collName string
	var update bson.M

	switch body.Type {
	case "consultant":
		collName = "consultants"
		status := models.StatusUnverified
		if body.IsVerified {
			status = models.StatusVerified
		}
		update = bson.M{"$set": bson.M{"verification_status": status, "updated_at": time.Now()}}
	case "farmer":
		collName = "users" // Assuming users collection
		// Verification field might not exist on Farmer yet, but we can add it dynamically
		update = bson.M{"$set": bson.M{"is_verified": body.IsVerified, "updated_at": time.Now()}}
	case "product":
		collName = "market_products"
		update = bson.M{"$set": bson.M{"is_verified": body.IsVerified, "updated_at": time.Now()}} // Assuming generic verified flag provided
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Type"})
		return
	}

	coll := database.GetCollection(collName)
	_, err = coll.UpdateOne(ctx, bson.M{"_id": objID}, update)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update verification status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification status updated", "entity": body.Type, "new_status": body.IsVerified})
}

func RegisterVerifyRoutes(router *gin.RouterGroup) {
	router.PUT("/verify", VerifyEntity) // /api/admin/finance/verify
}
