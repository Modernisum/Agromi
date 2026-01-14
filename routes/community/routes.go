package community

import (
	"Agromi/core/router"

	"github.com/gin-gonic/gin"
)

func init() {
	router.Register(func(r *gin.Engine) {
		group := r.Group("/api/community")
		{
			group.POST("/create", CreatePost)
			group.GET("/feed", GetFeed)
			group.DELETE("/delete/:id", DeletePost)
		}
	})
}
