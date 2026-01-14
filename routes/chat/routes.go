package chat

import (
	"Agromi/core/router"

	"github.com/gin-gonic/gin"
)

func init() {
	router.Register(func(r *gin.Engine) {
		group := r.Group("/api/chat")
		{
			group.POST("/send", SendMessage)
			group.GET("/history", GetHistory)

			group.POST("/group/create", CreateGroup)
			group.POST("/group/join", JoinGroup)
		}
	})
}
