package main

import (
	"log"

	"Agromi/routes"
	"Agromi/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Connect to MongoDB
	utils.ConnectDB()
	utils.InitFirebase() // Restore Firebase Init

	// 2. Initialize Gin Router
	app := gin.Default()

	// 3. Register Routes
	// All routes are auto-registered by the init() functions in the 'routes' package
	// utils.InitTwilio() // Removed (Reverted to Firebase)
	// utils.InitFirebase() // Removed (Trusted Frontend)
	log.Println("DEBUG: Calling routes.SetupRoutes...")
	routes.SetupRoutes(app)

	// 4. Start Server
	log.Println("🚜 Agromi Backend starting on :8080")
	if err := app.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
