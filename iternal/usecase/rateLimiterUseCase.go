// internal/usecase/rateLimiterUseCase.go
package usecase

import (
	"fmt"
	"sync"
	"time"

	"proxy/iternal/entity"
	"proxy/iternal/repository"
	ip2 "proxy/pkg/ip"
)

type RateLimiterUseCase struct {
	repo    repository.RateLimiterRepository
	cache   map[string]*entity.RateLimitData
	cacheMu sync.RWMutex
}

func NewRateLimiterUseCase(repo repository.RateLimiterRepository) *RateLimiterUseCase {
	uc := &RateLimiterUseCase{
		repo:  repo,
		cache: make(map[string]*entity.RateLimitData),
	}
	go uc.periodicCacheSave()
	return uc
}

func (uc *RateLimiterUseCase) GetRateLimitConfig() (*entity.RateLimitConfig, error) {
	return uc.repo.GetConfig()
}

func (uc *RateLimiterUseCase) getCachedData(ip string) *entity.RateLimitData {
	uc.cacheMu.RLock()
	data, exists := uc.cache[ip]
	uc.cacheMu.RUnlock()

	if exists {
		return data
	}

	uc.cacheMu.Lock()
	defer uc.cacheMu.Unlock()

	data, exists = uc.cache[ip]
	if exists {
		return data
	}

	data, _ = uc.repo.GetData(ip)
	uc.cache[ip] = data
	return data
}

func (uc *RateLimiterUseCase) saveCachedData(ip string, data *entity.RateLimitData) {
	uc.cacheMu.Lock()
	uc.cache[ip] = data
	uc.cacheMu.Unlock()

	go uc.repo.SaveData(data)
}

func (uc *RateLimiterUseCase) periodicCacheSave() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		uc.cacheMu.RLock()
		items := make([]*entity.RateLimitData, 0, len(uc.cache))
		for _, data := range uc.cache {
			items = append(items, data)
		}
		uc.cacheMu.RUnlock()

		for _, data := range items {
			uc.repo.SaveData(data)
		}
	}
}

func (uc *RateLimiterUseCase) CheckRequestLimit(ip string) (bool, string, error) {
	config, err := uc.repo.GetConfig()
	if err != nil || !config.Enabled {
		return true, "", err
	}

	data := uc.getCachedData(ip)
	now := time.Now()

	if data.IsBlocked && now.Before(data.BlockedUntil) {
		return false, fmt.Sprintf("IP blocked until %s: %s", data.BlockedUntil, data.BlockedReason), nil
	} else if data.IsBlocked && now.After(data.BlockedUntil) {
		data.IsBlocked = false
		data.BlockedReason = ""
	}

	uc.resetCountersIfNeeded(data, now)

	if config.RequestsPerSecond > 0 && data.RequestsThisSecond >= config.RequestsPerSecond {
		uc.blockIP(data, "RPS limit exceeded", 60*time.Second)
		uc.saveCachedData(ip, data)
		return false, "Rate limit exceeded: too many requests per second", nil
	}

	if config.RequestsPerMinute > 0 && data.RequestsThisMinute >= config.RequestsPerMinute {
		uc.blockIP(data, "RPM limit exceeded", 5*time.Minute)
		uc.saveCachedData(ip, data)
		return false, "Rate limit exceeded: too many requests per minute", nil
	}

	if config.RequestsPerHour > 0 && data.RequestsThisHour >= config.RequestsPerHour {
		uc.blockIP(data, "RPH limit exceeded", 30*time.Minute)
		uc.saveCachedData(ip, data)
		return false, "Rate limit exceeded: too many requests per hour", nil
	}

	if config.RequestsPerDay > 0 && data.RequestsThisDay >= config.RequestsPerDay {
		uc.blockIP(data, "RPD limit exceeded", 1*time.Hour)
		uc.saveCachedData(ip, data)
		return false, "Rate limit exceeded: too many requests per day", nil
	}

	data.RequestsThisSecond++
	data.RequestsThisMinute++
	data.RequestsThisHour++
	data.RequestsThisDay++

	uc.saveCachedData(ip, data)
	return true, "", nil
}

func (uc *RateLimiterUseCase) CheckTrafficLimit(ip string, downloadBytes, uploadBytes int64) (bool, string, error) {
	config, err := uc.repo.GetConfig()
	if err != nil {
		return true, "", err
	}

	data := uc.getCachedData(ip)

	data.DownloadBytes += downloadBytes
	data.UploadBytes += uploadBytes
	data.TotalBytes += downloadBytes + uploadBytes

	if config.DownloadLimitMB > 0 {
		downloadLimitBytes := int64(config.DownloadLimitMB) * 1024 * 1024
		if data.DownloadBytes > downloadLimitBytes {
			uc.blockIP(data, "Download limit exceeded", 1*time.Hour)
			uc.saveCachedData(ip, data)
			return false, "Download limit exceeded", nil
		}
	}

	if config.UploadLimitMB > 0 {
		uploadLimitBytes := int64(config.UploadLimitMB) * 1024 * 1024
		if data.UploadBytes > uploadLimitBytes {
			uc.blockIP(data, "Upload limit exceeded", 1*time.Hour)
			uc.saveCachedData(ip, data)
			return false, "Upload limit exceeded", nil
		}
	}

	if config.TotalTrafficMB > 0 {
		totalLimitBytes := int64(config.TotalTrafficMB) * 1024 * 1024
		if data.TotalBytes > totalLimitBytes {
			uc.blockIP(data, "Total traffic limit exceeded", 1*time.Hour)
			uc.saveCachedData(ip, data)
			return false, "Total traffic limit exceeded", nil
		}
	}

	uc.saveCachedData(ip, data)
	return true, "", nil
}

func (uc *RateLimiterUseCase) CheckConnectionLimit(ip string, increment bool) (bool, string, error) {
	config, err := uc.repo.GetConfig()
	if err != nil {
		return true, "", err
	}

	data := uc.getCachedData(ip)
	now := time.Now()

	if now.Sub(data.LastConnectionReset) >= time.Second {
		data.ConnectionsThisSec = 0
		data.LastConnectionReset = now
	}

	if increment {
		if config.MaxConcurrentConnections > 0 && data.ActiveConnections >= config.MaxConcurrentConnections {
			return false, "Too many concurrent connections", nil
		}

		if config.NewConnectionsPerSecond > 0 && data.ConnectionsThisSec >= config.NewConnectionsPerSecond {
			return false, "Too many new connections per second", nil
		}

		data.ActiveConnections++
		data.ConnectionsThisSec++
	} else {
		if data.ActiveConnections > 0 {
			data.ActiveConnections--
		}
	}

	uc.saveCachedData(ip, data)
	return true, "", nil
}

func (uc *RateLimiterUseCase) resetCountersIfNeeded(data *entity.RateLimitData, now time.Time) {
	if now.Sub(data.LastResetSecond) >= time.Second {
		data.RequestsThisSecond = 0
		data.LastResetSecond = now
	}

	if now.Sub(data.LastResetMinute) >= time.Minute {
		data.RequestsThisMinute = 0
		data.LastResetMinute = now
	}

	if now.Sub(data.LastResetHour) >= time.Hour {
		data.RequestsThisHour = 0
		data.LastResetHour = now
	}

	if now.Sub(data.LastResetDay) >= 24*time.Hour {
		data.RequestsThisDay = 0
		data.LastResetDay = now
	}
}

func (uc *RateLimiterUseCase) blockIP(data *entity.RateLimitData, reason string, duration time.Duration) {
	data.IsBlocked = true
	data.BlockedUntil = time.Now().Add(duration)
	data.BlockedReason = reason
}

func (uc *RateLimiterUseCase) GetSubnetLimits(ip string) (*entity.SubnetRateLimit, error) {
	config, err := uc.repo.GetConfig()
	if err != nil {
		return nil, err
	}

	for subnet, limits := range config.SubnetLimits {
		if ok, _ := ip2.IsIPInSubnet(ip, subnet); ok {
			return &limits, nil
		}
	}

	return nil, nil
}
