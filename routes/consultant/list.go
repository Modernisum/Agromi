package consultant

import (
	"context"
	"net/http"
	"sort"
	"time"

	"Agromi/database"
	"Agromi/routes/consultant/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ScoredConsultant struct {
	models.Consultant `json:",inline"`
	Score             float64 `json:"score"`
}

// ListConsultants returns a filtered and scored list
func ListConsultants(c *gin.Context) {
	typ := c.Query("type")
	verifiedOnly := c.Query("verified_only")
	// latStr := c.Query("lat") // Unused for now until model update
	// lonStr := c.Query("lon") // Unused for now until model update

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := database.GetCollection("consultants")
	filter := bson.M{"is_blocked": false}

	if typ != "" {
		filter["type"] = typ
	}
	if verifiedOnly == "true" {
		filter["verification_status"] = models.StatusVerified
	}

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer cursor.Close(ctx)

	var consultants []models.Consultant
	if err = cursor.All(ctx, &consultants); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing data"})
		return
	}

	// Parse User Location
	// Note: Consultant model currently stores Address string, not GeoJSON/Coords yet?
	// Checking previous model... keys were Address.
	// To support Distance, Consultant model needs Lat/Lon.
	// Assuming for now consultants don't have Lat/Lon in model, so Distance weight = 0.
	// Will add Lat/Lon to model if strictly required, but for now ignoring Distance in score if missing.

	// Scoring
	var scoredList []ScoredConsultant
	for _, cons := range consultants {
		score := 0.0

		// 1. Rating (Weight 30%)
		score += (cons.Rating / 5.0) * 30.0

		// 2. Experience (Weight 20%, capped at 20 years)
		exp := float64(cons.Experience)
		if exp > 20 {
			exp = 20
		}
		score += (exp / 20.0) * 20.0

		// 3. Low Cost (Weight 20%)
		// Formula: Higher Fee = Lower Score
		// Normalize: If fee 0 -> 20pts. If fee 1000 -> 0pts.
		if cons.ConsultationFee <= 1000 {
			score += ((1000.0 - cons.ConsultationFee) / 1000.0) * 20.0
		}

		// 4. Distance (Skipped - Model update needed)

		scoredList = append(scoredList, ScoredConsultant{
			Consultant: cons,
			Score:      score,
		})
	}

	// Sort by Score Descending
	sort.Slice(scoredList, func(i, j int) bool {
		return scoredList[i].Score > scoredList[j].Score
	})

	c.JSON(http.StatusOK, scoredList)
}

// GetConsultant returns a single consultant by ID
func GetConsultant(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	coll := database.GetCollection("consultants")

	var consultant models.Consultant
	err = coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&consultant)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Consultant not found"})
		return
	}

	c.JSON(http.StatusOK, consultant)
}

func RegisterListRoutes(router *gin.RouterGroup) {
	router.GET("/list", ListConsultants)
	router.GET("/profile/:id", GetConsultant)
}
