package entity

import ip2 "proxy/pkg/ip"

type List struct {
	Ips []string
}

func (l *List) AddIp(ip string) bool {
	for _, existingIp := range l.Ips {
		if existingIp == ip {
			return false
		}
	}
	l.Ips = append(l.Ips, ip)
	return true
}

func (l *List) RemoveIp(ip string) bool {
	for i, existingIp := range l.Ips {
		if existingIp == ip {
			l.Ips = append(l.Ips[:i], l.Ips[i+1:]...)
			return true
		}
	}
	return false
}

func (l *List) Contains(ip string) bool {
	for _, existingIp := range l.Ips {
		if existingIp == ip {
			return true
		}
	}
	return false
}

func (l *List) GetAll() []string {
	result := make([]string, len(l.Ips))
	copy(result, l.Ips)
	return result
}

func (l *List) GetAllCIDR() []string {
	result := make([]string, len(l.Ips))
	for i := 0; i < len(l.Ips); i++ {
		if ip2.CheckIpCIRS(l.Ips[i]) {
			result = append(result, l.Ips[i])
		}
	}
	return result
}

func (l *List) GetAllRangeIps() []string {
	result := make([]string, len(l.Ips))
	for i := 0; i < len(l.Ips); i++ {
		if ip2.CheckIpCIRS(l.Ips[i]) {
			result = append(result, l.Ips[i])
		}
	}
	return result
}
