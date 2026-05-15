package ip

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
)

var ipRegex1 = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
var ipRegex2 = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)-((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
var ipRegex3 = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)/(3[0-2]|[12]?[0-9])$`)

func CheckIpCIRS(ip string) bool {
	return ipRegex2.MatchString(ip)
}

func CheckIpRange(ip string) bool {
	return ipRegex3.MatchString(ip)
}

func CheckIp(ip string) bool {
	if ipRegex1.MatchString(ip) {
		return true
	}
	if ipRegex2.MatchString(ip) {
		return true
	}
	if ipRegex3.MatchString(ip) {
		return true
	}
	return false
}

func GetClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return strings.TrimSpace(realIP)
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}

	ip = strings.TrimSpace(ip)

	if ip == "::1" {
		ip = "127.0.0.1"
	}

	return ip
}

func IsIPInSubnet(ipStr string, cidr string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, fmt.Errorf("invalid CIDR: %s", cidr)
	}

	return subnet.Contains(ip), nil
}

func IsIPInRange(ipStr string, rangeStr string) (bool, error) {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid range format: %s (expected: startIP-endIP)", rangeStr)
	}

	startIPStr := strings.TrimSpace(parts[0])
	endIPStr := strings.TrimSpace(parts[1])

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	startIP := net.ParseIP(startIPStr)
	if startIP == nil {
		return false, fmt.Errorf("invalid start IP: %s", startIPStr)
	}

	endIP := net.ParseIP(endIPStr)
	if endIP == nil {
		return false, fmt.Errorf("invalid end IP: %s", endIPStr)
	}

	ipInt := ipToUint32(ip)
	startInt := ipToUint32(startIP)
	endInt := ipToUint32(endIP)

	return ipInt >= startInt && ipInt <= endInt, nil
}

func ipToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}
