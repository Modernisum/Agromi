package social

import (
	"Agromi/core/router"

	"github.com/gin-gonic/gin"
)

func init() {
	router.Register(func(r *gin.Engine) {
		socialGroup := r.Group("/api/social")
		{
			RegisterCommentRoutes(socialGroup)
			RegisterReactionRoutes(socialGroup)
			RegisterFollowRoutes(socialGroup)
			RegisterNotificationRoutes(socialGroup)
		}
	})
}
