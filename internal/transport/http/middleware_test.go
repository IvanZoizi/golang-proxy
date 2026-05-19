package http

import (
	"net/http"
	"net/http/httptest"
	"proxy/internal/dto"
	"proxy/internal/entity"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"proxy/configGolang"
)

type mockRateLimiterUCForMiddleware struct{}

func (m *mockRateLimiterUCForMiddleware) GetRateLimitConfig() (*entity.RateLimitConfig, error) {
	return nil, nil
}
func (m *mockRateLimiterUCForMiddleware) CheckRequestLimit(ip string) (bool, string, error) {
	return true, "", nil
}
func (m *mockRateLimiterUCForMiddleware) CheckTrafficLimit(ip string, d, u int64) (bool, string, error) {
	return true, "", nil
}
func (m *mockRateLimiterUCForMiddleware) CheckConnectionLimit(ip string, inc bool) (bool, string, error) {
	return true, "", nil
}
func (m *mockRateLimiterUCForMiddleware) GetSubnetLimits(ip string) (*entity.SubnetRateLimit, error) {
	return nil, nil
}
func (m *mockRateLimiterUCForMiddleware) GetIpsByRequest() ([]string, map[string]int) {
	return nil, nil
}
func (m *mockRateLimiterUCForMiddleware) UpdateConfigCache(req dto.UpdateCacheConfigRequest) error {
	return nil
}
func (m *mockRateLimiterUCForMiddleware) CheckNotCacheMethod(path string) (bool, error) {
	return false, nil
}
func (m *mockRateLimiterUCForMiddleware) getCachedData(string) *entity.RateLimitData   { return nil }
func (m *mockRateLimiterUCForMiddleware) saveCachedData(string, *entity.RateLimitData) {}

type mockRateLimiterUC struct {
	checkRequestOK    bool
	checkTrafficOK    bool
	checkConnectionOK bool
}

func (m *mockRateLimiterUC) GetRateLimitConfig() (*entity.RateLimitConfig, error) {
	return &entity.RateLimitConfig{Enabled: true}, nil
}
func (m *mockRateLimiterUC) CheckRequestLimit(ip string) (bool, string, error) {
	return m.checkRequestOK, "", nil
}
func (m *mockRateLimiterUC) CheckTrafficLimit(ip string, d, u int64) (bool, string, error) {
	return m.checkTrafficOK, "", nil
}
func (m *mockRateLimiterUC) CheckConnectionLimit(ip string, inc bool) (bool, string, error) {
	return m.checkConnectionOK, "", nil
}
func (m *mockRateLimiterUC) GetSubnetLimits(ip string) (*entity.SubnetRateLimit, error) {
	return nil, nil
}
func (m *mockRateLimiterUC) GetIpsByRequest() ([]string, map[string]int) { return nil, nil }

func (m *mockRateLimiterUC) getCachedData(ip string) *entity.RateLimitData          { return nil }
func (m *mockRateLimiterUC) saveCachedData(ip string, data *entity.RateLimitData)   {}
func (m *mockRateLimiterUC) resetCountersIfNeeded(*entity.RateLimitData, time.Time) {}
func (m *mockRateLimiterUC) blockIP(*entity.RateLimitData, string, time.Duration)   {}

func TestRateLimiterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := &mockRateLimiterUC{
		checkRequestOK:    true,
		checkTrafficOK:    true,
		checkConnectionOK: true,
	}

	handler := &IpHandler{
		RateLimiterUC: mockUC,
		Config:        &configGolang.Config{SecretKey: "secret"},
	}

	mw := handler.RateLimiterMiddleware()

	t.Run("Allow normal request", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.RemoteAddr = "1.2.3.4:12345"

		mw(c)

		if w.Code != 0 && w.Code != http.StatusOK {
			t.Errorf("expected continue or 200, got %d", w.Code)
		}
	})

	t.Run("Rate limit exceeded", func(t *testing.T) {
		mockUC.checkRequestOK = false

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.RemoteAddr = "1.2.3.4:12345"

		mw(c)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("expected 429, got %d", w.Code)
		}
	})
}

func TestSubnetRateLimiterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &IpHandler{
		RateLimiterUC: &mockRateLimiterUC{},
		Config:        &configGolang.Config{},
	}

	mw := handler.SubnetRateLimiterMiddleware()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	mw(c)
}

func TestIPCheckMiddleware_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := &mockListUseCase{
		contains: map[string]map[string]bool{
			"black": {"bad.ip": true},
			"white": {"good.ip": true},
		},
	}

	handler := &IpHandler{
		ListUseCase: mockUC,
		Config:      &configGolang.Config{SecretKey: "supersecret"},
	}

	mw := handler.IPCheckMiddleware()

	tests := []struct {
		name     string
		ip       string
		token    string
		wantCode int
	}{
		{"Secret Key Bypass", "any.ip", "supersecret", 0},
		{"Blacklisted IP", "bad.ip", "", http.StatusForbidden},
		{"Whitelisted IP", "good.ip", "", 0},
		{"Unknown IP (continue)", "unknown.ip", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Request.RemoteAddr = tt.ip + ":12345"
			if tt.token != "" {
				c.Request.Header.Set("Authorization", "Bearer "+tt.token)
			}

			mw(c)

			if tt.wantCode == 0 {
				if w.Code != 0 && w.Code != http.StatusOK {
					t.Errorf("expected continue, got %d", w.Code)
				}
			} else if w.Code != tt.wantCode {
				t.Errorf("expected %d, got %d", tt.wantCode, w.Code)
			}
		})
	}
}

func TestCacheMiddleware_MethodNotAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockCacheUseCaseForMiddleware{shouldCache: true}
	middleware := NewCacheMiddleware(uc)

	router := gin.New()
	router.Use(middleware.CacheMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.String(200, "post response")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
