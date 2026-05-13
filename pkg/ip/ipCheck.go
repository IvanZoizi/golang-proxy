package ip

import "regexp"

var ipRegex1 = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
var ipRegex2 = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)-((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
var ipRegex3 = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)/(3[0-2]|[12]?[0-9])$`)

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

func CheckInIp(ip string, ips []string) int {
	for i := 0; i < len(ips); i++ {
		if ip == ips[i] {
			return i
		}
	}
	return -1
}
