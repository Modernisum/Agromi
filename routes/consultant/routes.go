package consultant

import (
	"Agromi/core/router"

	"github.com/gin-gonic/gin"
)

func init() {
	router.Register(func(r *gin.Engine) {
		consultantGroup := r.Group("/api/consultant")
		{
			RegisterProfileRoutes(consultantGroup)
			RegisterListRoutes(consultantGroup)
		}
	})
}
