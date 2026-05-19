package entity

import (
	"testing"
)

func TestListAddIp(t *testing.T) {
	list := &List{Ips: []string{}}

	if !list.AddIp("192.168.1.1") {
		t.Error("expected true for new IP")
	}
	if list.AddIp("192.168.1.1") {
		t.Error("expected false for duplicate IP")
	}
	if len(list.Ips) != 1 {
		t.Errorf("expected 1 IP, got %d", len(list.Ips))
	}
}

func TestListRemoveIp(t *testing.T) {
	list := &List{Ips: []string{"192.168.1.1", "192.168.1.2"}}

	if !list.RemoveIp("192.168.1.1") {
		t.Error("expected true for existing IP")
	}
	if list.RemoveIp("10.0.0.1") {
		t.Error("expected false for non-existing IP")
	}
	if len(list.Ips) != 1 {
		t.Errorf("expected 1 IP, got %d", len(list.Ips))
	}
}

func TestListContains(t *testing.T) {
	list := &List{Ips: []string{"192.168.1.1"}}

	if !list.Contains("192.168.1.1") {
		t.Error("expected true for existing IP")
	}
	if list.Contains("10.0.0.1") {
		t.Error("expected false for non-existing IP")
	}
}

func TestListGetAll(t *testing.T) {
	list := &List{Ips: []string{"192.168.1.1", "192.168.1.2"}}
	result := list.GetAll()

	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
	result[0] = "modified"
	if list.Ips[0] == "modified" {
		t.Error("GetAll should return a copy")
	}
}

func TestListGetAllCIDR(t *testing.T) {
	list := &List{Ips: []string{"192.168.1.0/24", "10.0.0.1", "192.168.2.0/24"}}
	result := list.GetAllCIDR()

	for _, ip := range result {
		if ip == "" {
			t.Error("got empty string in result")
		}
	}

	if result == nil {
		t.Error("result should not be nil")
	}
}

func TestListGetAllRangeIps(t *testing.T) {
	list := &List{Ips: []string{"192.168.1.1-192.168.1.255", "10.0.0.1", "192.168.2.0/24"}}
	result := list.GetAllRangeIps()

	if result == nil {
		t.Error("result should not be nil")
	}
}
