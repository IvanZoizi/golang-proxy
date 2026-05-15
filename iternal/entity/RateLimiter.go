package entity

import "time"

type RateLimitConfig struct {
	Enabled bool `json:"enabled"`

	RequestsPerSecond int `json:"requests_per_second"`
	RequestsPerMinute int `json:"requests_per_minute"`
	RequestsPerHour   int `json:"requests_per_hour"`
	RequestsPerDay    int `json:"requests_per_day"`

	DownloadLimitMB int `json:"download_limit_mb"`
	UploadLimitMB   int `json:"upload_limit_mb"`
	TotalTrafficMB  int `json:"total_traffic_mb"`

	MaxConcurrentConnections int `json:"max_concurrent_connections"`
	NewConnectionsPerSecond  int `json:"new_connections_per_second"`

	SubnetLimits map[string]SubnetRateLimit `json:"subnet_limits"`
}

type SubnetRateLimit struct {
	RequestsPerSecond int `json:"requests_per_second"`
	RequestsPerMinute int `json:"requests_per_minute"`
	RequestsPerHour   int `json:"requests_per_hour"`
	RequestsPerDay    int `json:"requests_per_day"`
}

type RateLimitData struct {
	IP string `json:"ip"`

	RequestsThisSecond int `json:"requests_this_second"`
	RequestsThisMinute int `json:"requests_this_minute"`
	RequestsThisHour   int `json:"requests_this_hour"`
	RequestsThisDay    int `json:"requests_this_day"`

	LastResetSecond time.Time `json:"last_reset_second"`
	LastResetMinute time.Time `json:"last_reset_minute"`
	LastResetHour   time.Time `json:"last_reset_hour"`
	LastResetDay    time.Time `json:"last_reset_day"`

	DownloadBytes int64 `json:"download_bytes"`
	UploadBytes   int64 `json:"upload_bytes"`
	TotalBytes    int64 `json:"total_bytes"`

	ActiveConnections   int       `json:"active_connections"`
	ConnectionsThisSec  int       `json:"connections_this_sec"`
	LastConnectionReset time.Time `json:"last_connection_reset"`

	IsBlocked     bool      `json:"is_blocked"`
	BlockedUntil  time.Time `json:"blocked_until"`
	BlockedReason string    `json:"blocked_reason"`
}
