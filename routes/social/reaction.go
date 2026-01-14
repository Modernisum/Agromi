package social

import (
	"context"
	"net/http"
	"time"

	"Agromi/database"
	social_models "Agromi/routes/social/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ToggleLike
func ToggleLike(c *gin.Context) {
	var body struct {
		TargetID string `json:"target_id" binding:"required"`
		SenderID string `json:"sender_id" binding:"required"`
		Action   string `json:"action" binding:"required"` // "like" or "dislike"
		OwnerID  string `json:"owner_id"`                  // To notify
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetID, _ := primitive.ObjectIDFromHex(body.TargetID)
	senderID, _ := primitive.ObjectIDFromHex(body.SenderID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := database.GetCollection("likes")

	// Check if exists
	filter := bson.M{"target_id": targetID, "sender_id": senderID}
	var existing social_models.Like
	err := coll.FindOne(ctx, filter).Decode(&existing)

	if err == nil {
		// Update existing
		coll.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"action": body.Action}})
	} else {
		// Insert new
		like := social_models.Like{
			ID:        primitive.NewObjectID(),
			TargetID:  targetID,
			SenderID:  senderID,
			Action:    body.Action,
			CreatedAt: time.Now(),
		}
		coll.InsertOne(ctx, like)

		// Notify if new like
		if body.OwnerID != "" && body.OwnerID != body.SenderID && body.Action == "like" {
			ownerObjID, _ := primitive.ObjectIDFromHex(body.OwnerID)
			createNotification(ctx, ownerObjID, "like", "Someone liked your post.", targetID)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reaction updated"})
}

// AddReview
func AddReview(c *gin.Context) {
	var body struct {
		TargetID string  `json:"target_id" binding:"required"` // ConsultantID or ProductID
		SenderID string  `json:"sender_id" binding:"required"`
		Rating   float64 `json:"rating" binding:"required"`
		Text     string  `json:"text"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetID, _ := primitive.ObjectIDFromHex(body.TargetID)
	senderID, _ := primitive.ObjectIDFromHex(body.SenderID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := database.GetCollection("reviews")

	// Upsert Review
	filter := bson.M{"target_id": targetID, "sender_id": senderID}
	update := bson.M{
		"$set": bson.M{
			"rating":     body.Rating,
			"text":       body.Text,
			"updated_at": time.Now(),
		},
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
			"_id":        primitive.NewObjectID(),
		},
	}
	_, err := coll.UpdateOne(ctx, filter, update, database.UpsertOpt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save review"})
		return
	}

	// Recalculate Average (Simple approach)
	go updateAverageRating(targetID)

	c.JSON(http.StatusOK, gin.H{"message": "Review saved"})
}

func updateAverageRating(targetID primitive.ObjectID) {
	ctx := context.TODO()
	coll := database.GetCollection("reviews")

	pipeline := []bson.M{
		{"$match": bson.M{"target_id": targetID}},
		{"$group": bson.M{"_id": "$target_id", "avgRating": bson.M{"$avg": "$rating"}, "count": bson.M{"$sum": 1}}},
	}

	cursor, _ := coll.Aggregate(ctx, pipeline)
	var result struct {
		AvgRating float64 `bson:"avgRating"`
		Count     int     `bson:"count"`
	}
	if cursor.Next(ctx) {
		cursor.Decode(&result)

		// Update Target (Generic approach - try updating Consultant and Product collections)
		database.GetCollection("consultants").UpdateOne(ctx, bson.M{"_id": targetID}, bson.M{"$set": bson.M{"rating": result.AvgRating, "review_count": result.Count}})
		database.GetCollection("market_products").UpdateOne(ctx, bson.M{"_id": targetID}, bson.M{"$set": bson.M{"rating": result.AvgRating, "review_count": result.Count}})
	}
}

func RegisterReactionRoutes(router *gin.RouterGroup) {
	router.POST("/reaction/like", ToggleLike)
	router.POST("/reaction/review", AddReview)
}
