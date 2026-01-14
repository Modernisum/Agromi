package market

import (
	"Agromi/core/router"
	"Agromi/routes/market/buy"
	"Agromi/routes/market/rent"
	"Agromi/routes/market/sell"

	"github.com/gin-gonic/gin"
)

func init() {
	println("DEBUG: Market Routes Init called")
	router.Register(func(r *gin.Engine) {
		marketGroup := r.Group("/api/market")
		{
			buy.RegisterRoutes(marketGroup)
			rent.RegisterRoutes(marketGroup)
			sell.RegisterRoutes(marketGroup)
		}
	})
}
