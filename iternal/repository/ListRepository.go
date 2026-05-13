package repository

import (
	"encoding/json"
	"os"
	"proxy/iternal/entity"
	ip2 "proxy/pkg/ip"
	"proxy/pkg/utils"
)

type ListRepository interface {
	GetAllIps() []string
	AddIps(ip string) (string, bool)
	DeleteIp(ip string) (string, bool)
}

type ListRepositoryImpl struct {
	ListIps *entity.List
}

func CreateRepository(list *entity.List) ListRepository {
	return &ListRepositoryImpl{ListIps: list}
}

func (listRepo *ListRepositoryImpl) AddIps(ip string) (string, bool) {

	if !ip2.CheckIp(ip) {
		return "Text is not IP", false
	}

	if ip2.CheckInIp(ip, listRepo.ListIps.Ips) >= 0 {
		return "This IP already in list", false
	}

	listRepo.ListIps.Ips = append(listRepo.ListIps.Ips, ip)

	file, err := os.Create(listRepo.ListIps.Filename)
	if err != nil {
		panic("File Json List not found")
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(listRepo.ListIps)
	if err != nil {
		panic(err)
	}

	return "", true
}

func (listRepo *ListRepositoryImpl) DeleteIp(ip string) (string, bool) {
	index := ip2.CheckInIp(ip, listRepo.ListIps.Ips)
	if !ip2.CheckIp(ip) {
		return "Text is not IP", false
	}

	if index < 0 {
		return "IP not in list", false
	}

	listRepo.ListIps.Ips = utils.RemoveByIndex(listRepo.ListIps.Ips, index)

	return "", true
}

func (listRepo ListRepositoryImpl) GetAllIps() []string {
	return listRepo.ListIps.Ips
}
