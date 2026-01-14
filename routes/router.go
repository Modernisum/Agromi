package routes

import (
	core_router "Agromi/core/router"
	_ "Agromi/routes/admin/consultant"    // Trigger init() for Admin Consultant
	_ "Agromi/routes/admin/farmer"        // Trigger init() for farmer auth & profiles
	_ "Agromi/routes/admin/farmer/filter" // Trigger init() for farmer analytics
	_ "Agromi/routes/admin/finance"       // Trigger init() for Admin Finance
	_ "Agromi/routes/admin/market"        // Trigger init() for Admin Marketplace
	_ "Agromi/routes/admin/social"        // Trigger init() for Admin Social module
	_ "Agromi/routes/auth"                // Trigger init() for auth routes
	_ "Agromi/routes/chat"                // Trigger init() for Chat module
	_ "Agromi/routes/community"           // Trigger init() for Community module
	_ "Agromi/routes/consultant"          // Trigger init() for User Consultant interaction
	_ "Agromi/routes/farmer"              // Trigger init() for search, friends, suggestions
	_ "Agromi/routes/market"              // Trigger init() for User Marketplace
	_ "Agromi/routes/social"              // Trigger init() for Social module
	"fmt"

	"github.com/gin-gonic/gin"
)

// SetupRoutes applies all registered routes to the main application
func SetupRoutes(app *gin.Engine) {
	fmt.Println("DEBUG: SetupRoutes called. Registry size:", len(core_router.Registry))
	for _, routeFn := range core_router.Registry {
		routeFn(app)
	}
}
