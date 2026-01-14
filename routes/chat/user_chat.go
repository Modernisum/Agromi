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
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SendMessage handles 1-on-1 and Group messages
func SendMessage(c *gin.Context) {
	var body struct {
		SenderID   string `json:"sender_id" binding:"required"`
		ReceiverID string `json:"receiver_id"` // Optional (if 1-on-1)
		GroupID    string `json:"group_id"`    // Optional (if Group)
		Content    string `json:"content" binding:"required"`
		MediaURL   string `json:"media_url"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	senderOID, _ := primitive.ObjectIDFromHex(body.SenderID)

	var receiverOID primitive.ObjectID
	var groupOID primitive.ObjectID

	if body.GroupID != "" {
		groupOID, _ = primitive.ObjectIDFromHex(body.GroupID)
	} else if body.ReceiverID != "" {
		receiverOID, _ = primitive.ObjectIDFromHex(body.ReceiverID)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either receiver_id or group_id required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := database.GetCollection("messages")

	// 1. Auto-Pruning Logic (Admin Limit)
	// Check count of messages in this conversation
	filter := bson.M{}
	if body.GroupID != "" {
		filter = bson.M{"group_id": groupOID}
	} else {
		// 1-on-1: Match both sender and receiver pair (A->B OR B->A)
		// For simplicity/capacity, we prune based on conversation "Thread"?
		// Or simpler: Prune ANY sender/receiver combo? No, that deletes others' data.
		// Correct 1-on-1 filter:
		filter = bson.M{
			"$or": []bson.M{
				{"sender_id": senderOID, "receiver_id": receiverOID},
				{"sender_id": receiverOID, "receiver_id": senderOID},
			},
		}
	}

	count, _ := coll.CountDocuments(ctx, filter)
	if count >= chat_models.MaxMessagesPerChat {
		// Delete Oldest
		limit := int64(count - chat_models.MaxMessagesPerChat + 1) // Remove excess + 1 (for new msg)

		// Find oldest IDs to delete
		findOpts := options.Find().SetSort(bson.M{"created_at": 1}).SetLimit(limit).SetProjection(bson.M{"_id": 1})
		cursor, _ := coll.Find(ctx, filter, findOpts)

		var docs []struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		if err := cursor.All(ctx, &docs); err == nil {
			idsToDelete := make([]primitive.ObjectID, len(docs))
			for i, d := range docs {
				idsToDelete[i] = d.ID
			}

			if len(idsToDelete) > 0 {
				coll.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": idsToDelete}})
			}
		}
	}

	// 2. Insert New Message
	msg := chat_models.Message{
		ID:         primitive.NewObjectID(),
		SenderID:   senderOID,
		ReceiverID: receiverOID,
		GroupID:    groupOID,
		Content:    body.Content,
		MediaURL:   body.MediaURL,
		CreatedAt:  time.Now(),
	}

	_, err := coll.InsertOne(ctx, msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Sent", "id": msg.ID})
}

// GetHistory fetches messages
func GetHistory(c *gin.Context) {
	userIDStr := c.Query("user_id")
	otherIDStr := c.Query("other_id") // Can be UserID or GroupID
	isGroup := c.Query("is_group") == "true"

	userOID, _ := primitive.ObjectIDFromHex(userIDStr)
	otherOID, _ := primitive.ObjectIDFromHex(otherIDStr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	if isGroup {
		filter = bson.M{"group_id": otherOID}
	} else {
		filter = bson.M{
			"$or": []bson.M{
				{"sender_id": userOID, "receiver_id": otherOID},
				{"sender_id": otherOID, "receiver_id": userOID},
			},
		}
	}

	opts := options.Find().SetSort(bson.M{"created_at": 1}).SetLimit(100)
	cursor, _ := database.GetCollection("messages").Find(ctx, filter, opts)

	var messages []chat_models.Message
	cursor.All(ctx, &messages)

	c.JSON(http.StatusOK, messages)
}
