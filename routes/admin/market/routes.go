package admin_market

import (
	"Agromi/core/router"
	admin_buy "Agromi/routes/admin/market/buy"
	admin_rent "Agromi/routes/admin/market/rent"
	admin_sell "Agromi/routes/admin/market/sell"

	"github.com/gin-gonic/gin"
)

func init() {
	router.Register(func(r *gin.Engine) {
		marketGroup := r.Group("/api/admin/market")
		{
			admin_buy.RegisterRoutes(marketGroup)
			admin_rent.RegisterRoutes(marketGroup)
			admin_sell.RegisterRoutes(marketGroup)
			RegisterManageRoutes(marketGroup)
		}
	})
}
