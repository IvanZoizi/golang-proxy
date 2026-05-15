// internal/repository/RateLimiterRepository.go
package repository

import (
	"encoding/json"
	"os"
	"proxy/iternal/entity"
	"sync"
	"time"
)

type RateLimiterRepository interface {
	GetConfig() (*entity.RateLimitConfig, error)
	UpdateConfig(config *entity.RateLimitConfig) error
	GetData(ip string) (*entity.RateLimitData, error)
	SaveData(data *entity.RateLimitData) error
	DeleteData(ip string) error
	CleanupExpiredData() error
}

type RateLimiterRepositoryImpl struct {
	config     *entity.RateLimitConfig
	configFile string
	data       map[string]*entity.RateLimitData
	dataFile   string
	mu         sync.RWMutex
	dirtyData  map[string]bool
}

func NewRateLimiterRepository(configFile, dataFile string) (*RateLimiterRepositoryImpl, error) {
	repo := &RateLimiterRepositoryImpl{
		configFile: configFile,
		dataFile:   dataFile,
		data:       make(map[string]*entity.RateLimitData),
		dirtyData:  make(map[string]bool),
	}

	if err := repo.loadConfig(); err != nil {
		return nil, err
	}

	if err := repo.loadData(); err != nil {
		return nil, err
	}

	go repo.periodicCleanup()
	go repo.periodicSave()

	return repo, nil
}

func (r *RateLimiterRepositoryImpl) loadConfig() error {
	file, err := os.Open(r.configFile)
	if err != nil {
		r.config = &entity.RateLimitConfig{
			Enabled:                  true,
			RequestsPerSecond:        100,
			RequestsPerMinute:        1000,
			RequestsPerHour:          10000,
			RequestsPerDay:           100000,
			DownloadLimitMB:          100,
			UploadLimitMB:            50,
			TotalTrafficMB:           150,
			MaxConcurrentConnections: 10,
			NewConnectionsPerSecond:  5,
			SubnetLimits:             make(map[string]entity.SubnetRateLimit),
		}
		return r.saveConfig()
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&r.config)
}

func (r *RateLimiterRepositoryImpl) saveConfig() error {
	file, err := os.Create(r.configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(r.config)
}

func (r *RateLimiterRepositoryImpl) loadData() error {
	file, err := os.Open(r.dataFile)
	if err != nil {
		r.data = make(map[string]*entity.RateLimitData)
		return nil
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&r.data)
}

func (r *RateLimiterRepositoryImpl) saveData() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	file, err := os.Create(r.dataFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(r.data)
}

func (r *RateLimiterRepositoryImpl) periodicSave() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		r.mu.Lock()
		needSave := len(r.dirtyData) > 0
		if needSave {
			for k := range r.dirtyData {
				delete(r.dirtyData, k)
			}
		}
		r.mu.Unlock()

		if needSave {
			r.saveData()
		}
	}
}

func (r *RateLimiterRepositoryImpl) GetConfig() (*entity.RateLimitConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config, nil
}

func (r *RateLimiterRepositoryImpl) UpdateConfig(config *entity.RateLimitConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config = config
	return r.saveConfig()
}

func (r *RateLimiterRepositoryImpl) GetData(ip string) (*entity.RateLimitData, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, exists := r.data[ip]
	if !exists {
		data = &entity.RateLimitData{
			IP:                  ip,
			LastResetSecond:     time.Now(),
			LastResetMinute:     time.Now(),
			LastResetHour:       time.Now(),
			LastResetDay:        time.Now(),
			LastConnectionReset: time.Now(),
		}
		r.data[ip] = data
	}
	return data, nil
}

func (r *RateLimiterRepositoryImpl) SaveData(data *entity.RateLimitData) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[data.IP] = data
	r.dirtyData[data.IP] = true
	return nil
}

func (r *RateLimiterRepositoryImpl) DeleteData(ip string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, ip)
	delete(r.dirtyData, ip)
	return r.saveData()
}

func (r *RateLimiterRepositoryImpl) CleanupExpiredData() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for ip, data := range r.data {
		if !data.IsBlocked && now.Sub(data.LastResetDay) > 24*time.Hour {
			delete(r.data, ip)
			delete(r.dirtyData, ip)
		}
	}
	return r.saveData()
}

func (r *RateLimiterRepositoryImpl) periodicCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		r.CleanupExpiredData()
	}
}
