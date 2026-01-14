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

// GetNotifications
func GetNotifications(c *gin.Context) {
	userID := c.Query("user_id")
	sinceStr := c.Query("since")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}
	uID, _ := primitive.ObjectIDFromHex(userID)

	filter := bson.M{"recipient_id": uID}

	if sinceStr != "" {
		layout := time.RFC3339
		t, err := time.Parse(layout, sinceStr)
		if err == nil {
			filter["created_at"] = bson.M{"$gt": t}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := database.GetCollection("notifications")
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Error"})
		return
	}

	var notifs []social_models.Notification
	cursor.All(ctx, &notifs)

	c.JSON(http.StatusOK, notifs)
}

func RegisterNotificationRoutes(router *gin.RouterGroup) {
	router.GET("/notification/list", GetNotifications)
}
