package repository

import (
	"os"
	"path/filepath"
	"proxy/internal/entity"
	"testing"
)

func tempRateLimiterFiles(t *testing.T) (string, string) {
	dir := os.TempDir()
	return filepath.Join(dir, "test_rate_config.json"), filepath.Join(dir, "test_rate_data.json")
}

func TestNewRateLimiterRepository(t *testing.T) {
	configFile, dataFile := tempRateLimiterFiles(t)
	defer os.Remove(configFile)
	defer os.Remove(dataFile)

	repo, err := NewRateLimiterRepository(configFile, dataFile)
	if err != nil {
		t.Fatalf("NewRateLimiterRepository failed: %v", err)
	}
	if repo == nil || repo.config == nil {
		t.Fatal("Repository or config is nil")
	}
}

func TestRateLimiterRepository_Config(t *testing.T) {
	configFile, dataFile := tempRateLimiterFiles(t)
	defer os.Remove(configFile)
	defer os.Remove(dataFile)

	repo, _ := NewRateLimiterRepository(configFile, dataFile)

	cfg, err := repo.GetConfig()
	if err != nil || cfg == nil {
		t.Fatal("GetConfig failed")
	}
	if !cfg.Enabled {
		t.Error("Default config should be enabled")
	}

	newCfg := &entity.RateLimitConfig{
		Enabled:           false,
		RequestsPerSecond: 50,
	}
	err = repo.UpdateConfig(newCfg)
	if err != nil {
		t.Errorf("UpdateConfig failed: %v", err)
	}

	updated, _ := repo.GetConfig()
	if updated.Enabled != false || updated.RequestsPerSecond != 50 {
		t.Error("Config was not updated correctly")
	}
}

func TestRateLimiterRepository_GetIpsByRequest(t *testing.T) {
	configFile, dataFile := tempRateLimiterFiles(t)
	defer os.Remove(configFile)
	defer os.Remove(dataFile)

	repo, _ := NewRateLimiterRepository(configFile, dataFile)

	repo.SaveData(&entity.RateLimitData{IP: "1.1.1.1", RequestsThisDay: 500})
	repo.SaveData(&entity.RateLimitData{IP: "2.2.2.2", RequestsThisDay: 1500})
	repo.SaveData(&entity.RateLimitData{IP: "3.3.3.3", RequestsThisDay: 100})

	keys, counts := repo.GetIpsByRequest()

	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
	if counts[keys[0]] != 1500 {
		t.Error("Keys should be sorted by RequestsThisDay descending")
	}
}
