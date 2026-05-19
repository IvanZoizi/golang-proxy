package usecase

import (
	"net/http"
	"proxy/internal/entity"
	"proxy/internal/repository"
	"testing"
	"time"
)

type mockCacheRepository struct {
	data map[string]*entity.CacheEntry
}

func (m *mockCacheRepository) Get(key string) (*entity.CacheEntry, error) {
	if entry, ok := m.data[key]; ok {
		return entry, nil
	}
	return nil, nil
}

func (m *mockCacheRepository) Set(entry *entity.CacheEntry) error {
	m.data[entry.Key] = entry
	return nil
}

func (m *mockCacheRepository) Delete(key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockCacheRepository) DeleteByPrefix(prefix string) (int, error) {
	count := 0
	for k := range m.data {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(m.data, k)
			count++
		}
	}
	return count, nil
}

func (m *mockCacheRepository) DeleteByRegex(pattern string) (int, error) { return 0, nil }
func (m *mockCacheRepository) DeleteByTag(tag string) (int, error)       { return 0, nil }
func (m *mockCacheRepository) DeleteAll() (int, error) {
	count := len(m.data)
	m.data = make(map[string]*entity.CacheEntry)
	return count, nil
}
func (m *mockCacheRepository) GetStats() (*repository.CacheStats, error) {
	return &repository.CacheStats{}, nil
}
func (m *mockCacheRepository) Cleanup() (int, error)                           { return 0, nil }
func (m *mockCacheRepository) GetKeysByPrefix(prefix string) ([]string, error) { return nil, nil }
func (m *mockCacheRepository) UpdateAccess(key string) error                   { return nil }

func TestCacheUseCaseGetCacheKey(t *testing.T) {
	config := &entity.CacheConfig{CacheMethods: []string{"GET"}}
	uc := NewCacheUseCase(&mockCacheRepository{data: make(map[string]*entity.CacheEntry)}, config)

	req, _ := http.NewRequest("GET", "http://example.com/api/test", nil)
	key1 := uc.GetCacheKey(req)
	key2 := uc.GetCacheKey(req)

	if key1 != key2 {
		t.Error("same request should produce same key")
	}
}

func TestCacheUseCaseCheckMethod(t *testing.T) {
	config := &entity.CacheConfig{CacheMethods: []string{"GET", "HEAD"}}
	uc := NewCacheUseCase(&mockCacheRepository{data: make(map[string]*entity.CacheEntry)}, config)

	if !uc.CheckMethod("GET") {
		t.Error("expected true for GET")
	}
	if uc.CheckMethod("POST") {
		t.Error("expected false for POST")
	}
}

func TestCacheUseCaseShouldCacheDisabled(t *testing.T) {
	config := &entity.CacheConfig{Enabled: false}
	uc := NewCacheUseCase(&mockCacheRepository{data: make(map[string]*entity.CacheEntry)}, config)

	req, _ := http.NewRequest("GET", "http://test.com", nil)
	resp := &http.Response{StatusCode: 200}

	ok, _ := uc.ShouldCache(req, resp)
	if ok {
		t.Error("should not cache when disabled")
	}
}

func TestCacheUseCaseShouldCacheMethodNotAllowed(t *testing.T) {
	config := &entity.CacheConfig{Enabled: true, CacheMethods: []string{"GET"}}
	uc := NewCacheUseCase(&mockCacheRepository{data: make(map[string]*entity.CacheEntry)}, config)

	req, _ := http.NewRequest("POST", "http://test.com", nil)
	resp := &http.Response{StatusCode: 200}

	ok, _ := uc.ShouldCache(req, resp)
	if ok {
		t.Error("should not cache POST method")
	}
}

func TestCacheUseCaseInvalidateAll(t *testing.T) {
	config := &entity.CacheConfig{}
	mockRepo := &mockCacheRepository{data: make(map[string]*entity.CacheEntry)}
	uc := NewCacheUseCase(mockRepo, config)

	mockRepo.Set(&entity.CacheEntry{Key: "key1", Value: []byte("v1")})
	mockRepo.Set(&entity.CacheEntry{Key: "key2", Value: []byte("v2")})

	count, err := uc.Invalidate(&InvalidateRequest{Type: "all"})
	if err != nil {
		t.Errorf("invalidate failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 deleted, got %d", count)
	}
}

func TestCacheUseCaseInvalidateExact(t *testing.T) {
	config := &entity.CacheConfig{}
	mockRepo := &mockCacheRepository{data: make(map[string]*entity.CacheEntry)}
	uc := NewCacheUseCase(mockRepo, config)

	mockRepo.Set(&entity.CacheEntry{Key: "test", Value: []byte("v")})

	count, err := uc.Invalidate(&InvalidateRequest{Type: "exact", Key: "test"})
	if err != nil {
		t.Errorf("invalidate failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 deleted, got %d", count)
	}
}

func TestCacheUseCaseInvalidatePrefix(t *testing.T) {
	config := &entity.CacheConfig{}
	mockRepo := &mockCacheRepository{data: make(map[string]*entity.CacheEntry)}
	uc := NewCacheUseCase(mockRepo, config)

	mockRepo.Set(&entity.CacheEntry{Key: "user:1", Value: []byte("v")})
	mockRepo.Set(&entity.CacheEntry{Key: "user:2", Value: []byte("v")})
	mockRepo.Set(&entity.CacheEntry{Key: "admin:1", Value: []byte("v")})

	count, err := uc.Invalidate(&InvalidateRequest{Type: "prefix", Prefix: "user:"})
	if err != nil {
		t.Errorf("invalidate failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 deleted, got %d", count)
	}
}

func TestCacheUseCaseShouldCacheWithTTL(t *testing.T) {
	config := &entity.CacheConfig{
		Enabled:      true,
		CacheMethods: []string{"GET"},
		Cache2xx:     true,
		TTL2xx:       time.Minute,
	}
	mockRepo := &mockCacheRepository{data: make(map[string]*entity.CacheEntry)}
	uc := NewCacheUseCase(mockRepo, config)

	req, _ := http.NewRequest("GET", "http://test.com", nil)
	resp := &http.Response{StatusCode: 200}

	ok, ttl := uc.ShouldCache(req, resp)
	if !ok {
		t.Error("should cache 200 response")
	}
	if ttl != time.Minute {
		t.Errorf("expected %v, got %v", time.Minute, ttl)
	}
}

func TestCacheUseCaseShouldCacheRespectNoCache(t *testing.T) {
	config := &entity.CacheConfig{
		Enabled:        true,
		CacheMethods:   []string{"GET"},
		Cache2xx:       true,
		RespectNoCache: true,
	}
	mockRepo := &mockCacheRepository{data: make(map[string]*entity.CacheEntry)}
	uc := NewCacheUseCase(mockRepo, config)

	req, _ := http.NewRequest("GET", "http://test.com", nil)
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Cache-Control": []string{"no-cache"}},
	}

	ok, _ := uc.ShouldCache(req, resp)
	if ok {
		t.Error("should not cache when no-cache header present")
	}
}

func TestCacheUseCaseGetTTLForStatus(t *testing.T) {
	config := &entity.CacheConfig{
		Cache2xx: true,
		Cache3xx: true,
		Cache4xx: false,
		Cache5xx: false,
		TTL2xx:   time.Minute,
		TTL3xx:   30 * time.Second,
		TTL4xx:   0,
		TTL5xx:   0,
	}
	mockRepo := &mockCacheRepository{data: make(map[string]*entity.CacheEntry)}
	uc := NewCacheUseCase(mockRepo, config)

	tests := []struct {
		status int
		want   time.Duration
	}{
		{200, time.Minute},
		{302, 30 * time.Second},
		{404, 0},
		{500, 0},
	}

	for _, tt := range tests {
		ttl := uc.getTTLForStatus(tt.status)
		if ttl != tt.want {
			t.Errorf("status %d: expected %v, got %v", tt.status, tt.want, ttl)
		}
	}
}

func TestCacheUseCaseMatchPattern(t *testing.T) {
	config := &entity.CacheConfig{}
	mockRepo := &mockCacheRepository{data: make(map[string]*entity.CacheEntry)}
	uc := NewCacheUseCase(mockRepo, config)

	tests := []struct {
		url       string
		host      string
		pattern   string
		matchType string
		want      bool
	}{
		{"http://test.com/api", "", "http://test.com/api", "exact", true},
		{"http://test.com/api", "", "wrong", "exact", false},
		{"http://test.com/api", "test.com", "test.com", "domain", true},
		{"http://test.com/api", "", "/api", "path", true},
	}

	for _, tt := range tests {
		got := uc.matchPattern(tt.url, tt.host, tt.pattern, tt.matchType)
		if got != tt.want {
			t.Errorf("matchPattern(%q, %q, %q, %q) = %v, want %v",
				tt.url, tt.host, tt.pattern, tt.matchType, got, tt.want)
		}
	}
}

func TestCacheUseCaseParseMaxAge(t *testing.T) {
	tests := []struct {
		cacheControl string
		want         int
	}{
		{"max-age=3600", 3600},
		{"max-age=60, public", 60},
		{"no-cache", 0},
		{"", 0},
		{"public, max-age=120", 120},
	}

	for _, tt := range tests {
		got := parseMaxAge(tt.cacheControl)
		if got != tt.want {
			t.Errorf("parseMaxAge(%q) = %d, want %d", tt.cacheControl, got, tt.want)
		}
	}
}
