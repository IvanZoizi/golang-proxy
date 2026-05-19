package dto

// UpdateCacheConfigRequest godoc
// @Description Запрос на обновление конфигурации кэша
type UpdateCacheConfigRequest struct {
	Enabled           *bool    `json:"enabled,omitempty" example:"true"`
	DefaultTTL        *string  `json:"default_ttl,omitempty" example:"5m"`
	MaxEntrySizeBytes *int64   `json:"max_entry_size_bytes,omitempty" example:"1048576"`
	MinEntrySizeBytes *int64   `json:"min_entry_size_bytes,omitempty" example:"1024"`
	MaxCacheSizeBytes *int64   `json:"max_cache_size_bytes,omitempty" example:"1073741824"`
	MaxEntries        *int     `json:"max_entries,omitempty" example:"10000"`
	Cache2xx          *bool    `json:"cache_2xx,omitempty" example:"true"`
	Cache3xx          *bool    `json:"cache_3xx,omitempty" example:"true"`
	Cache4xx          *bool    `json:"cache_4xx,omitempty" example:"false"`
	Cache5xx          *bool    `json:"cache_5xx,omitempty" example:"false"`
	TTL2xx            *string  `json:"ttl_2xx,omitempty" example:"5m"`
	TTL3xx            *string  `json:"ttl_3xx,omitempty" example:"5m"`
	TTL4xx            *string  `json:"ttl_4xx,omitempty" example:"1m"`
	TTL5xx            *string  `json:"ttl_5xx,omitempty" example:"10s"`
	CacheMethods      []string `json:"cache_methods,omitempty" example:"GET,HEAD"`
	NotCacheURL       []string `json:"not_cache_url" example:"/metrics"`
	RespectNoCache    *bool    `json:"respect_no_cache,omitempty" example:"true"`
	RespectMaxAge     *bool    `json:"respect_max_age,omitempty" example:"true"`
}
