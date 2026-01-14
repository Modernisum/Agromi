package community

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"Agromi/database"
	community_models "Agromi/routes/community/models"
	"Agromi/utils" // Assuming Haversine is here

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// GetFeed returns scored community posts
func GetFeed(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	query := c.Query("query")

	userLat, _ := strconv.ParseFloat(latStr, 64)
	userLon, _ := strconv.ParseFloat(lonStr, 64)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := database.GetCollection("community_posts")

	// Basic filter (can extend to text search if query provided)
	filter := bson.M{}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var posts []community_models.Post
	if err = cursor.All(ctx, &posts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing posts"})
		return
	}

	// Constants for Weighting
	const (
		W_Rel   = 1.0 // Relevance
		W_Dist  = 0.3 // Distance
		W_Rate  = 0.3 // Rating
		W_Fresh = 0.4 // Freshness
	)

	for i := range posts {
		p := &posts[i]

		// 1. Relevance
		relevanceScore := 1.0
		if query != "" {
			// Simple check: if query in content, boost score
			// In prod: Use MongoDB Text Search Score
			// Here: Manual placeholder
			if contains(p.Content, query) {
				relevanceScore = 2.0
			} else {
				relevanceScore = 0.1
			}
		}

		// 2. Distance
		distScore := 0.0
		if p.Location != nil && len(p.Location.Coordinates) == 2 {
			dist := utils.Haversine(userLat, userLon, p.Location.Coordinates[1], p.Location.Coordinates[0])
			// Normalize: Closer is better. 100km max?
			if dist < 100 {
				distScore = (100.0 - dist) / 100.0
			}
		}

		// 3. Rating
		// Normalized 0-5 -> 0-1
		ratingScore := p.SenderRating / 5.0

		// 4. Freshness
		// Hours since posted. Decay.
		hours := time.Since(p.CreatedAt).Hours()
		freshnessScore := 1.0 / (1.0 + hours/24.0) // Drops over days

		// Total Score
		p.Score = (W_Rel * relevanceScore) + (W_Dist * distScore) + (W_Rate * ratingScore) + (W_Fresh * freshnessScore)
	}

	// Sort Descending by Score
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Score > posts[j].Score
	})

	c.JSON(http.StatusOK, posts)
}

func contains(text, sub string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(sub))
}

func RegisterRoutes(router *gin.RouterGroup) {
	commGroup := router.Group("/api/community")
	{
		commGroup.POST("/create", CreatePost)
		commGroup.GET("/feed", GetFeed)
		commGroup.DELETE("/delete/:id", DeletePost)
	}
}
