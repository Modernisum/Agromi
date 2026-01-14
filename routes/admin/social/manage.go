package admin_social

import (
	"Agromi/core/router"
	"context"
	"net/http"
	"time"

	"Agromi/database"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeleteCommentAdmin
func DeleteCommentAdmin(c *gin.Context) {
	id := c.Param("id")
	objID, _ := primitive.ObjectIDFromHex(id)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := database.GetCollection("comments")
	_, err := coll.DeleteOne(ctx, bson.M{"_id": objID})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted by admin"})
}

func init() {
	router.Register(func(r *gin.Engine) {
		group := r.Group("/api/admin/social")
		{
			group.DELETE("/manage/comment/:id", DeleteCommentAdmin)
			// Add review deletion here if needed
		}
	})
}
