package chat

import (
	"context"
	"net/http"
	"time"

	"Agromi/database"
	chat_models "Agromi/routes/chat/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateGroup
func CreateGroup(c *gin.Context) {
	var body struct {
		Name    string `json:"name" binding:"required"`
		AdminID string `json:"admin_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminOID, _ := primitive.ObjectIDFromHex(body.AdminID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	group := chat_models.ChatGroup{
		ID:        primitive.NewObjectID(),
		Name:      body.Name,
		AdminID:   adminOID,
		MemberIDs: []primitive.ObjectID{adminOID}, // Admin is first member
		CreatedAt: time.Now(),
	}

	_, err := database.GetCollection("chat_groups").InsertOne(ctx, group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Group created", "id": group.ID})
}

// JoinGroup
func JoinGroup(c *gin.Context) {
	var body struct {
		GroupID string `json:"group_id" binding:"required"`
		UserID  string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	groupOID, _ := primitive.ObjectIDFromHex(body.GroupID)
	userOID, _ := primitive.ObjectIDFromHex(body.UserID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Add user to member_ids if not exists
	filter := bson.M{"_id": groupOID}
	update := bson.M{"$addToSet": bson.M{"member_ids": userOID}}

	res, err := database.GetCollection("chat_groups").UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join"})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Joined group"})
}
