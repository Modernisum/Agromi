package admin_consultant

import (
	"Agromi/core/router"

	"github.com/gin-gonic/gin"
)

func init() {
	router.Register(func(r *gin.Engine) {
		adminGroup := r.Group("/api/admin/consultant")
		{
			RegisterAuthRoutes(adminGroup)
			RegisterAnalyticsRoutes(adminGroup)
		}
	})
}
