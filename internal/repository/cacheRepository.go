package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"proxy/internal/entity"
	"sync"
	"time"
)

type CacheRepository interface {
	Get(key string) (*entity.CacheEntry, error)
	Set(entry *entity.CacheEntry) error
	Delete(key string) error
	DeleteByPrefix(prefix string) (int, error)
	DeleteByRegex(pattern string) (int, error)
	DeleteByTag(tag string) (int, error)
	DeleteAll() (int, error)
	GetStats() (*CacheStats, error)
	Cleanup() (int, error)
	GetKeysByPrefix(prefix string) ([]string, error)
	UpdateAccess(key string) error
}

type CacheStats struct {
	TotalEntries int     `json:"total_entries"`
	TotalSize    int64   `json:"total_size_bytes"`
	HitCount     int64   `json:"hit_count"`
	MissCount    int64   `json:"miss_count"`
	HitRate      float64 `json:"hit_rate"`
	EvictedCount int64   `json:"evicted_count"`
}

type CacheRepositoryImpl struct {
	data       map[string]*entity.CacheEntry
	mu         sync.RWMutex
	filePath   string
	stats      *CacheStats
	statsMu    sync.RWMutex
	maxSize    int64
	maxEntries int
}

func NewCacheRepository(filePath string, maxSize int64, maxEntries int) (*CacheRepositoryImpl, error) {
	repo := &CacheRepositoryImpl{
		data:       make(map[string]*entity.CacheEntry),
		filePath:   filePath,
		maxSize:    maxSize,
		maxEntries: maxEntries,
		stats: &CacheStats{
			HitCount:     0,
			MissCount:    0,
			EvictedCount: 0,
		},
	}

	if err := repo.load(); err != nil {
		return nil, err
	}

	go repo.periodicSave()
	go repo.periodicCleanup()

	return repo, nil
}

func (r *CacheRepositoryImpl) Get(key string) (*entity.CacheEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fmt.Println("CacheRepositoryImpl ", r.data)

	entry, exists := r.data[key]
	if !exists {
		r.recordMiss()
		return nil, nil
	}

	if time.Now().After(entry.ExpiresAt) {
		go r.Delete(key)
		r.recordMiss()
		return nil, nil
	}

	entry.LastAccess = time.Now()
	entry.HitCount++
	r.recordHit()

	return entry, nil
}

func (r *CacheRepositoryImpl) Set(entry *entity.CacheEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if entry.Size > r.maxSize {
		return nil
	}

	for r.getTotalSizeLocked()+entry.Size > r.maxSize || len(r.data) >= r.maxEntries {
		r.evictOneLocked()
	}

	r.data[entry.Key] = entry
	return nil
}

func (r *CacheRepositoryImpl) Delete(key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, key)
	return nil
}

func (r *CacheRepositoryImpl) DeleteByPrefix(prefix string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for key := range r.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(r.data, key)
			count++
		}
	}
	return count, nil
}

func (r *CacheRepositoryImpl) DeleteByRegex(pattern string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for key := range r.data {
		if matchSimplePattern(key, pattern) {
			delete(r.data, key)
			count++
		}
	}
	return count, nil
}

func (r *CacheRepositoryImpl) DeleteByTag(tag string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for key, entry := range r.data {
		for _, t := range entry.Tags {
			if t == tag {
				delete(r.data, key)
				count++
				break
			}
		}
	}
	return count, nil
}

func (r *CacheRepositoryImpl) DeleteAll() (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := len(r.data)
	r.data = make(map[string]*entity.CacheEntry)
	return count, nil
}

func (r *CacheRepositoryImpl) GetStats() (*CacheStats, error) {
	r.statsMu.RLock()
	defer r.statsMu.RUnlock()

	statsCopy := *r.stats
	r.mu.RLock()
	statsCopy.TotalEntries = len(r.data)
	statsCopy.TotalSize = r.getTotalSizeLocked()
	r.mu.RUnlock()

	if statsCopy.HitCount+statsCopy.MissCount > 0 {
		statsCopy.HitRate = float64(statsCopy.HitCount) / float64(statsCopy.HitCount+statsCopy.MissCount)
	}

	return &statsCopy, nil
}

func (r *CacheRepositoryImpl) Cleanup() (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	now := time.Now()
	for key, entry := range r.data {
		if now.After(entry.ExpiresAt) {
			delete(r.data, key)
			count++
		}
	}
	return count, nil
}

func (r *CacheRepositoryImpl) GetKeysByPrefix(prefix string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys := make([]string, 0)
	for key := range r.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func (r *CacheRepositoryImpl) UpdateAccess(key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if entry, exists := r.data[key]; exists {
		entry.LastAccess = time.Now()
		entry.HitCount++
	}
	return nil
}

func (r *CacheRepositoryImpl) load() error {
	file, err := os.Open(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&r.data)
}

func (r *CacheRepositoryImpl) save() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	file, err := os.Create(r.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(r.data)
}

func (r *CacheRepositoryImpl) periodicSave() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		r.save()
	}
}

func (r *CacheRepositoryImpl) periodicCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		r.Cleanup()
	}
}

func (r *CacheRepositoryImpl) getTotalSizeLocked() int64 {
	var total int64
	for _, entry := range r.data {
		total += entry.Size
	}
	return total
}

func (r *CacheRepositoryImpl) evictOneLocked() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range r.data {
		if oldestKey == "" || entry.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastAccess
		}
	}

	if oldestKey != "" {
		delete(r.data, oldestKey)
		r.statsMu.Lock()
		r.stats.EvictedCount++
		r.statsMu.Unlock()
	}
}

func (r *CacheRepositoryImpl) recordHit() {
	r.statsMu.Lock()
	r.stats.HitCount++
	r.statsMu.Unlock()
}

func (r *CacheRepositoryImpl) recordMiss() {
	r.statsMu.Lock()
	r.stats.MissCount++
	r.statsMu.Unlock()
}

func matchSimplePattern(s, pattern string) bool {
	return len(s) >= len(pattern) && s[:len(pattern)] == pattern
}
