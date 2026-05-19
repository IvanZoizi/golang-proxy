package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"proxy/internal/entity"
	"proxy/internal/repository"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"proxy/internal/usecase"
)

type mockCacheUseCaseForMiddleware struct {
	shouldCache bool
}

func (m *mockCacheUseCaseForMiddleware) ShouldCache(req *http.Request, resp *http.Response) (bool, time.Duration) {
	return m.shouldCache, time.Minute
}

func (m *mockCacheUseCaseForMiddleware) CreateCacheEntry(key string, resp *http.Response, body []byte) (*entity.CacheEntry, error) {
	return nil, nil
}

func (m *mockCacheUseCaseForMiddleware) GetCachedResponse(key string) (*http.Response, []byte, error) {
	return nil, nil, nil
}

func (m *mockCacheUseCaseForMiddleware) Invalidate(req *usecase.InvalidateRequest) (int, error) {
	return 0, nil
}

func (m *mockCacheUseCaseForMiddleware) GetCacheStats() (*repository.CacheStats, error) {
	return nil, nil
}

func (m *mockCacheUseCaseForMiddleware) GetCacheKey(req *http.Request) string {
	return "test-key"
}

func (m *mockCacheUseCaseForMiddleware) CheckMethod(method string) bool {
	return method == "GET"
}

func TestNewCacheMiddleware(t *testing.T) {
	uc := &mockCacheUseCaseForMiddleware{}
	middleware := NewCacheMiddleware(uc)

	if middleware == nil {
		t.Error("expected non-nil middleware")
	}
}

func TestCacheMiddlewareCacheMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockCacheUseCaseForMiddleware{shouldCache: true}
	middleware := NewCacheMiddleware(uc)

	router := gin.New()
	router.Use(middleware.CacheMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "response")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestCacheResponseWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	writer := &CacheResponseWriter{
		ResponseWriter: c.Writer,
		body:           &bytes.Buffer{},
		statusCode:     http.StatusOK,
		headers:        make(http.Header),
	}

	writer.WriteHeader(http.StatusNotFound)
	if writer.Status() != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", writer.Status())
	}

	writer.Write([]byte("test body"))
	if string(writer.Body()) != "test body" {
		t.Errorf("expected 'test body', got %s", string(writer.Body()))
	}
}
