package http

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"proxy/internal/dto"
	"proxy/internal/entity"
	"proxy/internal/usecase"
	"proxy/pkg/logger"
)

type CacheHandler struct {
	cacheUC usecase.CacheUseCase
	cacheMd *CacheMiddleware
	config  *entity.CacheConfig
}

func NewCacheHandler(cacheUC usecase.CacheUseCase, config *entity.CacheConfig, cacheMd *CacheMiddleware) *CacheHandler {
	return &CacheHandler{
		cacheUC: cacheUC,
		config:  config,
		cacheMd: cacheMd,
	}
}

// GetCacheStats godoc
// @Summary Получить статистику кэша
// @Description Возвращает статистику использования кэша
// @Tags cache
// @Accept json
// @Produce json
// @Success 200 {object} repository.CacheStats
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/cache/stats [get]
func (h *CacheHandler) GetCacheStats(c *gin.Context) {
	stats, err := h.cacheUC.GetCacheStats()
	if err != nil {
		logger.Error("Failed to get cache stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// InvalidateCache godoc
// @Summary Инвалидировать кэш
// @Description Инвалидирует кэш по различным критериям
// @Tags cache
// @Accept json
// @Produce json
// @Param request body InvalidateRequestSwagger true "Параметры инвалидации"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/cache/invalidate [post]
func (h *CacheHandler) InvalidateCache(c *gin.Context) {
	var req struct {
		Type      string `json:"type" binding:"required" example:"prefix" enums:"exact,prefix,regex,tag,all"`
		Key       string `json:"key" example:"cache-key-123"`
		Prefix    string `json:"prefix" example:"/api/users/"`
		Regex     string `json:"regex" example:".*user.*"`
		Tag       string `json:"tag" example:"users"`
		StaleOnly bool   `json:"stale_only" example:"false"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	invalidateReq := &usecase.InvalidateRequest{
		Type:      req.Type,
		Key:       req.Key,
		Prefix:    req.Prefix,
		Regex:     req.Regex,
		Tag:       req.Tag,
		StaleOnly: req.StaleOnly,
	}

	count, err := h.cacheUC.Invalidate(invalidateReq)
	if err != nil {
		logger.Error("Failed to invalidate cache",
			zap.String("type", req.Type),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	logger.Info("Cache invalidated",
		zap.String("type", req.Type),
		zap.Int("count", count))

	c.JSON(http.StatusOK, gin.H{
		"message": "Cache invalidated successfully",
		"count":   count,
	})
}

// ClearCache godoc
// @Summary Очистить весь кэш
// @Description Полностью удаляет все записи из кэша
// @Tags cache
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/cache/clear [delete]
func (h *CacheHandler) ClearCache(c *gin.Context) {
	count, err := h.cacheUC.Invalidate(&usecase.InvalidateRequest{
		Type: "all",
	})

	if err != nil {
		logger.Error("Failed to clear cache", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	logger.Info("Cache cleared", zap.Int("count", count))
	c.JSON(http.StatusOK, gin.H{
		"message": "Cache cleared successfully",
		"count":   count,
	})
}

// GetCacheConfig godoc
// @Summary Получить конфигурацию кэша
// @Description Возвращает текущую конфигурацию кэширования
// @Tags cache
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/cache/config [get]
func (h *CacheHandler) GetCacheConfig(c *gin.Context) {
	// Возвращаем конфиг в читаемом формате
	config := map[string]interface{}{
		"enabled":              h.config.Enabled,
		"default_ttl":          h.config.DefaultTTL.String(),
		"max_entry_size_bytes": h.config.MaxEntrySize,
		"min_entry_size_bytes": h.config.MinEntrySize,
		"max_cache_size_bytes": h.config.MaxCacheSize,
		"max_entries":          h.config.MaxEntries,
		"cache_2xx":            h.config.Cache2xx,
		"cache_3xx":            h.config.Cache3xx,
		"cache_4xx":            h.config.Cache4xx,
		"cache_5xx":            h.config.Cache5xx,
		"ttl_2xx":              h.config.TTL2xx.String(),
		"ttl_3xx":              h.config.TTL3xx.String(),
		"ttl_4xx":              h.config.TTL4xx.String(),
		"ttl_5xx":              h.config.TTL5xx.String(),
		"cache_methods":        h.config.CacheMethods,
		"respect_no_cache":     h.config.RespectNoCache,
		"respect_max_age":      h.config.RespectMaxAge,
	}
	c.JSON(http.StatusOK, config)
}

// UpdateCacheConfig godoc
// @Summary Обновить конфигурацию кэша
// @Description Обновляет настройки кэширования. Можно передавать только изменяемые поля.
// @Tags cache
// @Accept json
// @Produce json
// @Param config body dto.UpdateCacheConfigRequest true "Новая конфигурация кэша (частичное обновление)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/cache/config [put]
func (h *CacheHandler) UpdateCacheConfig(c *gin.Context) {
	var req dto.UpdateCacheConfigRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	h.cacheUC.UpdateConfigCache(req)

	logger.Info("Cache configuration updated")
	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration updated successfully",
	})
}

// InvalidateRequestSwagger представляет запрос на инвалидацию для Swagger
type InvalidateRequestSwagger struct {
	Type      string `json:"type" binding:"required" example:"prefix" enums:"exact,prefix,regex,tag,all"`
	Key       string `json:"key" example:"cache-key-123"`
	Prefix    string `json:"prefix" example:"/api/users/"`
	Regex     string `json:"regex" example:".*user.*"`
	Tag       string `json:"tag" example:"users"`
	StaleOnly bool   `json:"stale_only" example:"false"`
}
