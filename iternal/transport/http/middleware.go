package http

import (
	"bytes"
	"io"
	"net/http"
	"proxy/pkg/utils"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	ip2 "proxy/pkg/ip"
	"proxy/pkg/logger"
)

type ResponseWriterWrapper struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
	size   int64
}

func (w *ResponseWriterWrapper) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += int64(size)
	if w.body != nil {
		w.body.Write(b)
	}
	return size, err
}

func (w *ResponseWriterWrapper) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *ResponseWriterWrapper) GetSize() int64 {
	return w.size
}

type SubnetLimiter struct {
	requestsThisSecond int
	requestsThisMinute int
	requestsThisHour   int
	requestsThisDay    int
	lastResetSecond    time.Time
	lastResetMinute    time.Time
	lastResetHour      time.Time
	lastResetDay       time.Time
}

func (h *IpHandler) IPCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		ipClient := ip2.GetClientIP(c.Request)

		proxyKey, _ := utils.ExtractBearerToken(c.Request)
		logger.Info("New request",
			zap.String("IP", ipClient),
			zap.String("path", c.Request.URL.Path),
			zap.String("X-Proxy-Key", proxyKey))

		if proxyKey == h.Config.SecretKey {
			logger.Info("[OFFICIAL] IP in whitelist",
				zap.String("ip", ipClient),
				zap.String("path", c.Request.URL.String()))
			c.Set("ip_status", "whitelist")
			c.Set("client_ip", ipClient)
			c.Next()
		}

		if flag, _ := h.ListUseCase.Contains(ipClient, "black"); flag {
			logger.Warn("[BLOCKED] IP in blacklist",
				zap.String("ip", ipClient),
				zap.String("path", c.Request.URL.String()))
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Access denied",
				"message": "Your IP is blocked",
				"ip":      ipClient,
			})
			c.Abort()
			return
		}

		blackIpsCIDR, _ := h.ListUseCase.GetAllCIDRIps("black")
		for _, str := range blackIpsCIDR {
			if flag, _ := ip2.IsIPInSubnet(ipClient, str); flag {
				logger.Warn("[BLOCKED] IP in blacklist CIDR",
					zap.String("ip", ipClient),
					zap.String("cidr", str),
					zap.String("path", c.Request.URL.String()))
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Access denied",
					"message": "Your IP is blocked",
					"ip":      ipClient,
				})
				c.Abort()
				return
			}
		}

		blackIpsRange, _ := h.ListUseCase.GetAllRangeIps("black")
		for _, str := range blackIpsRange {
			if flag, _ := ip2.IsIPInRange(ipClient, str); flag {
				logger.Warn("[BLOCKED] IP in blacklist range",
					zap.String("ip", ipClient),
					zap.String("range", str),
					zap.String("path", c.Request.URL.String()))
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Access denied",
					"message": "Your IP is blocked",
					"ip":      ipClient,
				})
				c.Abort()
				return
			}
		}
		if flag, _ := h.ListUseCase.Contains(ipClient, "white"); flag {
			logger.Info("[ALLOWED] IP in whitelist",
				zap.String("ip", ipClient),
				zap.String("path", c.Request.URL.String()))
			c.Set("ip_status", "whitelist")
			c.Set("client_ip", ipClient)
			c.Next()
			return
		}

		whiteIpsListCIDR, _ := h.ListUseCase.GetAllCIDRIps("white")
		for _, str := range whiteIpsListCIDR {
			if flag, _ := ip2.IsIPInSubnet(ipClient, str); flag {
				logger.Info("[ALLOWED] IP in whitelist CIDR",
					zap.String("ip", ipClient),
					zap.String("cidr", str),
					zap.String("path", c.Request.URL.String()))
				c.Set("ip_status", "whitelist")
				c.Set("client_ip", ipClient)
				c.Next()
				return
			}
		}

		whiteIpsRange, _ := h.ListUseCase.GetAllRangeIps("white")
		for _, str := range whiteIpsRange {
			if flag, _ := ip2.IsIPInRange(ipClient, str); flag {
				logger.Info("[ALLOWED] IP in whitelist range",
					zap.String("ip", ipClient),
					zap.String("range", str),
					zap.String("path", c.Request.URL.String()))
				c.Set("ip_status", "whitelist")
				c.Set("client_ip", ipClient)
				c.Next()
				return
			}
		}

		if flag, _ := h.ListUseCase.Contains(ipClient, "gray"); flag {
			logger.Info("[GRAYLIST] IP in graylist",
				zap.String("ip", ipClient),
				zap.String("path", c.Request.URL.String()))
			c.Set("ip_status", "graylist")
			c.Set("client_ip", ipClient)
			c.Next()
			return
		}

		grayIpsListCIDR, _ := h.ListUseCase.GetAllCIDRIps("gray")
		for _, str := range grayIpsListCIDR {
			if flag, _ := ip2.IsIPInSubnet(ipClient, str); flag {
				logger.Info("[GRAYLIST] IP in graylist CIDR",
					zap.String("ip", ipClient),
					zap.String("cidr", str),
					zap.String("path", c.Request.URL.String()))
				c.Set("ip_status", "graylist")
				c.Set("client_ip", ipClient)
				c.Next()
				return
			}
		}

		grayIpsRange, _ := h.ListUseCase.GetAllRangeIps("gray")
		for _, str := range grayIpsRange {
			if flag, _ := ip2.IsIPInRange(ipClient, str); flag {
				logger.Info("[GRAYLIST] IP in graylist range",
					zap.String("ip", ipClient),
					zap.String("range", str),
					zap.String("path", c.Request.URL.String()))
				c.Set("ip_status", "graylist")
				c.Set("client_ip", ipClient)
				c.Next()
				return
			}
		}

		logger.Info("[UNKNOWN] IP not in any list",
			zap.String("ip", ipClient),
			zap.String("path", c.Request.URL.String()))
		c.Set("ip_status", "unknown")
		c.Set("client_ip", ipClient)
		c.Next()
	}
}

func (h *IpHandler) RateLimiterMiddleware() gin.HandlerFunc {
	activeConns := sync.Map{}

	return func(c *gin.Context) {
		clientIP := ip2.GetClientIP(c.Request)

		proxyKey, _ := utils.ExtractBearerToken(c.Request)
		logger.Info("New request",
			zap.String("IP", clientIP),
			zap.String("path", c.Request.URL.Path),
			zap.String("X-Proxy-Key", proxyKey))

		if proxyKey == h.Config.SecretKey {
			c.Next()
		}

		if h.RateLimiterUC == nil {
			logger.Warn("Rate limiter not initialized, skipping")
			c.Next()
			return
		}

		logger.Debug("Rate limit check",
			zap.String("ip", clientIP),
			zap.String("path", c.Request.URL.Path))

		connKey := clientIP + ":in"
		if _, loaded := activeConns.LoadOrStore(connKey, true); !loaded {
			ok, msg, err := h.RateLimiterUC.CheckConnectionLimit(clientIP, true)
			if err != nil {
				logger.Error("Connection limit check error", zap.Error(err))
			}
			if !ok {
				logger.Warn("Connection limit exceeded",
					zap.String("ip", clientIP),
					zap.String("message", msg))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":   "Connection limit exceeded",
					"message": msg,
				})
				c.Abort()
				return
			}
			defer func() {
				activeConns.Delete(connKey)
				h.RateLimiterUC.CheckConnectionLimit(clientIP, false)
				logger.Debug("Connection released",
					zap.String("ip", clientIP))
			}()
		}

		ok, msg, err := h.RateLimiterUC.CheckRequestLimit(clientIP)
		if err != nil {
			logger.Error("Request limit check error", zap.Error(err))
		}
		if !ok {
			logger.Warn("Rate limit exceeded",
				zap.String("ip", clientIP),
				zap.String("path", c.Request.URL.Path),
				zap.String("reason", msg))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     msg,
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		var requestSize int64
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			requestSize = int64(len(bodyBytes))
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		wrapper := &ResponseWriterWrapper{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
			status:         http.StatusOK,
		}
		c.Writer = wrapper

		c.Next()

		ok, msg, err = h.RateLimiterUC.CheckTrafficLimit(clientIP, requestSize, wrapper.GetSize())
		if err != nil {
			logger.Error("Traffic limit check error", zap.Error(err))
		}
		if !ok {
			logger.Warn("Traffic limit exceeded",
				zap.String("ip", clientIP),
				zap.Int64("request_size", requestSize),
				zap.Int64("response_size", wrapper.GetSize()),
				zap.String("message", msg))
		}

		logger.Debug("Rate limit stats",
			zap.String("ip", clientIP),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", wrapper.status),
			zap.Int64("request_size", requestSize),
			zap.Int64("response_size", wrapper.GetSize()))
	}
}

func (h *IpHandler) SubnetRateLimiterMiddleware() gin.HandlerFunc {
	subnetLimiters := sync.Map{}

	return func(c *gin.Context) {
		clientIP := ip2.GetClientIP(c.Request)

		proxyKey, _ := utils.ExtractBearerToken(c.Request)
		logger.Info("New request",
			zap.String("IP", clientIP),
			zap.String("path", c.Request.URL.Path),
			zap.String("X-Proxy-Key", proxyKey))

		if proxyKey == h.Config.SecretKey {
			c.Next()
		}

		if h.RateLimiterUC == nil {
			c.Next()
			return
		}

		config, err := h.RateLimiterUC.GetRateLimitConfig()
		if err != nil {
			logger.Error("Failed to get rate limit config", zap.Error(err))
			c.Next()
			return
		}

		subnetLimits, err := h.RateLimiterUC.GetSubnetLimits(clientIP)
		if err != nil || subnetLimits == nil {
			c.Next()
			return
		}

		var subnetKey string
		for subnet := range config.SubnetLimits {
			if ok, _ := ip2.IsIPInSubnet(clientIP, subnet); ok {
				subnetKey = subnet
				break
			}
		}

		if subnetKey == "" {
			c.Next()
			return
		}

		limiterInterface, _ := subnetLimiters.LoadOrStore(subnetKey, &SubnetLimiter{
			requestsThisSecond: 0,
			requestsThisMinute: 0,
			requestsThisHour:   0,
			requestsThisDay:    0,
			lastResetSecond:    time.Now(),
			lastResetMinute:    time.Now(),
			lastResetHour:      time.Now(),
			lastResetDay:       time.Now(),
		})
		limiter := limiterInterface.(*SubnetLimiter)

		now := time.Now()

		if now.Sub(limiter.lastResetSecond) >= time.Second {
			limiter.requestsThisSecond = 0
			limiter.lastResetSecond = now
		}
		if now.Sub(limiter.lastResetMinute) >= time.Minute {
			limiter.requestsThisMinute = 0
			limiter.lastResetMinute = now
		}
		if now.Sub(limiter.lastResetHour) >= time.Hour {
			limiter.requestsThisHour = 0
			limiter.lastResetHour = now
		}
		if now.Sub(limiter.lastResetDay) >= 24*time.Hour {
			limiter.requestsThisDay = 0
			limiter.lastResetDay = now
		}

		if subnetLimits.RequestsPerSecond > 0 && limiter.requestsThisSecond >= subnetLimits.RequestsPerSecond {
			logger.Warn("Subnet rate limit exceeded",
				zap.String("subnet", subnetKey),
				zap.String("ip", clientIP),
				zap.Int("limit", subnetLimits.RequestsPerSecond))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Subnet rate limit exceeded",
				"message": "Too many requests from your subnet",
				"subnet":  subnetKey,
			})
			c.Abort()
			return
		}

		if subnetLimits.RequestsPerMinute > 0 && limiter.requestsThisMinute >= subnetLimits.RequestsPerMinute {
			logger.Warn("Subnet rate limit exceeded",
				zap.String("subnet", subnetKey),
				zap.String("ip", clientIP),
				zap.Int("limit", subnetLimits.RequestsPerMinute))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Subnet rate limit exceeded",
				"message": "Too many requests from your subnet",
				"subnet":  subnetKey,
			})
			c.Abort()
			return
		}

		limiter.requestsThisSecond++
		limiter.requestsThisMinute++
		limiter.requestsThisHour++
		limiter.requestsThisDay++

		logger.Debug("Subnet rate limit stats",
			zap.String("subnet", subnetKey),
			zap.Int("rps", limiter.requestsThisSecond),
			zap.Int("rpm", limiter.requestsThisMinute))

		c.Next()
	}
}
