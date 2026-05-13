package http

import (
	"github.com/gin-gonic/gin"
)

func SetupRoute(handler IpHandlerInterface) *gin.Engine {
	router := gin.Default()

	api := router.Group("/admin")
	{
		api.GET("/:list", handler.GetWhiteIps)
		api.POST("/:list", handler.NewIpInWhiteList)
		api.DELETE("/:list", handler.DeleteIpFromWhiteList)

	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	return router
}
