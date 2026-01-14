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

// Helper to Create Notification
func createNotification(ctx context.Context, recipientID primitive.ObjectID, notifType, message string, relatedID primitive.ObjectID) {
	coll := database.GetCollection("notifications")
	notif := social_models.Notification{
		ID:          primitive.NewObjectID(),
		RecipientID: recipientID,
		Type:        notifType,
		Message:     message,
		RelatedID:   relatedID,
		IsRead:      false,
		CreatedAt:   time.Now(),
	}
	coll.InsertOne(ctx, notif)
}

// CreateComment
func CreateComment(c *gin.Context) {
	var body struct {
		TargetID   string `json:"target_id" binding:"required"`
		SenderID   string `json:"sender_id" binding:"required"`
		SenderName string `json:"sender_name" binding:"required"`
		Text       string `json:"text" binding:"required"`
		MediaURL   string `json:"media_url"`
		OwnerID    string `json:"owner_id"` // ID of the user who owns the Target (Product/Profile) to notify
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetObjID, _ := primitive.ObjectIDFromHex(body.TargetID)
	senderObjID, _ := primitive.ObjectIDFromHex(body.SenderID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	comment := social_models.Comment{
		ID:         primitive.NewObjectID(),
		TargetID:   targetObjID,
		SenderID:   senderObjID,
		SenderName: body.SenderName,
		Text:       body.Text,
		MediaURL:   body.MediaURL,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	coll := database.GetCollection("comments")
	_, err := coll.InsertOne(ctx, comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to post comment"})
		return
	}

	// Notify Owner
	if body.OwnerID != "" && body.OwnerID != body.SenderID {
		ownerObjID, _ := primitive.ObjectIDFromHex(body.OwnerID)
		createNotification(ctx, ownerObjID, "comment", body.SenderName+" commented on your post.", comment.ID)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Comment posted", "id": comment.ID})
}

// ListComments
func ListComments(c *gin.Context) {
	targetID := c.Query("target_id")
	if targetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_id required"})
		return
	}

	objID, _ := primitive.ObjectIDFromHex(targetID)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := database.GetCollection("comments")
	cursor, err := coll.Find(ctx, bson.M{"target_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Error"})
		return
	}

	var comments []social_models.Comment
	if err = cursor.All(ctx, &comments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Parse Error"})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// UpdateComment (Sender Only)
func UpdateComment(c *gin.Context) {
	var body struct {
		ID       string `json:"id" binding:"required"`
		SenderID string `json:"sender_id" binding:"required"`
		Text     string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	commentID, _ := primitive.ObjectIDFromHex(body.ID)
	senderID, _ := primitive.ObjectIDFromHex(body.SenderID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := database.GetCollection("comments")
	res, err := coll.UpdateOne(ctx, bson.M{"_id": commentID, "sender_id": senderID}, bson.M{"$set": bson.M{"text": body.Text, "updated_at": time.Now()}})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized or comment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment updated"})
}

// DeleteComment (Sender or Admin)
func DeleteComment(c *gin.Context) {
	id := c.Param("id")              // Comment ID
	senderID := c.Query("sender_id") // If provided, ensures ownership. If empty, assumes Admin overrides (in real app, check context)

	commentID, _ := primitive.ObjectIDFromHex(id)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := database.GetCollection("comments")

	filter := bson.M{"_id": commentID}
	if senderID != "" {
		sID, _ := primitive.ObjectIDFromHex(senderID)
		filter["sender_id"] = sID
	}

	res, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}
	if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found or unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
}

func RegisterCommentRoutes(router *gin.RouterGroup) {
	router.POST("/comment/create", CreateComment)
	router.GET("/comment/list", ListComments)
	router.PUT("/comment/update", UpdateComment)
	router.DELETE("/comment/delete/:id", DeleteComment)
}
