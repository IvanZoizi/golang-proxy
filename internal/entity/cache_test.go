package entity

import (
	"testing"
	"time"
)

func TestCacheEntry(t *testing.T) {
	entry := &CacheEntry{
		Key:        "test-key",
		Value:      []byte("test-value"),
		StatusCode: 200,
		HitCount:   5,
	}

	if entry.Key != "test-key" {
		t.Errorf("expected test-key, got %s", entry.Key)
	}
	if entry.StatusCode != 200 {
		t.Errorf("expected 200, got %d", entry.StatusCode)
	}
	if string(entry.Value) != "test-value" {
		t.Errorf("expected test-value, got %s", string(entry.Value))
	}
}

func TestCacheConfig(t *testing.T) {
	config := &CacheConfig{
		Enabled:      true,
		DefaultTTL:   time.Minute,
		MaxEntrySize: 1024,
		CacheMethods: []string{"GET", "HEAD"},
	}

	if !config.Enabled {
		t.Error("expected enabled true")
	}
	if config.DefaultTTL != time.Minute {
		t.Errorf("expected %v, got %v", time.Minute, config.DefaultTTL)
	}
	if len(config.CacheMethods) != 2 {
		t.Errorf("expected 2 methods, got %d", len(config.CacheMethods))
	}
}

func TestCacheRule(t *testing.T) {
	ttl := time.Hour
	rule := &CacheRule{
		ID:        "rule-1",
		Priority:  10,
		MatchType: "domain",
		Pattern:   "example.com",
		Enabled:   true,
		TTL:       &ttl,
	}

	if rule.ID != "rule-1" {
		t.Errorf("expected rule-1, got %s", rule.ID)
	}
	if rule.Priority != 10 {
		t.Errorf("expected 10, got %d", rule.Priority)
	}
	if *rule.TTL != time.Hour {
		t.Errorf("expected %v, got %v", time.Hour, *rule.TTL)
	}
}

func TestInvalidationRequest(t *testing.T) {
	req := &InvalidationRequest{
		Type: "prefix",
		Key:  "test",
		Tag:  "users",
	}

	if req.Type != "prefix" {
		t.Errorf("expected prefix, got %s", req.Type)
	}
	if req.Tag != "users" {
		t.Errorf("expected users, got %s", req.Tag)
	}
}
