package http

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"proxy/configGolang"
	"proxy/internal/metrics"
	"proxy/internal/usecase"
	"proxy/pkg/logger"
)

type IpHandlerInterface interface {
	GetWhiteIps(*gin.Context)
	NewIpInWhiteList(*gin.Context)
	DeleteIpFromWhiteList(c *gin.Context)
	CheckIp(c *gin.Context)
	IPCheckMiddleware() gin.HandlerFunc
}

type IpHandler struct {
	Config        *configGolang.Config
	ListUseCase   usecase.ListUserCase
	RateLimiterUC usecase.RateLimiterUseCase
}

type IpConf struct {
	Ip string `json:"ip" example:"192.168.1.1"`
}

func CreateIpHandler(listUseCase usecase.ListUserCase, rateLimiterUC usecase.RateLimiterUseCase,
	config *configGolang.Config) *IpHandler {
	return &IpHandler{
		ListUseCase:   listUseCase,
		RateLimiterUC: rateLimiterUC,
		Config:        config,
	}
}

// GetWhiteIps godoc
// @Summary Получить список IP по типу списка
// @Description Возвращает все IP из указанного списка (white/black/gray)
// @Tags ip-lists
// @Accept json
// @Produce json
// @Param list path string true "Тип списка" enums(white,black,gray)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/{list} [get]
func (h *IpHandler) GetWhiteIps(c *gin.Context) {
	listName := c.Param("list")
	ips, err := h.ListUseCase.GetAllIps(listName)

	if err != nil {
		logger.Error("Failed to get IPs",
			zap.String("list", listName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err,
		})
		return
	}

	logger.Info("Get IPs",
		zap.String("list", listName),
		zap.Int("count", len(ips)))

	c.JSON(http.StatusOK, gin.H{
		"list":  listName,
		"items": ips,
		"count": len(ips),
	})
}

// CheckIp godoc
// @Summary Проверить наличие IP в списке
// @Description Проверяет, находится ли IP адрес в указанном списке
// @Tags ip-lists
// @Accept json
// @Produce json
// @Param list path string true "Тип списка" enums(white,black,gray)
// @Param request body IpConf true "IP адрес для проверки"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/{list}/check [post]
func (h IpHandler) CheckIp(c *gin.Context) {
	listName := c.Param("list")

	var ipStruct IpConf

	if err := c.ShouldBindJSON(&ipStruct); err != nil {
		logger.Warn("Check IP failed",
			zap.String("list", listName),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "IP is required",
		})
		return
	}

	flag, _ := h.ListUseCase.Contains(ipStruct.Ip, listName)

	logger.Info("Check IP",
		zap.String("list", listName),
		zap.String("ip", ipStruct.Ip),
		zap.Bool("exists", flag))

	c.JSON(http.StatusOK, gin.H{
		"list":  listName,
		"check": flag,
	})
}

// NewIpInWhiteList godoc
// @Summary Добавить IP в список
// @Description Добавляет новый IP адрес в указанный список (white/black/gray)
// @Tags ip-lists
// @Accept json
// @Produce json
// @Param list path string true "Тип списка" enums(white,black,gray)
// @Param request body IpConf true "IP адрес для добавления"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/{list} [post]
func (h *IpHandler) NewIpInWhiteList(c *gin.Context) {
	listName := c.Param("list")

	var ipStruct IpConf
	if err := c.ShouldBindJSON(&ipStruct); err != nil {
		logger.Warn("Add IP failed",
			zap.String("list", listName),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "IP is required",
		})
		return
	}

	err := h.ListUseCase.AddIp(ipStruct.Ip, listName)

	if err == nil {
		logger.Info("IP added",
			zap.String("list", listName),
			zap.String("ip", ipStruct.Ip))
		c.JSON(http.StatusOK, gin.H{
			"list":   listName,
			"status": "ok",
		})
		h.updateMetrics(listName)
	} else {
		logger.Error("Add IP failed",
			zap.String("list", listName),
			zap.String("ip", ipStruct.Ip),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"list":    listName,
			"status":  "error",
			"message": err.Error(),
		})
	}
}

// DeleteIpFromWhiteList godoc
// @Summary Удалить IP из списка
// @Description Удаляет IP адрес из указанного списка (white/black/gray)
// @Tags ip-lists
// @Accept json
// @Produce json
// @Param list path string true "Тип списка" enums(white,black,gray)
// @Param request body IpConf true "IP адрес для удаления"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/{list} [delete]
func (h *IpHandler) DeleteIpFromWhiteList(c *gin.Context) {
	listName := c.Param("list")

	var ipStruct IpConf

	if err := c.ShouldBindJSON(&ipStruct); err != nil {
		logger.Warn("Delete IP failed",
			zap.String("list", listName),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "IP is required",
		})
		return
	}

	err := h.ListUseCase.DeleteIp(ipStruct.Ip, listName)

	if err == nil {
		logger.Info("IP deleted",
			zap.String("list", listName),
			zap.String("ip", ipStruct.Ip))
		c.JSON(http.StatusOK, gin.H{
			"list":   listName,
			"status": "ok",
		})
		h.updateMetrics(listName)
	} else {
		logger.Error("Delete IP failed",
			zap.String("list", listName),
			zap.String("ip", ipStruct.Ip),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"list":    listName,
			"status":  "error",
			"message": err.Error(),
		})
	}
}

func (h *IpHandler) updateMetrics(listName string) {
	ips, _ := h.ListUseCase.GetAllIps(listName)
	switch listName {
	case "white":
		metrics.ActiveIpsWhite.Set(float64(len(ips)))
	case "black":
		metrics.ActiveIpsBlack.Set(float64(len(ips)))
	case "gray":
		metrics.ActiveIpsGray.Set(float64(len(ips)))
	}
}

// Health godoc
// @Summary Проверка здоровья сервиса
// @Description Возвращает статус сервиса
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *IpHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
