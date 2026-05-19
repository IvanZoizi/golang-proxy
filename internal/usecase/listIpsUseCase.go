package usecase

import (
	"fmt"
	"proxy/internal/repository"
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
	case "gray":
		return list.GrayRepo.GetAllIps()
	default:
		return make([]string, 0), fmt.Errorf("Invalid argument")
	}
}

func (list *ListsUseCaseImpl) AddIp(ip string, listName string) error {
	switch listName {
	case "white":
		return list.WhiteRepo.AddIp(ip)
	case "black":
		return list.BlackRepo.AddIp(ip)
	case "gray":
		return list.GrayRepo.AddIp(ip)
	default:
		return fmt.Errorf("Invalid argument")
	}
}

func (list *ListsUseCaseImpl) Contains(ip string, listName string) (bool, error) {
	switch listName {
	case "white":
		return list.WhiteRepo.Contains(ip), nil
	case "black":
		return list.BlackRepo.Contains(ip), nil
	case "gray":
		return list.GrayRepo.Contains(ip), nil
	default:

		return false, fmt.Errorf("Invalid argument")
	}
}

func (list *ListsUseCaseImpl) DeleteIp(ip string, listName string) error {
	switch listName {
	case "white":
		return list.WhiteRepo.DeleteIp(ip)
	case "black":
		return list.BlackRepo.DeleteIp(ip)
	case "gray":
		return list.GrayRepo.DeleteIp(ip)
	default:
		return fmt.Errorf("Invalid argument")
	}
}

func (list *ListsUseCaseImpl) GetAllCIDRIps(listName string) ([]string, error) {
	switch listName {
	case "white":
		return list.WhiteRepo.GetAllCIDRIps()
	case "black":
		return list.BlackRepo.GetAllCIDRIps()
	case "gray":
		return list.GrayRepo.GetAllCIDRIps()
	default:

		return make([]string, 0), fmt.Errorf("Invalid argument")
	}
}

func (list *ListsUseCaseImpl) GetAllRangeIps(listName string) ([]string, error) {
	switch listName {
	case "white":
		return list.WhiteRepo.GetAllRangeIps()
	case "black":
		return list.BlackRepo.GetAllRangeIps()
	case "gray":
		return list.GrayRepo.GetAllRangeIps()
	default:

		return make([]string, 0), fmt.Errorf("Invalid argument")
	}
}
