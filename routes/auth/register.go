package auth

import (
	"context"
	"net/http"
	"time"

	"Agromi/core/router"
	"Agromi/database"
	"Agromi/utils"

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
	var input struct {
		AuthToken        string   `json:"auth_token" binding:"required"`
		Name             string   `json:"name" binding:"required"`
		UserType         string   `json:"user_type" binding:"required,oneof=farmer consumer admin"`
		ProfilePhotoURL  string   `json:"profile_photo_url"`
		RegionalLanguage string   `json:"regional_language"`
		Crops            []Crop   `json:"crops"`
		Location         Location `json:"location"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Verify Firebase Token
	token, err := utils.AuthClient.VerifyIDToken(ctx, input.AuthToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Auth Token"})
		return
	}
	uid := token.UID
	// Extract phone/email from token if available
	phone, _ := token.Claims["phone_number"].(string)
	email, _ := token.Claims["email"].(string)

	usersColl := database.GetCollection("users")

	// 2. Check existence
	count, err := usersColl.CountDocuments(ctx, bson.M{"auth_token_num": uid})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// 3. Create User
	newUser := User{
		ID:               primitive.NewObjectID(),
		Phone:            phone,
		Email:            email,
		AuthTokenNum:     uid, // Store UID
		Name:             input.Name,
		UserType:         input.UserType,
		IsBlocked:        false,
		ProfilePhotoURL:  input.ProfilePhotoURL,
		RegionalLanguage: input.RegionalLanguage,
		Crops:            input.Crops,
		Location:         input.Location,
		CreatedAt:        time.Now(),
	}

	// GeoLocation
	if newUser.Location.Latitude != 0 || newUser.Location.Longitude != 0 {
		newUser.GeoLocation = &GeoJSON{
			Type:        "Point",
			Coordinates: []float64{newUser.Location.Longitude, newUser.Location.Latitude},
		}
	}

	_, err = usersColl.InsertOne(ctx, newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "id": newUser.ID})
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
