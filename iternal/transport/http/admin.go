package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"proxy/iternal/usecase"
)

type IpHandlerInterface interface {
	GetWhiteIps(*gin.Context)
	NewIpInWhiteList(*gin.Context)
	DeleteIpFromWhiteList(c *gin.Context)
}

type IpHandler struct {
	ListUseCase *usecase.ListsUseCase
}

type IpConf struct {
	Ip string `json:"ip"`
}

func CreateIpHandler(listUseCase *usecase.ListsUseCase) *IpHandler {
	return &IpHandler{ListUseCase: listUseCase}
}

func (h *IpHandler) GetWhiteIps(c *gin.Context) {

	listName := c.Param("list")
	var ips []string
	switch listName {
	case "white":
		ips = h.ListUseCase.WhiteRepo.GetAllIps()
	case "black":
		ips = h.ListUseCase.BlackRepo.GetAllIps()
	default:
		ips = h.ListUseCase.GrayRepo.GetAllIps()
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  listName,
		"items": ips,
		"count": len(ips),
	})
}

func (h *IpHandler) NewIpInWhiteList(c *gin.Context) {

	listName := c.Param("list")

	var ipStruct IpConf

	if err := c.ShouldBindJSON(&ipStruct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name is required",
		})
		return
	}
	var message string
	var flag bool
	switch listName {
	case "white":
		message, flag = h.ListUseCase.WhiteRepo.AddIps(ipStruct.Ip)
	case "black":
		message, flag = h.ListUseCase.BlackRepo.AddIps(ipStruct.Ip)
	default:
		message, flag = h.ListUseCase.GrayRepo.AddIps(ipStruct.Ip)
	}

	if flag {
		c.JSON(http.StatusOK, gin.H{
			"list":   listName,
			"status": "ok",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"list":    listName,
			"status":  "error",
			"message": message,
		})
	}

}

func (h *IpHandler) DeleteIpFromWhiteList(c *gin.Context) {
	listName := c.Param("list")

	var ipStruct IpConf

	if err := c.ShouldBindJSON(&ipStruct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name is required",
		})
		return
	}

	var message string
	var flag bool
	switch listName {
	case "white":
		message, flag = h.ListUseCase.WhiteRepo.DeleteIp(ipStruct.Ip)
	case "black":
		message, flag = h.ListUseCase.BlackRepo.DeleteIp(ipStruct.Ip)
	default:
		message, flag = h.ListUseCase.GrayRepo.DeleteIp(ipStruct.Ip)
	}

	if flag {
		c.JSON(http.StatusOK, gin.H{
			"list":   listName,
			"status": "ok",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"list":    listName,
			"status":  "error",
			"message": message,
		})
	}
}
