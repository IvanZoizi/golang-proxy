package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"proxy/configGolang"
	"proxy/internal/entity"
)

func TestSetupRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockCacheUC := &mockCacheUseCaseForMiddleware{shouldCache: true}
	cacheMd := NewCacheMiddleware(mockCacheUC)
	cacheConfig := &entity.CacheConfig{Enabled: true}

	cacheHandler := NewCacheHandler(mockCacheUC, cacheConfig, cacheMd)

	handler := &IpHandler{
		Config: &configGolang.Config{
			ServerPort:      ":8080",
			ProxyServerPort: ":3000",
			SecretKey:       "test-secret",
		},
		ListUseCase:   &mockListUseCase{},
		RateLimiterUC: &mockRateLimiterUCForMiddleware{},
	}

	router := SetupRoute(handler, cacheHandler, cacheConfig)

	tests := []struct {
		method string
		path   string
		code   int
	}{
		{"GET", "/health", http.StatusOK},
		{"GET", "/swagger/index.html", http.StatusOK},
		{"GET", "/metrics", http.StatusOK},
		{"GET", "/api/cache/stats", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)

			router.ServeHTTP(w, req)

			if w.Code != tt.code {
				t.Errorf("path %s: expected %d, got %d", tt.path, tt.code, w.Code)
			}
		})
	}
}
