package usecase

import (
	"proxy/internal/entity"
	"testing"
)

type mockListRepo struct {
	ips []string
}

func (m *mockListRepo) GetAllIps() ([]string, error)      { return m.ips, nil }
func (m *mockListRepo) GetAllCIDRIps() ([]string, error)  { return []string{}, nil }
func (m *mockListRepo) GetAllRangeIps() ([]string, error) { return []string{}, nil }
func (m *mockListRepo) AddIp(ip string) error {
	m.ips = append(m.ips, ip)
	return nil
}
func (m *mockListRepo) DeleteIp(ip string) error {
	for i, v := range m.ips {
		if v == ip {
			m.ips = append(m.ips[:i], m.ips[i+1:]...)
			return nil
		}
	}
	return nil
}
func (m *mockListRepo) Contains(ip string) bool {
	for _, v := range m.ips {
		if v == ip {
			return true
		}
	}
	return false
}

func TestListsUseCaseImplGetAllIps(t *testing.T) {
	whiteRepo := &mockListRepo{ips: []string{"192.168.1.1", "192.168.1.2"}}
	grayRepo := &mockListRepo{ips: []string{"10.0.0.1"}}
	blackRepo := &mockListRepo{ips: []string{}}

	uc := CreateListUseCase(whiteRepo, grayRepo, blackRepo)

	ips, _ := uc.GetAllIps("white")
	if len(ips) != 2 {
		t.Errorf("expected 2, got %d", len(ips))
	}

	ips, _ = uc.GetAllIps("gray")
	if len(ips) != 1 {
		t.Errorf("expected 1, got %d", len(ips))
	}

	ips, _ = uc.GetAllIps("black")
	if len(ips) != 0 {
		t.Errorf("expected 0, got %d", len(ips))
	}
}

func TestListsUseCaseImplAddIp(t *testing.T) {
	whiteRepo := &mockListRepo{ips: []string{}}
	grayRepo := &mockListRepo{ips: []string{}}
	blackRepo := &mockListRepo{ips: []string{}}

	uc := CreateListUseCase(whiteRepo, grayRepo, blackRepo)

	err := uc.AddIp("192.168.1.1", "white")
	if err != nil {
		t.Errorf("AddIp failed: %v", err)
	}

	ips, _ := uc.GetAllIps("white")
	if len(ips) != 1 {
		t.Errorf("expected 1, got %d", len(ips))
	}
}

func TestListsUseCaseImplContains(t *testing.T) {
	whiteRepo := &mockListRepo{ips: []string{"192.168.1.1"}}
	grayRepo := &mockListRepo{ips: []string{}}
	blackRepo := &mockListRepo{ips: []string{}}

	uc := CreateListUseCase(whiteRepo, grayRepo, blackRepo)

	exists, _ := uc.Contains("192.168.1.1", "white")
	if !exists {
		t.Error("expected true")
	}

	exists, _ = uc.Contains("10.0.0.1", "white")
	if exists {
		t.Error("expected false")
	}
}

func TestListsUseCaseImplDeleteIp(t *testing.T) {
	whiteRepo := &mockListRepo{ips: []string{"192.168.1.1"}}
	grayRepo := &mockListRepo{ips: []string{}}
	blackRepo := &mockListRepo{ips: []string{}}

	uc := CreateListUseCase(whiteRepo, grayRepo, blackRepo)

	err := uc.DeleteIp("192.168.1.1", "white")
	if err != nil {
		t.Errorf("DeleteIp failed: %v", err)
	}

	ips, _ := uc.GetAllIps("white")
	if len(ips) != 0 {
		t.Errorf("expected 0, got %d", len(ips))
	}
}

func TestRateLimiterUseCaseGetCachedData(t *testing.T) {
	repo := &mockRateLimiterRepo{
		config: &entity.RateLimitConfig{Enabled: true},
		data:   make(map[string]*entity.RateLimitData),
	}
	uc := newTestRateLimiterUseCase(repo)

	data := uc.getCachedData("192.168.1.1")
	if data == nil {
		t.Error("expected non-nil data")
	}
	if data.IP != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1, got %s", data.IP)
	}
}

func TestRateLimiterUseCaseSaveCachedData(t *testing.T) {
	repo := &mockRateLimiterRepo{
		config: &entity.RateLimitConfig{Enabled: true},
		data:   make(map[string]*entity.RateLimitData),
	}
	uc := newTestRateLimiterUseCase(repo)

	data := &entity.RateLimitData{IP: "192.168.1.1", RequestsThisSecond: 99}
	uc.saveCachedData("192.168.1.1", data)

	cached := uc.getCachedData("192.168.1.1")
	if cached.RequestsThisSecond != 99 {
		t.Errorf("expected 99, got %d", cached.RequestsThisSecond)
	}
}
