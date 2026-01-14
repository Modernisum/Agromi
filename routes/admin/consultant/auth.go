package admin_consultant

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

// AdminCreateConsultant allows admins to create accounts directly
func AdminCreateConsultant(c *gin.Context) {
	var body models.Consultant
	// Admin might provide minimal info, so we relax validation slightly
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Basic check
	if body.Phone == "" || body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name and Phone are required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	coll := database.GetCollection("consultants")

	// Check existing
	count, _ := coll.CountDocuments(ctx, bson.M{"phone": body.Phone})
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Consultant exists"})
		return
	}

	// Admin created accounts are verified by default if not specified?
	// Let's default to verified for admin convenience
	if body.VerificationStatus == "" {
		body.VerificationStatus = models.StatusVerified
	}

	body.ID = primitive.NewObjectID()
	body.CreatedAt = time.Now()
	body.UpdatedAt = time.Now()
	body.IsBlocked = false

	rand.Seed(time.Now().UnixNano())
	body.AuthTokenNum = fmt.Sprintf("ADMIN-CONS-%d-%d", time.Now().Unix(), rand.Intn(10000))

	_, err := coll.InsertOne(ctx, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create consultant"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Consultant created by admin", "id": body.ID})
}

// BlockConsultant toggles blocking
func BlockConsultant(c *gin.Context) {
	id := c.Param("id")
	// Action: block or unblock query param
	action := c.Query("action") // "block" or "unblock"

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	isBlocked := (action == "block")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	coll := database.GetCollection("consultants")

	_, err = coll.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"is_blocked": isBlocked}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Consultant %sed successfully", action)})
}

// DeleteConsultant immediately deletes the account
func DeleteConsultant(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	coll := database.GetCollection("consultants")

	_, err = coll.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete consultant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Consultant deleted permenantly"})
}

func RegisterAuthRoutes(router *gin.RouterGroup) {
	router.POST("/create", AdminCreateConsultant)
	router.PUT("/manage/block/:id", BlockConsultant)
	router.DELETE("/manage/delete/:id", DeleteConsultant)
}
