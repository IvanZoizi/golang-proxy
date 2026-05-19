package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http/httputil"
	"net/url"
	"proxy/internal/entity"
)

func SetupRoute(handler *IpHandler, cacheHandler *CacheHandler, cacheConfig *entity.CacheConfig) *gin.Engine {
	router := gin.Default()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Use(handler.IPCheckMiddleware())
	router.Use(handler.RateLimiterMiddleware())
	router.Use(handler.SubnetRateLimiterMiddleware())
	router.Use(cacheHandler.cacheMd.CacheMiddleware())

	router.GET("/health", handler.Health)
	router.GET("/metrics", MetricsHandler())

	cacheGroup := router.Group("/api/cache")
	{
		cacheGroup.GET("/stats", cacheHandler.GetCacheStats)
		cacheGroup.POST("/invalidate", cacheHandler.InvalidateCache)
		cacheGroup.DELETE("/clear", cacheHandler.ClearCache)
		cacheGroup.GET("/config", cacheHandler.GetCacheConfig)
		cacheGroup.PUT("/config", cacheHandler.UpdateCacheConfig)
	}

	api := router.Group("/api/:list")
	{
		api.GET("", handler.GetWhiteIps)
		api.POST("", handler.NewIpInWhiteList)
		api.DELETE("", handler.DeleteIpFromWhiteList)
		api.POST("/check", handler.CheckIp)
	}

	target, _ := url.Parse("http://localhost" + handler.Config.ProxyServerPort)
	proxy := httputil.NewSingleHostReverseProxy(target)

	router.NoRoute(func(c *gin.Context) {
		c.Request.URL.Scheme = target.Scheme
		c.Request.URL.Host = target.Host
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	return router
}
