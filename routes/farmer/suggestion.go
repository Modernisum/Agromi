package farmer

import (
	"context"
	"net/http"
	"time"

	"Agromi/core/router"
	"Agromi/database"
	"Agromi/routes/auth"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	// Register route
	router.Register(func(r *gin.Engine) {
		r.GET("/api/farmer/suggest-similar/:id", getSimilarFarmers)
	})
}

func getSimilarFarmers(c *gin.Context) {
	idStr := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	usersColl := database.GetCollection("users")

	// 1. Get Source Farmer
	var sourceFarmer auth.User
	err = usersColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&sourceFarmer)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Farmer not found"})
		return
	}

	if sourceFarmer.GeoLocation == nil || len(sourceFarmer.Crops) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Farmer needs location and crops for suggestions"})
		return
	}

	// 2. Find Similar: Same Crop within 50km
	// Extract crop names
	var cropNames []string
	for _, crop := range sourceFarmer.Crops {
		cropNames = append(cropNames, crop.Name)
	}

	filter := bson.M{
		"user_type":  "farmer",
		"_id":        bson.M{"$ne": objID},     // Exclude self
		"crops.name": bson.M{"$in": cropNames}, // Match any crop
		"geo_location": bson.M{
			"$near": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": sourceFarmer.GeoLocation.Coordinates,
				},
				"$maxDistance": 50000, // 50km
			},
		},
	}

	cursor, err := usersColl.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Recommendation search failed. " + err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var similar []auth.User
	// Limit to 5 manually for simplicity without options package import if not needed
	for cursor.Next(ctx) {
		var u auth.User
		if err := cursor.Decode(&u); err == nil {
			similar = append(similar, u)
			if len(similar) >= 5 {
				break
			}
		}
	}

	c.JSON(http.StatusOK, similar)
}
