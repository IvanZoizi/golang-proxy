package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func tempFile(t *testing.T) string {
	dir := os.TempDir()
	file := filepath.Join(dir, "test_list.json")
	return file
}

func TestNewListRepository(t *testing.T) {
	tmpFile := tempFile(t)
	defer os.Remove(tmpFile)

	repo, err := NewListRepository(tmpFile)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	if repo == nil {
		t.Error("expected non-nil repo")
	}
}

func TestListRepositoryAddAndGet(t *testing.T) {
	tmpFile := tempFile(t)
	defer os.Remove(tmpFile)

	repo, err := NewListRepository(tmpFile)
	if err != nil {
		t.Fatalf("NewListRepository failed: %v", err)
	}

	err = repo.AddIp("192.168.1.1")
	if err != nil {
		t.Errorf("AddIp failed: %v", err)
	}

	ips, _ := repo.GetAllIps()
	if len(ips) != 1 {
		t.Errorf("expected 1 IP, got %d", len(ips))
	}
	if ips[0] != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1, got %s", ips[0])
	}
}

func TestListRepositoryAddDuplicate(t *testing.T) {
	tmpFile := tempFile(t)
	defer os.Remove(tmpFile)

	repo, _ := NewListRepository(tmpFile)

	repo.AddIp("192.168.1.1")
	err := repo.AddIp("192.168.1.1")
	if err == nil {
		t.Error("expected error for duplicate IP")
	}
}

func TestListRepositoryDelete(t *testing.T) {
	tmpFile := tempFile(t)
	defer os.Remove(tmpFile)

	repo, _ := NewListRepository(tmpFile)

	repo.AddIp("192.168.1.1")
	err := repo.DeleteIp("192.168.1.1")
	if err != nil {
		t.Errorf("DeleteIp failed: %v", err)
	}

	ips, _ := repo.GetAllIps()
	if len(ips) != 0 {
		t.Errorf("expected 0 IPs, got %d", len(ips))
	}
}

func TestListRepositoryContains(t *testing.T) {
	tmpFile := tempFile(t)
	defer os.Remove(tmpFile)

	repo, _ := NewListRepository(tmpFile)

	repo.AddIp("192.168.1.1")

	if !repo.Contains("192.168.1.1") {
		t.Error("expected true for existing IP")
	}
	if repo.Contains("10.0.0.1") {
		t.Error("expected false for non-existing IP")
	}
}

func TestListRepository_CIDRAndRange(t *testing.T) {
	tmpFile := tempFile(t)
	defer os.Remove(tmpFile)

	repo, _ := NewListRepository(tmpFile)

	repo.AddIp("192.168.1.0/24")
	repo.AddIp("10.0.0.1-10.0.0.255")
	repo.AddIp("regular.ip")

	cidr, _ := repo.GetAllCIDRIps()
	rangeIps, _ := repo.GetAllRangeIps()

	if len(cidr) == 0 {
		t.Error("Should detect CIDR")
	}
	if len(rangeIps) == 0 {
		t.Error("Should detect range")
	}
}
