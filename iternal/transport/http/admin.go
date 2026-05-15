package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"proxy/iternal/usecase"
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
	ListUseCase   *usecase.ListsUseCase
	rateLimiterUC *usecase.RateLimiterUseCase
}

type IpConf struct {
	Ip string `json:"ip"`
}

func CreateIpHandler(listUseCase *usecase.ListsUseCase, rateLimiterUC *usecase.RateLimiterUseCase) *IpHandler {
	return &IpHandler{
		ListUseCase:   listUseCase,
		rateLimiterUC: rateLimiterUC,
	}
}

func (h *IpHandler) GetWhiteIps(c *gin.Context) {
	listName := c.Param("list")
	var ips []string
	var err error

	switch listName {
	case "white":
		ips, err = h.ListUseCase.WhiteRepo.GetAllIps()
	case "black":
		ips, err = h.ListUseCase.BlackRepo.GetAllIps()
	default:
		ips, err = h.ListUseCase.GrayRepo.GetAllIps()
	}

	if err != nil {
		logger.Error("Failed to get IPs",
			zap.String("list", listName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
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

	var flag bool
	switch listName {
	case "white":
		flag = h.ListUseCase.WhiteRepo.Contains(ipStruct.Ip)
	case "black":
		flag = h.ListUseCase.BlackRepo.Contains(ipStruct.Ip)
	default:
		flag = h.ListUseCase.GrayRepo.Contains(ipStruct.Ip)
	}

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

	var err error
	switch listName {
	case "white":
		err = h.ListUseCase.WhiteRepo.AddIp(ipStruct.Ip)
	case "black":
		err = h.ListUseCase.BlackRepo.AddIp(ipStruct.Ip)
	default:
		err = h.ListUseCase.GrayRepo.AddIp(ipStruct.Ip)
	}

	if err == nil {
		logger.Info("IP added",
			zap.String("list", listName),
			zap.String("ip", ipStruct.Ip))
		c.JSON(http.StatusOK, gin.H{
			"list":   listName,
			"status": "ok",
		})
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

	var err error
	switch listName {
	case "white":
		err = h.ListUseCase.WhiteRepo.DeleteIp(ipStruct.Ip)
	case "black":
		err = h.ListUseCase.BlackRepo.DeleteIp(ipStruct.Ip)
	default:
		err = h.ListUseCase.GrayRepo.DeleteIp(ipStruct.Ip)
	}

	if err == nil {
		logger.Info("IP deleted",
			zap.String("list", listName),
			zap.String("ip", ipStruct.Ip))
		c.JSON(http.StatusOK, gin.H{
			"list":   listName,
			"status": "ok",
		})
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

func (h *IpHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
