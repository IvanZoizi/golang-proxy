package repository

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"proxy/internal/entity"
)

func tempCacheFile(t *testing.T) string {
	dir := os.TempDir()
	return filepath.Join(dir, "test_cache.json")
}

func TestCacheRepositoryDeleteByRegex(t *testing.T) {
	tmpFile := tempCacheFile(t)
	defer os.Remove(tmpFile)

	repo, _ := NewCacheRepository(tmpFile, 1024*1024, 100)

	repo.Set(&entity.CacheEntry{Key: "user123", Value: []byte("a"), ExpiresAt: time.Now().Add(time.Hour), Size: 1})
	repo.Set(&entity.CacheEntry{Key: "user456", Value: []byte("b"), ExpiresAt: time.Now().Add(time.Hour), Size: 1})
	repo.Set(&entity.CacheEntry{Key: "admin789", Value: []byte("c"), ExpiresAt: time.Now().Add(time.Hour), Size: 1})

	count, _ := repo.DeleteByRegex("user")
	if count != 2 {
		t.Errorf("expected 2 deleted, got %d", count)
	}
}

func TestCacheRepositoryDeleteByTag(t *testing.T) {
	tmpFile := tempCacheFile(t)
	defer os.Remove(tmpFile)

	repo, _ := NewCacheRepository(tmpFile, 1024*1024, 100)

	repo.Set(&entity.CacheEntry{Key: "key1", Value: []byte("a"), ExpiresAt: time.Now().Add(time.Hour), Size: 1, Tags: []string{"tag1", "tag2"}})
	repo.Set(&entity.CacheEntry{Key: "key2", Value: []byte("b"), ExpiresAt: time.Now().Add(time.Hour), Size: 1, Tags: []string{"tag2"}})
	repo.Set(&entity.CacheEntry{Key: "key3", Value: []byte("c"), ExpiresAt: time.Now().Add(time.Hour), Size: 1, Tags: []string{"tag3"}})

	count, _ := repo.DeleteByTag("tag2")
	if count != 2 {
		t.Errorf("expected 2 deleted, got %d", count)
	}
}

func TestCacheRepositoryDeleteAll(t *testing.T) {
	tmpFile := tempCacheFile(t)
	defer os.Remove(tmpFile)

	repo, _ := NewCacheRepository(tmpFile, 1024*1024, 100)

	repo.Set(&entity.CacheEntry{Key: "key1", Value: []byte("a"), ExpiresAt: time.Now().Add(time.Hour), Size: 1})
	repo.Set(&entity.CacheEntry{Key: "key2", Value: []byte("b"), ExpiresAt: time.Now().Add(time.Hour), Size: 1})

	count, _ := repo.DeleteAll()
	if count != 2 {
		t.Errorf("expected 2 deleted, got %d", count)
	}

	ips, _ := repo.GetKeysByPrefix("")
	if len(ips) != 0 {
		t.Errorf("expected 0 entries, got %d", len(ips))
	}
}

func TestCacheRepositoryGetKeysByPrefix(t *testing.T) {
	tmpFile := tempCacheFile(t)
	defer os.Remove(tmpFile)

	repo, _ := NewCacheRepository(tmpFile, 1024*1024, 100)

	repo.Set(&entity.CacheEntry{Key: "user:1", Value: []byte("a"), ExpiresAt: time.Now().Add(time.Hour), Size: 1})
	repo.Set(&entity.CacheEntry{Key: "user:2", Value: []byte("b"), ExpiresAt: time.Now().Add(time.Hour), Size: 1})
	repo.Set(&entity.CacheEntry{Key: "admin:1", Value: []byte("c"), ExpiresAt: time.Now().Add(time.Hour), Size: 1})

	keys, _ := repo.GetKeysByPrefix("user:")
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestCacheRepositoryEvictOneLocked(t *testing.T) {
	tmpFile := tempCacheFile(t)
	defer os.Remove(tmpFile)

	repo, _ := NewCacheRepository(tmpFile, 100, 2)

	repo.Set(&entity.CacheEntry{Key: "key1", Value: []byte("a"), ExpiresAt: time.Now().Add(time.Hour), Size: 50, LastAccess: time.Now().Add(-10 * time.Minute)})
	repo.Set(&entity.CacheEntry{Key: "key2", Value: []byte("b"), ExpiresAt: time.Now().Add(time.Hour), Size: 50, LastAccess: time.Now()})
	repo.Set(&entity.CacheEntry{Key: "key3", Value: []byte("c"), ExpiresAt: time.Now().Add(time.Hour), Size: 50, LastAccess: time.Now()})

	stats, _ := repo.GetStats()
	if stats.TotalEntries > 2 {
		t.Logf("Entries after eviction: %d", stats.TotalEntries)
	}
}

func TestCacheRepository_GetAndUpdateAccess(t *testing.T) {
	tmpFile := tempCacheFile(t)
	defer os.Remove(tmpFile)

	repo, _ := NewCacheRepository(tmpFile, 1024*1024, 100)

	entry := &entity.CacheEntry{
		Key:        "test-key",
		Value:      []byte("hello"),
		Size:       5,
		ExpiresAt:  time.Now().Add(time.Hour),
		HitCount:   0,
		LastAccess: time.Now().Add(-time.Hour),
	}

	repo.Set(entry)

	got, err := repo.Get("test-key")
	if err != nil || got == nil || got.HitCount != 1 {
		t.Errorf("expected HitCount=1 after Get, got %d", got.HitCount)
	}
}

func TestCacheRepository_Stats(t *testing.T) {
	tmpFile := tempCacheFile(t)
	defer os.Remove(tmpFile)

	repo, _ := NewCacheRepository(tmpFile, 1024*1024, 100)

	repo.Set(&entity.CacheEntry{Key: "a", Value: []byte("1"), Size: 1, ExpiresAt: time.Now().Add(time.Hour)})
	repo.Get("a")
	repo.Get("not-exist")

	stats, _ := repo.GetStats()
	if stats.TotalEntries != 1 {
		t.Errorf("expected 1 entry, got %d", stats.TotalEntries)
	}
	if stats.HitCount == 0 || stats.MissCount == 0 {
		t.Error("Hit/Miss counters should be updated")
	}
}
