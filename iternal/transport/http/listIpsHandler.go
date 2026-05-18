package http

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"proxy/configGolang"
	"proxy/iternal/metrics"
	"proxy/iternal/usecase"
	"proxy/pkg/logger"
)

type IpHandlerInterface interface {
	GetWhiteIps(*gin.Context)
	NewIpInWhiteList(*gin.Context)
	DeleteIpFromWhiteList(c *gin.Context)
	CheckIp(c *gin.Context)
	IPCheckMiddleware() gin.HandlerFunc
	updateMetrics(listName string)
}

type IpHandler struct {
	Config        *configGolang.Config
	ListUseCase   usecase.ListUserCase
	RateLimiterUC *usecase.RateLimiterUseCase
}

type IpConf struct {
	Ip string `json:"ip"`
}

func CreateIpHandler(listUseCase usecase.ListUserCase, rateLimiterUC *usecase.RateLimiterUseCase,
	config *configGolang.Config) *IpHandler {
	return &IpHandler{
		ListUseCase:   listUseCase,
		RateLimiterUC: rateLimiterUC,
		Config:        config,
	}
}

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

func (h *IpHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
