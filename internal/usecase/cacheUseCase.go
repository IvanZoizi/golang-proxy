package usecase

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"proxy/internal/dto"
	"proxy/internal/entity"
	"proxy/internal/repository"
	"proxy/pkg/logger"
	"regexp"
	"strings"
	"sync"
	"time"
)

type CacheUseCase interface {
	ShouldCache(req *http.Request, resp *http.Response) (bool, time.Duration)
	CreateCacheEntry(key string, resp *http.Response, body []byte) (*entity.CacheEntry, error)
	GetCachedResponse(key string) (*http.Response, []byte, error)
	Invalidate(req *InvalidateRequest) (int, error)
	GetCacheStats() (*repository.CacheStats, error)
	GetCacheKey(req *http.Request) string
	CheckMethod(method string) bool
	UpdateConfigCache(req dto.UpdateCacheConfigRequest) error
}

type InvalidateRequest struct {
	Type      string
	Key       string
	Prefix    string
	Regex     string
	Tag       string
	StaleOnly bool
}

type CacheUseCaseImpl struct {
	repo   repository.CacheRepository
	config *entity.CacheConfig
	rules  []*entity.CacheRule
	ruleMu sync.RWMutex
}

func NewCacheUseCase(repo repository.CacheRepository, config *entity.CacheConfig) *CacheUseCaseImpl {
	uc := &CacheUseCaseImpl{
		repo:   repo,
		config: config,
		rules:  make([]*entity.CacheRule, 0),
	}
	return uc
}

func NewInvalidationRequest(typeReq string) *InvalidateRequest {
	return &InvalidateRequest{Type: typeReq}
}

func (uc *CacheUseCaseImpl) ShouldCache(req *http.Request, resp *http.Response) (bool, time.Duration) {
	if !uc.config.Enabled {
		return false, 0
	}

	methodAllowed := false
	for _, m := range uc.config.CacheMethods {
		if req.Method == m {
			methodAllowed = true
			break
		}
	}
	if !methodAllowed {
		return false, 0
	}

	ttl := uc.getTTLForStatus(resp.StatusCode)
	if ttl == 0 {
		return false, 0
	}

	if uc.config.RespectNoCache {
		if resp.Header.Get("Cache-Control") == "no-cache" ||
			resp.Header.Get("Cache-Control") == "no-store" ||
			resp.Header.Get("Pragma") == "no-cache" {
			return false, 0
		}
	}

	ruleTTL := uc.matchRules(req, resp)
	if ruleTTL != nil {
		ttl = *ruleTTL
	}

	if uc.config.RespectMaxAge {
		if maxAge := parseMaxAge(resp.Header.Get("Cache-Control")); maxAge > 0 {
			if maxAge < int(ttl.Seconds()) {
				ttl = time.Duration(maxAge) * time.Second
			}
		}
	}

	return true, ttl
}

func (uc *CacheUseCaseImpl) CreateCacheEntry(key string, resp *http.Response, body []byte) (*entity.CacheEntry, error) {
	size := int64(len(body))
	if size > uc.config.MaxEntrySize || size < uc.config.MinEntrySize {
		return nil, fmt.Errorf("entry size %d out of bounds [%d, %d]",
			size, uc.config.MinEntrySize, uc.config.MaxEntrySize)
	}

	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	_, ttl := uc.ShouldCache(resp.Request, resp)

	tags := uc.extractTags(resp.Request)

	entry := &entity.CacheEntry{
		Key:        key,
		Value:      body,
		Headers:    headers,
		StatusCode: resp.StatusCode,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(ttl),
		LastAccess: time.Now(),
		HitCount:   0,
		Tags:       tags,
		Size:       size,
	}

	uc.repo.Set(entry)

	return entry, nil
}

func (uc *CacheUseCaseImpl) GetCachedResponse(key string) (*http.Response, []byte, error) {
	entry, err := uc.repo.Get(key)
	if err != nil || entry == nil {
		return nil, nil, err
	}

	resp := &http.Response{
		StatusCode: entry.StatusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(entry.Value)),
	}

	for k, v := range entry.Headers {
		resp.Header.Set(k, v)
	}

	return resp, entry.Value, nil
}

func (uc *CacheUseCaseImpl) Invalidate(req *InvalidateRequest) (int, error) {
	switch req.Type {
	case "exact":
		if err := uc.repo.Delete(req.Key); err != nil {
			return 0, err
		}
		return 1, nil

	case "prefix":
		return uc.repo.DeleteByPrefix(req.Prefix)

	case "regex":
		return uc.repo.DeleteByRegex(req.Regex)

	case "tag":
		return uc.repo.DeleteByTag(req.Tag)

	case "all":
		return uc.repo.DeleteAll()

	default:
		return 0, fmt.Errorf("unknown invalidation type: %s", req.Type)
	}
}

func (uc *CacheUseCaseImpl) GetCacheStats() (*repository.CacheStats, error) {
	return uc.repo.GetStats()
}

func (uc *CacheUseCaseImpl) GetCacheKey(req *http.Request) string {
	hasher := sha256.New()
	hasher.Write([]byte(req.Method))
	hasher.Write([]byte(req.URL.String()))

	if accept := req.Header.Get("Accept"); accept != "" {
		hasher.Write([]byte(accept))
	}
	if acceptEncoding := req.Header.Get("Accept-Encoding"); acceptEncoding != "" {
		hasher.Write([]byte(acceptEncoding))
	}
	if acceptLanguage := req.Header.Get("Accept-Language"); acceptLanguage != "" {
		hasher.Write([]byte(acceptLanguage))
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

func (uc *CacheUseCaseImpl) AddRule(rule *entity.CacheRule) {
	uc.ruleMu.Lock()
	defer uc.ruleMu.Unlock()

	inserted := false
	for i, r := range uc.rules {
		if rule.Priority > r.Priority {
			uc.rules = append(uc.rules[:i], append([]*entity.CacheRule{rule}, uc.rules[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		uc.rules = append(uc.rules, rule)
	}
}

func (uc *CacheUseCaseImpl) CheckMethod(method string) bool {
	for _, mth := range uc.config.CacheMethods {
		if method == mth {
			return true
		}
	}
	return false
}

func (uc *CacheUseCaseImpl) matchRules(req *http.Request, resp *http.Response) *time.Duration {
	uc.ruleMu.RLock()
	defer uc.ruleMu.RUnlock()

	url := req.URL.String()
	host := req.URL.Host

	for _, rule := range uc.rules {
		if !rule.Enabled {
			continue
		}

		excluded := false
		for _, pattern := range rule.ExcludePatterns {
			if uc.matchPattern(url, host, pattern, rule.MatchType) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		if uc.matchPattern(url, host, rule.Pattern, rule.MatchType) {
			if len(rule.CacheStatus) > 0 {
				matched := false
				for _, code := range rule.CacheStatus {
					if resp.StatusCode == code {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}

			if rule.TTL != nil {
				return rule.TTL
			}
		}
	}

	return nil
}

func (uc *CacheUseCaseImpl) matchPattern(url, host, pattern, matchType string) bool {
	switch matchType {
	case "exact":
		return url == pattern || host == pattern
	case "domain":
		return strings.Contains(host, pattern)
	case "path":
		return strings.Contains(url, pattern)
	case "regex":
		matched, _ := regexp.MatchString(pattern, url)
		return matched
	default:
		return false
	}
}

func (uc *CacheUseCaseImpl) getTTLForStatus(statusCode int) time.Duration {
	if statusCode >= 200 && statusCode < 300 && uc.config.Cache2xx {
		return uc.config.TTL2xx
	}
	if statusCode >= 300 && statusCode < 400 && uc.config.Cache3xx {
		return uc.config.TTL3xx
	}
	if statusCode >= 400 && statusCode < 500 && uc.config.Cache4xx {
		return uc.config.TTL4xx
	}
	if statusCode >= 500 && statusCode < 600 && uc.config.Cache5xx {
		return uc.config.TTL5xx
	}
	return 0
}

func (uc *CacheUseCaseImpl) extractTags(req *http.Request) []string {
	uc.ruleMu.RLock()
	defer uc.ruleMu.RUnlock()

	tags := make([]string, 0)
	url := req.URL.String()
	host := req.URL.Host

	for _, rule := range uc.rules {
		if rule.Enabled && uc.matchPattern(url, host, rule.Pattern, rule.MatchType) {
			tags = append(tags, rule.Tags...)
		}
	}

	return tags
}

func parseMaxAge(cacheControl string) int {
	if cacheControl == "" {
		return 0
	}

	parts := strings.Split(cacheControl, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "max-age=") {
			var maxAge int
			fmt.Sscanf(part, "max-age=%d", &maxAge)
			return maxAge
		}
	}
	return 0
}

func (uc CacheUseCaseImpl) UpdateConfigCache(req dto.UpdateCacheConfigRequest) error {
	logger.Info("Update config cache")
	if req.Enabled != nil {
		uc.config.Enabled = *req.Enabled
	}
	if req.DefaultTTL != nil {
		if duration, err := time.ParseDuration(*req.DefaultTTL); err == nil {
			uc.config.DefaultTTL = duration
		}
	}
	if req.MaxEntrySizeBytes != nil {
		uc.config.MaxEntrySize = *req.MaxEntrySizeBytes
	}
	if req.MinEntrySizeBytes != nil {
		uc.config.MinEntrySize = *req.MinEntrySizeBytes
	}
	if req.MaxCacheSizeBytes != nil {
		uc.config.MaxCacheSize = *req.MaxCacheSizeBytes
	}
	if req.MaxEntries != nil {
		uc.config.MaxEntries = *req.MaxEntries
	}
	if req.Cache2xx != nil {
		uc.config.Cache2xx = *req.Cache2xx
	}
	if req.Cache3xx != nil {
		uc.config.Cache3xx = *req.Cache3xx
	}
	if req.Cache4xx != nil {
		uc.config.Cache4xx = *req.Cache4xx
	}
	if req.Cache5xx != nil {
		uc.config.Cache5xx = *req.Cache5xx
	}
	if req.TTL2xx != nil {
		if duration, err := time.ParseDuration(*req.TTL2xx); err == nil {
			uc.config.TTL2xx = duration
		}
	}
	if req.TTL3xx != nil {
		if duration, err := time.ParseDuration(*req.TTL3xx); err == nil {
			uc.config.TTL3xx = duration
		}
	}
	if req.TTL4xx != nil {
		if duration, err := time.ParseDuration(*req.TTL4xx); err == nil {
			uc.config.TTL4xx = duration
		}
	}
	if req.TTL5xx != nil {
		if duration, err := time.ParseDuration(*req.TTL5xx); err == nil {
			uc.config.TTL5xx = duration
		}
	}
	if req.CacheMethods != nil {
		uc.config.CacheMethods = req.CacheMethods
	}
	if req.RespectNoCache != nil {
		uc.config.RespectNoCache = *req.RespectNoCache
	}
	if req.RespectMaxAge != nil {
		uc.config.RespectMaxAge = *req.RespectMaxAge
	}
	return nil
}
