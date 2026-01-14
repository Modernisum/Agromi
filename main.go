package main

import (
	"log"

	"Agromi/database"
	"Agromi/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Connect to MongoDB
	database.Connect()

	// 2. Initialize Gin Router
	app := gin.Default()

	// 3. Register Routes
	// All routes are auto-registered by the init() functions in the 'routes' package
	log.Println("DEBUG: Calling routes.SetupRoutes...")
	routes.SetupRoutes(app)

	// 4. Start Server
	log.Println("🚜 Agromi Backend starting on :8080")
	if err := app.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
