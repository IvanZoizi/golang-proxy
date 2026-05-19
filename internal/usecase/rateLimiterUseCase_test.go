package usecase

import (
	"proxy/internal/entity"
	"proxy/internal/repository"
	"sync"
	"testing"
)

type mockRateLimiterRepo struct {
	config *entity.RateLimitConfig
	data   map[string]*entity.RateLimitData
	mu     sync.Mutex
}

func (m *mockRateLimiterRepo) GetConfig() (*entity.RateLimitConfig, error) {
	return m.config, nil
}

func (m *mockRateLimiterRepo) UpdateConfig(config *entity.RateLimitConfig) error {
	m.config = config
	return nil
}

func (m *mockRateLimiterRepo) GetData(ip string) (*entity.RateLimitData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if d, ok := m.data[ip]; ok {
		return d, nil
	}
	return &entity.RateLimitData{IP: ip}, nil
}

func (m *mockRateLimiterRepo) SaveData(data *entity.RateLimitData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[data.IP] = data
	return nil
}

func (m *mockRateLimiterRepo) DeleteData(ip string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, ip)
	return nil
}

func (m *mockRateLimiterRepo) CleanupExpiredData() error { return nil }

func (m *mockRateLimiterRepo) GetIpsByRequest() ([]string, map[string]int) {
	return nil, nil
}

func newTestRateLimiterUseCase(repo repository.RateLimiterRepository) *RateLimiterUseCaseImpl {
	uc := &RateLimiterUseCaseImpl{
		repo:  repo,
		cache: make(map[string]*entity.RateLimitData),
	}
	return uc
}

func TestRateLimiterUseCaseCheckRequestLimit(t *testing.T) {
	repo := &mockRateLimiterRepo{
		config: &entity.RateLimitConfig{
			Enabled:           true,
			RequestsPerSecond: 2,
			RequestsPerMinute: 10,
		},
		data: make(map[string]*entity.RateLimitData),
	}
	uc := newTestRateLimiterUseCase(repo)

	ok, _, _ := uc.CheckRequestLimit("192.168.1.1")
	if !ok {
		t.Error("expected true for first request")
	}

	ok, _, _ = uc.CheckRequestLimit("192.168.1.1")
	if !ok {
		t.Error("expected true for second request")
	}

	ok, _, _ = uc.CheckRequestLimit("192.168.1.1")
	if ok {
		t.Error("expected false for third request (exceeds RPS)")
	}
}

func TestRateLimiterUseCaseCheckRequestLimitDisabled(t *testing.T) {
	repo := &mockRateLimiterRepo{
		config: &entity.RateLimitConfig{Enabled: false},
		data:   make(map[string]*entity.RateLimitData),
	}
	uc := newTestRateLimiterUseCase(repo)

	for i := 0; i < 100; i++ {
		ok, _, _ := uc.CheckRequestLimit("192.168.1.1")
		if !ok {
			t.Error("expected true when disabled")
		}
	}
}

func TestRateLimiterUseCaseCheckTrafficLimit(t *testing.T) {
	repo := &mockRateLimiterRepo{
		config: &entity.RateLimitConfig{
			Enabled:         true,
			DownloadLimitMB: 1,
		},
		data: make(map[string]*entity.RateLimitData),
	}
	uc := newTestRateLimiterUseCase(repo)

	ok, msg, _ := uc.CheckTrafficLimit("192.168.1.1", 2*1024*1024, 0)
	if ok {
		t.Error("expected false for exceeding download limit")
	}
	if msg == "" {
		t.Error("expected error message")
	}
}

func TestRateLimiterUseCaseCheckConnectionLimit(t *testing.T) {
	repo := &mockRateLimiterRepo{
		config: &entity.RateLimitConfig{
			MaxConcurrentConnections: 2,
		},
		data: make(map[string]*entity.RateLimitData),
	}
	uc := newTestRateLimiterUseCase(repo)

	ok, _, _ := uc.CheckConnectionLimit("192.168.1.1", true)
	if !ok {
		t.Error("expected true for first connection")
	}

	ok, _, _ = uc.CheckConnectionLimit("192.168.1.1", true)
	if !ok {
		t.Error("expected true for second connection")
	}

	ok, _, _ = uc.CheckConnectionLimit("192.168.1.1", true)
	if ok {
		t.Error("expected false for third connection")
	}
}

func TestRateLimiterUseCaseBlockIP(t *testing.T) {
	repo := &mockRateLimiterRepo{
		config: &entity.RateLimitConfig{
			Enabled:           true,
			RequestsPerSecond: 1,
		},
		data: make(map[string]*entity.RateLimitData),
	}
	uc := newTestRateLimiterUseCase(repo)

	ok, _, _ := uc.CheckRequestLimit("192.168.1.1")
	if !ok {
		t.Error("first request should be allowed")
	}

	ok, _, _ = uc.CheckRequestLimit("192.168.1.1")
	if ok {
		t.Error("second request should be blocked")
	}

	data := uc.getCachedData("192.168.1.1")
	if !data.IsBlocked {
		t.Error("IP should be blocked after exceeding limit")
	}
}
