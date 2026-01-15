package auth

import (
	"context"
	"net/http"
	"time"

	"Agromi/core/router"
	"Agromi/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// User & Session Models
type User struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Phone            string             `bson:"phone" json:"phone"` // Not required binding if Email provided
	Email            string             `bson:"email" json:"email"`
	Name             string             `bson:"name" json:"name" binding:"required"`
	IsBlocked        bool               `bson:"is_blocked" json:"is_blocked"`
	UserType         string             `bson:"user_type" json:"user_type" binding:"required,oneof=farmer consumer admin"`
	ProfilePhotoURL  string             `bson:"profile_photo_url,omitempty" json:"profile_photo_url"`
	AuthTokenNum     string             `bson:"auth_token_num,omitempty" json:"auth_token_num"`
	RegionalLanguage string             `bson:"regional_language,omitempty" json:"regional_language"`
	Crops            []Crop             `bson:"crops,omitempty" json:"crops"`
	Location         Location           `bson:"location,omitempty" json:"location"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	LastActiveAt     time.Time          `bson:"last_active_at,omitempty" json:"last_active_at"`
	// GeoLocation for MongoDB 2dsphere index
	GeoLocation *GeoJSON `bson:"geo_location,omitempty" json:"geo_location,omitempty"`
}

type GeoJSON struct {
	Type        string    `bson:"type" json:"type"`
	Coordinates []float64 `bson:"coordinates" json:"coordinates"` // [longitude, latitude]
}

type Crop struct {
	Name string `bson:"name" json:"name"`
	Area string `bson:"area" json:"area"`
	Age  string `bson:"age" json:"age"`
}

type Location struct {
	Latitude  float64 `bson:"latitude" json:"latitude"`
	Longitude float64 `bson:"longitude" json:"longitude"`
}

type Session struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Token     string             `bson:"token" json:"token"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

var jwtSecret = []byte("YOUR_SUPER_SECRET_KEY") // In prod, use Env Var

func init() {
	router.Register(func(r *gin.Engine) {
		r.POST("/api/auth/login", handleLogin)
		r.POST("/api/auth/logout", handleLogout)
	})
}

func handleLogin(c *gin.Context) {
	var input struct {
		Phone string `json:"phone"`
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var userID primitive.ObjectID
	var userType string
	var userName string
	var isBlocked bool

	// 1. Search in Users (Farmers/Consumers)
	usersColl := database.GetCollection("users")
	var user User

	filter := bson.M{}
	if input.Phone != "" {
		filter["phone"] = input.Phone
	} else if input.Email != "" {
		filter["email"] = input.Email
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone or Email required"})
		return
	}

	err := usersColl.FindOne(ctx, filter).Decode(&user)

	if err == nil {
		// Found in Users
		userID = user.ID
		userType = user.UserType
		userName = user.Name
		isBlocked = user.IsBlocked
	} else if err == mongo.ErrNoDocuments {
		// 2. Search in Consultants
		consultantsColl := database.GetCollection("consultants")
		var consultant struct { // Minimal struct for auth check
			ID        primitive.ObjectID `bson:"_id"`
			Name      string             `bson:"name"`
			Phone     string             `bson:"phone"`
			Type      string             `bson:"type"`
			IsBlocked bool               `bson:"is_blocked"`
		}
		// Consultants mostly use Phone, but we can try generic filter if we add Email later
		errCons := consultantsColl.FindOne(ctx, filter).Decode(&consultant)
		if errCons == nil {
			// Found in Consultants
			userID = consultant.ID
			userType = "consultant" // Generalized type, specific type is in consultant profile search
			userName = consultant.Name
			isBlocked = consultant.IsBlocked
		} else {
			// Not found in either
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// 3. Check Blocked Status
	if isBlocked {
		c.JSON(http.StatusForbidden, gin.H{"error": "Your account has been blocked by Admin"})
		return
	}

	// 4. Generate JWT Token
	tokenString, _ := generateJWT(userID.Hex())

	// 5. Create Session
	session := Session{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Token:     tokenString,
		CreatedAt: time.Now(),
	}
	_, _ = database.GetCollection("sessions").InsertOne(ctx, session)

	// Update LastActiveAt (User or Consultant)
	if userType == "consultant" {
		database.GetCollection("consultants").UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{"updated_at": time.Now()}})
	} else {
		database.GetCollection("users").UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{"last_active_at": time.Now()}})
	}

	c.JSON(http.StatusOK, gin.H{
		"token":     tokenString,
		"user_id":   userID,
		"user_type": userType,
		"name":      userName,
	})
}

func handleLogout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Auto-delete session
	_, err := database.GetCollection("sessions").DeleteOne(ctx, bson.M{"token": token})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Logout failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// Helper: Generate JWT
func generateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
