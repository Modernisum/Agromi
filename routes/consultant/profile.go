package consultant

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"Agromi/database"
	"Agromi/routes/consultant/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RegisterConsultant creates a new consultant profile
func RegisterConsultant(c *gin.Context) {
	var body models.Consultant
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Basic Validation
	if body.Phone == "" || body.Name == "" || body.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name, Phone, and Type are required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	coll := database.GetCollection("consultants")

	// Check existing phone
	count, _ := coll.CountDocuments(ctx, bson.M{"phone": body.Phone})
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Consultant with this phone already exists"})
		return
	}

	// Set Defaults
	body.ID = primitive.NewObjectID()
	body.CreatedAt = time.Now()
	body.UpdatedAt = time.Now()
	body.VerificationStatus = models.StatusPending // Default pending
	body.IsBlocked = false
	body.Rating = 0
	body.ReviewCount = 0

	// Generate Auth Token
	rand.Seed(time.Now().UnixNano())
	body.AuthTokenNum = fmt.Sprintf("CONS-%d-%d", time.Now().Unix(), rand.Intn(10000))

	_, err := coll.InsertOne(ctx, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register consultant"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Consultant registered successfully", "id": body.ID, "auth_token": body.AuthTokenNum})
}

// UpdateProfile updates consultant details
func UpdateProfile(c *gin.Context) {
	// For MVP, passing ID in body or query. In prod, extract from JWT context.
	var body struct {
		ID      string                 `json:"id" binding:"required"`
		Updates map[string]interface{} `json:"updates" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objID, err := primitive.ObjectIDFromHex(body.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Consultant ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	coll := database.GetCollection("consultants")

	// Filter allowed fields to update
	allowedUpdates := bson.M{}
	for k, v := range body.Updates {
		// Prevent updating critical fields like ID, Phone (without verification), Ratings
		if k != "id" && k != "phone" && k != "rating" && k != "review_count" && k != "is_blocked" && k != "verification_status" {
			allowedUpdates[k] = v
		}
	}
	allowedUpdates["updated_at"] = time.Now()

	result, err := coll.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": allowedUpdates})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Consultant not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// DELETE /request-delete
// Schedule deletion after 30 days
func RequestDeletion(c *gin.Context) {
	var body struct {
		ID string `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objID, err := primitive.ObjectIDFromHex(body.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	coll := database.GetCollection("consultants")

	scheduledTime := time.Now().Add(30 * 24 * time.Hour) // 30 Days

	_, err = coll.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{
		"deletion_scheduled_at": scheduledTime,
		"updated_at":            time.Now(),
	}})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule deletion"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account scheduled for deletion in 30 days"})
}

func RegisterProfileRoutes(router *gin.RouterGroup) {
	// Public or Auth protected (assuming public for registration, auth for update/delete in real app)
	router.POST("/create", RegisterConsultant)
	router.PUT("/update", UpdateProfile)
	router.POST("/delete-request", RequestDeletion)
}
