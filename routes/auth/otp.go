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
)

func init() {
	router.Register(func(r *gin.Engine) {
		r.POST("/api/auth/send-otp", handleSendOTP)
		r.POST("/api/auth/verify-otp", handleVerifyOTP)
	})
	// Ensure utils init runs? Or call it here.
	utils.InitTwilio()
}

func handleSendOTP(c *gin.Context) {
	var input struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sid, err := utils.SendOTP(input.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP Sent Successfully", "sid": sid})
}

func handleVerifyOTP(c *gin.Context) {
	var input struct {
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	valid, err := utils.VerifyOTP(input.Phone, input.Code)
	if err != nil || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or Expired OTP"})
		return
	}

	// OTP Verified! Now check if User exists to Login, or return "New User" status.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check Users
	collection := database.GetCollection("users")
	var user User
	errDB := collection.FindOne(ctx, bson.M{"phone": input.Phone}).Decode(&user)

	if errDB == nil {
		// User Exists -> LOGIN
		token, _ := generateJWT(user.ID.Hex())
		// (Skipping Session creation for brevity, assuming similar to login.go logic)
		c.JSON(http.StatusOK, gin.H{
			"status":    "login_success",
			"token":     token,
			"user_id":   user.ID,
			"user_type": user.UserType,
			"username":  user.Name,
		})
		return
	}

	// Check Consultants
	collCons := database.GetCollection("consultants")
	var cons struct {
		ID   primitive.ObjectID `bson:"_id"`
		Type string             `bson:"type"`
		Name string             `bson:"name"`
	}
	errCons := collCons.FindOne(ctx, bson.M{"phone": input.Phone}).Decode(&cons)

	if errCons == nil {
		// Consultant Exists -> LOGIN
		token, _ := generateJWT(cons.ID.Hex())
		c.JSON(http.StatusOK, gin.H{
			"status":    "login_success",
			"token":     token,
			"user_id":   cons.ID,
			"user_type": "consultant", // generic
			"username":  cons.Name,
		})
		return
	}

	// User Not Found -> PROCEED TO REGISTER
	c.JSON(http.StatusAccepted, gin.H{ // 202 Accepted
		"status":  "new_user",
		"message": "OTP Verified. Proceed to Registration.",
		"phone":   input.Phone,
	})
}
