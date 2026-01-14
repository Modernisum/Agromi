package admin_consultant

import (
	"context"
	"net/http"
	"time"

	"Agromi/database"
	"Agromi/routes/consultant/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// GetAnalytics returns simplified stats
func GetAnalytics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll := database.GetCollection("consultants")

	total, _ := coll.CountDocuments(ctx, bson.M{})
	verified, _ := coll.CountDocuments(ctx, bson.M{"verification_status": models.StatusVerified})
	blocked, _ := coll.CountDocuments(ctx, bson.M{"is_blocked": true})

	// Group by Type (simple aggregation)
	pipeline := []bson.M{
		{"$group": bson.M{"_id": "$type", "count": bson.M{"$sum": 1}}},
	}
	cursor, _ := coll.Aggregate(ctx, pipeline)

	var typeStats []bson.M
	if cursor != nil {
		cursor.All(ctx, &typeStats)
	}

	c.JSON(http.StatusOK, gin.H{
		"total_consultants": total,
		"verified_count":    verified,
		"blocked_count":     blocked,
		"by_type":           typeStats,
	})
}

func RegisterAnalyticsRoutes(router *gin.RouterGroup) {
	router.GET("/analytics", GetAnalytics)
}
