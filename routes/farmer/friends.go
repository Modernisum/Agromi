package farmer

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"Agromi/core/router"
	"Agromi/database"
	"Agromi/routes/auth"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	router.Register(func(r *gin.Engine) {
		r.GET("/api/farmer/nearby", getNearbyFarmers)
	})
}

func getNearbyFarmers(c *gin.Context) {
	latStr := c.Query("lat")
	longStr := c.Query("long")
	distStr := c.Query("maxDistance") // in meters

	if latStr == "" || longStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat and long are required"})
		return
	}

	lat, _ := strconv.ParseFloat(latStr, 64)
	long, _ := strconv.ParseFloat(longStr, 64)
	maxDist, _ := strconv.ParseFloat(distStr, 64)
	if maxDist == 0 {
		maxDist = 10000 // Default 10km
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersColl := database.GetCollection("users")

	// GeoJSON Point for query
	// $near operator automatically sorts by distance
	filter := bson.M{
		"user_type": "farmer",
		"geo_location": bson.M{
			"$near": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{long, lat},
				},
				"$maxDistance": maxDist,
			},
		},
	}

	cursor, err := usersColl.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Also ensure 2dsphere index exists. " + err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var farmers []auth.User
	if err = cursor.All(ctx, &farmers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Decoding failed"})
		return
	}

	c.JSON(http.StatusOK, farmers)
}
