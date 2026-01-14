package community

import (
	"context"
	"net/http"
	"time"

	"Agromi/database"
	community_models "Agromi/routes/community/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreatePost
func CreatePost(c *gin.Context) {
	var body struct {
		SenderID     string   `json:"sender_id" binding:"required"`
		SenderName   string   `json:"sender_name" binding:"required"`
		Content      string   `json:"content" binding:"required"`
		MediaURL     string   `json:"media_url"`
		Tags         []string `json:"tags"`
		Lat          float64  `json:"lat" binding:"required"`
		Lon          float64  `json:"lon" binding:"required"`
		SenderRating float64  `json:"sender_rating"` // Optional: Client can send or we fetch
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	senderObjID, _ := primitive.ObjectIDFromHex(body.SenderID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	post := community_models.Post{
		ID:           primitive.NewObjectID(),
		SenderID:     senderObjID,
		SenderName:   body.SenderName,
		SenderRating: body.SenderRating,
		Content:      body.Content,
		MediaURL:     body.MediaURL,
		Tags:         body.Tags,
		Location: &community_models.GeoJSON{
			Type:        "Point",
			Coordinates: []float64{body.Lon, body.Lat},
		},
		LikesCount: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	coll := database.GetCollection("community_posts")
	_, err := coll.InsertOne(ctx, post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Post created", "id": post.ID})
}

// DeletePost
func DeletePost(c *gin.Context) {
	id := c.Param("id")
	postID, _ := primitive.ObjectIDFromHex(id)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := database.GetCollection("community_posts")
	res, err := coll.DeleteOne(ctx, bson.M{"_id": postID})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}
	if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted"})
}
