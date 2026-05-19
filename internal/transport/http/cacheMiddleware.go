package http

import (
	"bytes"
	"net/http"
	"proxy/internal/usecase"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"proxy/pkg/logger"
)

type CacheMiddleware struct {
	cacheUC usecase.CacheUseCase
}

func NewCacheMiddleware(cacheUC usecase.CacheUseCase) *CacheMiddleware {
	return &CacheMiddleware{
		cacheUC: cacheUC,
	}
}

type CacheResponseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	headers    http.Header
}

func (w *CacheResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *CacheResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *CacheResponseWriter) Status() int {
	return w.statusCode
}

func (w *CacheResponseWriter) Body() []byte {
	return w.body.Bytes()
}

func (m *CacheMiddleware) CacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.cacheUC.CheckMethod(c.Request.Method) {
			m.cacheUC.Invalidate(usecase.NewInvalidationRequest("all"))
			c.Next()
			return
		}

		cacheKey := m.cacheUC.GetCacheKey(c.Request)

		cachedResp, cachedBody, err := m.cacheUC.GetCachedResponse(cacheKey)
		if err == nil && cachedResp != nil {
			logger.Debug("Cache HIT",
				zap.String("key", cacheKey),
				zap.String("path", c.Request.URL.Path))

			c.Header("X-Cache", "HIT")

			for k, v := range cachedResp.Header {
				for _, val := range v {
					c.Header(k, val)
				}
			}
			c.Status(cachedResp.StatusCode)
			c.Writer.Write(cachedBody)
			c.Abort()
			return
		}

		logger.Debug("Cache MISS",
			zap.String("key", cacheKey),
			zap.String("path", c.Request.URL.Path))

		wrapper := &CacheResponseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
			statusCode:     http.StatusOK,
			headers:        make(http.Header),
		}
		c.Writer = wrapper

		c.Next()

		statusCode := wrapper.Status()

		resp := &http.Response{
			StatusCode: statusCode,
			Header:     wrapper.headers,
			Request:    c.Request,
		}

		shouldCache, ttl := m.cacheUC.ShouldCache(c.Request, resp)

		if shouldCache && statusCode < 400 {
			body := wrapper.Body()

			_, err := m.cacheUC.CreateCacheEntry(cacheKey, resp, body)
			if err == nil {
				logger.Info("Response cached",
					zap.String("key", cacheKey),
					zap.Duration("ttl", ttl),
					zap.Int("size", len(body)))
				c.Header("X-Cache-TTL", ttl.String())
			}
		}

		c.Header("X-Cache", "MISS")
	}
}
