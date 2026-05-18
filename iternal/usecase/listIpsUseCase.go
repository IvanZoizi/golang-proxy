package usecase

import (
	"proxy/iternal/repository"
)

type ListUserCase interface {
	GetAllIps(listName string) ([]string, error)
	AddIp(ip string, listName string) error
	Contains(ip string, listName string) (bool, error)
	DeleteIp(ip string, listName string) error
	GetAllCIDRIps(listName string) ([]string, error)
	GetAllRangeIps(listName string) ([]string, error)
}

type ListsUseCaseImpl struct {
	WhiteRepo repository.ListRepository
	GrayRepo  repository.ListRepository
	BlackRepo repository.ListRepository
}

func CreateListUseCase(whiteRepo, grayRepo, blackRepo repository.ListRepository) ListUserCase {
	return &ListsUseCaseImpl{WhiteRepo: whiteRepo, GrayRepo: grayRepo, BlackRepo: blackRepo}
}

func (list *ListsUseCaseImpl) GetAllIps(listName string) ([]string, error) {
	switch listName {
	case "white":
		return list.WhiteRepo.GetAllIps()
	case "black":
		return list.BlackRepo.GetAllIps()
	default:
		return list.GrayRepo.GetAllIps()
	}
}

func (list *ListsUseCaseImpl) AddIp(ip string, listName string) error {
	switch listName {
	case "white":
		return list.WhiteRepo.AddIp(ip)
	case "black":
		return list.BlackRepo.AddIp(ip)
	default:
		return list.GrayRepo.AddIp(ip)
	}
}

func (list *ListsUseCaseImpl) Contains(ip string, listName string) (bool, error) {
	switch listName {
	case "white":
		return list.WhiteRepo.Contains(ip), nil
	case "black":
		return list.BlackRepo.Contains(ip), nil
	default:
		return list.GrayRepo.Contains(ip), nil
	}
}

func (list *ListsUseCaseImpl) DeleteIp(ip string, listName string) error {
	switch listName {
	case "white":
		return list.WhiteRepo.DeleteIp(ip)
	case "black":
		return list.BlackRepo.DeleteIp(ip)
	default:
		return list.GrayRepo.DeleteIp(ip)
	}
}

func (list *ListsUseCaseImpl) GetAllCIDRIps(listName string) ([]string, error) {
	switch listName {
	case "white":
		return list.WhiteRepo.GetAllCIDRIps()
	case "black":
		return list.BlackRepo.GetAllCIDRIps()
	default:
		return list.GrayRepo.GetAllCIDRIps()
	}
}

func (list *ListsUseCaseImpl) GetAllRangeIps(listName string) ([]string, error) {
	switch listName {
	case "white":
		return list.WhiteRepo.GetAllRangeIps()
	case "black":
		return list.BlackRepo.GetAllRangeIps()
	default:
		return list.GrayRepo.GetAllRangeIps()
	}
}
