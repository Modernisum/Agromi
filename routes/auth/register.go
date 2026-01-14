package auth

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"Agromi/core/router"
	"Agromi/database"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	router.Register(func(r *gin.Engine) {
		r.POST("/api/auth/register", handleRegister)
	})

	// Create Index for Optimized Search on Startup
	go createPhoneIndex()
}

func handleRegister(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate Farmer Details
	if user.UserType == "farmer" {
		if user.ProfilePhotoURL == "" || user.RegionalLanguage == "" || len(user.Crops) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Farmers must provide profile photo, language, and at least one crop"})
			return
		}
		// Basic check for location (0,0 is valid but unlikely for a real user, but let's just check if it was provided if possible?
		// Since it's a struct value, it defaults to 0. We can add a check if needed, but strict 0,0 check might be checking default.
		// Let's assume frontend sends it.
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	collection := database.GetCollection("users")

	// Check existing (handled by Unique Index mostly, but good for custom error)
	count, _ := collection.CountDocuments(ctx, bson.M{"phone": user.Phone})
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	user.ID = primitive.NewObjectID()
	user.IsBlocked = false
	user.CreatedAt = time.Now()

	// Populate GeoLocation for Geospatial Index
	if user.Location.Latitude != 0 || user.Location.Longitude != 0 {
		user.GeoLocation = &GeoJSON{
			Type:        "Point",
			Coordinates: []float64{user.Location.Longitude, user.Location.Latitude},
		}
	}

	// Auto-generate Auth Token Number (e.g., random 8-digit string or UUID)
	// For simplicity, using a timestamp-based random string
	rand.Seed(time.Now().UnixNano())
	user.AuthTokenNum = fmt.Sprintf("AG-%d-%d", time.Now().Unix(), rand.Intn(10000))

	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "id": user.ID})
}

// createPhoneIndex ensures fast lookups by Phone
func createPhoneIndex() {
	// Wait for DB connection
	for i := 0; i < 20; i++ { // Retry for 20 attempts (approx 2 seconds)
		if database.Client != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if database.Client == nil {
		// Log error or just return if database never connected (unlikely in correct flow)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := database.GetCollection("users")

	// 1. Unique Phone Index
	modelPhone := mongo.IndexModel{
		Keys:    bson.D{{Key: "phone", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, _ = collection.Indexes().CreateOne(ctx, modelPhone)

	// 2. Text Index for Name Search
	modelText := mongo.IndexModel{
		Keys: bson.D{{Key: "name", Value: "text"}},
	}
	_, _ = collection.Indexes().CreateOne(ctx, modelText)

	// 3. 2dsphere Index for Geospatial Queries
	modelGeo := mongo.IndexModel{
		Keys: bson.D{{Key: "geo_location", Value: "2dsphere"}},
	}
	_, _ = collection.Indexes().CreateOne(ctx, modelGeo)
}
