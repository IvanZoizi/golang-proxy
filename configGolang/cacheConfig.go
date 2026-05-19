package configGolang

import (
	"proxy/internal/entity"
	"time"
)

func GetDefaultCacheConfig() *entity.CacheConfig {
	return &entity.CacheConfig{
		Enabled:        true,
		DefaultTTL:     5 * time.Minute,
		MaxEntrySize:   10 * 1024 * 1024,
		MinEntrySize:   1,
		MaxCacheSize:   100 * 1024 * 1024,
		MaxEntries:     10000,
		Cache2xx:       true,
		Cache3xx:       true,
		Cache4xx:       false,
		Cache5xx:       false,
		TTL2xx:         5 * time.Minute,
		TTL3xx:         1 * time.Minute,
		TTL4xx:         0,
		TTL5xx:         0,
		CacheMethods:   []string{"GET", "HEAD"},
		NotCacheURL:    []string{"/metrics"},
		RespectNoCache: true,
		RespectMaxAge:  true,
	}
}
