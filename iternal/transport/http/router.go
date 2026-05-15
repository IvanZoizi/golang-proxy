package http

import (
	"github.com/gin-gonic/gin"
)

func SetupRoute(handler *IpHandler) *gin.Engine {
	router := gin.Default()

	router.Use(handler.RateLimiterMiddleware())
	router.Use(handler.SubnetRateLimiterMiddleware())
	router.Use(handler.IPCheckMiddleware())

	router.GET("/health", handler.Health)
	router.GET("/metrics", MetricsHandler())

	api := router.Group("/api/:list")
	{
		api.GET("", handler.GetWhiteIps)
		api.POST("", handler.NewIpInWhiteList)
		api.DELETE("", handler.DeleteIpFromWhiteList)
		api.POST("/check", handler.CheckIp)
	}

	return router
}
