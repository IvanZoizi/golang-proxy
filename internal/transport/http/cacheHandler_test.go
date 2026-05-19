package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"proxy/internal/entity"
	"proxy/internal/repository"
	"proxy/internal/usecase"
	"proxy/pkg/logger"
)

type mockCacheUseCase struct {
	stats           *repository.CacheStats
	invalidateCount int
	err             error
}

func (m *mockCacheUseCase) ShouldCache(req *http.Request, resp *http.Response) (bool, time.Duration) {
	return true, time.Minute
}
func (m *mockCacheUseCase) CreateCacheEntry(key string, resp *http.Response, body []byte) (*entity.CacheEntry, error) {
	return nil, nil
}
func (m *mockCacheUseCase) GetCachedResponse(key string) (*http.Response, []byte, error) {
	return nil, nil, nil
}
func (m *mockCacheUseCase) Invalidate(req *usecase.InvalidateRequest) (int, error) {
	return m.invalidateCount, m.err
}
func (m *mockCacheUseCase) GetCacheStats() (*repository.CacheStats, error) {
	return m.stats, nil
}
func (m *mockCacheUseCase) GetCacheKey(req *http.Request) string { return "test-key" }
func (m *mockCacheUseCase) CheckMethod(method string) bool       { return method == "GET" }

func init() {
	gin.SetMode(gin.TestMode)

	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	testLogger, err := config.Build()
	if err != nil {
		panic("failed to create test logger: " + err.Error())
	}

	logger.Log = testLogger
	zap.ReplaceGlobals(testLogger)
}

func TestCacheHandler(t *testing.T) {
	stats := &repository.CacheStats{
		TotalEntries: 42,
		HitCount:     100,
		MissCount:    20,
		HitRate:      0.833,
	}

	mockUC := &mockCacheUseCase{
		stats:           stats,
		invalidateCount: 5,
	}

	cacheMd := NewCacheMiddleware(mockUC)
	config := &entity.CacheConfig{Enabled: true}

	handler := NewCacheHandler(mockUC, config, cacheMd)

	t.Run("GetCacheStats", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/cache/stats", nil)

		handler.GetCacheStats(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("InvalidateCache", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := map[string]interface{}{"type": "all"}
		jsonBody, _ := json.Marshal(body)

		c.Request = httptest.NewRequest("POST", "/api/cache/invalidate", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.InvalidateCache(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GetCacheConfig", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/cache/config", nil)

		handler.GetCacheConfig(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("ClearCache", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/api/cache/clear", nil)

		handler.ClearCache(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("UpdateCacheConfig", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := map[string]interface{}{"enabled": true}
		jsonBody, _ := json.Marshal(body)

		c.Request = httptest.NewRequest("PUT", "/api/cache/config", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.UpdateCacheConfig(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("InvalidateCache_InvalidJSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/cache/invalidate", bytes.NewBuffer([]byte("invalid json")))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.InvalidateCache(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
}
