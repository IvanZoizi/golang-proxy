package entity

import "time"

type CacheEntry struct {
	Key        string            `json:"key"`
	Value      []byte            `json:"value"`
	Headers    map[string]string `json:"headers"`
	StatusCode int               `json:"status_code"`
	CreatedAt  time.Time         `json:"created_at"`
	ExpiresAt  time.Time         `json:"expires_at"`
	LastAccess time.Time         `json:"last_access"`
	HitCount   int64             `json:"hit_count"`
	Tags       []string          `json:"tags"`
	Size       int64             `json:"size"`
}

type CacheConfig struct {
	Enabled bool `json:"enabled"`

	DefaultTTL   time.Duration `json:"default_ttl"`
	MaxEntrySize int64         `json:"max_entry_size_bytes"`
	MinEntrySize int64         `json:"min_entry_size_bytes"`
	MaxCacheSize int64         `json:"max_cache_size_bytes"`
	MaxEntries   int           `json:"max_entries"`

	Cache2xx bool `json:"cache_2xx"`
	Cache3xx bool `json:"cache_3xx"`
	Cache4xx bool `json:"cache_4xx"`
	Cache5xx bool `json:"cache_5xx"`

	TTL2xx time.Duration `json:"ttl_2xx"`
	TTL3xx time.Duration `json:"ttl_3xx"`
	TTL4xx time.Duration `json:"ttl_4xx"`
	TTL5xx time.Duration `json:"ttl_5xx"`

	CacheMethods []string `json:"cache_methods"`

	RespectNoCache bool `json:"respect_no_cache"`
	RespectMaxAge  bool `json:"respect_max_age"`
}

type CacheRule struct {
	ID        string `json:"id"`
	Priority  int    `json:"priority"`
	MatchType string `json:"match_type"`
	Pattern   string `json:"pattern"`
	Enabled   bool   `json:"enabled"`

	TTL         *time.Duration `json:"ttl,omitempty"`
	MaxSize     *int64         `json:"max_size,omitempty"`
	CacheStatus []int          `json:"cache_status,omitempty"`

	ExcludePatterns []string `json:"exclude_patterns,omitempty"`

	Tags []string `json:"tags,omitempty"`
}

type InvalidationRequest struct {
	Type      string `json:"type"`
	Key       string `json:"key,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
	Regex     string `json:"regex,omitempty"`
	Tag       string `json:"tag,omitempty"`
	StaleOnly bool   `json:"stale_only,omitempty"`
}
