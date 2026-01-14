package finance_routes

import (
	"Agromi/core/router"
	sponsor "Agromi/routes/admin/finance/sponsor"

	"github.com/gin-gonic/gin"
)

func init() {
	router.Register(func(r *gin.Engine) {
		adminGroup := r.Group("/api/admin")
		// Sponsor
		sponsor.RegisterRoutes(adminGroup)

		// Verify
		financeGroup := r.Group("/api/admin/finance")
		RegisterVerifyRoutes(financeGroup) // Direct call, same package
	})
}
