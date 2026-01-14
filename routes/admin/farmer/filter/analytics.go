package filter

import (
	"context"
	"net/http"
	"time"

	"Agromi/core/router"
	"Agromi/database"

	// "Agromi/routes/auth" // Only if we need User struct, otherwise bson.M is fine for counts

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	router.Register(func(r *gin.Engine) {
		group := r.Group("/api/admin/filter")
		{
			group.GET("/stats", getStats)
			group.GET("/active-users", getActiveUsersList)
		}
	})
}

func getStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	usersColl := database.GetCollection("users")

	// Total Registered Farmers
	totalFarmers, _ := usersColl.CountDocuments(ctx, bson.M{"user_type": "farmer"})

	// Active User Logic
	// Logic: Active defined as LastActiveAt within defined periods
	now := time.Now()
	oneDayAgo := now.Add(-24 * time.Hour)
	oneWeekAgo := now.Add(-7 * 24 * time.Hour)
	oneMonthAgo := now.Add(-30 * 24 * time.Hour)

	dailyActive, _ := usersColl.CountDocuments(ctx, bson.M{"last_active_at": bson.M{"$gte": oneDayAgo}})
	weeklyActive, _ := usersColl.CountDocuments(ctx, bson.M{"last_active_at": bson.M{"$gte": oneWeekAgo}})
	monthlyActive, _ := usersColl.CountDocuments(ctx, bson.M{"last_active_at": bson.M{"$gte": oneMonthAgo}})

	// All Active Users (Generic definition, maybe logged in recently? Using Weekly as "Active")
	totalActive := weeklyActive

	c.JSON(http.StatusOK, gin.H{
		"total_farmers":        totalFarmers,
		"active_users_total":   totalActive,
		"daily_active_users":   dailyActive,
		"weekly_active_users":  weeklyActive,
		"monthly_active_users": monthlyActive,
	})
}

func getActiveUsersList(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Logic: Get users active in last 24 hours (or week?)
	// Let's assume daily active list
	sinceStr := c.Query("since") // optional duration
	var sinceTime time.Time

	now := time.Now()
	if sinceStr == "weekly" {
		sinceTime = now.Add(-7 * 24 * time.Hour)
	} else if sinceStr == "monthly" {
		sinceTime = now.Add(-30 * 24 * time.Hour)
	} else {
		sinceTime = now.Add(-24 * time.Hour) // Default Daily
	}

	usersColl := database.GetCollection("users")
	cursor, err := usersColl.Find(ctx, bson.M{"last_active_at": bson.M{"$gte": sinceTime}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer cursor.Close(ctx)

	var users []bson.M // Returning raw BSON to avoid cyclic dep or strict struct if not needed
	if err = cursor.All(ctx, &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding users"})
		return
	}

	c.JSON(http.StatusOK, users)
}
