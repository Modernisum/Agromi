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
	router.Register(func(r *gin.Engine) {
		group := r.Group("/api/admin/farmer/profile")
		{
			group.POST("/create", createFarmerDirect)
			group.PUT("/update/:id", updateFarmer)
			group.GET("/:id", getFarmer)
			group.GET("/all", getAllFarmers) // Moving basic list here or in filter? filter says "get all farmerlist". Let's put specific ID getters here.
		}
	})
}

// createFarmerDirect allows admin to create a farmer without phone verification flow
func createFarmerDirect(c *gin.Context) {
	// Use a specific input struct to avoid "required" validation failure on UserType since we set it manually
	var input struct {
		Phone            string        `json:"phone" binding:"required"`
		Name             string        `json:"name" binding:"required"`
		RegionalLanguage string        `json:"regional_language"`
		ProfilePhotoURL  string        `json:"profile_photo_url"`
		Crops            []auth.Crop   `json:"crops"`
		Location         auth.Location `json:"location"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := auth.User{
		ID:               primitive.NewObjectID(),
		Phone:            input.Phone,
		Name:             input.Name,
		UserType:         "farmer",
		IsBlocked:        false,
		CreatedAt:        time.Now(),
		RegionalLanguage: input.RegionalLanguage,
		ProfilePhotoURL:  input.ProfilePhotoURL,
		Crops:            input.Crops,
		Location:         input.Location,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersColl := database.GetCollection("users")

	// Check if user already exists
	count, _ := usersColl.CountDocuments(ctx, bson.M{"phone": input.Phone})
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "User already registered with this phone number"})
		return
	}

	_, err := usersColl.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create farmer"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func updateFarmer(c *gin.Context) {
	idStr := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prevent updating ID
	delete(updateData, "id")
	delete(updateData, "_id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = database.GetCollection("users").UpdateOne(
		ctx,
		bson.M{"_id": objID, "user_type": "farmer"},
		bson.M{"$set": updateData},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update farmer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Farmer updated successfully"})
}

func getFarmer(c *gin.Context) {
	idStr := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user auth.User
	err = database.GetCollection("users").FindOne(ctx, bson.M{"_id": objID, "user_type": "farmer"}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Farmer not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func getAllFarmers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.GetCollection("users").Find(ctx, bson.M{"user_type": "farmer"})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer cursor.Close(ctx)

	var farmers []auth.User
	if err = cursor.All(ctx, &farmers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding farmers"})
		return
	}

	c.JSON(http.StatusOK, farmers)
}
