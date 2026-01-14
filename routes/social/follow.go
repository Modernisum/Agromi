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

// FollowUser
func FollowUser(c *gin.Context) {
	var body struct {
		FollowerID string `json:"follower_id" binding:"required"`
		FolloweeID string `json:"followee_id" binding:"required"` // Target User
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	followerID, _ := primitive.ObjectIDFromHex(body.FollowerID)
	followeeID, _ := primitive.ObjectIDFromHex(body.FolloweeID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := database.GetCollection("follows")

	// Check if already following
	count, _ := coll.CountDocuments(ctx, bson.M{"follower_id": followerID, "followee_id": followeeID})
	if count > 0 {
		// Unfollow
		coll.DeleteOne(ctx, bson.M{"follower_id": followerID, "followee_id": followeeID})
		c.JSON(http.StatusOK, gin.H{"message": "Unfollowed"})
		return
	}

	// Follow
	follow := social_models.Follow{
		ID:         primitive.NewObjectID(),
		FollowerID: followerID,
		FolloweeID: followeeID,
		CreatedAt:  time.Now(),
	}
	coll.InsertOne(ctx, follow)

	createNotification(ctx, followeeID, "follow", "You have a new follower!", followerID)

	c.JSON(http.StatusOK, gin.H{"message": "Followed"})
}

// NotifyFollowersOfNewPost (Internal Helper)
func NotifyFollowersOfNewPost(userIDStr string, postIDStr string, postTitle string) {
	userID, _ := primitive.ObjectIDFromHex(userIDStr)
	postID, _ := primitive.ObjectIDFromHex(postIDStr)
	ctx := context.TODO()

	coll := database.GetCollection("follows")
	cursor, _ := coll.Find(ctx, bson.M{"followee_id": userID})

	var follows []social_models.Follow
	cursor.All(ctx, &follows)

	for _, f := range follows {
		createNotification(ctx, f.FollowerID, "new_post", "New post by someone you follow: "+postTitle, postID)
	}
}

func RegisterFollowRoutes(router *gin.RouterGroup) {
	router.POST("/follow", FollowUser)
	// Add /list followers/following endpoints as needed
}
