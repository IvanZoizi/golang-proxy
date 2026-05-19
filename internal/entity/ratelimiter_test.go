package entity

import (
	"testing"
	"time"
)

func TestRateLimitConfig(t *testing.T) {
	config := &RateLimitConfig{
		Enabled:                  true,
		RequestsPerSecond:        100,
		RequestsPerMinute:        1000,
		MaxConcurrentConnections: 10,
	}

	if !config.Enabled {
		t.Error("expected enabled true")
	}
	if config.RequestsPerSecond != 100 {
		t.Errorf("expected 100, got %d", config.RequestsPerSecond)
	}
}

func TestSubnetRateLimit(t *testing.T) {
	limit := SubnetRateLimit{
		RequestsPerSecond: 50,
		RequestsPerMinute: 500,
	}

	if limit.RequestsPerSecond != 50 {
		t.Errorf("expected 50, got %d", limit.RequestsPerSecond)
	}
}

func TestRateLimitData(t *testing.T) {
	now := time.Now()
	data := &RateLimitData{
		IP:                 "192.168.1.1",
		RequestsThisSecond: 10,
		IsBlocked:          false,
		BlockedUntil:       now.Add(time.Hour),
	}

	if data.IP != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1, got %s", data.IP)
	}
	if data.RequestsThisSecond != 10 {
		t.Errorf("expected 10, got %d", data.RequestsThisSecond)
	}
	if data.IsBlocked {
		t.Error("expected false")
	}
}
