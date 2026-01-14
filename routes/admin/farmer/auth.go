package farmer

import (
	"context"
	"net/http"
	"time"

	"Agromi/core/router"
	"Agromi/database"

	// Importing User model
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	router.Register(func(r *gin.Engine) {
		adminGroup := r.Group("/api/admin/farmer")
		// Middleware to check admin auth should be here, skipping for now as per instructions
		{
			adminGroup.PUT("/block/:id", blockFarmer)
			adminGroup.DELETE("/delete/:id", deleteFarmer)
			adminGroup.POST("/revoke-tokens/:id", revokeTokens)
		}
	})
}

func blockFarmer(c *gin.Context) {
	idStr := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input struct {
		Block bool `json:"block"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		// Default to true (block) if body not allowed/empty?
		// Actually let's assume toggle or specific value.
		// For simplicity, let's just say this endpoint BLOCKS.
		// To Unblock, we might need another param or endpoint.
		// Let's assume input determines it.
		// If binding fails, default to block=true
		input.Block = true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersColl := database.GetCollection("users")
	_, err = usersColl.UpdateOne(ctx, bson.M{"_id": objID, "user_type": "farmer"}, bson.M{"$set": bson.M{"is_blocked": input.Block}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	msg := "Farmer blocked"
	if !input.Block {
		msg = "Farmer unblocked"
	}
	c.JSON(http.StatusOK, gin.H{"message": msg})
}

func deleteFarmer(c *gin.Context) {
	idStr := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersColl := database.GetCollection("users")
	res, err := usersColl.DeleteOne(ctx, bson.M{"_id": objID, "user_type": "farmer"})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete farmer"})
		return
	}
	if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Farmer not found"})
		return
	}

	// Also delete sessions
	database.GetCollection("sessions").DeleteMany(ctx, bson.M{"user_id": objID})

	c.JSON(http.StatusOK, gin.H{"message": "Farmer deleted successfully"})
}

func revokeTokens(c *gin.Context) {
	idStr := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Delete all sessions for this user
	_, err = database.GetCollection("sessions").DeleteMany(ctx, bson.M{"user_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All tokens revoked for farmer"})
}
